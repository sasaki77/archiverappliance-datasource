package archiverappliance

import (
	"context"
	"runtime"
	"testing"
	"time"

	"github.com/sasaki77/archiverappliance-datasource/pkg/models"
)

// blockingClient holds every request until it is released, so that a query can
// be cut short with work still outstanding.
type blockingClient struct {
	release chan struct{}
	entered chan struct{}
}

func (c blockingClient) FetchRegexTargetPVs(ctx context.Context, regex string, limit int) ([]string, error) {
	return []string{regex}, nil
}

func (c blockingClient) ExecuteSingleQuery(ctx context.Context, target string, qm models.ArchiverQueryModel) (models.SingleData, error) {
	c.entered <- struct{}{}
	<-c.release

	// Big enough that holding onto it would matter.
	v := models.NewSclars(1000)
	base := time.Date(2021, time.January, 10, 0, 0, 0, 0, time.UTC)
	for i := 0; i < 1000; i++ {
		v.AppendConcrete(float64(i), base.Add(time.Duration(i)*time.Second))
	}

	return models.SingleData{Name: target, PVname: target, Values: v}, nil
}

// waitForGoroutines waits for the count to come back down to want, which it
// only does once every request goroutine has delivered and exited.
func waitForGoroutines(t *testing.T, want int) {
	t.Helper()

	deadline := time.Now().Add(5 * time.Second)
	for {
		if runtime.NumGoroutine() <= want {
			return
		}
		if time.Now().After(deadline) {
			t.Fatalf("goroutines did not finish: wanted %v or fewer, still %v", want, runtime.NumGoroutine())
		}
		time.Sleep(10 * time.Millisecond)
	}
}

// A query that gives up before its requests finish used to leave each one
// blocked on an unbuffered send, holding its goroutine and its samples for the
// life of the process.
func TestSingleQueryLeavesNothingBehindWhenCutShort(t *testing.T) {
	const pvs = 4
	client := blockingClient{
		release: make(chan struct{}),
		entered: make(chan struct{}, pvs),
	}

	qm := models.ArchiverQueryModel{
		Target:        "(PV:A|PV:B|PV:C|PV:D)",
		MaxDataPoints: 100,
	}

	baseline := runtime.NumGoroutine()

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		defer close(done)
		singleQuery(ctx, qm, client, models.DatasourceSettings{})
	}()

	// Cut the query short with every request still in flight.
	for i := 0; i < pvs; i++ {
		<-client.entered
	}
	cancel()

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("singleQuery did not return after its context was cancelled")
	}

	// Let the requests finish. Their sends have to complete for the goroutines
	// to exit, which is what an unbuffered channel would prevent.
	close(client.release)
	waitForGoroutines(t, baseline)
}

package models

import (
	"testing"
	"time"
)

// A window of one averages each value with itself, and anything smaller has no
// meaning, so both leave the series alone. functions rejects the parameter
// before it gets here; this only pins down that nothing is destroyed if it does.
func TestMovingAverageWindowBelowTwo(t *testing.T) {
	for _, windowSize := range []int{1, 0, -1} {
		v := NewSclars(3)
		base := time.Date(2021, time.January, 10, 0, 0, 0, 0, time.UTC)
		v.AppendConcrete(1, base)
		v.Append(nil, base.Add(time.Second))
		v.AppendConcrete(3, base.Add(2*time.Second))

		v.MovingAverage(windowSize)

		if *v.Values[0] != 1 || *v.Values[2] != 3 {
			t.Errorf("window %d: wanted the values left alone, got %v and %v", windowSize, *v.Values[0], *v.Values[2])
		}
		if v.Values[1] != nil {
			t.Errorf("window %d: a nil sample should stay nil", windowSize)
		}
	}
}

// Delta used to index Times[0] to build its one-element fallback, which panicked
// when there was nothing to index.
func TestDeltaOnEmptyScalars(t *testing.T) {
	v := NewSclars(0)

	v.Delta()

	if len(v.Values) != 0 || len(v.Times) != 0 {
		t.Errorf("wanted an empty series, got %d values and %d times", len(v.Values), len(v.Times))
	}
}

package models

import (
	"testing"
	"time"
)

// Benchmarks for the Scalars container.
//
// These methods run once per data point of every series in a query, so
// allocs/op is the metric that matters: the container currently stores
// []*float64, and each write through SetValConcrete allocates a fresh float64.

const benchPoints = 100000

// makeBenchScalars builds a series shaped like parser output: individually
// allocated float64 values, with a nil every nilEvery points to represent
// invalid samples (nilEvery <= 0 means no nils).
func makeBenchScalars(n int, nilEvery int) *Scalars {
	v := NewSclars(n)
	base := time.Date(2021, time.January, 10, 0, 0, 0, 0, time.UTC)
	for i := 0; i < n; i++ {
		t := base.Add(time.Duration(i) * time.Millisecond)
		if nilEvery > 0 && i%nilEvery == 0 {
			v.Append(nil, t)
			continue
		}
		v.AppendConcrete(float64(i%1000), t)
	}
	return v
}

func benchmarkInPlace(b *testing.B, nilEvery int, fn func(v *Scalars)) {
	v := makeBenchScalars(benchPoints, nilEvery)
	b.ReportAllocs()
	for b.Loop() {
		fn(v)
	}
}

func BenchmarkScalarsScale(b *testing.B) {
	// factor 1 keeps the values finite across iterations without changing the
	// amount of work done.
	benchmarkInPlace(b, 0, func(v *Scalars) { v.Scale(1) })
}

func BenchmarkScalarsScaleWithNils(b *testing.B) {
	benchmarkInPlace(b, 10, func(v *Scalars) { v.Scale(1) })
}

func BenchmarkScalarsOffset(b *testing.B) {
	benchmarkInPlace(b, 0, func(v *Scalars) { v.Offset(0) })
}

func BenchmarkScalarsFluctuation(b *testing.B) {
	benchmarkInPlace(b, 0, func(v *Scalars) { v.Fluctuation() })
}

func BenchmarkScalarsRankAvg(b *testing.B) {
	benchmarkInPlace(b, 0, func(v *Scalars) {
		if _, err := v.Rank(RANKTYPE_AVG); err != nil {
			b.Fatal(err)
		}
	})
}

// Delta drops the first point on every call, so the input has to be rebuilt
// each iteration to keep the measured work constant.
func BenchmarkScalarsDelta(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		b.StopTimer()
		v := makeBenchScalars(benchPoints, 0)
		b.StartTimer()
		v.Delta()
	}
}

// MovingAverage is currently O(n * windowSize); the window sizes below show
// how the cost scales with a user-supplied window.
func BenchmarkScalarsMovingAverage(b *testing.B) {
	for _, window := range []int{10, 100, 1000} {
		b.Run(windowName(window), func(b *testing.B) {
			benchmarkInPlace(b, 0, func(v *Scalars) { v.MovingAverage(window) })
		})
	}
}

func windowName(window int) string {
	switch window {
	case 10:
		return "Window10"
	case 100:
		return "Window100"
	default:
		return "Window1000"
	}
}

// ToFields is where the SDK deep-copies both the timestamps and the values
// into the frame, so it is a full extra traversal of the dataset.
func BenchmarkScalarsToFields(b *testing.B) {
	benchmarkInPlace(b, 0, func(v *Scalars) {
		_ = v.ToFields("TEST:PV:NAME", "TEST:PV:NAME", FormatOption(FORMAT_TIMESERIES))
	})
}

// Extrapolation appends a single point, which forces a reallocation and copy
// of both Times and Values whenever the slices are at capacity.
func BenchmarkScalarsExtrapolation(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		b.StopTimer()
		v := makeBenchScalars(benchPoints, 0)
		last := v.Times[len(v.Times)-1].Add(time.Second)
		b.StartTimer()
		v.Extrapolation(last)
	}
}

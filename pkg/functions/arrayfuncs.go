package functions

import (
	"math"
	"slices"
)

// The reductions behind the toScalarBy* functions. Each one takes one waveform
// and returns the single value that replaces it.
//
// An empty waveform has no value to reduce to, so these report NaN. The caller
// discards the error and stores what it gets, which puts a gap in the series.

func sum(values []float64) float64 {
	if len(values) == 0 {
		return math.NaN()
	}

	var total float64
	for _, v := range values {
		total += v
	}

	return total
}

func mean(values []float64) float64 {
	if len(values) == 0 {
		return math.NaN()
	}

	return sum(values) / float64(len(values))
}

func maximum(values []float64) float64 {
	if len(values) == 0 {
		return math.NaN()
	}

	m := values[0]
	for _, v := range values[1:] {
		if v > m {
			m = v
		}
	}

	return m
}

func minimum(values []float64) float64 {
	if len(values) == 0 {
		return math.NaN()
	}

	m := values[0]
	for _, v := range values[1:] {
		if v < m {
			m = v
		}
	}

	return m
}

// standardDeviation is the population deviation, dividing by n rather than
// n-1. The frontend's own toScalarByStd uses the sample deviation, so the two
// query paths disagree; this keeps the backend as it was.
func standardDeviation(values []float64) float64 {
	if len(values) == 0 {
		return math.NaN()
	}

	m := mean(values)
	var variance float64
	for _, v := range values {
		variance += (v - m) * (v - m)
	}

	return math.Sqrt(variance / float64(len(values)))
}

// medianOver sorts into scratch so that one buffer serves every waveform of a
// series rather than one allocation per waveform. The caller keeps the buffer
// and passes back what it gets.
func medianOver(values []float64, scratch []float64) (float64, []float64) {
	if len(values) == 0 {
		return math.NaN(), scratch
	}

	scratch = append(scratch[:0], values...)
	slices.Sort(scratch)

	mid := len(scratch) / 2
	if len(scratch)%2 != 0 {
		return scratch[mid], scratch
	}

	return (scratch[mid-1] + scratch[mid]) / 2, scratch
}

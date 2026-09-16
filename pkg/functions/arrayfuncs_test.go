package functions

import (
	"math"
	"testing"
)

// Tests

func equalFloat(a, b float64) bool {
	if math.IsNaN(a) && math.IsNaN(b) {
		return true
	}
	return a == b
}

func TestReductions(t *testing.T) {
	var tests = []struct {
		name   string
		f      func([]float64) float64
		input  []float64
		output float64
	}{
		{name: "sum", f: sum, input: []float64{1, 2, 3, 4}, output: 10},
		{name: "sum of one", f: sum, input: []float64{2.5}, output: 2.5},
		{name: "sum of none", f: sum, input: []float64{}, output: math.NaN()},
		{name: "mean", f: mean, input: []float64{1, 2, 3, 4}, output: 2.5},
		{name: "mean of one", f: mean, input: []float64{2.5}, output: 2.5},
		{name: "mean of none", f: mean, input: []float64{}, output: math.NaN()},
		{name: "maximum", f: maximum, input: []float64{3, 1, 4, 1, 5}, output: 5},
		{name: "maximum at the front", f: maximum, input: []float64{5, 1, 4}, output: 5},
		{name: "maximum of one", f: maximum, input: []float64{-2}, output: -2},
		{name: "maximum of none", f: maximum, input: []float64{}, output: math.NaN()},
		{name: "minimum", f: minimum, input: []float64{3, 1, 4, 1, 5}, output: 1},
		{name: "minimum at the front", f: minimum, input: []float64{-1, 1, 4}, output: -1},
		{name: "minimum of one", f: minimum, input: []float64{-2}, output: -2},
		{name: "minimum of none", f: minimum, input: []float64{}, output: math.NaN()},
		// Population deviation: 2 divides by 4, not by 3.
		{name: "standardDeviation", f: standardDeviation, input: []float64{1, 3, 5, 7}, output: math.Sqrt(5)},
		{name: "standardDeviation of one", f: standardDeviation, input: []float64{4}, output: 0},
		{name: "standardDeviation of none", f: standardDeviation, input: []float64{}, output: math.NaN()},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			result := testCase.f(testCase.input)
			if !equalFloat(result, testCase.output) {
				t.Errorf("got %v, want %v", result, testCase.output)
			}
		})
	}
}

func TestMedianOver(t *testing.T) {
	var tests = []struct {
		name   string
		input  []float64
		output float64
	}{
		{name: "odd length takes the middle", input: []float64{3, 1, 2}, output: 2},
		{name: "even length averages the middle two", input: []float64{4, 1, 3, 2}, output: 2.5},
		{name: "one sample", input: []float64{7}, output: 7},
		{name: "no samples", input: []float64{}, output: math.NaN()},
	}

	// The same buffer serves every call, as it does for a whole series.
	var scratch []float64
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			input := append([]float64(nil), testCase.input...)

			var result float64
			result, scratch = medianOver(input, scratch)
			if !equalFloat(result, testCase.output) {
				t.Errorf("got %v, want %v", result, testCase.output)
			}

			for idx := range input {
				if input[idx] != testCase.input[idx] {
					t.Errorf("input was sorted in place at %v: got %v, want %v", idx, input[idx], testCase.input[idx])
				}
			}
		})
	}
}

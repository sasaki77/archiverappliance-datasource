package main

import (
	"fmt"
	"testing"
)

func TestOperatorValidator(t *testing.T) {
	var tests = []struct {
		input  string
		output bool
	}{
		{input: "firstSample", output: true},
		{input: "lastFill", output: true},
		{input: "lastFill_16", output: false},
		{input: "snakes", output: false},
	}
	for idx, testCase := range tests {
		testName := fmt.Sprintf("%d: %s, %t", idx, testCase.input, testCase.output)
		t.Run(testName, func(t *testing.T) {
			result := OperatorValidator(testCase.input)
			if testCase.output != result {
				t.Errorf("got %v, want %v", result, testCase.output)
			}
		})
	}
}

func TestCreateOperatorQuery(t *testing.T) {
	var tests = []struct {
		name   string
		input  ArchiverQueryModel
		output string
	}{
		{
			name: "mean with 10 second Interval",
			input: ArchiverQueryModel{
				Operator: "mean",
				Interval: 10,
			},
			output: "mean_10",
		},
		{
			name: "raw with binInterval function",
			input: ArchiverQueryModel{
				Operator: "raw",
				Interval: 16,
			},
			output: "",
		},
		{
			name: "empty operator with 0 second interval",
			input: ArchiverQueryModel{
				IntervalMs: InitIntPointer(100),
				Operator:   "",
				Interval:   0,
			},
			output: "",
		},
		{
			name: "empty operator with 0 second interval and DisableAutoRaw true",
			input: ArchiverQueryModel{
				IntervalMs:     InitIntPointer(100),
				Operator:       "",
				Interval:       0,
				DisableAutoRaw: true,
			},
			output: "mean_1",
		},
		{
			name: "max operator with 10 second interval",
			input: ArchiverQueryModel{
				Operator: "max",
				Interval: 10,
			},
			output: "max_10",
		},
		{
			name: "max operator with 0.1 second interval",
			input: ArchiverQueryModel{
				Operator: "max",
				Interval: 0,
			},
			output: "",
		},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			result, err := CreateOperatorQuery(testCase.input)
			if err != nil {
				t.Errorf("Error received %v", err)
			}
			if testCase.output != result {
				t.Errorf("got %v, want %v", result, testCase.output)
			}
		})
	}
}

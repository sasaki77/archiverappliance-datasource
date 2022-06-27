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
			name: "mean with binInterval function",
			input: ArchiverQueryModel{
				Functions: []FunctionDescriptorQueryModel{
					{
						Def: FuncDefQueryModel{
							Category:      "Options",
							DefaultParams: InitRawMsg(`[16]`),
							Name:          "binInterval",
							Params: []FuncDefParamQueryModel{
								{Name: "interval", Type: "int"},
							},
						},
						Params: []string{"16"},
					},
				},
				Operator: "mean",
			},
			output: "mean_16",
		},
		{
			name: "raw with binInterval function",
			input: ArchiverQueryModel{
				Functions: []FunctionDescriptorQueryModel{
					{
						Def: FuncDefQueryModel{
							Category:      "Options",
							DefaultParams: InitRawMsg(`[16]`),
							Name:          "binInterval",
							Params: []FuncDefParamQueryModel{
								{Name: "interval", Type: "int"},
							},
						},
						Params: []string{"16"},
					},
				},
				Operator: "raw",
			},
			output: "",
		},
		{
			name: "empty operator with 10.1 second interval",
			input: ArchiverQueryModel{
				IntervalMs: InitIntPointer(10100),
				Operator:   "",
			},
			output: "mean_10",
		},
		{
			name: "empty operator with 0.1 second interval",
			input: ArchiverQueryModel{
				IntervalMs: InitIntPointer(100),
				Operator:   "",
			},
			output: "",
		},
		{
			name: "max operator with 10 second interval",
			input: ArchiverQueryModel{
				IntervalMs: InitIntPointer(10100),
				Operator:   "max",
			},
			output: "max_10",
		},
		{
			name: "max operator with 0.1 second interval",
			input: ArchiverQueryModel{
				IntervalMs: InitIntPointer(100),
				Operator:   "max",
			},
			output: "",
		},
		{
			name: "max operator with 0.1 second interval",
			input: ArchiverQueryModel{
				Operator: "max",
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

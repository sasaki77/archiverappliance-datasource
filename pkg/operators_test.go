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

func TestLoadInterval(t *testing.T) {
	var tests = []struct {
		name   string
		input  ArchiverQueryModel
		output int
	}{
		{
			name: "Empty Operator with binInterval function",
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
				Operator: "",
			},
			output: 16,
		},
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
			output: 16,
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
			output: 0,
		},
		{
			name: "last with binInterval function",
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
				Operator: "last",
			},
			output: 0,
		},
		{
			name: "empty operator with 10.1 second interval",
			input: ArchiverQueryModel{
				IntervalMs: InitIntPointer(10100),
				Operator:   "",
			},
			output: 10,
		},
		{
			name: "empty operator with 0.1 second interval",
			input: ArchiverQueryModel{
				IntervalMs: InitIntPointer(100),
				Operator:   "",
			},
			output: 0,
		},
		{
			name: "max operator with 10 second interval",
			input: ArchiverQueryModel{
				IntervalMs: InitIntPointer(10100),
				Operator:   "max",
			},
			output: 10,
		},
		{
			name: "raw operator with 10 second interval",
			input: ArchiverQueryModel{
				IntervalMs: InitIntPointer(10100),
				Operator:   "raw",
			},
			output: 0,
		},
		{
			name: "last operator with 10 second interval",
			input: ArchiverQueryModel{
				IntervalMs: InitIntPointer(10100),
				Operator:   "raw",
			},
			output: 0,
		},
		{
			name: "max operator without IntervalMs",
			input: ArchiverQueryModel{
				Operator: "max",
			},
			output: 0,
		},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			result, err := loadInterval(testCase.input)
			if err != nil {
				t.Errorf("Error received %v", err)
			}
			if testCase.output != result {
				t.Errorf("got %v, want %v", result, testCase.output)
			}
		})
	}
}

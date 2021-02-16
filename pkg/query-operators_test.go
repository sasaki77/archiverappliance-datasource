package main

import (
    "fmt"
    "testing"
)

func TestOperatorValidator(t *testing.T) {
    var tests = []struct{
        input string
        output bool
    }{
        {input: "firstSample",     output: true},
        {input: "lastFill",        output: true},
        {input: "lastFill_16",     output: false},
        {input: "snakes",          output: false},
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
    var tests = []struct{
        input ArchiverQueryModel
        output string
    }{
        {
            input: ArchiverQueryModel{
                Functions: []FunctionDescriptorQueryModel{
                    {
                        Def: FuncDefQueryModel{
                            Category: "Options",
                            DefaultParams: InitRawMsg(`[16]`),
                            Name: "binInterval",
                            Params:[]FuncDefParamQueryModel{
                                {Name:"interval", Type: "int"},
                            },
                        },
                        Params: []string{"16",},
                    },
                },
                Operator: "mean",
            },
            output: "mean_16",
        },
        {
            input: ArchiverQueryModel{
                Functions: []FunctionDescriptorQueryModel{
                    {
                        Def: FuncDefQueryModel{
                            Category: "Options",
                            DefaultParams: InitRawMsg(`[16]`),
                            Name: "binInterval",
                            Params:[]FuncDefParamQueryModel{
                                {Name:"interval", Type: "int"},
                            },
                        },
                        Params: []string{"16",},
                    },
                },
                Operator: "raw",
            },
            output: "",
        },
    }
    for idx, testCase := range tests {
        testName := fmt.Sprintf("%d: %v, %v", idx, testCase.input, testCase.output)
        t.Run(testName, func(t *testing.T) {
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

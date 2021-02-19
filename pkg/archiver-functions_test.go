package main

import (
    "fmt"
    "testing"
	//"github.com/grafana/grafana-plugin-sdk-go/backend/log"
)

// Tests

func TestIdentifyFunctionsByName(t *testing.T) {
    var tests = []struct{
        input ArchiverQueryModel
        inputString string
        output []FunctionDescriptorQueryModel
    }{
        {
            input: ArchiverQueryModel{Functions: []FunctionDescriptorQueryModel{{Def: FuncDefQueryModel{Name: "binInterval"}}}},
            inputString: "binInterval",
            output: []FunctionDescriptorQueryModel{{Def: FuncDefQueryModel{Name: "binInterval"}}},
        },
        {
            input: ArchiverQueryModel{Functions: []FunctionDescriptorQueryModel{{Def: FuncDefQueryModel{Name: "binInterval"}}}},
            inputString: "binInterval-fake",
            output: []FunctionDescriptorQueryModel{},
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
                        Params: []string{"[16]",},
                    },
                },
                Operator: "mean",
            },
            inputString: "binInterval",
            output: []FunctionDescriptorQueryModel{{Def: FuncDefQueryModel{Name: "binInterval"}}},
        },
    }
    for tdx, testCase := range tests {
        testName := fmt.Sprintf("%d: %v, %v", tdx, testCase.input, testCase.output)
        t.Run(testName, func(t *testing.T) {
            result := testCase.input.IdentifyFunctionsByName(testCase.inputString)
            if len(result) != len(testCase.output) {
                t.Errorf("lengths differ: got %v, want %v", len(result), len(testCase.output))
            }
            for idx, out := range result {
                if out.Def.Name != testCase.output[idx].Def.Name {
                    t.Errorf("got %v, want %v", out, result[idx])
                }
            }
        })
    }
}

func TestGetParametersByName(t *testing.T) {
    var tests = []struct{
        input FunctionDescriptorQueryModel
        targetArg string
        output *string
    }{
		{
		    input: FunctionDescriptorQueryModel{
		        Def: FuncDefQueryModel{
		            Fake: nil,
		            Category: "Options",
		            DefaultParams: InitRawMsg(`[900]`),
		            Name: "binInterval",
                    Params: []FuncDefParamQueryModel{
		                {
		                    Name: "interval",
		                    Options: nil,
		                    Type: "int",
		                },
		            },
		        },
		        Params:[]string{"900"},
		    },
            targetArg: "interval",
            output: InitString("900"),
		},
        {
		    input: FunctionDescriptorQueryModel{
		        Def: FuncDefQueryModel{
		            Fake: nil,
		            Category: "Filter Series",
		            DefaultParams: InitRawMsg(`[5 avg]`),
		            Name: "bottom",
		            Params: []FuncDefParamQueryModel{
		                {
		                    Name: "number",
		                    Options: nil,
		                    Type: "int",
		                },
		                {
		                    Name: "value",
		                    Options: &[]string{"avg", "min", "max", "absoluteMin", "absoluteMax" ,"sum"},
		                    Type: "string",
		                },
		            },
		        },
		        Params: []string{"5", "avg"},
		    },
            targetArg: "value",
            output: InitString("avg"),
        },
        {
		    input: FunctionDescriptorQueryModel{
		        Def: FuncDefQueryModel{
		            Fake: nil,
		            Category: "Transform",
		            DefaultParams: InitRawMsg(`[100]`),
		            Name: "offset",
		            Params: []FuncDefParamQueryModel{
		                {
		                    Name: "delta",
		                    Options: nil,
		                    Type: "float",
		                },
		            },
		        },
		        Params:[]string{"100"},
		    },
            targetArg: "delta",
            output: InitString("100"),
        },
        {
		    input: FunctionDescriptorQueryModel{
		        Def: FuncDefQueryModel{
		            Fake: nil,
		            Category: "Transform",
		            DefaultParams: InitRawMsg(`[]`),
		            Name: "delta",
		            Params: []FuncDefParamQueryModel{},
		        },
		        Params:[]string{},
		    },
            targetArg: "delta",
            output: nil,
		},
    }

    for tdx, testCase := range tests {
        testName := fmt.Sprintf("case %d: %v", tdx, testCase.output)
        t.Run(testName, func(t *testing.T) {
            result, err := testCase.input.GetParametersByName(testCase.targetArg)
            if testCase.output == nil {
                if err == nil {
                    t.Errorf("An error was expected but not received. Bad output: %v", result)
                }
                if result != "" {
                    t.Errorf("Expected output was \"\" but something else was received. Bad output: %v", result)
                }
            } else {
                if err != nil {
                    t.Errorf("An error was received but not expected. Bad output: %v", result)
                }
                if result != *testCase.output {
                    t.Errorf("Incorrect result: got %v, want %v", result, *testCase.output)
                }
            }
        })
    }
}

func TestGetParamTypeByName(t *testing.T) {
    var tests = []struct{
        input FunctionDescriptorQueryModel
        targetArg string
        output *string
    }{
		{
		    input: FunctionDescriptorQueryModel{
		        Def: FuncDefQueryModel{
		            Fake: nil,
		            Category: "Options",
		            DefaultParams: InitRawMsg(`[900]`),
		            Name: "binInterval",
                    Params: []FuncDefParamQueryModel{
		                {
		                    Name: "interval",
		                    Options: nil,
		                    Type: "int",
		                },
		            },
		        },
		        Params:[]string{"900"},
		    },
            targetArg: "interval",
            output: InitString("int"),
		},
        {
		    input: FunctionDescriptorQueryModel{
		        Def: FuncDefQueryModel{
		            Fake: nil,
		            Category: "Filter Series",
		            DefaultParams: InitRawMsg(`[5 avg]`),
		            Name: "bottom",
		            Params: []FuncDefParamQueryModel{
		                {
		                    Name: "number",
		                    Options: nil,
		                    Type: "int",
		                },
		                {
		                    Name: "value",
		                    Options: &[]string{"avg", "min", "max", "absoluteMin", "absoluteMax" ,"sum"},
		                    Type: "string",
		                },
		            },
		        },
		        Params: []string{"5", "avg"},
		    },
            targetArg: "value",
            output: InitString("string"),
        },
        {
		    input: FunctionDescriptorQueryModel{
		        Def: FuncDefQueryModel{
		            Fake: nil,
		            Category: "Transform",
		            DefaultParams: InitRawMsg(`[100]`),
		            Name: "offset",
		            Params: []FuncDefParamQueryModel{
		                {
		                    Name: "delta",
		                    Options: nil,
		                    Type: "float",
		                },
		            },
		        },
		        Params:[]string{"100"},
		    },
            targetArg: "delta",
            output: InitString("float"),
        },
        {
		    input: FunctionDescriptorQueryModel{
		        Def: FuncDefQueryModel{
		            Fake: nil,
		            Category: "Transform",
		            DefaultParams: InitRawMsg(`[]`),
		            Name: "delta",
		            Params: []FuncDefParamQueryModel{},
		        },
		        Params:[]string{},
		    },
            targetArg: "delta",
            output: nil,
		},
    }

    for tdx, testCase := range tests {
        testName := fmt.Sprintf("case %d: %v", tdx, testCase.output)
        t.Run(testName, func(t *testing.T) {
            result, err := testCase.input.GetParamTypeByName(testCase.targetArg)
            if testCase.output == nil {
                if err == nil {
                    t.Errorf("An error was expected but not received. Bad output: %v", result)
                }
                if result != "" {
                    t.Errorf("Expected output was \"\" but something else was received. Bad output: %v", result)
                }
            } else {
                if err != nil {
                    t.Errorf("An error was received but not expected. Bad output: %v", result)
                }
                if result != *testCase.output {
                    t.Errorf("Incorrect result: got %v, want %v", result, *testCase.output)
                }
            }
        })
    }
}

func TestExtractParamInt(t *testing.T) {
    var tests = []struct{
        input FunctionDescriptorQueryModel
        targetArg string
        output int
    }{
        {
		    input: FunctionDescriptorQueryModel{
		        Def: FuncDefQueryModel{
		            Fake: nil,
		            Category: "Filter Series",
		            DefaultParams: InitRawMsg(`[5 avg]`),
		            Name: "bottom",
		            Params: []FuncDefParamQueryModel{
		                {
		                    Name: "number",
		                    Options: nil,
		                    Type: "int",
		                },
		                {
		                    Name: "value",
		                    Options: &[]string{"avg", "min", "max", "absoluteMin", "absoluteMax" ,"sum"},
		                    Type: "string",
		                },
		            },
		        },
		        Params: []string{"5", "avg"},
		    },
            targetArg: "number",
            output: 5,
        },
    }
    for tdx, testCase := range tests {
        testName := fmt.Sprintf("case %d: %v", tdx, testCase.output)
        t.Run(testName, func(t *testing.T) {
            result, err := testCase.input.ExtractParamInt(testCase.targetArg)
            if err != nil {
                t.Errorf("Error received")
            }
            if result != testCase.output {
                t.Errorf(fmt.Sprintf("Got %v, wanted %v", result, testCase.output))
            }
        })
    }
}

func TestExtractParamFloat64(t *testing.T) {
    var tests = []struct{
        input FunctionDescriptorQueryModel
        targetArg string
        output float64
    }{
        {
		    input: FunctionDescriptorQueryModel{
		        Def: FuncDefQueryModel{
		            Fake: nil,
		            Category: "Transform",
		            DefaultParams: InitRawMsg(`[100]`),
		            Name: "offset",
		            Params: []FuncDefParamQueryModel{
		                {
		                    Name: "delta",
		                    Options: nil,
		                    Type: "float",
		                },
		            },
		        },
		        Params:[]string{"100"},
		    },
            targetArg: "delta",
            output: 100,
        },
    }
    for tdx, testCase := range tests {
        testName := fmt.Sprintf("case %d: %v", tdx, testCase.output)
        t.Run(testName, func(t *testing.T) {
            result, err := testCase.input.ExtractParamFloat64(testCase.targetArg)
            if err != nil {
                t.Errorf("Error received")
            }
            if result != testCase.output {
                t.Errorf(fmt.Sprintf("Got %v, wanted %v", result, testCase.output))
            }
        })
    }
}

func TestExtractParamString(t *testing.T) {
    var tests = []struct{
        input FunctionDescriptorQueryModel
        targetArg string
        output string
    }{
        {
		    input: FunctionDescriptorQueryModel{
		        Def: FuncDefQueryModel{
		            Fake: nil,
		            Category: "Filter Series",
		            DefaultParams: InitRawMsg(`[5 avg]`),
		            Name: "bottom",
		            Params: []FuncDefParamQueryModel{
		                {
		                    Name: "number",
		                    Options: nil,
		                    Type: "int",
		                },
		                {
		                    Name: "value",
		                    Options: &[]string{"avg", "min", "max", "absoluteMin", "absoluteMax" ,"sum"},
		                    Type: "string",
		                },
		            },
		        },
		        Params: []string{"5", "avg"},
		    },
            targetArg: "value",
            output: "avg",
        },
    }
    for tdx, testCase := range tests {
        testName := fmt.Sprintf("case %d: %v", tdx, testCase.output)
        t.Run(testName, func(t *testing.T) {
            result, err := testCase.input.ExtractParamString(testCase.targetArg)
            if err != nil {
                t.Errorf("Error received")
            }
            if result != testCase.output {
                t.Errorf(fmt.Sprintf("Got %v, wanted %v", result, testCase.output))
            }
        })
    }
}

func TestApplyFunctions(t *testing.T) {
    var tests = []struct{
        inputSd []SingleData
        inputAqm ArchiverQueryModel
        output []SingleData
    }{
        {
            inputSd: []SingleData{
                {
                    Times: TimeArrayHelper(0,6),
                    Values: []float64{1,1,2,3,5,8},
                },
            },
            inputAqm: ArchiverQueryModel{
                Functions: []FunctionDescriptorQueryModel{
                    {
                        Def: FuncDefQueryModel{
                            Category: "Transform",
                            Name: "offset",
                            Params: []FuncDefParamQueryModel{
                                    {
                                        Name: "delta",
                                        Type: "float",
                                    },
                            },
                        },
                        Params: []string{"2"},
                    },
                },
            },
            output: []SingleData{
                {
                    Times: TimeArrayHelper(0,6),
                    Values: []float64{3,3,4,5,7,10},
                },
            },
        },
        {
            inputSd: []SingleData{
                {
                    Times: TimeArrayHelper(0,6),
                    Values: []float64{1,1,2,3,5,8},
                },
            },
            inputAqm: ArchiverQueryModel{
                Functions: []FunctionDescriptorQueryModel{
                    {
                        Def: FuncDefQueryModel{
                            Category: "Transform",
                            Name: "offset",
                            Params: []FuncDefParamQueryModel{
                                    {
                                        Name: "delta",
                                        Type: "float",
                                    },
                            },
                        },
                        Params: []string{"2"},
                    },
                    {
                        Def: FuncDefQueryModel{
                            Category: "Transform",
                            Name: "scale",
                            Params: []FuncDefParamQueryModel{
                                {
                                    Name: "factor",
                                    Type: "float",
                                },
                            },
                        },
                        Params: []string{"3"},
                    },
                },
            },
            output: []SingleData{
                {
                    Times: TimeArrayHelper(0,6),
                    Values: []float64{9,9,12,15,21,30},
                },
            },
        },
        {
            inputSd: []SingleData{
                {
                    Times: TimeArrayHelper(0,6),
                    Values: []float64{1,1,2,3,5,8},
                },
                {
                    Times: TimeArrayHelper(0,6),
                    Values: []float64{100, 100, 100, 100, 100, 100},
                },
            },
            inputAqm: ArchiverQueryModel{
                Functions: []FunctionDescriptorQueryModel{
                    {
                        Def: FuncDefQueryModel{
                            Category: "Transform",
                            Name: "offset",
                            Params: []FuncDefParamQueryModel{
                                    {
                                        Name: "delta",
                                        Type: "float",
                                    },
                            },
                        },
                        Params: []string{"2"},
                    },
                    {
                        Def: FuncDefQueryModel{
                            Category: "Transform",
                            Name: "scale",
                            Params: []FuncDefParamQueryModel{
                                {
                                    Name: "factor",
                                    Type: "float",
                                },
                            },
                        },
                        Params: []string{"3"},
                    },
                    {
		                Def: FuncDefQueryModel{
		                    Fake: nil,
		                    Category: "Filter Series",
		                    DefaultParams: InitRawMsg(`[1 avg]`),
		                    Name: "bottom",
                            Params: []FuncDefParamQueryModel{
		                        {
		                            Name: "number",
		                            Options: nil,
		                            Type: "int",
		                        },
		                        {
		                            Name: "value",
		                            Options: &[]string{"avg", "min", "max", "absoluteMin", "absoluteMax" ,"sum"},
		                            Type: "string",
		                        },
		                    },
                        },
		                Params: []string{"1", "avg"},
                    },
                },
            },
            output: []SingleData{
                {
                    Times: TimeArrayHelper(0,6),
                    Values: []float64{9,9,12,15,21,30},
                },
            },
        },
    }

    for tdx, testCase := range tests {
        testName := fmt.Sprintf("case %d: %v", tdx, testCase.output)
        t.Run(testName, func(t *testing.T) {
            result, err := ApplyFunctions(tests[tdx].inputSd, testCase.inputAqm)
            if err != nil {
                t.Errorf("An error has been generated")
            }
            SingleDataCompareHelper(result, testCase.output, t)
        })
    }
}

func TestFunctionSelector(t *testing.T) {
    var tests = []struct{
        inputSd []SingleData
        inputFdqm FunctionDescriptorQueryModel
        output []SingleData
    }{
        {
            inputSd: []SingleData{
                {
                    Times: TimeArrayHelper(0,6),
                    Values: []float64{1,1,2,3,5,8},
                },
            },
            inputFdqm: FunctionDescriptorQueryModel{
                Def: FuncDefQueryModel{
                    Category: "Transform",
                    Name: "offset",
                    Params: []FuncDefParamQueryModel{
                        {
                            Name: "delta",
                            Type: "float",
                        },
                    },
                },
                Params: []string{"2"},
            },
            output: []SingleData{
                {
                    Times: TimeArrayHelper(0,6),
                    Values: []float64{3,3,4,5,7,10},
                },
            },
        },
    }

    for tdx, testCase := range tests {
        testName := fmt.Sprintf("case %d: %v", tdx, testCase.output)
        t.Run(testName, func(t *testing.T) {
            result, err := FunctionSelector(tests[tdx].inputSd, testCase.inputFdqm)
            if err != nil {
                t.Errorf("An error has been generated")
            }
            SingleDataCompareHelper(result, testCase.output, t)
        })
    }
}




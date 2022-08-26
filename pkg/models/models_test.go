package models

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/sasaki77/archiverappliance-datasource/pkg/testhelper"

	"github.com/google/go-cmp/cmp"
	"github.com/grafana/grafana-plugin-sdk-go/backend"
)

func TestReadQueryModel(t *testing.T) {
	TIME_FORMAT := "2006-01-02T15:04:05.000-07:00"
	var tests = []struct {
		name   string
		input  backend.DataQuery
		output ArchiverQueryModel
	}{
		{
			name: "Frontend query test",
			input: backend.DataQuery{
				Interval: testhelper.MultiReturnHelperParseDuration(time.ParseDuration("0s")),
				JSON: json.RawMessage(`{
                    		"alias": "test alias",
                    		"aliasPattern": "(.*):(.*)",
                    		"operator": "max",
                    		"refId":"A" ,
                    		"regex":true ,
                    		"target":".*(1|2)" ,
                    		"IntervalMs":5000 ,
							"functions":[
							]
						}`),
				MaxDataPoints: 1000,
				QueryType:     "",
				RefID:         "A",
				TimeRange: backend.TimeRange{
					From: testhelper.MultiReturnHelperParse(time.Parse(TIME_FORMAT, "2021-01-27T14:30:41.678-08:00")),
					To:   testhelper.MultiReturnHelperParse(time.Parse(TIME_FORMAT, "2021-01-28T14:30:41.678-08:00")),
				},
			},
			output: ArchiverQueryModel{
				Target:       ".*(1|2)",
				Alias:        "test alias",
				AliasPattern: "(.*):(.*)",
				Operator:     "max",
				Regex:        true,
				Functions:    []FunctionDescriptorQueryModel{},
				IntervalMs:   testhelper.InitIntPointer(5000),
				RefId:        "A",
				TimeRange: backend.TimeRange{
					From: testhelper.MultiReturnHelperParse(time.Parse(TIME_FORMAT, "2021-01-27T14:30:41.678-08:00")),
					To:   testhelper.MultiReturnHelperParse(time.Parse(TIME_FORMAT, "2021-01-28T14:30:41.678-08:00")),
				},
				Interval:        5,
				MaxNumPVs:       1000,
				BackendQuery:    false,
				DisableExtrapol: false,
			},
		},
		{
			name: "Backend query test",
			input: backend.DataQuery{
				Interval: testhelper.MultiReturnHelperParseDuration(time.ParseDuration("0s")),
				JSON: json.RawMessage(`{
                    		"alias": "",
                    		"aliasPattern": "",
                    		"operator": "",
                    		"refId":"A" ,
                    		"regex":false ,
                    		"target":"PV:TEST" ,
							"functions":[
								{
									"params": [
										"true"
									],
									"def": {
										"category": "Options",
										"name": "disableExtrapol",
										"params": [
											{
												"name": "boolean",
												"options": ["true", "false"],
												"type": "string"
											}
										]
									}
								},
								{
									"params": [
										"true"
									],
									"def": {
										"category": "Options",
										"name": "disableAutoRaw",
										"params": [
											{
												"name": "boolean",
												"options": ["true", "false"],
												"type": "string"
											}
										]
									}
								},
								{
									"params": [
										"100"
									],
									"def": {
										"category": "Options",
										"name": "maxNumPVs",
										"params": [
											{
												"name": "number",
												"type": "int"
											}
										]
									}
								}
							]
						}`),
				MaxDataPoints: 1000,
				QueryType:     "",
				RefID:         "A",
				TimeRange: backend.TimeRange{
					From: testhelper.MultiReturnHelperParse(time.Parse(TIME_FORMAT, "2021-01-27T14:30:41.678-08:00")),
					To:   testhelper.MultiReturnHelperParse(time.Parse(TIME_FORMAT, "2021-01-28T14:30:41.678-08:00")),
				},
			},
			output: ArchiverQueryModel{
				Target:       "PV:TEST",
				Alias:        "",
				AliasPattern: "",
				Operator:     "",
				Regex:        false,
				Functions: []FunctionDescriptorQueryModel{
					{
						Params: []string{"true"},
						Def: FuncDefQueryModel{
							Category: "Options",
							Name:     "disableExtrapol",
							Params: []FuncDefParamQueryModel{
								{
									Name:    "boolean",
									Options: &[]string{"true", "false"},
									Type:    "string",
								},
							},
						},
					},
					{
						Params: []string{"true"},
						Def: FuncDefQueryModel{
							Category: "Options",
							Name:     "disableAutoRaw",
							Params: []FuncDefParamQueryModel{
								{
									Name:    "boolean",
									Options: &[]string{"true", "false"},
									Type:    "string",
								},
							},
						},
					},
					{
						Params: []string{"100"},
						Def: FuncDefQueryModel{
							Category: "Options",
							Name:     "maxNumPVs",
							Params: []FuncDefParamQueryModel{
								{
									Name: "number",
									Type: "int",
								},
							},
						},
					},
				},
				IntervalMs: nil,
				RefId:      "A",
				TimeRange: backend.TimeRange{
					From: testhelper.MultiReturnHelperParse(time.Parse(TIME_FORMAT, "2021-01-27T14:30:41.678-08:00")),
					To:   testhelper.MultiReturnHelperParse(time.Parse(TIME_FORMAT, "2021-01-28T14:30:41.678-08:00")),
				},
				Interval:        0,
				BackendQuery:    true,
				MaxNumPVs:       100,
				DisableAutoRaw:  true,
				DisableExtrapol: true,
			},
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			result, _ := ReadQueryModel(testCase.input)
			if diff := cmp.Diff(testCase.output, result); diff != "" {
				t.Errorf("Compare value is mismatch (-v1 +v2):%s\n", diff)
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
							DefaultParams: testhelper.InitRawMsg(`[16]`),
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
							DefaultParams: testhelper.InitRawMsg(`[16]`),
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
							DefaultParams: testhelper.InitRawMsg(`[16]`),
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
							DefaultParams: testhelper.InitRawMsg(`[16]`),
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
				IntervalMs: testhelper.InitIntPointer(10100),
				Operator:   "",
			},
			output: 10,
		},
		{
			name: "empty operator with 0.1 second interval",
			input: ArchiverQueryModel{
				IntervalMs: testhelper.InitIntPointer(100),
				Operator:   "",
			},
			output: 0,
		},
		{
			name: "max operator with 10 second interval",
			input: ArchiverQueryModel{
				IntervalMs: testhelper.InitIntPointer(10100),
				Operator:   "max",
			},
			output: 10,
		},
		{
			name: "raw operator with 10 second interval",
			input: ArchiverQueryModel{
				IntervalMs: testhelper.InitIntPointer(10100),
				Operator:   "raw",
			},
			output: 0,
		},
		{
			name: "last operator with 10 second interval",
			input: ArchiverQueryModel{
				IntervalMs: testhelper.InitIntPointer(10100),
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

func TestPickFunsByCategories(t *testing.T) {
	var tests = []struct {
		name     string
		input    ArchiverQueryModel
		category []FunctionCategory
		output   []FunctionDescriptorQueryModel
	}{
		{
			name: "Pick Transform and Options",
			input: ArchiverQueryModel{
				Functions: []FunctionDescriptorQueryModel{
					{
						Def: FuncDefQueryModel{
							Category: "Transform",
							Name:     "scale",
						},
					},
					{
						Def: FuncDefQueryModel{
							Category: "Transform",
							Name:     "offset",
						},
					},
					{
						Def: FuncDefQueryModel{
							Category: "Options",
							Name:     "binInterval",
						},
					},
					{
						Def: FuncDefQueryModel{
							Category: "Filter",
							Name:     "Top",
						},
					},
				},
			},
			category: []FunctionCategory{FUNC_CATEGORY_TRANSFORM, FUNC_CATEGORY_OPTIONS},
			output: []FunctionDescriptorQueryModel{
				{Def: FuncDefQueryModel{Name: "scale"}},
				{Def: FuncDefQueryModel{Name: "offset"}},
				{Def: FuncDefQueryModel{Name: "binInterval"}},
			},
		},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			result := testCase.input.PickFuncsByCategories(testCase.category)
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

func TestIdentifyFunctionsByName(t *testing.T) {
	var tests = []struct {
		input       ArchiverQueryModel
		inputString string
		output      []FunctionDescriptorQueryModel
	}{
		{
			input:       ArchiverQueryModel{Functions: []FunctionDescriptorQueryModel{{Def: FuncDefQueryModel{Name: "binInterval"}}}},
			inputString: "binInterval",
			output:      []FunctionDescriptorQueryModel{{Def: FuncDefQueryModel{Name: "binInterval"}}},
		},
		{
			input:       ArchiverQueryModel{Functions: []FunctionDescriptorQueryModel{{Def: FuncDefQueryModel{Name: "binInterval"}}}},
			inputString: "binInterval-fake",
			output:      []FunctionDescriptorQueryModel{},
		},
		{
			input: ArchiverQueryModel{
				Functions: []FunctionDescriptorQueryModel{
					{
						Def: FuncDefQueryModel{
							Category:      "Options",
							DefaultParams: testhelper.InitRawMsg(`[16]`),
							Name:          "binInterval",
							Params: []FuncDefParamQueryModel{
								{Name: "interval", Type: "int"},
							},
						},
						Params: []string{"[16]"},
					},
				},
				Operator: "mean",
			},
			inputString: "binInterval",
			output:      []FunctionDescriptorQueryModel{{Def: FuncDefQueryModel{Name: "binInterval"}}},
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

func TestLoadBooleanOption(t *testing.T) {
	var tests = []struct {
		name     string
		input    ArchiverQueryModel
		option   FunctionOption
		defaultv bool
		disable  bool
		err      bool
	}{
		{
			name: "Test disableExtrapol: true",
			input: ArchiverQueryModel{
				Functions: []FunctionDescriptorQueryModel{
					{
						Def: FuncDefQueryModel{
							Category:      "Options",
							DefaultParams: testhelper.InitRawMsg(`true`),
							Name:          "disableExtrapol",
							Params: []FuncDefParamQueryModel{
								{Name: "boolean", Type: "string"},
							},
						},
						Params: []string{"true"},
					},
				},
			},
			option:  FunctionOption(FUNC_OPTION_DISABLEEXTRAPOL),
			disable: true,
			err:     false,
		},
		{
			name: "Test disableExtrapol: false",
			input: ArchiverQueryModel{
				Functions: []FunctionDescriptorQueryModel{
					{
						Def: FuncDefQueryModel{
							Category:      "Options",
							DefaultParams: testhelper.InitRawMsg(`true`),
							Name:          "disableExtrapol",
							Params: []FuncDefParamQueryModel{
								{Name: "boolean", Type: "string"},
							},
						},
						Params: []string{"false"},
					},
				},
			},
			option:  FunctionOption(FUNC_OPTION_DISABLEEXTRAPOL),
			disable: false,
			err:     false,
		},
		{
			name: "Test disableExtrapol: bad parameter",
			input: ArchiverQueryModel{
				Functions: []FunctionDescriptorQueryModel{
					{
						Def: FuncDefQueryModel{
							Category:      "Options",
							DefaultParams: testhelper.InitRawMsg(`true`),
							Name:          "disableExtrapol",
							Params: []FuncDefParamQueryModel{
								{Name: "boolean", Type: "string"},
							},
						},
						Params: []string{"bad"},
					},
				},
			},
			option:  FunctionOption(FUNC_OPTION_DISABLEEXTRAPOL),
			disable: false,
			err:     true,
		},
		{
			name: "Test no option function: default false",
			input: ArchiverQueryModel{
				Functions: []FunctionDescriptorQueryModel{},
			},
			disable: false,
			err:     false,
		},
		{
			name: "Test no option function: default true",
			input: ArchiverQueryModel{
				Functions: []FunctionDescriptorQueryModel{},
			},
			option:   FunctionOption(FUNC_OPTION_DISABLEEXTRAPOL),
			defaultv: true,
			disable:  true,
			err:      false,
		},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			result, err := testCase.input.LoadBooleanOption(testCase.option, testCase.defaultv)
			if result != testCase.disable {
				t.Errorf("got %v, want %v", result, testCase.disable)
			}
			if (err != nil && testCase.err == false) || (err == nil && testCase.err == true) {
				t.Errorf("Incorrect error state: got %v, want %v", (err != nil), testCase.err)
			}
		})
	}
}

func TestGetParametersByName(t *testing.T) {
	var tests = []struct {
		input     FunctionDescriptorQueryModel
		targetArg string
		output    *string
	}{
		{
			input: FunctionDescriptorQueryModel{
				Def: FuncDefQueryModel{
					Fake:          nil,
					Category:      "Options",
					DefaultParams: testhelper.InitRawMsg(`[900]`),
					Name:          "binInterval",
					Params: []FuncDefParamQueryModel{
						{
							Name:    "interval",
							Options: nil,
							Type:    "int",
						},
					},
				},
				Params: []string{"900"},
			},
			targetArg: "interval",
			output:    testhelper.InitString("900"),
		},
		{
			input: FunctionDescriptorQueryModel{
				Def: FuncDefQueryModel{
					Fake:          nil,
					Category:      "Filter Series",
					DefaultParams: testhelper.InitRawMsg(`[5 avg]`),
					Name:          "bottom",
					Params: []FuncDefParamQueryModel{
						{
							Name:    "number",
							Options: nil,
							Type:    "int",
						},
						{
							Name:    "value",
							Options: &[]string{"avg", "min", "max", "absoluteMin", "absoluteMax", "sum"},
							Type:    "string",
						},
					},
				},
				Params: []string{"5", "avg"},
			},
			targetArg: "value",
			output:    testhelper.InitString("avg"),
		},
		{
			input: FunctionDescriptorQueryModel{
				Def: FuncDefQueryModel{
					Fake:          nil,
					Category:      "Transform",
					DefaultParams: testhelper.InitRawMsg(`[100]`),
					Name:          "offset",
					Params: []FuncDefParamQueryModel{
						{
							Name:    "delta",
							Options: nil,
							Type:    "float",
						},
					},
				},
				Params: []string{"100"},
			},
			targetArg: "delta",
			output:    testhelper.InitString("100"),
		},
		{
			input: FunctionDescriptorQueryModel{
				Def: FuncDefQueryModel{
					Fake:          nil,
					Category:      "Transform",
					DefaultParams: testhelper.InitRawMsg(`[]`),
					Name:          "delta",
					Params:        []FuncDefParamQueryModel{},
				},
				Params: []string{},
			},
			targetArg: "delta",
			output:    nil,
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
	var tests = []struct {
		input     FunctionDescriptorQueryModel
		targetArg string
		output    *string
	}{
		{
			input: FunctionDescriptorQueryModel{
				Def: FuncDefQueryModel{
					Fake:          nil,
					Category:      "Options",
					DefaultParams: testhelper.InitRawMsg(`[900]`),
					Name:          "binInterval",
					Params: []FuncDefParamQueryModel{
						{
							Name:    "interval",
							Options: nil,
							Type:    "int",
						},
					},
				},
				Params: []string{"900"},
			},
			targetArg: "interval",
			output:    testhelper.InitString("int"),
		},
		{
			input: FunctionDescriptorQueryModel{
				Def: FuncDefQueryModel{
					Fake:          nil,
					Category:      "Filter Series",
					DefaultParams: testhelper.InitRawMsg(`[5 avg]`),
					Name:          "bottom",
					Params: []FuncDefParamQueryModel{
						{
							Name:    "number",
							Options: nil,
							Type:    "int",
						},
						{
							Name:    "value",
							Options: &[]string{"avg", "min", "max", "absoluteMin", "absoluteMax", "sum"},
							Type:    "string",
						},
					},
				},
				Params: []string{"5", "avg"},
			},
			targetArg: "value",
			output:    testhelper.InitString("string"),
		},
		{
			input: FunctionDescriptorQueryModel{
				Def: FuncDefQueryModel{
					Fake:          nil,
					Category:      "Transform",
					DefaultParams: testhelper.InitRawMsg(`[100]`),
					Name:          "offset",
					Params: []FuncDefParamQueryModel{
						{
							Name:    "delta",
							Options: nil,
							Type:    "float",
						},
					},
				},
				Params: []string{"100"},
			},
			targetArg: "delta",
			output:    testhelper.InitString("float"),
		},
		{
			input: FunctionDescriptorQueryModel{
				Def: FuncDefQueryModel{
					Fake:          nil,
					Category:      "Transform",
					DefaultParams: testhelper.InitRawMsg(`[]`),
					Name:          "delta",
					Params:        []FuncDefParamQueryModel{},
				},
				Params: []string{},
			},
			targetArg: "delta",
			output:    nil,
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
	var tests = []struct {
		input     FunctionDescriptorQueryModel
		targetArg string
		output    int
	}{
		{
			input: FunctionDescriptorQueryModel{
				Def: FuncDefQueryModel{
					Fake:          nil,
					Category:      "Filter Series",
					DefaultParams: testhelper.InitRawMsg(`[5 avg]`),
					Name:          "bottom",
					Params: []FuncDefParamQueryModel{
						{
							Name:    "number",
							Options: nil,
							Type:    "int",
						},
						{
							Name:    "value",
							Options: &[]string{"avg", "min", "max", "absoluteMin", "absoluteMax", "sum"},
							Type:    "string",
						},
					},
				},
				Params: []string{"5", "avg"},
			},
			targetArg: "number",
			output:    5,
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
	var tests = []struct {
		input     FunctionDescriptorQueryModel
		targetArg string
		output    float64
	}{
		{
			input: FunctionDescriptorQueryModel{
				Def: FuncDefQueryModel{
					Fake:          nil,
					Category:      "Transform",
					DefaultParams: testhelper.InitRawMsg(`[100]`),
					Name:          "offset",
					Params: []FuncDefParamQueryModel{
						{
							Name:    "delta",
							Options: nil,
							Type:    "float",
						},
					},
				},
				Params: []string{"100"},
			},
			targetArg: "delta",
			output:    100,
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
	var tests = []struct {
		input     FunctionDescriptorQueryModel
		targetArg string
		output    string
	}{
		{
			input: FunctionDescriptorQueryModel{
				Def: FuncDefQueryModel{
					Fake:          nil,
					Category:      "Filter Series",
					DefaultParams: testhelper.InitRawMsg(`[5 avg]`),
					Name:          "bottom",
					Params: []FuncDefParamQueryModel{
						{
							Name:    "number",
							Options: nil,
							Type:    "int",
						},
						{
							Name:    "value",
							Options: &[]string{"avg", "min", "max", "absoluteMin", "absoluteMax", "sum"},
							Type:    "string",
						},
					},
				},
				Params: []string{"5", "avg"},
			},
			targetArg: "value",
			output:    "avg",
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

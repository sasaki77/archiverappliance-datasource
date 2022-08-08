package main

import (
	"encoding/json"
	"testing"
	"time"

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
				Interval: MultiReturnHelperParseDuration(time.ParseDuration("0s")),
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
					From: MultiReturnHelperParse(time.Parse(TIME_FORMAT, "2021-01-27T14:30:41.678-08:00")),
					To:   MultiReturnHelperParse(time.Parse(TIME_FORMAT, "2021-01-28T14:30:41.678-08:00")),
				},
			},
			output: ArchiverQueryModel{
				Target:       ".*(1|2)",
				Alias:        "test alias",
				AliasPattern: "(.*):(.*)",
				Operator:     "max",
				Regex:        true,
				Functions:    []FunctionDescriptorQueryModel{},
				IntervalMs:   InitIntPointer(5000),
				RefId:        "A",
				TimeRange: backend.TimeRange{
					From: MultiReturnHelperParse(time.Parse(TIME_FORMAT, "2021-01-27T14:30:41.678-08:00")),
					To:   MultiReturnHelperParse(time.Parse(TIME_FORMAT, "2021-01-28T14:30:41.678-08:00")),
				},
				Interval:        5,
				BackendQuery:    false,
				DisableExtrapol: false,
			},
		},
		{
			name: "Backend query test",
			input: backend.DataQuery{
				Interval: MultiReturnHelperParseDuration(time.ParseDuration("0s")),
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
								}
							]
						}`),
				MaxDataPoints: 1000,
				QueryType:     "",
				RefID:         "A",
				TimeRange: backend.TimeRange{
					From: MultiReturnHelperParse(time.Parse(TIME_FORMAT, "2021-01-27T14:30:41.678-08:00")),
					To:   MultiReturnHelperParse(time.Parse(TIME_FORMAT, "2021-01-28T14:30:41.678-08:00")),
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
				},
				IntervalMs: nil,
				RefId:      "A",
				TimeRange: backend.TimeRange{
					From: MultiReturnHelperParse(time.Parse(TIME_FORMAT, "2021-01-27T14:30:41.678-08:00")),
					To:   MultiReturnHelperParse(time.Parse(TIME_FORMAT, "2021-01-28T14:30:41.678-08:00")),
				},
				Interval:        0,
				BackendQuery:    true,
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

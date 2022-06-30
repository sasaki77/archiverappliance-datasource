package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"testing"
	"time"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
)

func TestBuildQueryUrl(t *testing.T) {
	//            "2021-01-27T14:25:41.678-08:00"
	TIME_FORMAT := "2006-01-02T15:04:05.000-07:00"
	var tests = []struct {
		name      string
		target    string
		query     backend.DataQuery
		pluginctx backend.PluginContext
		qm        ArchiverQueryModel
		output    string
	}{
		{
			name:   "URL for auto raw data with empty operator (interval is less than 1 second)",
			target: "MR1K1:BEND:PIP:1:PMON",
			query: backend.DataQuery{
				Interval: MultiReturnHelperParseDuration(time.ParseDuration("0s")),
				JSON: json.RawMessage(`{
                    "alias": null,
                    "aliasPattern": null,
                    "constant":6.5, 
                    "functions":[], 
                    "hide":false ,
                    "operator": null,
                    "refId":"A" ,
                    "regex":true ,
                    "target":"MR1K1:BEND:PIP:1:PMON"}`),
				MaxDataPoints: 1000,
				QueryType:     "",
				RefID:         "A",
				TimeRange: backend.TimeRange{
					From: MultiReturnHelperParse(time.Parse(TIME_FORMAT, "2021-01-27T14:25:41.678-08:00")),
					To:   MultiReturnHelperParse(time.Parse(TIME_FORMAT, "2021-01-27T14:30:41.678-08:00")),
				},
			},
			pluginctx: backend.PluginContext{
				DataSourceInstanceSettings: &backend.DataSourceInstanceSettings{URL: "http://localhost:3396/retrieval"},
			},
			qm: ArchiverQueryModel{
				IntervalMs: InitIntPointer(300),
				// alias: ,
				// aliasPattern: ,
				// constant: 6.5,
				// dataTopic: nil,
				// datasource: nil,
				// format: ,
				Functions: []FunctionDescriptorQueryModel{},
				// Hide: false,
				Operator:  "",
				QueryText: "",
				QueryType: nil,
				RefId:     "A",
				Regex:     true,
				// String: nil,
				Target: "MR1K1:BEND:PIP:1:PMON",
			},
			output: "http://localhost:3396/retrieval/data/getData.qw?donotchunk=&from=2021-01-27T14%3A25%3A41.678-08%3A00&pv=MR1K1%3ABEND%3APIP%3A1%3APMON&to=2021-01-27T14%3A30%3A41.678-08%3A00",
		},
		{
			name:   "URL for empty operator (interval is higher than 1 second)",
			target: "MR1K1:BEND:PIP:1:PMON",
			query: backend.DataQuery{
				Interval: MultiReturnHelperParseDuration(time.ParseDuration("0s")),
				JSON: json.RawMessage(`{
                    "alias": null,
                    "aliasPattern": null,
                    "constant":6.5, 
                    "functions":[], 
                    "hide":false ,
                    "operator": null,
                    "refId":"A" ,
                    "regex":true ,
                    "target":"MR1K1:BEND:PIP:1:PMON"}`),
				MaxDataPoints: 1000,
				QueryType:     "",
				RefID:         "A",
				TimeRange: backend.TimeRange{
					From: MultiReturnHelperParse(time.Parse(TIME_FORMAT, "2021-01-27T14:25:41.678-08:00")),
					To:   MultiReturnHelperParse(time.Parse(TIME_FORMAT, "2021-01-27T16:25:41.678-08:00")),
				},
			},
			pluginctx: backend.PluginContext{
				DataSourceInstanceSettings: &backend.DataSourceInstanceSettings{URL: "http://localhost:3396/retrieval"},
			},
			qm: ArchiverQueryModel{
				IntervalMs: InitIntPointer(7200),
				Functions:  []FunctionDescriptorQueryModel{},
				Operator:   "",
				QueryText:  "",
				QueryType:  nil,
				RefId:      "A",
				Regex:      true,
				Target:     "MR1K1:BEND:PIP:1:PMON",
			},
			output: "http://localhost:3396/retrieval/data/getData.qw?donotchunk=&from=2021-01-27T14%3A25%3A41.678-08%3A00&pv=mean_7%28MR1K1%3ABEND%3APIP%3A1%3APMON%29&to=2021-01-27T16%3A25%3A41.678-08%3A00",
		},
		{
			name:   "URL for max operator (interval is less than 1 second)",
			target: "MR1K1:BEND:PIP:1:PMON",
			query: backend.DataQuery{
				Interval: MultiReturnHelperParseDuration(time.ParseDuration("0s")),
				JSON: json.RawMessage(`{
                    "alias": null,
                    "aliasPattern": null,
                    "constant":6.5, 
                    "functions":[], 
                    "hide":false ,
                    "operator": null,
                    "refId":"A" ,
                    "regex":true ,
                    "target":"MR1K1:BEND:PIP:1:PMON"}`),
				MaxDataPoints: 1000,
				QueryType:     "",
				RefID:         "A",
				TimeRange: backend.TimeRange{
					From: MultiReturnHelperParse(time.Parse(TIME_FORMAT, "2021-01-27T14:25:41.678-08:00")),
					To:   MultiReturnHelperParse(time.Parse(TIME_FORMAT, "2021-01-27T14:30:41.678-08:00")),
				},
			},
			pluginctx: backend.PluginContext{
				DataSourceInstanceSettings: &backend.DataSourceInstanceSettings{URL: "http://localhost:3396/retrieval"},
			},
			qm: ArchiverQueryModel{
				IntervalMs: InitIntPointer(300),
				Functions:  []FunctionDescriptorQueryModel{},
				Operator:   "max",
				QueryText:  "",
				QueryType:  nil,
				RefId:      "A",
				Regex:      true,
				Target:     "MR1K1:BEND:PIP:1:PMON",
			},
			output: "http://localhost:3396/retrieval/data/getData.qw?donotchunk=&from=2021-01-27T14%3A25%3A41.678-08%3A00&pv=MR1K1%3ABEND%3APIP%3A1%3APMON&to=2021-01-27T14%3A30%3A41.678-08%3A00",
		},
		{
			name:   "URL for max operator (interval is higher than 1 second)",
			target: "MR1K1:BEND:PIP:1:PMON",
			query: backend.DataQuery{
				Interval: MultiReturnHelperParseDuration(time.ParseDuration("0s")),
				JSON: json.RawMessage(`{
                    "alias": null,
                    "aliasPattern": null,
                    "constant":6.5, 
                    "functions":[], 
                    "hide":false ,
                    "operator": null,
                    "refId":"A" ,
                    "regex":true ,
                    "target":"MR1K1:BEND:PIP:1:PMON"}`),
				MaxDataPoints: 1000,
				QueryType:     "",
				RefID:         "A",
				TimeRange: backend.TimeRange{
					From: MultiReturnHelperParse(time.Parse(TIME_FORMAT, "2021-01-27T14:25:41.678-08:00")),
					To:   MultiReturnHelperParse(time.Parse(TIME_FORMAT, "2021-01-27T16:25:41.678-08:00")),
				},
			},
			pluginctx: backend.PluginContext{
				DataSourceInstanceSettings: &backend.DataSourceInstanceSettings{URL: "http://localhost:3396/retrieval"},
			},
			qm: ArchiverQueryModel{
				IntervalMs: InitIntPointer(7200),
				Functions:  []FunctionDescriptorQueryModel{},
				Operator:   "max",
				QueryText:  "",
				QueryType:  nil,
				RefId:      "A",
				Regex:      true,
				Target:     "MR1K1:BEND:PIP:1:PMON",
			},
			output: "http://localhost:3396/retrieval/data/getData.qw?donotchunk=&from=2021-01-27T14%3A25%3A41.678-08%3A00&pv=max_7%28MR1K1%3ABEND%3APIP%3A1%3APMON%29&to=2021-01-27T16%3A25%3A41.678-08%3A00",
		},
		{
			name:   "URL for median operator and fixed interval",
			target: "MR1K1:BEND:PIP:1:PMON",
			query: backend.DataQuery{
				Interval: MultiReturnHelperParseDuration(time.ParseDuration("0s")),
				JSON: json.RawMessage(`{
                    "alias": null,
                    "aliasPattern": null,
                    "constant":6.5, 
                    "functions":[
                        {
                            "def":{
                                "category":"Options",
                                "defaultParams":[900], 
                                "name":"binInterval",
                                "params":[
                                    {
                                        "name":"interval",
                                        "type":"int"
                                    }
                                ]
                            }, 
                            "params":[900]
                            }, 
                        {
                            "def":{
                                "category":"Transform",
                                "defaultParams":[],
                                "name":"delta",
                                "params":[]
                            },
                            "params":[]
                        }
                    ], 
                    "hide":false ,
                    "operator": "median",
                    "refId":"A" ,
                    "regex":true ,
                    "target":"MR1K1:BEND:PIP:1:PMON"}`),
				MaxDataPoints: 1000,
				QueryType:     "",
				RefID:         "A",
				TimeRange: backend.TimeRange{
					From: MultiReturnHelperParse(time.Parse(TIME_FORMAT, "2021-01-27T14:25:41.678-08:00")),
					To:   MultiReturnHelperParse(time.Parse(TIME_FORMAT, "2021-01-27T14:30:41.678-08:00")),
				},
			},
			pluginctx: backend.PluginContext{
				DataSourceInstanceSettings: &backend.DataSourceInstanceSettings{URL: "http://localhost:3396/retrieval"},
			},
			qm: ArchiverQueryModel{
				IntervalMs: InitIntPointer(300),
				// alias: ,
				// aliasPattern: ,
				// constant: 6.5,
				// dataTopic: nil,
				// datasource: nil,
				// format: ,
				Functions: []FunctionDescriptorQueryModel{
					{
						Def: FuncDefQueryModel{
							// Fake:<nil>
							Category:      "Options",
							DefaultParams: InitRawMsg(`[900]`),
							Name:          "binInterval",
							Params: []FuncDefParamQueryModel{
								{Name: "interval", Type: "int"},
							},
						},
						Params: []string{"900"},
					},
					{
						Def: FuncDefQueryModel{
							// Fake:<nil>
							Category:      "Transform",
							DefaultParams: InitRawMsg(`[]`),
							Name:          "delta",
							Params:        []FuncDefParamQueryModel{},
						},
						Params: []string{},
					},
				},
				// Hide: false,
				Operator:  "median",
				QueryText: "",
				QueryType: nil,
				RefId:     "A",
				Regex:     true,
				// String: nil,
				Target: "MR1K1:BEND:PIP:1:PMON",
			},
			output: "http://localhost:3396/retrieval/data/getData.qw?donotchunk=&from=2021-01-27T14%3A25%3A41.678-08%3A00&pv=median_900%28MR1K1%3ABEND%3APIP%3A1%3APMON%29&to=2021-01-27T14%3A30%3A41.678-08%3A00",
		},
	}
	// fmt.Println(tests)
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			result := BuildQueryUrl(testCase.target, testCase.query, testCase.pluginctx, testCase.qm)
			if testCase.output != result {
				t.Errorf("got %v, want %v", result, testCase.output)
			}
		})
	}
}

func TestArchiverSingleQuery(t *testing.T) {
	t.Skipf("Test not implemented")
}

func TestArchiverSingleQueryParser(t *testing.T) {
	ARCHIVER_FLOAT_PRECISION := 1e-18
	type responseParams struct {
		length   int
		name     string
		firstVal float64
		lastVal  float64
	}

	var dataNames = []struct {
		fileName string
		output   responseParams
	}{
		{
			fileName: "test_data/good_query_response_01.JSON",
			output:   responseParams{length: 612, name: "EM2K0:XGMD:GPI:10:PRESS_RBV", firstVal: 0.005249832756817341, lastVal: 0.005262143909931183},
		},
	}

	type testData struct {
		input  []byte
		output responseParams
	}

	var tests []testData
	for _, entry := range dataNames {
		fileData, err := ioutil.ReadFile(entry.fileName)
		if err != nil {
			t.Fatalf("Failed to load test data: %v", err)
		}
		tests = append(tests, testData{input: fileData, output: entry.output})
	}

	for idx, testCase := range tests {
		testName := fmt.Sprintf("Case: %d", idx)
		t.Run(testName, func(t *testing.T) {
			// result := testCase.output
			result, err := ArchiverSingleQueryParser(testCase.input)
			if err != nil {
				t.Fatalf("An unexpected error has occurred")
			}
			if len(result.Times) != len(result.Values) {
				t.Fatalf("Lengths of Times and Values differ - Times: %v Values: %v", len(result.Times), len(result.Values))
			}
			resultLength := len(result.Times)
			if resultLength != testCase.output.length {
				t.Fatalf("Lengths differ - Wanted: %v Got: %v", testCase.output.length, resultLength)
			}
			if result.Name != testCase.output.name {
				t.Fatalf("Names differ - Wanted: %v Got: %v", testCase.output.name, result.Name)
			}
			if math.Abs(result.Values[0]-testCase.output.firstVal) > ARCHIVER_FLOAT_PRECISION {
				t.Fatalf("First values differ - Wanted: %v Got: %v", testCase.output.firstVal, result.Values[0])
			}
			if math.Abs(result.Values[resultLength-1]-testCase.output.lastVal) > ARCHIVER_FLOAT_PRECISION {
				t.Fatalf("Last values differ - Wanted: %v Got: %v", testCase.output.lastVal, result.Values[resultLength-1])
			}
		})
	}
}

func TestBuildRegexUrl(t *testing.T) {
	t.Skipf("Test not implemented")
}

func TestArchiverRegexQueryParser(t *testing.T) {
	var tests = []struct {
		input  []byte
		output []string
	}{
		{input: []byte("[\"MR1K1:BEND:PIP:1:PMON\",\"MR1K3:BEND:PIP:1:PMON\"]"), output: []string{"MR1K1:BEND:PIP:1:PMON", "MR1K3:BEND:PIP:1:PMON"}},
		{input: []byte("[\"MR1K3:BEND:PIP:1:PMON\"]"), output: []string{"MR1K3:BEND:PIP:1:PMON"}},
		{input: []byte("[]"), output: []string{}},
	}

	for idx, testCase := range tests {
		testName := fmt.Sprintf("%d: %s, %s", idx, testCase.input, testCase.output)
		t.Run(testName, func(t *testing.T) {
			// result := testCase.output
			result, err := ArchiverRegexQueryParser(testCase.input)
			if err != nil {
				t.Fatalf("An unexpected error has occurred")
			}
			if len(result) != len(testCase.output) {
				t.Fatalf("Lengths differ - Wanted: %v Got: %v", testCase.output, result)
			}
			for idx := range testCase.output {
				if testCase.output[idx] != result[idx] {
					t.Errorf("got %v, want %v", result, testCase.output)
				}
			}
		})
	}
}

func TestIsolateBasicQuery(t *testing.T) {
	var tests = []struct {
		inputUnparsed string
		output        []string
	}{
		{inputUnparsed: "(this:is:1|this:is:2)", output: []string{"this:is:1", "this:is:2"}},
		{inputUnparsed: "(this:is:1)", output: []string{"this:is:1"}},
		{inputUnparsed: "this:is:1", output: []string{"this:is:1"}},
		{inputUnparsed: "(this):is:1", output: []string{"this:is:1"}},
		{inputUnparsed: "before:(this)", output: []string{"before:this"}},
		{inputUnparsed: "before:(this|that):is:1", output: []string{"before:this:is:1", "before:that:is:1"}},
		{inputUnparsed: "before:(this|that):(is|was):1", output: []string{"before:this:is:1", "before:this:was:1", "before:that:is:1", "before:that:was:1"}},
		{inputUnparsed: "()", output: []string{""}},
		{inputUnparsed: "((this|that):is:1|this:is:2)", output: []string{"this:is:2", "this:is:1", "that:is:1"}},
		{inputUnparsed: "prefix:((this|that):is:1|this:is:2)", output: []string{"prefix:this:is:2", "prefix:this:is:1", "prefix:that:is:1"}},
	}

	for idx, testCase := range tests {
		testName := fmt.Sprintf("%d: %s, %s", idx, testCase.inputUnparsed, testCase.output)
		t.Run(testName, func(t *testing.T) {
			// result := testCase.output
			result := IsolateBasicQuery(testCase.inputUnparsed)
			if len(result) != len(testCase.output) {
				t.Fatalf("Lengths differ - Wanted: %v Got: %v", testCase.output, result)
			}
			for idx := range testCase.output {
				if testCase.output[idx] != result[idx] {
					t.Errorf("got %v, want %v", result, testCase.output)
				}
			}
		})
	}
}

func TestSplitLowestLevelOnly(t *testing.T) {
	var tests = []struct {
		input  string
		output []string
	}{
		{
			input:  "one (two (three)) (four five) six",
			output: []string{"one (two (three)) (four five) six"},
		},
		{
			input:  "one two |three",
			output: []string{"one two ", "three"},
		},
		{
			input:  "one two |three | four",
			output: []string{"one two ", "three ", " four"},
		},
		{
			input:  "one (two |three) | four",
			output: []string{"one (two |three) ", " four"},
		},
		// {
		//     input:   "one (two (three)) (four five) six",
		//     output:  []string{"one (two (three)) (four five) six"},
		// },
	}
	for idx, testCase := range tests {
		testName := fmt.Sprintf("%d: %s", idx, testCase.input)
		t.Run(testName, func(t *testing.T) {
			result := SplitLowestLevelOnly(testCase.input)
			if len(result) != len(testCase.output) {
				t.Fatalf("Lengths differ - Wanted: %v Got: %v", testCase.output, result)
			}
			for idx := range testCase.output {
				if testCase.output[idx] != result[idx] {
					t.Errorf("got %v, want %v", result, testCase.output)
				}
			}
		})
	}
}

func TestLocateOuterParen(t *testing.T) {
	var tests = []struct {
		input         string
		outputPhrases []string
		outputIdxs    [][]int
	}{
		{
			input:         "one (two (three)) (four five) six",
			outputPhrases: []string{"(two (three))", "(four five)"},
			outputIdxs:    [][]int{{4, 17}, {18, 29}},
		},
		{
			input:         "one (match) (here)",
			outputPhrases: []string{"(match)", "(here)"},
			outputIdxs:    [][]int{{4, 11}, {12, 18}},
		},
		{
			input:         "one (match) here",
			outputPhrases: []string{"(match)"},
			outputIdxs:    [][]int{{4, 11}},
		},
		{
			input:         "no matches here",
			outputPhrases: []string{},
			outputIdxs:    [][]int{},
		},
		// {
		//     input:           "",
		//     outputPhrases:  []string{},
		//     outputIdxs:     [][]int{},
		// },
	}
	for idx, testCase := range tests {
		testName := fmt.Sprintf("%d: %s, ", idx, testCase.input)
		t.Run(testName, func(t *testing.T) {
			result := LocateOuterParen(testCase.input)
			// Check Phrases
			if len(result.Phrases) != len(testCase.outputPhrases) {
				t.Fatalf("Lengths differ - Wanted: %v Got: %v", testCase.outputPhrases, result.Phrases)
			}
			for idx := range testCase.outputPhrases {
				if testCase.outputPhrases[idx] != result.Phrases[idx] {
					t.Errorf("got %v, want %v", result.Phrases, testCase.outputPhrases)
				}
			}

			// Check indices
			if len(result.Idxs) != len(testCase.outputIdxs) {
				t.Fatalf("Lengths differ - Wanted: %v Got: %v", testCase.outputIdxs, result.Idxs)
			}
			for idx := range testCase.outputPhrases {
				for i := 0; i < 2; i++ {
					if testCase.outputIdxs[idx][i] != result.Idxs[idx][i] {
						t.Errorf("got %v, want %v", result.Idxs[idx], testCase.outputIdxs[idx])
					}
				}
			}
		})
	}
}

func TestPermuteQuery(t *testing.T) {
	var tests = []struct {
		input  [][]string
		output [][]string
	}{
		{input: [][]string{{"a", "d"}, {"b", "c"}}, output: [][]string{{"a", "b"}, {"a", "c"}, {"d", "b"}, {"d", "c"}}},
		{input: [][]string{{"a"}, {"b", "c"}}, output: [][]string{{"a", "b"}, {"a", "c"}}},
		{input: [][]string{{"a"}, {"b"}}, output: [][]string{{"a", "b"}}},
		{input: [][]string{{"a"}}, output: [][]string{{"a"}}},
		{input: [][]string{{"a", "b"}}, output: [][]string{{"a"}, {"b"}}},
		{input: [][]string{{}}, output: [][]string{}},
		{input: [][]string{}, output: [][]string{{}}},
	}
	for idx, testCase := range tests {
		testName := fmt.Sprintf("%d: %s, %s", idx, testCase.input, testCase.output)
		t.Run(testName, func(t *testing.T) {
			result := PermuteQuery(testCase.input)
			if len(result) != len(testCase.output) {
				t.Fatalf("Lengths differ (0th index) - Wanted: %v (%v) Got: %v (%v)", testCase.output, len(testCase.output), result, len(result))
			}
			for idx := range testCase.output {
				if len(result[idx]) != len(testCase.output[idx]) {
					t.Fatalf("Lengths differ (1st index) - Wanted: %v (%v) Got: %v (%v)", testCase.output[idx], len(testCase.output[idx]), result[idx], len(result[idx]))
				}
				for jdx := range testCase.output[idx] {
					if testCase.output[idx][jdx] != result[idx][jdx] {
						t.Errorf("got %v, want %v at [%v][%v]", result, testCase.output, idx, jdx)
					}
				}
			}
		})
	}
}

func TestSelectiveInsert(t *testing.T) {
	var tests = []struct {
		input   string
		idxs    [][]int
		inserts []string
		output  string
	}{
		{
			input:   "hello there",
			idxs:    [][]int{{0, 5}},
			inserts: []string{"be"},
			output:  "be there",
		},
		{
			input:   "hello there",
			idxs:    [][]int{{6, 11}},
			inserts: []string{"friend"},
			output:  "hello friend",
		},
		{
			input:   "hello there",
			idxs:    [][]int{{0, 4}, {6, 11}},
			inserts: []string{"y", "what's up"},
			output:  "yo what's up",
		},
		{
			input:   "This won't work",
			idxs:    [][]int{{0, 4}, {6, 11}},
			inserts: []string{"y"},
			output:  "",
		},
		// {
		//     input: ,
		//     idxs: ,
		//     inserts: ,
		//     outpu:,
		// },
	}
	for idx, testCase := range tests {
		testName := fmt.Sprintf("%d: %s, %s", idx, testCase.input, testCase.output)
		t.Run(testName, func(t *testing.T) {
			result := SelectiveInsert(testCase.input, testCase.idxs, testCase.inserts)
			if testCase.output != result {
				t.Errorf("Incorrect output - Wanted: %v Got: %v", testCase.output, result)
			}
		})
	}
}

func TestApplyAlias(t *testing.T) {
	var tests = []struct {
		name    string
		inputSd []SingleData
		qm      ArchiverQueryModel
		output  []SingleData
	}{
		{
			name: "normal alias",
			inputSd: []SingleData{
				{
					Name:   "PV:NAME",
					PVname: "PV:NAME",
				},
			},
			qm: ArchiverQueryModel{
				Alias: "alias",
			},
			output: []SingleData{
				{
					Name:   "alias",
					PVname: "PV:NAME",
				},
			},
		},
		{
			name: "empty alias",
			inputSd: []SingleData{
				{
					Name:   "PV:NAME",
					PVname: "PV:NAME",
				},
			},
			qm: ArchiverQueryModel{
				Alias: "",
			},
			output: []SingleData{
				{
					Name:   "PV:NAME",
					PVname: "PV:NAME",
				},
			},
		},
		{
			name: "alias pattern",
			inputSd: []SingleData{
				{
					Name:   "PV:NAME",
					PVname: "PV:NAME",
				},
			},
			qm: ArchiverQueryModel{
				Alias:        "$2:$1",
				AliasPattern: "(.*):(.*)",
			},
			output: []SingleData{
				{
					Name:   "NAME:PV",
					PVname: "PV:NAME",
				},
			},
		},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			result, _ := ApplyAlias(testCase.inputSd, testCase.qm)
			SingleDataCompareHelper(result, testCase.output, t)
		})
	}
}

func TestFrameBuilder(t *testing.T) {
	var tests = []struct {
		sD   SingleData
		name string
	}{
		{
			sD: SingleData{
				Name: "testing_name",
			},
			name: "testing_name",
		},
	}
	for idx, testCase := range tests {
		testName := fmt.Sprintf("%d: %s", idx, testCase.name)
		t.Run(testName, func(t *testing.T) {
			result := FrameBuilder(testCase.sD)
			if testCase.name != result.Name {
				t.Errorf("got %v, want %v", result.Name, testCase.name)
			}
			if testCase.name != result.Fields[1].Name {
				t.Errorf("got %v, want %v", result.Fields[1].Name, testCase.name)
			}
		})
	}
}

func TestDataExtrapol(t *testing.T) {
	var tests = []struct {
		sDIn  SingleData
		query backend.DataQuery
		qm    ArchiverQueryModel
		sDOut SingleData
	}{
		{
			sDIn: SingleData{
				Times:  []time.Time{TimeHelper(0)},
				Values: []float64{1},
			},
			query: backend.DataQuery{
				TimeRange: backend.TimeRange{
					From: TimeHelper(1),
					To:   TimeHelper(5),
				},
			},
			qm: ArchiverQueryModel{
				Operator: "raw",
			},
			sDOut: SingleData{
				Times:  []time.Time{TimeHelper(0), TimeHelper(5)},
				Values: []float64{1, 1},
			},
		},
		{
			sDIn: SingleData{
				Times:  []time.Time{TimeHelper(0)},
				Values: []float64{1},
			},
			query: backend.DataQuery{
				TimeRange: backend.TimeRange{
					From: TimeHelper(1),
					To:   TimeHelper(5),
				},
			},
			qm: ArchiverQueryModel{},
			sDOut: SingleData{
				Times:  []time.Time{TimeHelper(0)},
				Values: []float64{1},
			},
		},
		{
			sDIn: SingleData{
				Times:  []time.Time{TimeHelper(0)},
				Values: []float64{1},
			},
			query: backend.DataQuery{
				TimeRange: backend.TimeRange{
					From: TimeHelper(1),
					To:   TimeHelper(5),
				},
			},
			qm: ArchiverQueryModel{
				Functions: []FunctionDescriptorQueryModel{
					{
						Def: FuncDefQueryModel{
							Category:      "Options",
							DefaultParams: InitRawMsg(`true`),
							Name:          "disableExtrapol",
							Params: []FuncDefParamQueryModel{
								{Name: "boolean", Type: "string"},
							},
						},
						Params: []string{"false"},
					},
				},
				Operator: "raw",
			},
			sDOut: SingleData{
				Times:  []time.Time{TimeHelper(0), TimeHelper(5)},
				Values: []float64{1, 1},
			},
		},
		{
			sDIn: SingleData{
				Times:  []time.Time{TimeHelper(0)},
				Values: []float64{1},
			},
			query: backend.DataQuery{
				TimeRange: backend.TimeRange{
					From: TimeHelper(1),
					To:   TimeHelper(5),
				},
			},
			qm: ArchiverQueryModel{
				Functions: []FunctionDescriptorQueryModel{
					{
						Def: FuncDefQueryModel{
							Category:      "Options",
							DefaultParams: InitRawMsg(`true`),
							Name:          "disableExtrapol",
							Params: []FuncDefParamQueryModel{
								{Name: "boolean", Type: "string"},
							},
						},
						Params: []string{"true"},
					},
				},
				Operator: "raw",
			},
			sDOut: SingleData{
				Times:  []time.Time{TimeHelper(0)},
				Values: []float64{1},
			},
		},
		{
			sDIn: SingleData{
				Times:  []time.Time{TimeHelper(0), TimeHelper(3)},
				Values: []float64{1, 2},
			},
			query: backend.DataQuery{
				TimeRange: backend.TimeRange{
					From: TimeHelper(1),
					To:   TimeHelper(5),
				},
			},
			sDOut: SingleData{
				Times:  []time.Time{TimeHelper(0), TimeHelper(3)},
				Values: []float64{1, 2},
			},
		},
	}
	for idx, testCase := range tests {
		testName := fmt.Sprintf("%d:", idx)
		t.Run(testName, func(t *testing.T) {
			result := DataExtrapol(testCase.sDIn, testCase.qm, testCase.query)
			SingleDataCompareHelper(
				[]SingleData{result},
				[]SingleData{testCase.sDOut},
				t,
			)
		})
	}
}

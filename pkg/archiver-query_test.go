package main

import (
    "encoding/json"
    "fmt"
    "math"
    "time"
    "io/ioutil"
    "testing"
    "github.com/grafana/grafana-plugin-sdk-go/backend"
)

func TestBuildQueryUrl(t *testing.T) {
    //            "2021-01-27T14:25:41.678-08:00"
    TIME_FORMAT := "2006-01-02T15:04:05.000-07:00"
    var tests = []struct {
        target string
        query backend.DataQuery
        pluginctx backend.PluginContext
        qm ArchiverQueryModel
        output string
    }{
        {
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
                    "target":"MR1K[1,3]:BEND:PIP:1:PMON"}`),
                MaxDataPoints:0,
                QueryType: "",
                RefID:"A",
                TimeRange: backend.TimeRange{
                    From: MultiReturnHelperParse(time.Parse(TIME_FORMAT, "2021-01-27T14:25:41.678-08:00")), 
                    To: MultiReturnHelperParse(time.Parse(TIME_FORMAT, "2021-01-27T14:30:41.678-08:00")),
                },
            },
            pluginctx: backend.PluginContext{
                 DataSourceInstanceSettings: &backend.DataSourceInstanceSettings{URL: "http://localhost:3396/retrieval",},
            },
            qm: ArchiverQueryModel{
                // alias: ,
                // aliasPattern: ,
                // constant: 6.5,
                // dataTopic: nil,
                // datasource: nil,
                // format: ,
                Functions: []FunctionDescriptorQueryModel{},
                // Hide: false,
                Operator: "",
                QueryText: "",
                QueryType: nil,
                RefId: "A",
                Regex: true,
                // String: nil, 
                Target: "MR1K[1,3]:BEND:PIP:1:PMON",
            },
            output: "http://localhost:3396/retrieval/data/getData.qw?donotchunk=&from=2021-01-27T14%3A25%3A41.678-08%3A00&pv=MR1K1%3ABEND%3APIP%3A1%3APMON&to=2021-01-27T14%3A30%3A41.678-08%3A00",
        },
        {
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
                    "target":"MR1K[1,3]:BEND:PIP:1:PMON"}`),
                MaxDataPoints:0,
                QueryType: "",
                RefID:"A",
                TimeRange: backend.TimeRange{
                    From: MultiReturnHelperParse(time.Parse(TIME_FORMAT, "2021-01-27T14:25:41.678-08:00")),
                    To: MultiReturnHelperParse(time.Parse(TIME_FORMAT, "2021-01-27T14:30:41.678-08:00")),
                },
            },
            pluginctx: backend.PluginContext{
                 DataSourceInstanceSettings: &backend.DataSourceInstanceSettings{URL: "http://localhost:3396/retrieval",},
            },
            qm: ArchiverQueryModel{
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
                                Category: "Options",
                                DefaultParams: InitRawMsg(`[900]`),
                                Name: "binInterval",
                                Params:[]FuncDefParamQueryModel{
                                    {Name:"interval", Type: "int"},
                                },
                            },
                            Params: []string{"900",},
                        },
                        {
                            Def: FuncDefQueryModel{
                                // Fake:<nil>
                                Category: "Transform",
                                DefaultParams: InitRawMsg(`[]`),
                                Name: "delta",
                                Params:[]FuncDefParamQueryModel{},
                            },
                            Params: []string{},
                        },
                },
                // Hide: false,
                Operator: "median",
                QueryText: "",
                QueryType: nil,
                RefId: "A",
                Regex: true,
                // String: nil, 
                Target: "MR1K[1,3]:BEND:PIP:1:PMON",
            },
            output: "http://localhost:3396/retrieval/data/getData.qw?donotchunk=&from=2021-01-27T14%3A25%3A41.678-08%3A00&pv=median_900%28MR1K1%3ABEND%3APIP%3A1%3APMON%29&to=2021-01-27T14%3A30%3A41.678-08%3A00",
        },
    }
    // fmt.Println(tests)
    for idx, testCase := range tests {
        testName := fmt.Sprintf("%d: %v", idx, testCase.output)
        t.Run(testName, func(t *testing.T) {
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
    type responseParams struct{
        length int
        name string
        firstVal float64
        lastVal float64
    }

    var dataNames = []struct{
        fileName string
        output responseParams
    }{
        {
            fileName: "test_data/good_query_response_01.JSON",
            output: responseParams{length: 612, name: "EM2K0:XGMD:GPI:10:PRESS_RBV", firstVal: 0.005249832756817341, lastVal: 0.005262143909931183,},
        },
    }

    type testData struct {
        input []byte
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
            if len(result.Times) != len(result.Values){
                t.Fatalf("Lengths of Times and Values differ - Times: %v Values: %v", len(result.Times), len(result.Values))
            }
            resultLength := len(result.Times)
            if resultLength != testCase.output.length {
                t.Fatalf("Lengths differ - Wanted: %v Got: %v", testCase.output.length, resultLength)
            }
            if result.Name != testCase.output.name {
                t.Fatalf("Names differ - Wanted: %v Got: %v", testCase.output.name, result.Name)
            }
            if math.Abs(result.Values[0] - testCase.output.firstVal) > ARCHIVER_FLOAT_PRECISION {
                t.Fatalf("First values differ - Wanted: %v Got: %v", testCase.output.firstVal, result.Values[0])
            }
            if math.Abs(result.Values[resultLength-1] - testCase.output.lastVal) > ARCHIVER_FLOAT_PRECISION {
                t.Fatalf("Last values differ - Wanted: %v Got: %v", testCase.output.lastVal, result.Values[resultLength-1])
            }
        })
    }

}

func TestBuildRegexUrl(t *testing.T) {
    t.Skipf("Test not implemented")
}

func TestArchiverRegexQueryParser(t *testing.T) {
    var tests = []struct{
        input []byte
        output []string
    }{
        {input: []byte("[\"MR1K1:BEND:PIP:1:PMON\",\"MR1K3:BEND:PIP:1:PMON\"]"), output: []string{"MR1K1:BEND:PIP:1:PMON","MR1K3:BEND:PIP:1:PMON"}},
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
            for idx, _ := range(testCase.output) {
                if testCase.output[idx] != result[idx] {
                    t.Errorf("got %v, want %v", result, testCase.output)
                }
            }
        })
    }
}

func TestIsolateBasicQuery(t *testing.T) {
    var tests = []struct{
        inputUnparsed string
        output []string
    }{
        {inputUnparsed: "(this:is:1|this:is:2)", output: []string{"this:is:1", "this:is:2"}},
        {inputUnparsed: "(this:is:1)", output: []string{"this:is:1"}},
        {inputUnparsed: "this:is:1", output: []string{"this:is:1"}},
    }

    for idx, testCase := range tests {
        testName := fmt.Sprintf("%d: %s, %s", idx, testCase.inputUnparsed, testCase.output)
        t.Run(testName, func(t *testing.T) {
            // result := testCase.output
            result := IsolateBasicQuery(testCase.inputUnparsed)
            if len(result) != len(testCase.output) {
                t.Fatalf("Lengths differ - Wanted: %v Got: %v", testCase.output, result)
            }
            for idx, _ := range(testCase.output) {
                if testCase.output[idx] != result[idx] {
                    t.Errorf("got %v, want %v", result, testCase.output)
                }
            }
        })
    }
}

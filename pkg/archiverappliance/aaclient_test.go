package archiverappliance

import (
	"fmt"
	"io/ioutil"
	"math"
	"testing"
	"time"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/sasaki77/archiverappliance-datasource/pkg/models"
	"github.com/sasaki77/archiverappliance-datasource/pkg/testhelper"
)

func TestBuildQueryUrl(t *testing.T) {
	//            "2021-01-27T14:25:41.678-08:00"
	TIME_FORMAT := "2006-01-02T15:04:05.000-07:00"
	var tests = []struct {
		name   string
		target string
		url    string
		qm     models.ArchiverQueryModel
		output string
	}{
		{
			name:   "URL for auto raw data with empty operator (interval is less than 1 second)",
			target: "MR1K1:BEND:PIP:1:PMON",
			url:    "http://localhost:3396/retrieval",
			qm: models.ArchiverQueryModel{
				IntervalMs: testhelper.InitIntPointer(300),
				// alias: ,
				// aliasPattern: ,
				// constant: 6.5,
				// dataTopic: nil,
				// datasource: nil,
				// format: ,
				Functions: []models.FunctionDescriptorQueryModel{},
				// Hide: false,
				Operator:  "",
				QueryText: "",
				QueryType: nil,
				RefId:     "A",
				Regex:     true,
				// String: nil,
				Target: "MR1K1:BEND:PIP:1:PMON",
				TimeRange: backend.TimeRange{
					From: testhelper.MultiReturnHelperParse(time.Parse(TIME_FORMAT, "2021-01-27T14:25:41.678-08:00")),
					To:   testhelper.MultiReturnHelperParse(time.Parse(TIME_FORMAT, "2021-01-27T14:30:41.678-08:00")),
				},
				Interval: 0,
			},
			output: "http://localhost:3396/retrieval/data/getData.qw?donotchunk=&from=2021-01-27T14%3A25%3A41.678-08%3A00&pv=MR1K1%3ABEND%3APIP%3A1%3APMON&to=2021-01-27T14%3A30%3A41.678-08%3A00",
		},
		{
			name:   "URL for empty operator (interval is higher than 1 second)",
			target: "MR1K1:BEND:PIP:1:PMON",
			url:    "http://localhost:3396/retrieval",
			qm: models.ArchiverQueryModel{
				IntervalMs: testhelper.InitIntPointer(7200),
				Functions:  []models.FunctionDescriptorQueryModel{},
				Operator:   "",
				QueryText:  "",
				QueryType:  nil,
				RefId:      "A",
				Regex:      true,
				Target:     "MR1K1:BEND:PIP:1:PMON",
				TimeRange: backend.TimeRange{
					From: testhelper.MultiReturnHelperParse(time.Parse(TIME_FORMAT, "2021-01-27T14:25:41.678-08:00")),
					To:   testhelper.MultiReturnHelperParse(time.Parse(TIME_FORMAT, "2021-01-27T16:25:41.678-08:00")),
				},
				Interval: 7,
			},
			output: "http://localhost:3396/retrieval/data/getData.qw?donotchunk=&from=2021-01-27T14%3A25%3A41.678-08%3A00&pv=mean_7%28MR1K1%3ABEND%3APIP%3A1%3APMON%29&to=2021-01-27T16%3A25%3A41.678-08%3A00",
		},
		{
			name:   "URL for max operator (interval is less than 1 second)",
			target: "MR1K1:BEND:PIP:1:PMON",
			url:    "http://localhost:3396/retrieval",
			qm: models.ArchiverQueryModel{
				IntervalMs: testhelper.InitIntPointer(300),
				Functions:  []models.FunctionDescriptorQueryModel{},
				Operator:   "max",
				QueryText:  "",
				QueryType:  nil,
				RefId:      "A",
				Regex:      true,
				Target:     "MR1K1:BEND:PIP:1:PMON",
				TimeRange: backend.TimeRange{
					From: testhelper.MultiReturnHelperParse(time.Parse(TIME_FORMAT, "2021-01-27T14:25:41.678-08:00")),
					To:   testhelper.MultiReturnHelperParse(time.Parse(TIME_FORMAT, "2021-01-27T14:30:41.678-08:00")),
				},
				Interval: 0,
			},
			output: "http://localhost:3396/retrieval/data/getData.qw?donotchunk=&from=2021-01-27T14%3A25%3A41.678-08%3A00&pv=MR1K1%3ABEND%3APIP%3A1%3APMON&to=2021-01-27T14%3A30%3A41.678-08%3A00",
		},
		{
			name:   "URL for max operator (interval is higher than 1 second)",
			target: "MR1K1:BEND:PIP:1:PMON",
			url:    "http://localhost:3396/retrieval",
			qm: models.ArchiverQueryModel{
				IntervalMs: testhelper.InitIntPointer(7200),
				Functions:  []models.FunctionDescriptorQueryModel{},
				Operator:   "max",
				QueryText:  "",
				QueryType:  nil,
				RefId:      "A",
				Regex:      true,
				Target:     "MR1K1:BEND:PIP:1:PMON",
				TimeRange: backend.TimeRange{
					From: testhelper.MultiReturnHelperParse(time.Parse(TIME_FORMAT, "2021-01-27T14:25:41.678-08:00")),
					To:   testhelper.MultiReturnHelperParse(time.Parse(TIME_FORMAT, "2021-01-27T16:25:41.678-08:00")),
				},
				Interval: 7,
			},
			output: "http://localhost:3396/retrieval/data/getData.qw?donotchunk=&from=2021-01-27T14%3A25%3A41.678-08%3A00&pv=max_7%28MR1K1%3ABEND%3APIP%3A1%3APMON%29&to=2021-01-27T16%3A25%3A41.678-08%3A00",
		},
		{
			name:   "URL for last operator (interval is less than 1 second)",
			target: "MR1K1:BEND:PIP:1:PMON",
			url:    "http://localhost:3396/retrieval",
			qm: models.ArchiverQueryModel{
				IntervalMs: testhelper.InitIntPointer(100),
				Functions:  []models.FunctionDescriptorQueryModel{},
				Operator:   "last",
				QueryText:  "",
				QueryType:  nil,
				RefId:      "A",
				Regex:      true,
				Target:     "MR1K1:BEND:PIP:1:PMON",
				TimeRange: backend.TimeRange{
					From: testhelper.MultiReturnHelperParse(time.Parse(TIME_FORMAT, "2021-01-27T14:25:41.678-08:00")),
					To:   testhelper.MultiReturnHelperParse(time.Parse(TIME_FORMAT, "2021-01-27T16:25:41.678-08:00")),
				},
				Interval: 0,
			},
			output: "http://localhost:3396/retrieval/data/getData.qw?donotchunk=&from=2021-01-27T16%3A25%3A41.678-08%3A00&pv=MR1K1%3ABEND%3APIP%3A1%3APMON&to=2021-01-27T16%3A25%3A41.678-08%3A00",
		},
	}
	// fmt.Println(tests)
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			result := buildQueryUrl(testCase.target, testCase.url, testCase.qm)
			if testCase.output != result {
				t.Errorf("got %v, want %v", result, testCase.output)
			}
		})
	}
}

func TestBuildRegexUrl(t *testing.T) {
	base_url := string("http://localhost:3396/retrieval")
	var tests = []struct {
		name   string
		regex  string
		limit  int
		output string
	}{
		{
			name:   "Maximum number of PVs is 1000",
			regex:  ".*",
			limit:  1000,
			output: "http://localhost:3396/retrieval/bpl/getMatchingPVs?limit=1000&regex=.%2A",
		},
		{
			name:   "Maximum number of PVs is 100",
			regex:  ".*",
			limit:  100,
			output: "http://localhost:3396/retrieval/bpl/getMatchingPVs?limit=100&regex=.%2A",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			result := buildRegexUrl(testCase.regex, base_url, testCase.limit)
			if testCase.output != result {
				t.Errorf("got %v, want %v", result, testCase.output)
			}
		})
	}
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
			fileName: "../test_data/good_query_response_01.JSON",
			output:   responseParams{length: 612, name: "EM2K0:XGMD:GPI:10:PRESS_RBV", firstVal: 0.005249832756817341, lastVal: 0.005262143909931183},
		},
		{
			fileName: "../test_data/good_query_response_waveform_01.JSON",
			output:   responseParams{length: 61, name: "PLC:LFE:MOTION:FFO:02:FF:011:Info:InfoString_RBV", firstVal: 112, lastVal: 88},
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
			result, err := archiverSingleQueryParser(testCase.input)
			if err != nil {
				t.Fatalf("An unexpected error has occurred")
			}
			if result.Name != testCase.output.name {
				t.Fatalf("Names differ - Wanted: %v Got: %v", testCase.output.name, result.Name)
			}

			switch v := result.Values.(type) {
			case *models.Scalars:
				if len(v.Times) != len(v.Values) {
					t.Fatalf("Lengths of Times and Values differ - Times: %v Values: %v", len(v.Times), len(v.Values))
				}
				resultLength := len(v.Times)
				if resultLength != testCase.output.length {
					t.Fatalf("Lengths differ - Wanted: %v Got: %v", testCase.output.length, resultLength)
				}
				if math.Abs(v.Values[0]-testCase.output.firstVal) > ARCHIVER_FLOAT_PRECISION {
					t.Fatalf("First values differ - Wanted: %v Got: %v", testCase.output.firstVal, v.Values[0])
				}
				if math.Abs(v.Values[resultLength-1]-testCase.output.lastVal) > ARCHIVER_FLOAT_PRECISION {
					t.Fatalf("Last values differ - Wanted: %v Got: %v", testCase.output.lastVal, v.Values[resultLength-1])
				}
			case *models.Arrays:
				if len(v.Times) != len(v.Values) {
					t.Fatalf("Lengths of Times and Values differ - Times: %v Values: %v", len(v.Times), len(v.Values))
				}
				resultLength := len(v.Times)
				if resultLength != testCase.output.length {
					t.Fatalf("Lengths differ - Wanted: %v Got: %v", testCase.output.length, resultLength)
				}
				if math.Abs(v.Values[0][0]-testCase.output.firstVal) > ARCHIVER_FLOAT_PRECISION {
					t.Fatalf("First values differ - Wanted: %v Got: %v", testCase.output.firstVal, v.Values[0])
				}
				vlen := len(v.Values[resultLength-1])
				if math.Abs(v.Values[resultLength-1][vlen-1]-testCase.output.lastVal) > ARCHIVER_FLOAT_PRECISION {
					t.Fatalf("Last values differ - Wanted: %v Got: %v", testCase.output.lastVal, v.Values[resultLength-1])
				}
			default:
				t.Fatalf("Response Values are invalid")
			}
		})
	}
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
			result, err := archiverRegexQueryParser(testCase.input)
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

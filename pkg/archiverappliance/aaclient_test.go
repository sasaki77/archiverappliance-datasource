package archiverappliance

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
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
			name:   "URL for auto raw data with mean operator (interval is less than 1 second)",
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
				Operator:  "mean",
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
			output: "http://localhost:3396/retrieval/data/getData.raw?donotchunk=&from=2021-01-27T14%3A25%3A41.678-08%3A00&pv=MR1K1%3ABEND%3APIP%3A1%3APMON&to=2021-01-27T14%3A30%3A41.678-08%3A00",
		},
		{
			name:   "URL for mean operator (interval is higher than 1 second)",
			target: "MR1K1:BEND:PIP:1:PMON",
			url:    "http://localhost:3396/retrieval",
			qm: models.ArchiverQueryModel{
				IntervalMs: testhelper.InitIntPointer(7200),
				Functions:  []models.FunctionDescriptorQueryModel{},
				Operator:   "mean",
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
			output: "http://localhost:3396/retrieval/data/getData.raw?donotchunk=&from=2021-01-27T14%3A25%3A41.678-08%3A00&pv=mean_7%28MR1K1%3ABEND%3APIP%3A1%3APMON%29&to=2021-01-27T16%3A25%3A41.678-08%3A00",
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
			output: "http://localhost:3396/retrieval/data/getData.raw?donotchunk=&from=2021-01-27T14%3A25%3A41.678-08%3A00&pv=MR1K1%3ABEND%3APIP%3A1%3APMON&to=2021-01-27T14%3A30%3A41.678-08%3A00",
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
			output: "http://localhost:3396/retrieval/data/getData.raw?donotchunk=&from=2021-01-27T14%3A25%3A41.678-08%3A00&pv=max_7%28MR1K1%3ABEND%3APIP%3A1%3APMON%29&to=2021-01-27T16%3A25%3A41.678-08%3A00",
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
			output: "http://localhost:3396/retrieval/data/getData.raw?donotchunk=&from=2021-01-27T16%3A25%3A41.678-08%3A00&pv=MR1K1%3ABEND%3APIP%3A1%3APMON&to=2021-01-27T16%3A25%3A41.678-08%3A00",
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

func TestSingleQueryWithout200StatusCode(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		},
	))
	defer mockServer.Close()
	_, err := archiverSingleQuery(mockServer.URL)

	if err == nil {
		t.Errorf("archiverSingleQuery should return error for error status code")
	}
}

func TestLiveOnly(t *testing.T) {
	//            "2021-01-27T14:25:41.678-08:00"
	TIME_FORMAT := "2006-01-02T15:04:05.000-07:00"
	var tests = []struct {
		name   string
		target string
		qm     models.ArchiverQueryModel
	}{
		{
			target: "PV:NAME",
			qm: models.ArchiverQueryModel{
				IntervalMs: testhelper.InitIntPointer(300),
				Functions:  []models.FunctionDescriptorQueryModel{},
				// Hide: false,
				Operator:  "mean",
				QueryText: "",
				QueryType: nil,
				RefId:     "A",
				Regex:     true,
				LiveOnly:  true,
				// String: nil,
				Target: "MR1K1:BEND:PIP:1:PMON",
				TimeRange: backend.TimeRange{
					From: testhelper.MultiReturnHelperParse(time.Parse(TIME_FORMAT, "2021-01-27T14:25:41.678-08:00")),
					To:   testhelper.MultiReturnHelperParse(time.Parse(TIME_FORMAT, "2021-01-27T14:30:41.678-08:00")),
				},
				Interval: 0,
			},
		},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			ctx := context.Background()
			client, _ := NewAAClient(ctx, "url")
			result, _ := client.ExecuteSingleQuery(testCase.target, testCase.qm)
			pvname := "PV:NAME"

			if result.Name != pvname {
				t.Errorf("got %v, want %v", result.Name, pvname)
			}
			if result.PVname != pvname {
				t.Errorf("got %v, want %v", result.Name, pvname)
			}
			if len(result.Values.(*models.Scalars).Values) != 0 {
				t.Errorf("Values should be empty")
			}
			if len(result.Values.(*models.Scalars).Times) != 0 {
				t.Errorf("Times should be empty")
			}
		})
	}
}

package main

import (
	"fmt"
	"regexp"
	"testing"
	"time"
)

func TestApplyAliasSd(t *testing.T) {
	var tests = []struct {
		name    string
		inputSd SingleData
		alias   string
		pattern string
		result  string
	}{
		{
			name: "normal alias",
			inputSd: SingleData{
				Name:   "PV:NAME",
				PVname: "PV:NAME",
			},
			alias:  "alias",
			result: "alias",
		},
		{
			name: "alias pattern",
			inputSd: SingleData{
				Name:   "PV:NAME",
				PVname: "PV:NAME",
			},
			alias:   "$2:$1",
			pattern: "(.*):(.*)",
			result:  "NAME:PV",
		},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			var rep *regexp.Regexp
			if testCase.pattern != "" {
				rep, _ = regexp.Compile(testCase.pattern)
			}
			testCase.inputSd.ApplyAlias(testCase.alias, rep)
			if testCase.inputSd.Name != testCase.result {
				t.Errorf("got %v, want %v", testCase.inputSd.Name, testCase.result)
			}
		})
	}
}

func TestToFrameScalar(t *testing.T) {
	var tests = []struct {
		sD       SingleData
		name     string
		pvname   string
		values   []float64
		dataSize int
	}{
		{
			sD: SingleData{
				Name:   "testing_name",
				PVname: "pvname",
				Values: &Scalars{
					Times:  []time.Time{TimeHelper(0), TimeHelper(1), TimeHelper(2)},
					Values: []float64{1, 2, 3},
				},
			},
			name:     "testing_name",
			pvname:   "pvname",
			values:   []float64{1, 2, 3},
			dataSize: 3,
		},
	}
	for idx, testCase := range tests {
		testName := fmt.Sprintf("%d: %s", idx, testCase.name)
		t.Run(testName, func(t *testing.T) {
			result := testCase.sD.ToFrame()
			if testCase.name != result.Name {
				t.Errorf("got %v, want %v", result.Name, testCase.name)
			}
			if result.Fields[0].Name != "time" {
				t.Errorf("got %v, want time", result.Fields[0].Name)
			}
			if result.Fields[0].Len() != testCase.dataSize {
				t.Errorf("got %d, want %d", result.Fields[0].Len(), testCase.dataSize)
			}
			if testCase.name != result.Fields[1].Config.DisplayName {
				t.Errorf("got %v, want %v", result.Fields[1].Config.DisplayName, testCase.name)
			}
			if testCase.pvname != result.Fields[1].Labels["pvname"] {
				t.Errorf("got %v, want %v", result.Fields[1].Labels["pvname"], testCase.pvname)
			}
			if testCase.name != result.Fields[1].Name {
				t.Errorf("got %v, want %v", result.Fields[1].Name, testCase.name)
			}
			for i := 0; i < result.Fields[1].Len(); i++ {
				if testCase.values[i] != result.Fields[1].CopyAt(i) {
					t.Errorf("got %v, want %v", result.Fields[1].CopyAt(i), testCase.values[i])
				}
			}
		})
	}
}

func TestToFrameArray(t *testing.T) {
	var tests = []struct {
		sD         SingleData
		name       string
		pvname     string
		fieldNames []string
		values     [][]float64
		dataSize   int
		fieldsSize int
	}{
		{
			sD: SingleData{
				Name:   "testing_name",
				PVname: "pvname",
				Values: &Arrays{
					Times:  []time.Time{TimeHelper(0), TimeHelper(1), TimeHelper(2)},
					Values: [][]float64{{1, 2}, {3, 4}, {5, 6}},
				},
			},
			name:       "testing_name",
			pvname:     "pvname",
			fieldNames: []string{"testing_name[0]", "testing_name[1]"},
			values:     [][]float64{{1, 3, 5}, {2, 4, 6}},
			dataSize:   3,
			fieldsSize: 3,
		},
	}
	for idx, testCase := range tests {
		testName := fmt.Sprintf("%d: %s", idx, testCase.name)
		t.Run(testName, func(t *testing.T) {
			result := testCase.sD.ToFrame()
			if len(result.Fields) != testCase.fieldsSize {
				t.Errorf("got %d, want %d", len(result.Fields), testCase.fieldsSize)
			}
			if testCase.name != result.Name {
				t.Errorf("got %v, want %v", result.Name, testCase.name)
			}
			if result.Fields[0].Name != "time" {
				t.Errorf("got %v, want time", result.Fields[0].Name)
			}
			if result.Fields[0].Len() != testCase.dataSize {
				t.Errorf("got %d, want %d", result.Fields[0].Len(), testCase.dataSize)
			}
			for idx, v := range result.Fields[1:] {
				if testCase.fieldNames[idx] != v.Config.DisplayName {
					t.Errorf("got %v, want %v", v.Config.DisplayName, testCase.fieldNames[idx])
				}
				if testCase.pvname != v.Labels["pvname"] {
					t.Errorf("got %v, want %v", v.Labels["pvname"], testCase.pvname)
				}
				if testCase.fieldNames[idx] != v.Name {
					t.Errorf("got %v, want %v", v.Name, testCase.fieldNames[idx])
				}
				if v.Len() != testCase.dataSize {
					t.Errorf("got %d, want %d", v.Len(), testCase.dataSize)
				}
				fmt.Printf("test")
				for idy := 0; idy < v.Len(); idy++ {
					if testCase.values[idx][idy] != v.CopyAt(idy) {
						t.Errorf("got %v, want %v", v.CopyAt(idy), testCase.values[idx][idy])
					}
				}
			}
		})
	}
}

func TestExtrapolation(t *testing.T) {
	var tests = []struct {
		sDIn  SingleData
		name  string
		t     time.Time
		sDOut SingleData
	}{
		{
			sDIn: SingleData{
				Values: &Scalars{
					Times:  []time.Time{TimeHelper(0)},
					Values: []float64{1},
				},
			},
			name: "extrapolation",
			t:    TimeHelper(5),
			sDOut: SingleData{
				Values: &Scalars{
					Times:  []time.Time{TimeHelper(0), TimeHelper(5)},
					Values: []float64{1, 1},
				},
			},
		},
		{
			sDIn: SingleData{
				Values: &Arrays{
					Times:  []time.Time{TimeHelper(0)},
					Values: [][]float64{{1, 1}},
				},
			},
			name: "extrapolation",
			t:    TimeHelper(5),
			sDOut: SingleData{
				Values: &Arrays{
					Times:  []time.Time{TimeHelper(0), TimeHelper(5)},
					Values: [][]float64{{1, 1}, {1, 1}},
				},
			},
		},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			result := testCase.sDIn.Extrapolation(testCase.t)
			SingleDataCompareHelper(
				[]*SingleData{result},
				[]*SingleData{&testCase.sDOut},
				t,
			)
		})
	}
}

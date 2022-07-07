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

func TestToFrame(t *testing.T) {
	var tests = []struct {
		sD     SingleData
		name   string
		pvname string
	}{
		{
			sD: SingleData{
				Name:   "testing_name",
				PVname: "pvname",
			},
			name:   "testing_name",
			pvname: "pvname",
		},
	}
	for idx, testCase := range tests {
		testName := fmt.Sprintf("%d: %s", idx, testCase.name)
		t.Run(testName, func(t *testing.T) {
			result := testCase.sD.ToFrame()
			if testCase.name != result.Name {
				t.Errorf("got %v, want %v", result.Name, testCase.name)
			}
			if testCase.name != result.Fields[1].Name {
				t.Errorf("got %v, want %v", result.Fields[1].Name, testCase.name)
			}
			if testCase.name != result.Fields[1].Config.DisplayName {
				t.Errorf("got %v, want %v", result.Fields[1].Config.DisplayName, testCase.name)
			}
			if testCase.pvname != result.Fields[1].Labels["pvname"] {
				t.Errorf("got %v, want %v", result.Fields[1].Labels["pvname"], testCase.pvname)
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
				Times:  []time.Time{TimeHelper(0)},
				Values: []float64{1},
			},
			name: "extrapolation",
			t:    TimeHelper(5),
			sDOut: SingleData{
				Times:  []time.Time{TimeHelper(0), TimeHelper(5)},
				Values: []float64{1, 1},
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

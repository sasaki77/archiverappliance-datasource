package main

import (
	"fmt"
	"testing"
	"time"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
)

func TestArchiverSingleQuery(t *testing.T) {
	t.Skipf("Test not implemented")
}

func TestBuildRegexUrl(t *testing.T) {
	t.Skipf("Test not implemented")
}

func TestApplyAlias(t *testing.T) {
	var tests = []struct {
		name    string
		inputSd []*SingleData
		qm      ArchiverQueryModel
		output  []*SingleData
	}{
		{
			name: "normal alias",
			inputSd: []*SingleData{
				{
					Name:   "PV:NAME",
					PVname: "PV:NAME",
				},
			},
			qm: ArchiverQueryModel{
				Alias: "alias",
			},
			output: []*SingleData{
				{
					Name:   "alias",
					PVname: "PV:NAME",
				},
			},
		},
		{
			name: "empty alias",
			inputSd: []*SingleData{
				{
					Name:   "PV:NAME",
					PVname: "PV:NAME",
				},
			},
			qm: ArchiverQueryModel{
				Alias: "",
			},
			output: []*SingleData{
				{
					Name:   "PV:NAME",
					PVname: "PV:NAME",
				},
			},
		},
		{
			name: "alias pattern",
			inputSd: []*SingleData{
				{
					Name:   "PV:NAME",
					PVname: "PV:NAME",
				},
			},
			qm: ArchiverQueryModel{
				Alias:        "$2:$1",
				AliasPattern: "(.*):(.*)",
			},
			output: []*SingleData{
				{
					Name:   "NAME:PV",
					PVname: "PV:NAME",
				},
			},
		},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			result, _ := applyAlias(testCase.inputSd, testCase.qm)
			SingleDataCompareHelper(result, testCase.output, t)
		})
	}
}

func TestDataExtrapol(t *testing.T) {
	var tests = []struct {
		sDIn  SingleData
		qm    ArchiverQueryModel
		sDOut SingleData
	}{
		{
			sDIn: SingleData{
				Times:  []time.Time{TimeHelper(0)},
				Values: []float64{1},
			},
			qm: ArchiverQueryModel{
				Operator: "raw",
				TimeRange: backend.TimeRange{
					From: TimeHelper(1),
					To:   TimeHelper(5),
				},
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
			qm: ArchiverQueryModel{
				TimeRange: backend.TimeRange{
					From: TimeHelper(1),
					To:   TimeHelper(5),
				},
			},
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
				TimeRange: backend.TimeRange{
					From: TimeHelper(1),
					To:   TimeHelper(5),
				},
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
				TimeRange: backend.TimeRange{
					From: TimeHelper(1),
					To:   TimeHelper(5),
				},
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
			qm: ArchiverQueryModel{
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
			result := dataExtrapol(&testCase.sDIn, testCase.qm)
			SingleDataCompareHelper(
				[]*SingleData{result},
				[]*SingleData{&testCase.sDOut},
				t,
			)
		})
	}
}

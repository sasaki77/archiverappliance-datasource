package main

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/data"
)

type fakeClient struct {
}

func (f fakeClient) FetchRegexTargetPVs(regex string) ([]string, error) {
	pvs := []string{"PV:NAME1", "PV:NAME2"}
	return pvs, nil
}

func (f fakeClient) ExecuteSingleQuery(target string, qm ArchiverQueryModel) (SingleData, error) {
	var values []float64
	if target == "PV:NAME1" {
		values = []float64{0, 1, 2}
	} else {
		values = []float64{3, 4, 5}
	}

	v := &Scalars{Times: TimeArrayHelper(0, 3), Values: values}

	sd := SingleData{
		Name:   target,
		PVname: target,
		Values: v,
	}
	return sd, nil
}

func TestQuery(t *testing.T) {
	TIME_FORMAT := "2006-01-02T15:04:05.000-07:00"
	var tests = []struct {
		name string
		ctx  context.Context
		req  *backend.QueryDataRequest
		out  *backend.QueryDataResponse
	}{
		{
			name: "test",
			req: &backend.QueryDataRequest{
				Queries: []backend.DataQuery{
					{
						Interval: MultiReturnHelperParseDuration(time.ParseDuration("0s")),
						JSON: json.RawMessage(`{
                    		"alias": "$2:$1",
                    		"aliasPattern": "(.*):(.*)",
                    		"constant":6.5, 
                    		"functions":[], 
                    		"hide":false ,
                    		"operator": "max",
                    		"refId":"A" ,
                    		"regex":true ,
                    		"target":"PV:NAME.*" ,
							"functions":[
								{
									"params": [
										"desc"
									],
									"def": {
										"defaultParams": "",
										"shortName": "",
										"version": "",
										"category": "Sort",
										"description": "",
										"fake": false,
										"name": "sortByMax",
										"params": [
											{
												"name": "order",
												"options": ["desc", "asc"],
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
				},
			},
			out: &backend.QueryDataResponse{
				Responses: map[string]backend.DataResponse{
					"A": {
						Frames: data.Frames{
							&data.Frame{
								Name:  "NAME2:PV",
								RefID: "",
								Fields: []*data.Field{
									{
										Name: "Time",
									},
									{
										Name: "NAME2:PV",
										Labels: data.Labels{
											"pvname": "PV:NAME2",
										},
										Config: &data.FieldConfig{
											DisplayName: "NAME2:PV",
										},
									},
								},
								Meta: &data.FrameMeta{},
							},
							&data.Frame{
								Name:  "NAME1:PV",
								RefID: "",
								Fields: []*data.Field{
									{
										Name: "Time",
									},
									{
										Name: "NAME1:PV",
										Labels: data.Labels{
											"pvname": "PV:NAME1",
										},
										Config: &data.FieldConfig{
											DisplayName: "NAME1:PV",
										},
									},
								},
								Meta: &data.FrameMeta{},
							},
						},
					},
				},
			},
		},
	}
	f := fakeClient{}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			result := Query(testCase.ctx, f, testCase.req)
			for i, frame := range result.Responses["A"].Frames {
				out := testCase.out.Responses["A"].Frames[i]

				if frame.Name != out.Name {
					t.Errorf("got %v, want %v", frame.Name, out.Name)
				}
				if frame.RefID != out.RefID {
					t.Errorf("got %v, want %v", frame.RefID, out.RefID)
				}

				tf := frame.Fields[0]
				if tf.Name != "time" {
					t.Errorf("got %v, want %v", tf.Name, "time")
				}
				if tf.Len() != 3 {
					t.Errorf("got %v, want %v", tf.Len(), 3)
				}

				vf := frame.Fields[1]
				outvf := out.Fields[1]
				if vf.Name != outvf.Name {
					t.Errorf("got %v, want %v", vf.Name, outvf.Name)
				}
				if vf.Labels["pvname"] != outvf.Labels["pvname"] {
					t.Errorf("got %v, want %v", vf.Labels["pvname"], outvf.Labels["pvname"])
				}
				if vf.Config.DisplayName != outvf.Config.DisplayName {
					t.Errorf("got %v, want %v", vf.Config.DisplayName, outvf.Config.DisplayName)
				}
				if vf.Len() != 3 {
					t.Errorf("got %v, want %v", vf.Len(), 3)
				}
			}
		})
	}
}

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
					Values: &Scalars{},
				},
			},
			qm: ArchiverQueryModel{
				Alias: "alias",
			},
			output: []*SingleData{
				{
					Name:   "alias",
					PVname: "PV:NAME",
					Values: &Scalars{},
				},
			},
		},
		{
			name: "empty alias",
			inputSd: []*SingleData{
				{
					Name:   "PV:NAME",
					PVname: "PV:NAME",
					Values: &Scalars{},
				},
			},
			qm: ArchiverQueryModel{
				Alias: "",
			},
			output: []*SingleData{
				{
					Name:   "PV:NAME",
					PVname: "PV:NAME",
					Values: &Scalars{},
				},
			},
		},
		{
			name: "alias pattern",
			inputSd: []*SingleData{
				{
					Name:   "PV:NAME",
					PVname: "PV:NAME",
					Values: &Scalars{},
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
					Values: &Scalars{},
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
				Values: &Scalars{
					Times:  []time.Time{TimeHelper(0)},
					Values: []float64{1},
				},
			},
			qm: ArchiverQueryModel{
				Operator: "raw",
				TimeRange: backend.TimeRange{
					From: TimeHelper(1),
					To:   TimeHelper(5),
				},
			},
			sDOut: SingleData{
				Values: &Scalars{
					Times:  []time.Time{TimeHelper(0), TimeHelper(5)},
					Values: []float64{1, 1},
				},
			},
		},
		{
			sDIn: SingleData{
				Values: &Scalars{
					Times:  []time.Time{TimeHelper(0)},
					Values: []float64{1},
				},
			},
			qm: ArchiverQueryModel{
				TimeRange: backend.TimeRange{
					From: TimeHelper(1),
					To:   TimeHelper(5),
				},
			},
			sDOut: SingleData{
				Values: &Scalars{
					Times:  []time.Time{TimeHelper(0)},
					Values: []float64{1},
				},
			},
		},
		{
			sDIn: SingleData{
				Values: &Scalars{
					Times:  []time.Time{TimeHelper(0)},
					Values: []float64{1},
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
				TimeRange: backend.TimeRange{
					From: TimeHelper(1),
					To:   TimeHelper(5),
				},
			},
			sDOut: SingleData{
				Values: &Scalars{
					Times:  []time.Time{TimeHelper(0), TimeHelper(5)},
					Values: []float64{1, 1},
				},
			},
		},
		{
			sDIn: SingleData{
				Values: &Scalars{
					Times:  []time.Time{TimeHelper(0)},
					Values: []float64{1},
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
				TimeRange: backend.TimeRange{
					From: TimeHelper(1),
					To:   TimeHelper(5),
				},
			},
			sDOut: SingleData{
				Values: &Scalars{
					Times:  []time.Time{TimeHelper(0)},
					Values: []float64{1},
				},
			},
		},
		{
			sDIn: SingleData{
				Values: &Scalars{
					Times:  []time.Time{TimeHelper(0), TimeHelper(3)},
					Values: []float64{1, 2},
				},
			},
			qm: ArchiverQueryModel{
				TimeRange: backend.TimeRange{
					From: TimeHelper(1),
					To:   TimeHelper(5),
				},
			},
			sDOut: SingleData{
				Values: &Scalars{
					Times:  []time.Time{TimeHelper(0), TimeHelper(3)},
					Values: []float64{1, 2},
				},
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

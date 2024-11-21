package archiverappliance

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/data"
	"github.com/sasaki77/archiverappliance-datasource/pkg/models"
	"github.com/sasaki77/archiverappliance-datasource/pkg/testhelper"
)

type fakeClient struct {
}

func (f fakeClient) FetchRegexTargetPVs(regex string, limit int) ([]string, error) {
	if regex == ".*1" {
		return []string{"PV:NAME1"}, nil
	} else if regex == ".*2" {
		return []string{"PV:NAME2"}, nil
	} else {
		return []string{}, nil
	}
}

func (f fakeClient) ExecuteSingleQuery(target string, qm models.ArchiverQueryModel) (models.SingleData, error) {
	if target == "invalid" || target == "invalid2" {
		return models.SingleData{}, errors.New("test error")
	}
	if target == "emptyErr" {
		return models.SingleData{}, errEmptyResponse
	}

	var v models.Values

	switch target {
	case "string":
		s := []string{"test1", "test2", "test3"}
		v = &models.Strings{Times: testhelper.TimeArrayHelper(0, 3), Values: s}
	case "PV:NAME1":
		values := testhelper.InitFloat64SlicePointer([]float64{0, 1, 2})
		v = &models.Scalars{Times: testhelper.TimeArrayHelper(0, 3), Values: values}
	default:
		values := testhelper.InitFloat64SlicePointer([]float64{3, 4, 5})
		v = &models.Scalars{Times: testhelper.TimeArrayHelper(0, 3), Values: values}
	}

	sd := models.SingleData{
		Name:   target,
		PVname: target,
		Values: v,
	}
	return sd, nil
}

func TestQuery(t *testing.T) {
	TIME_FORMAT := "2006-01-02T15:04:05.000-07:00"
	var tests = []struct {
		name   string
		ctx    context.Context
		req    backend.DataQuery
		config models.DatasourceSettings
		out    *backend.QueryDataResponse
	}{
		{
			name: "test",
			req: backend.DataQuery{
				Interval: testhelper.MultiReturnHelperParseDuration(time.ParseDuration("0s")),
				JSON: json.RawMessage(`{
                    		"alias": "$2:$1",
                    		"aliasPattern": "(.*):(.*)",
                    		"constant":6.5,
                    		"functions":[],
                    		"hide":false ,
                    		"operator": "max",
                    		"refId":"A" ,
                    		"regex":true ,
                    		"target":".*(1|2)" ,
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
					From: testhelper.MultiReturnHelperParse(time.Parse(TIME_FORMAT, "2021-01-27T14:30:41.678-08:00")),
					To:   testhelper.MultiReturnHelperParse(time.Parse(TIME_FORMAT, "2021-01-28T14:30:41.678-08:00")),
				},
			},
			config: models.DatasourceSettings{
				DefaultOperator: "mean",
				UID:             "uuid",
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
		{
			name: "test without sour function",
			req: backend.DataQuery{
				Interval: testhelper.MultiReturnHelperParseDuration(time.ParseDuration("0s")),
				JSON: json.RawMessage(`{
                    		"alias": "$2:$1",
                    		"aliasPattern": "(.*):(.*)",
                    		"constant":6.5,
                    		"functions":[],
                    		"hide":false ,
                    		"operator": "max",
                    		"refId":"A" ,
                    		"regex":true ,
                    		"target":".*(1|2)"
						}`),
				MaxDataPoints: 1000,
				QueryType:     "",
				RefID:         "A",
				TimeRange: backend.TimeRange{
					From: testhelper.MultiReturnHelperParse(time.Parse(TIME_FORMAT, "2021-01-27T14:30:41.678-08:00")),
					To:   testhelper.MultiReturnHelperParse(time.Parse(TIME_FORMAT, "2021-01-28T14:30:41.678-08:00")),
				},
			},
			config: models.DatasourceSettings{
				DefaultOperator: "mean",
				UID:             "uuid",
			},
			out: &backend.QueryDataResponse{
				Responses: map[string]backend.DataResponse{
					"A": {
						Frames: data.Frames{
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
						},
					},
				},
			},
		},
		{
			name: "test string response",
			req: backend.DataQuery{
				Interval: testhelper.MultiReturnHelperParseDuration(time.ParseDuration("0s")),
				JSON: json.RawMessage(`{
                    		"alias": "",
                    		"aliasPattern": "",
                    		"constant":6.5,
                    		"functions":[],
                    		"hide":false ,
                    		"operator": "",
                    		"refId":"A" ,
                    		"regex":false ,
                    		"target":"string"
						}`),
				MaxDataPoints: 1000,
				QueryType:     "",
				RefID:         "A",
				TimeRange: backend.TimeRange{
					From: testhelper.MultiReturnHelperParse(time.Parse(TIME_FORMAT, "2021-01-27T14:30:41.678-08:00")),
					To:   testhelper.MultiReturnHelperParse(time.Parse(TIME_FORMAT, "2021-01-28T14:30:41.678-08:00")),
				},
			},
			config: models.DatasourceSettings{
				DefaultOperator: "mean",
				UID:             "uuid",
			},
			out: &backend.QueryDataResponse{
				Responses: map[string]backend.DataResponse{
					"A": {
						Frames: data.Frames{
							&data.Frame{
								Name:  "string",
								RefID: "",
								Fields: []*data.Field{
									{
										Name: "Time",
									},
									{
										Name: "string",
										Labels: data.Labels{
											"pvname": "string",
										},
										Config: &data.FieldConfig{
											DisplayName: "string",
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
			result := Query(testCase.ctx, testCase.req, f, testCase.config)
			for i, frame := range result.Frames {
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

func TestLiveQuery(t *testing.T) {
	TIME_FORMAT := "2006-01-02T15:04:05.000-07:00"
	var tests = []struct {
		name   string
		ctx    context.Context
		req    backend.DataQuery
		config models.DatasourceSettings
	}{
		{
			name: "Live Test",
			req: backend.DataQuery{
				Interval: testhelper.MultiReturnHelperParseDuration(time.ParseDuration("0s")),
				JSON: json.RawMessage(`{
                    		"alias": "",
                    		"aliasPattern": "",
                    		"constant":6.5,
                    		"hide":false ,
                    		"operator": "max",
                    		"refId":"A" ,
                    		"regex":false,
							"live": true,
                    		"target":"PV:NAME1" ,
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
			config: models.DatasourceSettings{
				DefaultOperator: "mean",
				UID:             "uuid",
				UseLiveUpdate:   true,
				LiveUpdateURI:   "uri",
			},
		},
	}

	f := fakeClient{}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			result := Query(testCase.ctx, testCase.req, f, testCase.config)
			for _, frame := range result.Frames {
				path := "ds/uuid/PV=NAME1"
				if frame.Meta.Channel != path {
					t.Errorf("got %v, want %v", frame.Meta.Channel, path)
				}
			}
		})
	}
}

func TestQueryWithInvalidResponse(t *testing.T) {
	TIME_FORMAT := "2006-01-02T15:04:05.000-07:00"
	var tests = []struct {
		name   string
		ctx    context.Context
		req    backend.DataQuery
		config models.DatasourceSettings
	}{
		{
			name: "invalid response",
			req: backend.DataQuery{
				Interval: testhelper.MultiReturnHelperParseDuration(time.ParseDuration("0s")),
				JSON: json.RawMessage(`{
                    		"alias": "",
                    		"aliasPattern": "",
                    		"constant":6.5,
                    		"functions":[],
                    		"hide":false ,
                    		"operator": "",
                    		"refId":"A" ,
                    		"regex":false ,
                    		"target":"(invalid|invalid2|PV:NAME1)" ,
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
			config: models.DatasourceSettings{
				DefaultOperator: "mean",
				UID:             "uuid",
			},
		},
	}

	f := fakeClient{}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			result := Query(testCase.ctx, testCase.req, f, testCase.config)
			if result.Error == nil {
				t.Errorf("An unexpected error has occurred")
			}
			if len(result.Frames) != 1 {
				t.Errorf("resposne length sould be 1")
			}
			if result.Frames[0].Name != "PV:NAME1" {
				t.Errorf("got %v, want PV:NAME1", result.Frames[0].Name)
			}
		})
	}
}

func TestQueryWithEmptyResponse(t *testing.T) {
	TIME_FORMAT := "2006-01-02T15:04:05.000-07:00"
	var tests = []struct {
		name   string
		ctx    context.Context
		req    backend.DataQuery
		config models.DatasourceSettings
	}{
		{
			name: "invalid response",
			req: backend.DataQuery{
				Interval: testhelper.MultiReturnHelperParseDuration(time.ParseDuration("0s")),
				JSON: json.RawMessage(`{
                    		"alias": "",
                    		"aliasPattern": "",
                    		"constant":6.5,
                    		"hide":false ,
                    		"operator": "",
                    		"refId":"A" ,
                    		"regex":false ,
                    		"target":"(emptyErr|PV:NAME1)" ,
							"functions":[
								{
									"params": [
										"boolean"
									],
									"def": {
										"defaultParams": "true",
										"shortName": "",
										"version": "",
										"category": "Options",
										"description": "",
										"fake": false,
										"name": "ignoreEmptyErr",
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
					From: testhelper.MultiReturnHelperParse(time.Parse(TIME_FORMAT, "2021-01-27T14:30:41.678-08:00")),
					To:   testhelper.MultiReturnHelperParse(time.Parse(TIME_FORMAT, "2021-01-28T14:30:41.678-08:00")),
				},
			},
			config: models.DatasourceSettings{
				DefaultOperator: "mean",
				UID:             "uuid",
			},
		},
	}

	f := fakeClient{}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			result := Query(testCase.ctx, testCase.req, f, testCase.config)
			if !errors.Is(result.Error, errEmptyResponse) {
				t.Errorf("An unexpected error has occurred")
			}
			if len(result.Frames) != 1 {
				t.Errorf("resposne length sould be 1")
			}
			if result.Frames[0].Name != "PV:NAME1" {
				t.Errorf("got %v, want PV:NAME1", result.Frames[0].Name)
			}
		})
	}
}

func TestArchiverSingleQuery(t *testing.T) {
	t.Skipf("Test not implemented")
}

func TestApplyAlias(t *testing.T) {
	var tests = []struct {
		name    string
		inputSd []*models.SingleData
		qm      models.ArchiverQueryModel
		output  []*models.SingleData
	}{
		{
			name: "normal alias",
			inputSd: []*models.SingleData{
				{
					Name:   "PV:NAME",
					PVname: "PV:NAME",
					Values: &models.Scalars{},
				},
			},
			qm: models.ArchiverQueryModel{
				Alias: "alias",
			},
			output: []*models.SingleData{
				{
					Name:   "alias",
					PVname: "PV:NAME",
					Values: &models.Scalars{},
				},
			},
		},
		{
			name: "empty alias",
			inputSd: []*models.SingleData{
				{
					Name:   "PV:NAME",
					PVname: "PV:NAME",
					Values: &models.Scalars{},
				},
			},
			qm: models.ArchiverQueryModel{
				Alias: "",
			},
			output: []*models.SingleData{
				{
					Name:   "PV:NAME",
					PVname: "PV:NAME",
					Values: &models.Scalars{},
				},
			},
		},
		{
			name: "alias pattern",
			inputSd: []*models.SingleData{
				{
					Name:   "PV:NAME",
					PVname: "PV:NAME",
					Values: &models.Scalars{},
				},
			},
			qm: models.ArchiverQueryModel{
				Alias:        "$2:$1",
				AliasPattern: "(.*):(.*)",
			},
			output: []*models.SingleData{
				{
					Name:   "NAME:PV",
					PVname: "PV:NAME",
					Values: &models.Scalars{},
				},
			},
		},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			result, _ := applyAlias(testCase.inputSd, testCase.qm)
			models.SingleDataCompareHelper(result, testCase.output, t)
		})
	}
}

func TestDataExtrapol(t *testing.T) {
	var tests = []struct {
		name  string
		sDIn  models.SingleData
		qm    models.ArchiverQueryModel
		sDOut models.SingleData
	}{
		{
			name: "Interval is 0: raw mode",
			sDIn: models.SingleData{
				Values: &models.Scalars{
					Times:  []time.Time{testhelper.TimeHelper(0)},
					Values: testhelper.InitFloat64SlicePointer([]float64{1}),
				},
			},
			qm: models.ArchiverQueryModel{
				TimeRange: backend.TimeRange{
					From: testhelper.TimeHelper(1),
					To:   testhelper.TimeHelper(5),
				},
				Interval: 0,
			},
			sDOut: models.SingleData{
				Values: &models.Scalars{
					Times:  []time.Time{testhelper.TimeHelper(0), testhelper.TimeHelper(5)},
					Values: testhelper.InitFloat64SlicePointer([]float64{1, 1}),
				},
			},
		},
		{
			name: "Interval is 1: normal mode",
			sDIn: models.SingleData{
				Values: &models.Scalars{
					Times:  []time.Time{testhelper.TimeHelper(0)},
					Values: testhelper.InitFloat64SlicePointer([]float64{1}),
				},
			},
			qm: models.ArchiverQueryModel{
				TimeRange: backend.TimeRange{
					From: testhelper.TimeHelper(1),
					To:   testhelper.TimeHelper(5),
				},
				Interval: 1,
			},
			sDOut: models.SingleData{
				Values: &models.Scalars{
					Times:  []time.Time{testhelper.TimeHelper(0)},
					Values: testhelper.InitFloat64SlicePointer([]float64{1}),
				},
			},
		},
		{
			name: "Disable Extrapolation flag is false",
			sDIn: models.SingleData{
				Values: &models.Scalars{
					Times:  []time.Time{testhelper.TimeHelper(0)},
					Values: testhelper.InitFloat64SlicePointer([]float64{1}),
				},
			},
			qm: models.ArchiverQueryModel{
				Functions: []models.FunctionDescriptorQueryModel{
					{
						Def: models.FuncDefQueryModel{
							Category:      "Options",
							DefaultParams: testhelper.InitRawMsg(`true`),
							Name:          "disableExtrapol",
							Params: []models.FuncDefParamQueryModel{
								{Name: "boolean", Type: "string"},
							},
						},
						Params: []string{"false"},
					},
				},
				TimeRange: backend.TimeRange{
					From: testhelper.TimeHelper(1),
					To:   testhelper.TimeHelper(5),
				},
				Interval: 0,
			},
			sDOut: models.SingleData{
				Values: &models.Scalars{
					Times:  []time.Time{testhelper.TimeHelper(0), testhelper.TimeHelper(5)},
					Values: testhelper.InitFloat64SlicePointer([]float64{1, 1}),
				},
			},
		},
		{
			name: "Disable Extrapolation flag is true",
			sDIn: models.SingleData{
				Values: &models.Scalars{
					Times:  []time.Time{testhelper.TimeHelper(0)},
					Values: testhelper.InitFloat64SlicePointer([]float64{1}),
				},
			},
			qm: models.ArchiverQueryModel{
				Functions: []models.FunctionDescriptorQueryModel{
					{
						Def: models.FuncDefQueryModel{
							Category:      "Options",
							DefaultParams: testhelper.InitRawMsg(`true`),
							Name:          "disableExtrapol",
							Params: []models.FuncDefParamQueryModel{
								{Name: "boolean", Type: "string"},
							},
						},
						Params: []string{"true"},
					},
				},
				Operator: "raw",
				TimeRange: backend.TimeRange{
					From: testhelper.TimeHelper(1),
					To:   testhelper.TimeHelper(5),
				},
				Interval:        0,
				DisableExtrapol: true,
			},
			sDOut: models.SingleData{
				Values: &models.Scalars{
					Times:  []time.Time{testhelper.TimeHelper(0)},
					Values: testhelper.InitFloat64SlicePointer([]float64{1}),
				},
			},
		},
		{
			name: "last operator",
			sDIn: models.SingleData{
				Values: &models.Scalars{
					Times:  []time.Time{testhelper.TimeHelper(0), testhelper.TimeHelper(3)},
					Values: testhelper.InitFloat64SlicePointer([]float64{1, 2}),
				},
			},
			qm: models.ArchiverQueryModel{
				TimeRange: backend.TimeRange{
					From: testhelper.TimeHelper(1),
					To:   testhelper.TimeHelper(5),
				},
				Operator: "last",
				Interval: 0,
			},
			sDOut: models.SingleData{
				Values: &models.Scalars{
					Times:  []time.Time{testhelper.TimeHelper(0), testhelper.TimeHelper(3)},
					Values: testhelper.InitFloat64SlicePointer([]float64{1, 2}),
				},
			},
		},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			result := dataExtrapol(&testCase.sDIn, testCase.qm)
			models.SingleDataCompareHelper(
				[]*models.SingleData{result},
				[]*models.SingleData{&testCase.sDOut},
				t,
			)
		})
	}
}

func TestMakeTargetPVList(t *testing.T) {
	var tests = []struct {
		name   string
		input  string
		regex  bool
		output []string
	}{
		{
			name:   "without regex",
			input:  "PV:NAME(1|1)",
			regex:  false,
			output: []string{"PV:NAME1"},
		},
		{
			name:   "with regex",
			input:  ".*(1|1)",
			regex:  true,
			output: []string{"PV:NAME1"},
		},
	}
	f := fakeClient{}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			result := makeTargetPVList(f, testCase.input, testCase.regex, 100)
			if diff := cmp.Diff(testCase.output, result); diff != "" {
				t.Errorf("Compare value is mismatch (-v1 +v2):%s\n", diff)
			}
		})
	}
}

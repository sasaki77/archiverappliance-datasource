package models

import (
	"fmt"
	"regexp"
	"testing"
	"time"

	"github.com/grafana/grafana-plugin-sdk-go/data"
	"github.com/sasaki77/archiverappliance-datasource/pkg/testhelper"
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

func TestToFrameNil(t *testing.T) {
	var sD SingleData
	result := sD.ToFrame(FormatOption(FORMAT_TIMESERIES))
	if result.Name != "" {
		t.Errorf("Name should be empty: got %v", result.Name)
	}
	if len(result.Fields) > 1 {
		t.Errorf("Fields should be empty: got %d", len(result.Fields))
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
					Times:  []time.Time{testhelper.TimeHelper(0), testhelper.TimeHelper(1), testhelper.TimeHelper(2)},
					Values: testhelper.InitFloat64SlicePointer([]float64{1, 2, 3}),
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
			result := testCase.sD.ToFrame(FormatOption(FORMAT_TIMESERIES))
			if testCase.name != result.Name {
				t.Errorf("got %v, want %v", result.Name, testCase.name)
			}
			if result.Fields[0].Name != "time" {
				t.Errorf("got %v, want time", result.Fields[0].Name)
			}
			if result.Fields[0].Len() != testCase.dataSize {
				t.Errorf("got %d, want %d", result.Fields[0].Len(), testCase.dataSize)
			}
			if testCase.name != result.Fields[1].Config.DisplayNameFromDS {
				t.Errorf("got %v, want %v", result.Fields[1].Config.DisplayNameFromDS, testCase.name)
			}
			if testCase.pvname != result.Fields[1].Labels["pvname"] {
				t.Errorf("got %v, want %v", result.Fields[1].Labels["pvname"], testCase.pvname)
			}
			if testCase.name != result.Fields[1].Name {
				t.Errorf("got %v, want %v", result.Fields[1].Name, testCase.name)
			}
			for i := 0; i < result.Fields[1].Len(); i++ {
				rv, _ := result.Fields[1].ConcreteAt(i)
				if testCase.values[i] != rv {
					t.Errorf("got %v, want %v", rv, testCase.values[i])
				}
			}
		})
	}
}

func TestToFrameString(t *testing.T) {
	var tests = []struct {
		sD       SingleData
		name     string
		pvname   string
		values   []string
		dataSize int
	}{
		{
			sD: SingleData{
				Name:   "testing_name",
				PVname: "pvname",
				Values: &Strings{
					Times:  []time.Time{testhelper.TimeHelper(0), testhelper.TimeHelper(1), testhelper.TimeHelper(2)},
					Values: []string{"1", "2", "3"},
				},
			},
			name:     "testing_name",
			pvname:   "pvname",
			values:   []string{"1", "2", "3"},
			dataSize: 3,
		},
	}
	for idx, testCase := range tests {
		testName := fmt.Sprintf("%d: %s", idx, testCase.name)
		t.Run(testName, func(t *testing.T) {
			result := testCase.sD.ToFrame(FormatOption(FORMAT_TIMESERIES))
			if testCase.name != result.Name {
				t.Errorf("got %v, want %v", result.Name, testCase.name)
			}
			if result.Fields[0].Name != "time" {
				t.Errorf("got %v, want time", result.Fields[0].Name)
			}
			if result.Fields[0].Len() != testCase.dataSize {
				t.Errorf("got %d, want %d", result.Fields[0].Len(), testCase.dataSize)
			}
			if testCase.name != result.Fields[1].Config.DisplayNameFromDS {
				t.Errorf("got %v, want %v", result.Fields[1].Config.DisplayNameFromDS, testCase.name)
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
					Times:  []time.Time{testhelper.TimeHelper(0), testhelper.TimeHelper(1), testhelper.TimeHelper(2)},
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
			result := testCase.sD.ToFrame(FormatOption(FORMAT_TIMESERIES))
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
				if testCase.fieldNames[idx] != v.Config.DisplayNameFromDS {
					t.Errorf("got %v, want %v", v.Config.DisplayNameFromDS, testCase.fieldNames[idx])
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
				for idy := 0; idy < v.Len(); idy++ {
					if testCase.values[idx][idy] != v.CopyAt(idy) {
						t.Errorf("got %v, want %v", v.CopyAt(idy), testCase.values[idx][idy])
					}
				}
			}
		})
	}
}

func TestToFrameDtSpaceArray(t *testing.T) {
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
					Times:  []time.Time{testhelper.TimeHelperMilli(0), testhelper.TimeHelperMilli(3), testhelper.TimeHelperMilli(6)},
					Values: [][]float64{{1, 2, 3}, {4, 5, 6}, {7, 8, 9, 10}},
				},
			},
			name:       "testing_name",
			pvname:     "pvname",
			fieldNames: []string{"testing_name"},
			values:     [][]float64{{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}},
			dataSize:   10,
			fieldsSize: 2,
		},
	}
	for idx, testCase := range tests {
		testName := fmt.Sprintf("%d: %s", idx, testCase.name)
		t.Run(testName, func(t *testing.T) {
			result := testCase.sD.ToFrame(FormatOption(FORMAT_DTSPACE))
			if len(result.Fields) != testCase.fieldsSize {
				t.Errorf("got %d, want %d", len(result.Fields), testCase.fieldsSize)
			}
			if testCase.name != result.Name {
				t.Errorf("got %v, want %v", result.Name, testCase.name)
			}
			if result.Fields[0].Name != "time" {
				t.Errorf("got %v, want index", result.Fields[0].Name)
			}
			if result.Fields[0].Len() != testCase.dataSize {
				t.Errorf("got %d, want %d", result.Fields[0].Len(), testCase.dataSize)
			}
			for i := 0; i < result.Fields[0].Len(); i++ {
				if result.Fields[0].CopyAt(i) != testhelper.TimeHelperMilli(i) {
					t.Errorf("got %v, want %v", result.Fields[0].CopyAt(i), testhelper.TimeHelperMilli(i))
				}
			}
			for idx, v := range result.Fields[1:] {
				if testCase.fieldNames[idx] != v.Config.DisplayNameFromDS {
					t.Errorf("got %v, want %v", v.Config.DisplayNameFromDS, testCase.fieldNames[idx])
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
				for idy := 0; idy < v.Len(); idy++ {
					if testCase.values[idx][idy] != v.CopyAt(idy) {
						t.Errorf("got %v, want %v", v.CopyAt(idy), testCase.values[idx][idy])
					}
				}
			}
		})
	}
}

func TestToFrameIndexArray(t *testing.T) {
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
					Times:  []time.Time{testhelper.TimeHelper(0), testhelper.TimeHelper(1), testhelper.TimeHelper(2)},
					Values: [][]float64{{1, 2, 3}, {4, 5, 6}, {7, 8, 9, 10}},
				},
			},
			name:       "testing_name",
			pvname:     "pvname",
			fieldNames: []string{"2021-01-10T01:00:00.000Z", "2021-01-10T01:01:00.000Z", "2021-01-10T01:02:00.000Z"},
			values:     [][]float64{{1, 2, 3}, {4, 5, 6}, {7, 8, 9}},
			dataSize:   3,
			fieldsSize: 4,
		},
	}
	time.Local = time.FixedZone("UTC", 0)
	for idx, testCase := range tests {
		testName := fmt.Sprintf("%d: %s", idx, testCase.name)
		t.Run(testName, func(t *testing.T) {
			result := testCase.sD.ToFrame(FormatOption(FORMAT_INDEX))
			if len(result.Fields) != testCase.fieldsSize {
				t.Errorf("got %d, want %d", len(result.Fields), testCase.fieldsSize)
			}
			if testCase.name != result.Name {
				t.Errorf("got %v, want %v", result.Name, testCase.name)
			}
			if result.Fields[0].Name != "index" {
				t.Errorf("got %v, want index", result.Fields[0].Name)
			}
			if result.Fields[0].Len() != testCase.dataSize {
				t.Errorf("got %d, want %d", result.Fields[0].Len(), testCase.dataSize)
			}
			for idx, v := range result.Fields[1:] {
				if testCase.fieldNames[idx] != v.Config.DisplayNameFromDS {
					t.Errorf("got %v, want %v", v.Config.DisplayNameFromDS, testCase.fieldNames[idx])
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
				for idy := 0; idy < v.Len(); idy++ {
					if testCase.values[idx][idy] != v.CopyAt(idy) {
						t.Errorf("got %v, want %v", v.CopyAt(idy), testCase.values[idx][idy])
					}
				}
			}
		})
	}
}

func TestToFrameIndexArrayInJST(t *testing.T) {
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
					Times:  []time.Time{testhelper.TimeHelper(0), testhelper.TimeHelper(1), testhelper.TimeHelper(2)},
					Values: [][]float64{{1, 2, 3}, {4, 5, 6}, {7, 8, 9, 10}},
				},
			},
			name:       "testing_name",
			pvname:     "pvname",
			fieldNames: []string{"2021-01-10T10:00:00.000+09:00", "2021-01-10T10:01:00.000+09:00", "2021-01-10T10:02:00.000+09:00"},
			values:     [][]float64{{1, 2, 3}, {4, 5, 6}, {7, 8, 9}},
			dataSize:   3,
			fieldsSize: 4,
		},
	}
	time.Local = time.FixedZone("Asia/Tokyo", 9*60*60)
	for idx, testCase := range tests {
		testName := fmt.Sprintf("%d: %s", idx, testCase.name)
		t.Run(testName, func(t *testing.T) {
			result := testCase.sD.ToFrame(FormatOption(FORMAT_INDEX))
			if len(result.Fields) != testCase.fieldsSize {
				t.Errorf("got %d, want %d", len(result.Fields), testCase.fieldsSize)
			}
			if testCase.name != result.Name {
				t.Errorf("got %v, want %v", result.Name, testCase.name)
			}
			if result.Fields[0].Name != "index" {
				t.Errorf("got %v, want index", result.Fields[0].Name)
			}
			if result.Fields[0].Len() != testCase.dataSize {
				t.Errorf("got %d, want %d", result.Fields[0].Len(), testCase.dataSize)
			}
			for idx, v := range result.Fields[1:] {
				if testCase.fieldNames[idx] != v.Config.DisplayNameFromDS {
					t.Errorf("got %v, want %v", v.Config.DisplayNameFromDS, testCase.fieldNames[idx])
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
				for idy := 0; idy < v.Len(); idy++ {
					if testCase.values[idx][idy] != v.CopyAt(idy) {
						t.Errorf("got %v, want %v", v.CopyAt(idy), testCase.values[idx][idy])
					}
				}
			}
		})
	}
}

func TestToFrameEnum(t *testing.T) {
	var tests = []struct {
		sD       SingleData
		name     string
		pvname   string
		values   []data.EnumItemIndex
		dataSize int
	}{
		{
			sD: SingleData{
				Name:   "testing_name",
				PVname: "pvname",
				Values: &Enums{
					Times:  []time.Time{testhelper.TimeHelper(0), testhelper.TimeHelper(1), testhelper.TimeHelper(2)},
					Values: []data.EnumItemIndex{1, 2, 3},
					EnumConfig: data.EnumFieldConfig{
						Text:  []string{"NO_ALARM", "MINOR"},
						Color: []string{"rgb(86, 166, 75)", "rgb(255, 120, 10)"},
					},
				},
			},
			name:     "testing_name",
			pvname:   "pvname",
			values:   []data.EnumItemIndex{1, 2, 3},
			dataSize: 3,
		},
	}
	for idx, testCase := range tests {
		testName := fmt.Sprintf("%d: %s", idx, testCase.name)
		t.Run(testName, func(t *testing.T) {
			result := testCase.sD.ToFrame(FormatOption(FORMAT_TIMESERIES))
			if testCase.name != result.Name {
				t.Errorf("got %v, want %v", result.Name, testCase.name)
			}
			if result.Fields[0].Name != "time" {
				t.Errorf("got %v, want time", result.Fields[0].Name)
			}
			if result.Fields[0].Len() != testCase.dataSize {
				t.Errorf("got %d, want %d", result.Fields[0].Len(), testCase.dataSize)
			}
			if testCase.name != result.Fields[1].Config.DisplayNameFromDS {
				t.Errorf("got %v, want %v", result.Fields[1].Config.DisplayNameFromDS, testCase.name)
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
			if result.Fields[1].Config.TypeConfig.Enum.Text[0] != "NO_ALARM" {
				t.Errorf("got %v, want NO_ALARM", result.Fields[1].Config.TypeConfig.Enum.Text[0])
			}
			if result.Fields[1].Config.TypeConfig.Enum.Color[1] != "rgb(255, 120, 10)" {
				t.Errorf("got %v, want rgb(255, 120, 10)", result.Fields[1].Config.TypeConfig.Enum.Color[1])
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
					Times:  []time.Time{testhelper.TimeHelper(0), testhelper.TimeHelper(1), testhelper.TimeHelper(2)},
					Values: testhelper.InitFloat64SlicePointer([]float64{1, 2, 3}),
				},
			},
			name: "scalars extrapolation",
			t:    testhelper.TimeHelper(5),
			sDOut: SingleData{
				Values: &Scalars{
					Times:  []time.Time{testhelper.TimeHelper(0), testhelper.TimeHelper(1), testhelper.TimeHelper(2), testhelper.TimeHelper(5)},
					Values: testhelper.InitFloat64SlicePointer([]float64{1, 2, 3, 3}),
				},
			},
		},
		{
			sDIn: SingleData{
				Values: &Scalars{
					Times:  []time.Time{testhelper.TimeHelper(0), testhelper.TimeHelper(1)},
					Values: append([]*float64{nil}, testhelper.InitFloat64SlicePointer([]float64{1})...),
				},
			},
			name: "scalars extrapolation",
			t:    testhelper.TimeHelper(5),
			sDOut: SingleData{
				Values: &Scalars{
					Times:  []time.Time{testhelper.TimeHelper(0), testhelper.TimeHelper(1), testhelper.TimeHelper(5)},
					Values: append([]*float64{nil}, testhelper.InitFloat64SlicePointer([]float64{1, 1})...),
				},
			},
		},
		{
			sDIn: SingleData{
				Values: &Strings{
					Times:  []time.Time{testhelper.TimeHelper(0)},
					Values: []string{"1"},
				},
			},
			name: "strings extrapolation",
			t:    testhelper.TimeHelper(5),
			sDOut: SingleData{
				Values: &Strings{
					Times:  []time.Time{testhelper.TimeHelper(0)},
					Values: []string{"1"},
				},
			},
		},
		{
			sDIn: SingleData{
				Values: &Arrays{
					Times:  []time.Time{testhelper.TimeHelper(0)},
					Values: [][]float64{{1, 1}},
				},
			},
			name: "arrays extrapolation",
			t:    testhelper.TimeHelper(5),
			sDOut: SingleData{
				Values: &Arrays{
					Times:  []time.Time{testhelper.TimeHelper(0), testhelper.TimeHelper(5)},
					Values: [][]float64{{1, 1}, {1, 1}},
				},
			},
		},
		{
			sDIn: SingleData{
				Values: &Enums{
					Times:  []time.Time{testhelper.TimeHelper(0)},
					Values: []data.EnumItemIndex{1},
				},
			},
			name: "enums extrapolation",
			t:    testhelper.TimeHelper(5),
			sDOut: SingleData{
				Values: &Enums{
					Times:  []time.Time{testhelper.TimeHelper(0), testhelper.TimeHelper(5)},
					Values: []data.EnumItemIndex{1, 1},
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

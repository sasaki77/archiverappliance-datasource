package models

import (
	"fmt"
	"time"

	"github.com/grafana/grafana-plugin-sdk-go/data"
)

type Arrays struct {
	Times  []time.Time
	Values [][]float64
}

func NewArrays(length int) *Arrays {
	return &Arrays{
		Times:  make([]time.Time, 0, length),
		Values: make([][]float64, 0, length),
	}
}

func (v *Arrays) Append(val []float64, t time.Time) {
	v.Values = append(v.Values, val)
	v.Times = append(v.Times, t)
}

func (v *Arrays) ToFields(pvname string, name string, format FormatOption) []*data.Field {
	if format == FormatOption(FORMAT_DTSPACE) {
		fields := v.makeDtSpaceFields(pvname, name)
		return fields
	}

	if format == FormatOption(FORMAT_INDEX) {
		fields := v.makeIndexFields(pvname, name)
		return fields
	}

	// Default: Timeseries
	fields := v.makeTimeseriesFields(pvname, name)
	return fields
}

func (v *Arrays) Extrapolation(t time.Time) {
	if len(v.Values) == 0 {
		return
	}

	v.Values = append(v.Values, v.Values[len(v.Values)-1])
	v.Times = append(v.Times, t)
}

func (v *Arrays) makeDtSpaceFields(pvname string, name string) []*data.Field {
	var times []time.Time
	var vals []float64

	for i, row := range v.Values {
		for j, column := range row {
			vals = append(vals, column)
			times = append(times, v.Times[i].Add(time.Duration(j)*time.Millisecond))
		}
	}

	var fields []*data.Field

	//add the time dimension
	fields = append(fields, data.NewField("time", nil, times))

	// add values
	labels := make(data.Labels, 1)
	labels["pvname"] = pvname

	valueField := data.NewField(name, labels, vals)
	valueField.Config = &data.FieldConfig{DisplayName: name}
	fields = append(fields, valueField)

	return fields
}

func (v *Arrays) makeIndexFields(pvname string, name string) []*data.Field {
	var fields []*data.Field

	//add the index field
	dataLen := len(v.Values[0])
	numbers := make([]int64, dataLen)
	for i := 0; i < dataLen; i++ {
		numbers[i] = int64(i)
	}
	fields = append(fields, data.NewField("index", nil, numbers))

	// add values
	for idx, datapoint := range v.Values {
		labels := make(data.Labels, 1)
		labels["pvname"] = pvname

		n := v.Times[idx].Local().Format("2006-01-02T15:04:05.000Z07:00")
		valueField := data.NewField(n, labels, datapoint[0:dataLen])
		valueField.Config = &data.FieldConfig{DisplayName: n}
		fields = append(fields, valueField)
	}

	return fields
}

func (v *Arrays) makeTimeseriesFields(pvname string, name string) []*data.Field {
	var fields []*data.Field

	//add the time dimension
	fields = append(fields, data.NewField("time", nil, v.Times))

	// add values
	datapoints := transpose(v.Values)
	for idx, datapoint := range datapoints {
		labels := make(data.Labels, 1)
		labels["pvname"] = pvname

		n := fmt.Sprintf("%s[%d]", name, idx)
		valueField := data.NewField(n, labels, datapoint)
		valueField.Config = &data.FieldConfig{DisplayName: n}
		fields = append(fields, valueField)
	}

	return fields
}

func transpose(slice [][]float64) [][]float64 {
	x := len(slice[0])
	y := len(slice)

	result := make([][]float64, x)
	for i := range result {
		result[i] = make([]float64, y)
	}

	for i := 0; i < x; i++ {
		for j := 0; j < y; j++ {
			result[i][j] = slice[j][i]
		}
	}

	return result
}

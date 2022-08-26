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

func (v *Arrays) ToFields(pvname string, name string) []*data.Field {
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

func (v *Arrays) Extrapolation(t time.Time) {
	v.Values = append(v.Values, v.Values[len(v.Values)-1])
	v.Times = append(v.Times, t)
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

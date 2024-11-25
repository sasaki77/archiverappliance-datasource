package models

import (
	"time"

	"github.com/grafana/grafana-plugin-sdk-go/data"
)

type Enums struct {
	Times      []time.Time
	Values     []data.EnumItemIndex
	EnumConfig data.EnumFieldConfig
}

func NewEnums(length int) *Enums {
	return &Enums{
		Times:  make([]time.Time, 0, length),
		Values: make([]data.EnumItemIndex, 0, length),
	}
}

func NewSevirityEnums(length int) *Enums {
	c := data.EnumFieldConfig{
		Text:  []string{"NO_ALARM", "MINOR", "MAJOR", "INVALID"},
		Color: []string{"rgb(86, 166, 75)", "rgb(255, 120, 10)", "rgb(224, 47, 68)", "rgb(163, 82, 204)"},
	}

	return &Enums{
		Times:      make([]time.Time, 0, length),
		Values:     make([]data.EnumItemIndex, 0, length),
		EnumConfig: c,
	}
}

func NewStatusEnums(length int) *Enums {
	c := data.EnumFieldConfig{
		Text: []string{"NO_ALARM", "READ", "WRITE", "HIHI", "HIGH", "LOLO", "LOW", "STATE", "COS", "COMM", "TIMEOUT", "HWLIMIT", "CALC", "SCAN", "LINK", "SOFT", "BAD_SUB", "UDF", "DISABLE", "SIMM", "READ_ACCESS", "WRITE_ACCESS"},
	}

	return &Enums{
		Times:      make([]time.Time, 0, length),
		Values:     make([]data.EnumItemIndex, 0, length),
		EnumConfig: c,
	}
}

func (v *Enums) Append(val int16, t time.Time) {
	v.Values = append(v.Values, data.EnumItemIndex(val))
	v.Times = append(v.Times, t)
}

func (v *Enums) ToFields(pvname string, name string, format FormatOption) []*data.Field {
	// ToFields doesn't use FormatOption in Enums for now

	var fields []*data.Field

	//add the time dimension
	fields = append(fields, data.NewField("time", nil, v.Times))

	// add values
	labels := make(data.Labels, 1)
	labels["pvname"] = pvname

	valueField := data.NewField(name, labels, v.Values)
	tc := &data.FieldTypeConfig{Enum: &v.EnumConfig}
	valueField.Config = &data.FieldConfig{DisplayNameFromDS: name, TypeConfig: tc}
	fields = append(fields, valueField)

	return fields
}

func (v *Enums) Extrapolation(t time.Time) {
	if len(v.Values) == 0 {
		return
	}
	v.Values = append(v.Values, v.Values[len(v.Values)-1])
	v.Times = append(v.Times, t)
}

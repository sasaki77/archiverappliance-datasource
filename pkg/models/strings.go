package models

import (
	"time"

	"github.com/grafana/grafana-plugin-sdk-go/data"
)

type Strings struct {
	Times  []time.Time
	Values []string
}

func NewStrings(length int) *Strings {
	return &Strings{
		Times:  make([]time.Time, 0, length),
		Values: make([]string, 0, length),
	}
}

func (v *Strings) Append(val string, t time.Time) {
	v.Values = append(v.Values, val)
	v.Times = append(v.Times, t)
}

func (v *Strings) ToFields(pvname string, name string, format FormatOption) []*data.Field {
	// ToFields doesn't use FormatOption in Strings for now

	var fields []*data.Field

	//add the time dimension
	fields = append(fields, data.NewField("time", nil, v.Times))

	// add values
	labels := make(data.Labels, 1)
	labels["pvname"] = pvname

	valueField := data.NewField(name, labels, v.Values)
	valueField.Config = &data.FieldConfig{DisplayNameFromDS: name}
	fields = append(fields, valueField)

	return fields
}

func (v *Strings) Extrapolation(t time.Time) {
}

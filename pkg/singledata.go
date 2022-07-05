package main

import (
	"regexp"
	"time"

	"github.com/grafana/grafana-plugin-sdk-go/data"
)

type SingleData struct {
	Name   string
	PVname string
	Times  []time.Time
	Values []float64
}

func (sd *SingleData) ToFrame() *data.Frame {
	// create data frame response
	frame := data.NewFrame(sd.Name)

	//add the time dimension
	frame.Fields = append(frame.Fields,
		data.NewField("time", nil, sd.Times),
	)

	// add values
	labels := make(data.Labels, 1)
	labels["pvname"] = sd.PVname

	valueField := data.NewField(sd.Name, labels, sd.Values)
	valueField.Config = &data.FieldConfig{DisplayName: sd.Name}
	frame.Fields = append(frame.Fields, valueField)

	return frame
}

func (sd *SingleData) ApplyAlias(alias string, rep *regexp.Regexp) {
	a := alias
	if rep != nil {
		a = rep.ReplaceAllString(sd.Name, alias)
	}
	sd.Name = a
}

func (sd *SingleData) Extrapolation(t time.Time) *SingleData {
	newValue := sd.Values[len(sd.Values)-1]

	sd.Values = append(sd.Values, newValue)
	sd.Times = append(sd.Times, t)

	return sd
}

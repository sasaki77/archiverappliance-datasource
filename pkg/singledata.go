package main

import (
	"regexp"
	"time"

	"github.com/grafana/grafana-plugin-sdk-go/data"
)

type Arrays [][]float64

type Values interface {
	Extrapolation(t time.Time)
	ToFields(pvname string, name string) []*data.Field
}

type SingleData struct {
	Name   string
	PVname string
	Values Values
}

func (sd *SingleData) ToFrame() *data.Frame {
	// create data frame response
	frame := data.NewFrame(sd.Name)

	//add the time dimension
	v := sd.Values.ToFields(sd.PVname, sd.Name)
	frame.Fields = append(frame.Fields, v[0], v[1])

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
	//newValue := sd.Values[len(sd.Values)-1]

	//sd.Values = append(sd.Values, newValue)
	sd.Values.Extrapolation(t)

	return sd
}

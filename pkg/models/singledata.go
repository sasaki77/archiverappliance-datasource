package models

import (
	"regexp"
	"time"

	"github.com/grafana/grafana-plugin-sdk-go/data"
)

type Values interface {
	Extrapolation(t time.Time)
	ToFields(pvname string, name string, format FormatOption) []*data.Field
}

type SingleData struct {
	Name   string
	PVname string
	Values Values
}

type FormatOption string

const (
	FORMAT_TIMESERIES = FunctionCategory("timeseries")
	FORMAT_INDEX      = FunctionCategory("index")
	FORMAT_DTSPACE    = FunctionCategory("dt-space")
)

func (sd *SingleData) ToFrame(format FormatOption) *data.Frame {
	// create data frame response
	frame := data.NewFrame(sd.Name)

	if sd.Values == nil {
		return frame
	}

	v := sd.Values.ToFields(sd.PVname, sd.Name, format)
	frame.Fields = append(frame.Fields, v...)

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
	sd.Values.Extrapolation(t)

	return sd
}

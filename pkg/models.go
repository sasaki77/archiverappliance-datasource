package main

import (
	"encoding/json"
	"fmt"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
)

type ArchiverQueryModel struct {
	// It's not apparent to me where these two originate from but they do appear to be necessary
	Format    string      `json:"format"`
	Constant  json.Number `json:"constant"`  // I don't know what this is for yet
	QueryText string      `json:"queryText"` // deprecated

	// Parameters added in AAQuery's extension of DataQuery
	Target       string                         `json:"target"`       //This will be the PV as entered by the user, or regex searching for PVs
	Alias        string                         `json:"alias"`        // What to refer to the data as in the table - I think this only for the frontend rn
	AliasPattern string                         `json:"aliasPattern"` // use for collecting a large number of returned values
	Operator     string                         `json:"operator"`     // ?
	Regex        bool                           `json:"regex"`        // configured by the user's setting of the "Regex" field in the panel
	Functions    []FunctionDescriptorQueryModel `json:"functions"`    // collection of functions to be applied to the data by the archiver

	// Only appears for visualization queries
	IntervalMs *int `json:"intervalMs,omitempty"`
	/*
	   Functions is a few layers deep:
	   Functions []FunctionDescriptorQueryModel contains:
	       Def FuncDefQueryModel contains:
	           Params []FuncDefParamQueryModel
	*/

	// Parameters from DataQuery
	RefId     string  `json:"refId"`
	Hide      *bool   `json:"hide"`
	Key       *string `json:"string"`
	QueryType *string `json:"queryType"`

	// Not from JSON
	TimeRange backend.TimeRange `json:"-"`
}

type FunctionDescriptorQueryModel struct {
	// Matched to FunctionDescriptor in types.ts
	Params []string          `json:"params"`
	Def    FuncDefQueryModel `json:"def"`
}

type FuncDefQueryModel struct {
	// Matched to FuncDef in types.ts
	DefaultParams *json.RawMessage         `json:"defaultParams,omitempty"`
	ShortName     *json.RawMessage         `json:"shortName,omitempty"`
	Version       *json.RawMessage         `json:"version,omitempty"`
	Category      string                   `json:"category"`
	Description   *string                  `json:"description,omitempty"`
	Fake          *bool                    `json:"fake,omitempty"`
	Name          string                   `json:"name"`
	Params        []FuncDefParamQueryModel `json:"params"`
}

type FuncDefParamQueryModel struct {
	Name    string    `json:"name"`
	Options *[]string `json:"options"`
	Type    string    `json:"type"`
}

type ArchiverResponseModel struct {
	// Structure for unpacking the JSON response from the Archiver
	Meta struct {
		Name     string      `json:"name"`
		Waveform bool        `json:"waveform"`
		EGU      string      `json:"EGU"`
		PREC     json.Number `json:"PREC"`
	} `json:"meta"`
	Data []struct {
		Millis *json.Number `json:"millis,omitempty"`
		Nanos  *json.Number `json:"nanos,omitempty"`
		Secs   *json.Number `json:"secs,omitempty"`
		Val    json.Number  `json:"val"`
	} `json:"data"`
}

type DatasourceSettings struct {
	URL string `json:"-"`
}

func ReadQueryModel(query backend.DataQuery) (ArchiverQueryModel, error) {
	model := ArchiverQueryModel{}

	err := json.Unmarshal(query.JSON, &model)
	if err != nil {
		return model, fmt.Errorf("error reading query: %s", err.Error())
	}

	// Copy directly from the well typed query
	model.TimeRange = query.TimeRange
	return model, nil
}

func LoadSettings(ctx backend.PluginContext) (DatasourceSettings, error) {
	model := DatasourceSettings{}

	model.URL = ctx.DataSourceInstanceSettings.URL

	return model, nil
}

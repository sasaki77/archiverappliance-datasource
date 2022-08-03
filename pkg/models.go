package main

import (
	"encoding/json"
	"fmt"
	"math"
	"time"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/backend/log"
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
	TimeRange       backend.TimeRange `json:"-"`
	Interval        int               `json:"-"`
	BackendQuery    bool              `json:"-"`
	DisableExtrapol bool              `json:"-"`
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
	Category      Category                 `json:"category"`
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
	Data json.RawMessage `json:"data"`
}

type SingleArrayResponseModel struct {
	Millis *json.Number `json:"millis,omitempty"`
	Nanos  *json.Number `json:"nanos,omitempty"`
	Secs   *json.Number `json:"secs,omitempty"`
	Val    []float64    `json:"val"`
}

type SingleScalarResponseModel struct {
	Millis *json.Number `json:"millis,omitempty"`
	Nanos  *json.Number `json:"nanos,omitempty"`
	Secs   *json.Number `json:"secs,omitempty"`
	Val    json.Number  `json:"val"`
}

type ScalarResponseModel []SingleScalarResponseModel
type ArrayResponseModel []SingleArrayResponseModel

type dataResponse interface {
	ToSingleDataValues() (Values, error)
}

func (response ScalarResponseModel) ToSingleDataValues() (Values, error) {
	// Build output data block
	dataSize := len(response)

	// initialize the slices with their final size so append operations are not necessary
	times := make([]time.Time, dataSize)
	values := make([]float64, dataSize)

	for idx, dataPt := range response {
		times[idx] = convertNanosec(dataPt.Millis)

		valCache, valErr := dataPt.Val.Float64()
		if valErr != nil {
			log.DefaultLogger.Warn("Conversion of val to float64 has failed", "Error", valErr)
		}
		values[idx] = valCache
	}

	return &Scalars{Times: times, Values: values}, nil
}

func (response ArrayResponseModel) ToSingleDataValues() (Values, error) {
	// Build output data block
	dataSize := len(response)

	// initialize the slices with their final size so append operations are not necessary
	times := make([]time.Time, dataSize)
	values := make([][]float64, dataSize)

	arraySize := len(response[0].Val)
	for i := range values {
		values[i] = make([]float64, arraySize)
	}

	for idx, dataPt := range response {
		times[idx] = convertNanosec(dataPt.Millis)
		values[idx] = dataPt.Val
	}

	return &Arrays{Times: times, Values: values}, nil
}

func convertNanosec(number *json.Number) time.Time {
	millisCache, millisErr := number.Int64()
	if millisErr != nil {
		log.DefaultLogger.Warn("Conversion of millis to int64 has failed", "Error", millisErr)
	}

	// use convert to nanoseconds
	return time.Unix(0, 1e6*millisCache)
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

	// Parameters Not from JSON
	model.TimeRange = query.TimeRange
	model.Interval, err = loadInterval(model)
	if err != nil {
		model.Interval = 0
	}
	if model.IntervalMs == nil {
		model.BackendQuery = true
	}
	model.DisableExtrapol, _ = model.LoadBooleanOption(OptionName(DisableExtrapol), false)
	return model, nil
}

func LoadSettings(ctx backend.PluginContext) (DatasourceSettings, error) {
	model := DatasourceSettings{}

	model.URL = ctx.DataSourceInstanceSettings.URL

	return model, nil
}

func loadInterval(qm ArchiverQueryModel) (int, error) {
	// No operators are necessary in this case
	if qm.Operator == "raw" || qm.Operator == "last" {
		return 0, nil
	}

	var defaultv int
	if qm.IntervalMs != nil {
		// if interval is not given by user, interval is determined by IntervalMs
		intervalMs := float64(*qm.IntervalMs)
		defaultv = int(math.Floor(intervalMs / 1000)) // convert to seconds
	}
	interval, err := qm.LoadIntOption(OptionName(BinInterval), defaultv)

	return interval, err
}

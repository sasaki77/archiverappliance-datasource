package models

import (
	"encoding/json"
	"fmt"
	"math"
	"time"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/backend/log"
)

type QueryMgr struct {
	Res    backend.DataResponse
	QRefID string
}

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
	Live         bool                           `json:"live"`         // configured by the user's setting of the "Live" field in the panel
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
	LiveOnly        bool              `json:"-"`
	MaxNumPVs       int               `json:"-"`
	DisableAutoRaw  bool              `json:"-"`
	DisableExtrapol bool              `json:"-"`
	FormatOption    FormatOption      `json:"-"`
	IgnoreEmptyErr  bool              `json:"-"`
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
	Category      FunctionCategory         `json:"category"`
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

type SingleStringResponseModel struct {
	Millis *json.Number `json:"millis,omitempty"`
	Nanos  *json.Number `json:"nanos,omitempty"`
	Secs   *json.Number `json:"secs,omitempty"`
	Val    string       `json:"val"`
}

type ScalarResponseModel []SingleScalarResponseModel
type StringResponseModel []SingleStringResponseModel
type ArrayResponseModel []SingleArrayResponseModel

type DataResponse interface {
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

func (response StringResponseModel) ToSingleDataValues() (Values, error) {
	// Build output data block
	dataSize := len(response)

	// initialize the slices with their final size so append operations are not necessary
	times := make([]time.Time, dataSize)
	values := make([]string, dataSize)

	for idx, dataPt := range response {
		times[idx] = convertNanosec(dataPt.Millis)
		values[idx] = dataPt.Val
	}

	return &Strings{Times: times, Values: values}, nil
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
	return time.Unix(0, 1e6*millisCache).UTC()
}

type DatasourceSettings struct {
	DefaultOperator string `json:"defaultOperator"`
	UseLiveUpdate   bool   `json:"useLiveUpdate"`
	LiveUpdateURI   string `json:"liveUpdateURI"`

	URL string `json:"-"`
	UID string `json:"-"`
}

func ReadQueryModel(query backend.DataQuery, config DatasourceSettings) (ArchiverQueryModel, error) {
	const REGEX_MAXIMUM_MATCHES = 1000
	model := ArchiverQueryModel{}

	err := json.Unmarshal(query.JSON, &model)
	if err != nil {
		return model, fmt.Errorf("error reading query: %s", err.Error())
	}

	if model.Operator == "" {
		if config.DefaultOperator != "" {
			model.Operator = config.DefaultOperator
		} else {
			model.Operator = "mean"
		}
	}

	// Parameters Not from JSON
	model.TimeRange = query.TimeRange
	// If "from" == "to" in seconds then "to" should be "to + 1 second"
	if model.TimeRange.To.Sub(model.TimeRange.From) < time.Second {
		model.TimeRange.To = model.TimeRange.To.Add(time.Second)
	}
	model.Interval, err = loadInterval(model)
	if err != nil {
		model.Interval = 0
	}
	if model.IntervalMs == nil {
		model.BackendQuery = true
	}
	model.MaxNumPVs, _ = model.LoadIntOption(FunctionOption(FUNC_OPTION_MAXNUMPVS), REGEX_MAXIMUM_MATCHES)
	model.DisableAutoRaw, _ = model.LoadBooleanOption(FunctionOption(FUNC_OPTION_DISABLEAUTORAW), false)
	model.DisableExtrapol, _ = model.LoadBooleanOption(FunctionOption(FUNC_OPTION_DISABLEEXTRAPOL), false)
	model.LiveOnly, _ = model.LoadBooleanOption(FunctionOption(FUNC_OPTION_LIVEONLY), false)
	model.IgnoreEmptyErr, _ = model.LoadBooleanOption(FunctionOption(FUNC_OPTION_IGNOREEMPTYERR), false)

	f, _ := model.LoadStrOption(FUNC_OPTION_ARRAY_FORMAT, string(FORMAT_TIMESERIES))
	model.FormatOption = FormatOption(f)

	return model, nil
}

func LoadSettings(ctx backend.PluginContext) (DatasourceSettings, error) {
	model := DatasourceSettings{}

	err := json.Unmarshal(ctx.DataSourceInstanceSettings.JSONData, &model)

	if err != nil {
		return model, fmt.Errorf("error reading datasource instance settings: %s", err.Error())
	}

	model.URL = ctx.DataSourceInstanceSettings.URL
	model.UID = ctx.DataSourceInstanceSettings.UID

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
	interval, err := qm.LoadIntOption(FunctionOption(FUNC_OPTION_BININTERVAL), defaultv)

	return interval, err
}

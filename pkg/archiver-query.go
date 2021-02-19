package main

import (
	"encoding/json"
	"net/http"
    "net/url"
	"time"
    "strings"
    "strconv"
    "io/ioutil"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/backend/log"
)

type ArchiverQueryModel struct {
    // It's not apparent to me where these two originate from but they do appear to be necessary
    Format string `json:"format"`
    Constant json.Number `json:"constant"` // I don't know what this is for yet
    QueryText string `json:"queryText"` // deprecated

    // Parameters added in AAQuery's extension of DataQuery
    Target string `json:"target"` //This will be the PV as entered by the user, or regex searching for PVs 
    Alias string `json:"alias"` // What to refer to the data as in the table - I think this only for the frontend rn
    AliasPattern string `json:"aliasPattern"` // use for collecting a large number of returned values 
    Operator string `json:"operator"` // ?
    Regex bool `json:"regex"` // configured by the user's setting of the "Regex" field in the panel
    Functions []FunctionDescriptorQueryModel `json:"functions"` // collection of functions to be applied to the data by the archiver
    /*
        Functions is a few layers deep:
        Functions []FunctionDescriptorQueryModel contains: 
            Def FuncDefQueryModel contains:
                Params []FuncDefParamQueryModel
    */

    // Parameters from DataQuery
    RefId string `json:"refId"`
    Hide *bool `json:"hide"`
    Key *string `json:"string"`
    QueryType *string `json:"queryType"`
    DataTopic *string `json:"dataTopic"` //??
    Datasource *string `json:"datasource"` // comes back empty -- investigate further 
}

type FunctionDescriptorQueryModel struct {
    // Matched to FunctionDescriptor in types.ts
    Params []string `json:"params"`
    Def FuncDefQueryModel `json:"def"`
}

type FuncDefQueryModel struct {
    // Matched to FuncDef in types.ts
    DefaultParams *json.RawMessage `json:"defaultParams,omitempty"`
    ShortName *json.RawMessage `json:"shortName,omitempty"`
    Version *json.RawMessage `json:"version,omitempty"`
    Category string `json:"category"`
    Description *string `json:"description,omitempty"`
    Fake *bool `"json:"fake,omitempty"`
    Name string `json:"name"`
    Params []FuncDefParamQueryModel `json:"params"`
}

type FuncDefParamQueryModel struct {
    Name string `json"name"`
    Options *[]string `json:"options"`
    Type string `json:"type"`
}

func BuildQueryUrl(target string, query backend.DataQuery, pluginctx backend.PluginContext, qm ArchiverQueryModel) string {
    // Build the URL to query the archiver built from Grafana's configuration
    // Set some constants

    const TIME_FORMAT = "2006-01-02T15:04:05.000-07:00"
    const JSON_DATA_URL = "data/getData.qw"

    // Unpack the configured URL for the datasource and use that as the base for assembling the query URL
    u, err := url.Parse(pluginctx.DataSourceInstanceSettings.URL)
    if err != nil {
        log.DefaultLogger.Warn("err", "err", err)
    }

    // apply an operator to the PV string if one (not "raw" or "last") is provided
    opQuery, opErr := CreateOperatorQuery(qm)
    if opErr != nil {
        log.DefaultLogger.Warn("Operator has not been properly created")
    }

    var targetPv string
    if len(opQuery) > 0 {
        var opBuilder strings.Builder
        opBuilder.WriteString(opQuery)
        opBuilder.WriteString("(")
        opBuilder.WriteString(target)
        opBuilder.WriteString(")")
        targetPv = opBuilder.String()

    } else {
        targetPv = target
    }

    // amend the incomplete path
    var pathBuilder strings.Builder
    pathBuilder.WriteString(u.Path)
    pathBuilder.WriteString("/")
    pathBuilder.WriteString(JSON_DATA_URL)
    u.Path = pathBuilder.String()

    // assemble the query of the URL and attach it to u
    query_vals :=  make(url.Values)
    query_vals["pv"] = []string{targetPv}
    query_vals["from"] = []string{query.TimeRange.From.Format(TIME_FORMAT)}
    query_vals["to"] = []string{query.TimeRange.To.Format(TIME_FORMAT)}
    query_vals["donotchunk"] = []string{""}
    u.RawQuery = query_vals.Encode()

    // Display the result
    return u.String()
}

type SingleData struct {
    Name string
    Times []time.Time
    Values []float64
}

type ArchiverResponseModel struct {
    // Structure for unpacking the JSON response from the Archiver
    Meta struct{
        Name string `json:"name"`
        Waveform bool `json:"waveform"`
        EGU string `json:"EGU"`
        PREC json.Number `json:"PREC"`
    } `json:"meta"`
    Data []struct{
        Millis *json.Number`json:"millis,omitempty"`
        Nanos *json.Number`json:"nanos,omitempty"`
        Secs *json.Number`json:"secs,omitempty"`
        Val json.Number `json:"val"`
    } `json:"data"`
}

func ArchiverSingleQuery(queryUrl string) ([]byte, error){
    // Take the unformatted response from the http GET request and turn it into rows of timeseries data
    var jsonAsBytes []byte

    // Make the GET request
    httpResponse, getErr := http.Get(queryUrl)
    if getErr != nil {
        log.DefaultLogger.Warn("Get request has failed", "Error", getErr)
        return jsonAsBytes, getErr
    }

    // Convert get request response to variable and close the file
    jsonAsBytes, ioErr := ioutil.ReadAll(httpResponse.Body)
    httpResponse.Body.Close()
    if ioErr != nil {
        log.DefaultLogger.Warn("Parsing of incoming data has failed", "Error", ioErr)
        return jsonAsBytes, ioErr
    }

    return jsonAsBytes, nil
}

func ArchiverSingleQueryParser(jsonAsBytes []byte) (SingleData, error){
    // Convert received data to JSON
    var sD SingleData
    var data []ArchiverResponseModel
    jsonErr := json.Unmarshal(jsonAsBytes, &data)
    if jsonErr != nil {
        log.DefaultLogger.Warn("Conversion of incoming data to JSON has failed", "Error", jsonErr)
        return sD, jsonErr
    }

    // Obtain PV name 
    sD.Name = data[0].Meta.Name

    // Build output data block
    dataSize := len(data[0].Data)

    // initialize the slices with their final size so append operations are not necessary
    sD.Times = make([]time.Time, dataSize, dataSize)
    sD.Values = make([]float64, dataSize, dataSize)

    for idx, dataPt := range data[0].Data {

        millisCache, millisErr := dataPt.Millis.Int64()
        if millisErr != nil {
            log.DefaultLogger.Warn("Conversion of millis to int64 has failed", "Error", millisErr)
        }
        // use convert to nanoseconds
        sD.Times[idx] = time.Unix(0, 1e6 * millisCache)
        valCache, valErr := dataPt.Val.Float64()
        if valErr != nil {
            log.DefaultLogger.Warn("Conversion of val to float64 has failed", "Error", valErr)
        }
        sD.Values[idx] = valCache
    }
    return sD, nil
}

func BuildRegexUrl(regex string, pluginctx backend.PluginContext) string {
    // Construct the request URL for the regex search of PVs and return it as a string
    const REGEX_URL = "bpl/getMatchingPVs"
    const REGEX_MAXIMUM_MATCHES = 1000

    // Unpack the configured URL for the datasource and use that as the base for assembling the query URL
    u, err := url.Parse(pluginctx.DataSourceInstanceSettings.URL)
        if err != nil {
        log.DefaultLogger.Warn("err", "err", err)
    }

    // amend the incomplete path
    var pathBuilder strings.Builder
    pathBuilder.WriteString(u.Path)
    pathBuilder.WriteString("/")
    pathBuilder.WriteString(REGEX_URL)
    u.Path = pathBuilder.String()

    // assemble the query of the URL and attach it to u
    query_vals :=  make(url.Values)
    query_vals["regex"] = []string{regex}
    query_vals["limit"] = []string{strconv.Itoa(REGEX_MAXIMUM_MATCHES)}
    u.RawQuery = query_vals.Encode()

    // log.DefaultLogger.Debug("u.String", "value", u.String())
    return u.String()
}

func ArchiverRegexQuery(queryUrl string) ([]byte, error){
    // Make the GET request  for the JSON list of matching PVs, parse it, and return a list of strings
    var jsonAsBytes []byte

    // Make the GET request
    httpResponse, getErr := http.Get(queryUrl)
    if getErr != nil {
        log.DefaultLogger.Warn("Get request has failed", "Error", getErr)
        return jsonAsBytes, getErr
    }

    // Convert get request response to variable and close the file
    jsonAsBytes, ioErr := ioutil.ReadAll(httpResponse.Body)
    httpResponse.Body.Close()
    if ioErr != nil {
        log.DefaultLogger.Warn("Parsing of incoming data has failed", "Error", ioErr)
        return jsonAsBytes, ioErr
    }
    return jsonAsBytes, nil
}

func ArchiverRegexQueryParser(jsonAsBytes []byte) ([]string, error){
    // Convert received data to JSON
    var pvList []string
    jsonErr := json.Unmarshal(jsonAsBytes, &pvList)
    if jsonErr != nil {
        log.DefaultLogger.Warn("Conversion of incoming data to JSON has failed", "Error", jsonErr)
        return pvList, jsonErr
    }

    return pvList, nil
}

func ExecuteSingleQuery(target string, query backend.DataQuery, pluginctx backend.PluginContext, qm ArchiverQueryModel) (SingleData, error) {
    // wrap together the individual operations build a query, execute the query, and compile the data into a singleData structure
    // target: This is the PV to be queried for. As the "query" argument may be a regular expression, the specific PV desired must be specified
    queryUrl := BuildQueryUrl(target, query, pluginctx, qm)
    queryResponse, _ := ArchiverSingleQuery(queryUrl)
    parsedResponse, _ := ArchiverSingleQueryParser(queryResponse)
    return parsedResponse, nil
}

func IsolateBasicQuery(unparsed string) []string {
    // Non-regex queries can request multiple PVs using this syntax: (PV:NAME:1|PV:NAME:2|...)
    // This function takes queries in this format and breaks them up into a list of individual PVs
    unparsed_clean := strings.TrimSpace(unparsed)
    if unparsed_clean[0] != '(' || unparsed_clean[len(unparsed_clean)-1] != ')' {
        // if the statement doesn't have the parentheses, no parsing is necessary
        return []string{unparsed_clean}
    }
    // remove leading and following parentheses
    unparsed_clean = unparsed_clean[1:len(unparsed_clean)-1]
    result := strings.Split(unparsed_clean, "|")
    return result
}


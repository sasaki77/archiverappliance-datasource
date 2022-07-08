package main

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/grafana/grafana-plugin-sdk-go/backend/log"
)

type AAclient struct {
	baseURL string
}

type client interface {
	FetchRegexTargetPVs(regex string) ([]string, error)
	ExecuteSingleQuery(target string, qm ArchiverQueryModel) (SingleData, error)
}

func NewAAClient(ctx context.Context, config DatasourceSettings) (*AAclient, error) {
	return &AAclient{
		baseURL: config.URL,
	}, nil
}

func (client AAclient) FetchRegexTargetPVs(regex string) ([]string, error) {
	regexUrl := buildRegexUrl(regex, client.baseURL)
	regexQueryResponse, _ := archiverRegexQuery(regexUrl)
	pvList, _ := archiverRegexQueryParser(regexQueryResponse)

	return pvList, nil
}

func (client AAclient) ExecuteSingleQuery(target string, qm ArchiverQueryModel) (SingleData, error) {
	// wrap together the individual operations build a query, execute the query, and compile the data into a singleData structure
	// target: This is the PV to be queried for. As the "query" argument may be a regular expression, the specific PV desired must be specified
	queryUrl := buildQueryUrl(target, client.baseURL, qm)
	queryResponse, _ := archiverSingleQuery(queryUrl)
	parsedResponse, _ := archiverSingleQueryParser(queryResponse)
	return parsedResponse, nil
}

func buildQueryUrl(target string, baseURL string, qm ArchiverQueryModel) string {
	// Build the URL to query the archiver built from Grafana's configuration
	// Set some constants

	const TIME_FORMAT = "2006-01-02T15:04:05.000-07:00"
	const JSON_DATA_URL = "data/getData.qw"

	// Unpack the configured URL for the datasource and use that as the base for assembling the query URL
	u, err := url.Parse(baseURL)
	if err != nil {
		log.DefaultLogger.Warn("err", "err", err)
	}

	// apply an operator to the PV string if one (not "raw" or "last") is provided
	opQuery, opErr := CreateOperatorQuery(qm)
	if opErr != nil {
		log.DefaultLogger.Warn("Operator has not been properly created")
	}

	// log.DefaultLogger.Debug("pluginctx","pluginctx", pluginctx)
	// log.DefaultLogger.Debug("query","query", query)

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
	query_vals := make(url.Values)
	query_vals["pv"] = []string{targetPv}
	query_vals["from"] = []string{qm.TimeRange.From.Format(TIME_FORMAT)}
	query_vals["to"] = []string{qm.TimeRange.To.Format(TIME_FORMAT)}
	query_vals["donotchunk"] = []string{""}
	u.RawQuery = query_vals.Encode()

	// Display the result
	return u.String()
}

func archiverSingleQuery(queryUrl string) ([]byte, error) {
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

func archiverSingleQueryParser(jsonAsBytes []byte) (SingleData, error) {
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
	sD.PVname = data[0].Meta.Name

	// Build output data block
	dataSize := len(data[0].Data)

	// initialize the slices with their final size so append operations are not necessary
	times := make([]time.Time, dataSize)
	values := make([]float64, dataSize)

	for idx, dataPt := range data[0].Data {

		millisCache, millisErr := dataPt.Millis.Int64()
		if millisErr != nil {
			log.DefaultLogger.Warn("Conversion of millis to int64 has failed", "Error", millisErr)
		}
		// use convert to nanoseconds
		times[idx] = time.Unix(0, 1e6*millisCache)
		valCache, valErr := dataPt.Val.Float64()
		if valErr != nil {
			log.DefaultLogger.Warn("Conversion of val to float64 has failed", "Error", valErr)
		}
		values[idx] = valCache
	}
	sD.Values = &Scalars{Times: times, Values: values}
	return sD, nil
}

func buildRegexUrl(regex string, baseURL string) string {
	// Construct the request URL for the regex search of PVs and return it as a string
	const REGEX_URL = "bpl/getMatchingPVs"
	const REGEX_MAXIMUM_MATCHES = 1000

	// Unpack the configured URL for the datasource and use that as the base for assembling the query URL
	u, err := url.Parse(baseURL)
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
	query_vals := make(url.Values)
	query_vals["regex"] = []string{regex}
	query_vals["limit"] = []string{strconv.Itoa(REGEX_MAXIMUM_MATCHES)}
	u.RawQuery = query_vals.Encode()

	// log.DefaultLogger.Debug("u.String", "value", u.String())
	return u.String()
}

func archiverRegexQuery(queryUrl string) ([]byte, error) {
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

func archiverRegexQueryParser(jsonAsBytes []byte) ([]string, error) {
	// Convert received data to JSON
	var pvList []string
	jsonErr := json.Unmarshal(jsonAsBytes, &pvList)
	if jsonErr != nil {
		log.DefaultLogger.Warn("Conversion of incoming data to JSON has failed", "Error", jsonErr)
		return pvList, jsonErr
	}

	return pvList, nil
}

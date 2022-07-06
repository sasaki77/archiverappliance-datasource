package main

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/backend/log"
)

func IsBackendQuery(pluginctx backend.PluginContext) bool {
	// Return true if this query was created by the backend as opposed to visualization query for the frontend
	return pluginctx.User != nil
}

func BuildQueryUrl(target string, config DatasourceSettings, qm ArchiverQueryModel) string {
	// Build the URL to query the archiver built from Grafana's configuration
	// Set some constants

	const TIME_FORMAT = "2006-01-02T15:04:05.000-07:00"
	const JSON_DATA_URL = "data/getData.qw"

	// Unpack the configured URL for the datasource and use that as the base for assembling the query URL
	u, err := url.Parse(config.URL)
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

func ArchiverSingleQuery(queryUrl string) ([]byte, error) {
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

func ArchiverSingleQueryParser(jsonAsBytes []byte) (SingleData, error) {
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
	sD.Times = make([]time.Time, dataSize)
	sD.Values = make([]float64, dataSize)

	for idx, dataPt := range data[0].Data {

		millisCache, millisErr := dataPt.Millis.Int64()
		if millisErr != nil {
			log.DefaultLogger.Warn("Conversion of millis to int64 has failed", "Error", millisErr)
		}
		// use convert to nanoseconds
		sD.Times[idx] = time.Unix(0, 1e6*millisCache)
		valCache, valErr := dataPt.Val.Float64()
		if valErr != nil {
			log.DefaultLogger.Warn("Conversion of val to float64 has failed", "Error", valErr)
		}
		sD.Values[idx] = valCache
	}
	return sD, nil
}

func BuildRegexUrl(regex string, config DatasourceSettings) string {
	// Construct the request URL for the regex search of PVs and return it as a string
	const REGEX_URL = "bpl/getMatchingPVs"
	const REGEX_MAXIMUM_MATCHES = 1000

	// Unpack the configured URL for the datasource and use that as the base for assembling the query URL
	u, err := url.Parse(config.URL)
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

func ArchiverRegexQuery(queryUrl string) ([]byte, error) {
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

func ArchiverRegexQueryParser(jsonAsBytes []byte) ([]string, error) {
	// Convert received data to JSON
	var pvList []string
	jsonErr := json.Unmarshal(jsonAsBytes, &pvList)
	if jsonErr != nil {
		log.DefaultLogger.Warn("Conversion of incoming data to JSON has failed", "Error", jsonErr)
		return pvList, jsonErr
	}

	return pvList, nil
}

func FetchRegexTargetPVs(regex string, config DatasourceSettings) ([]string, error) {
	regexUrl := BuildRegexUrl(regex, config)
	regexQueryResponse, _ := ArchiverRegexQuery(regexUrl)
	pvList, _ := ArchiverRegexQueryParser(regexQueryResponse)

	return pvList, nil
}

func ExecuteSingleQuery(target string, config DatasourceSettings, qm ArchiverQueryModel) (SingleData, error) {
	// wrap together the individual operations build a query, execute the query, and compile the data into a singleData structure
	// target: This is the PV to be queried for. As the "query" argument may be a regular expression, the specific PV desired must be specified
	queryUrl := BuildQueryUrl(target, config, qm)
	queryResponse, _ := ArchiverSingleQuery(queryUrl)
	parsedResponse, _ := ArchiverSingleQueryParser(queryResponse)
	return parsedResponse, nil
}

func ApplyAlias(sD []*SingleData, qm ArchiverQueryModel) ([]*SingleData, error) {
	// Alias is not set. Return data as is is.
	if qm.Alias == "" {
		return sD, nil
	}

	var rep *regexp.Regexp
	if qm.AliasPattern != "" {
		var err error
		rep, err = regexp.Compile(qm.AliasPattern)
		if err != nil {
			return sD, err
		}
	}

	for _, d := range sD {
		d.ApplyAlias(qm.Alias, rep)
	}

	return sD, nil
}

func DataExtrapol(singleResponse *SingleData, qm ArchiverQueryModel) *SingleData {
	disableExtrapol, err := qm.DisableExtrapol()
	if err != nil {
		disableExtrapol = false
	}

	if (qm.Operator != "raw") || disableExtrapol {
		return singleResponse
	}

	newResponse := singleResponse.Extrapolation(qm.TimeRange.To)

	return newResponse
}

func Query(ctx context.Context, qm ArchiverQueryModel, config DatasourceSettings) backend.DataResponse {
	response := backend.DataResponse{}

	// make the query and compile the results into a SingleData instance
	var targetPvList []string
	if qm.Regex {
		// If the user is using a regex to specify the PVs, parse and resolve the regex expression first
		// assemble the list of PVs to be queried for
		targetPvList, _ = FetchRegexTargetPVs(qm.Target, config)
	} else {
		// If a regex is not being used, only check for listed PVs
		targetPvList = IsolateBasicQuery(qm.Target)
	}

	// execute the individual queries
	responseData := make([]*SingleData, 0, len(targetPvList))
	responsePipe := make(chan SingleData)

	// Create timeout. If any request routines take longer than timeoutDurationSeconds to execute, they will be dropped.
	timeoutDurationSeconds := 30 // units are seconds
	timeoutDuration, _ := time.ParseDuration(strconv.Itoa(timeoutDurationSeconds) + "s")
	timeoutPipe := time.After(timeoutDuration)

	// create goroutines for individual requests
	for _, targetPv := range targetPvList {
		go func(targetPv string, pipe chan SingleData) {
			parsedResponse, _ := ExecuteSingleQuery(targetPv, config, qm)
			pipe <- parsedResponse
		}(targetPv, responsePipe)
	}

	// Collect responses from the request goroutines
responseCollector:
	for range targetPvList {
		select {
		case response := <-responsePipe:
			responseData = append(responseData, &response)
		case <-timeoutPipe:
			log.DefaultLogger.Warn("Timeout limit for query has been reached")
			break responseCollector
		}
	}

	// Apply Alias to the data
	var aliasErr error
	responseData, aliasErr = ApplyAlias(responseData, qm)
	if aliasErr != nil {
		log.DefaultLogger.Warn("Error applying alias")
	}

	// Apply Functions to the data
	var funcErr error
	responseData, funcErr = ApplyFunctions(responseData, qm)
	if funcErr != nil {
		log.DefaultLogger.Warn("Error applying functions")
	}

	// Extrapolate data as necessary
	for idx, data := range responseData {
		responseData[idx] = DataExtrapol(data, qm)
	}

	// for each query response, compile the data into response.Frames
	for _, singleResponse := range responseData {

		frame := singleResponse.ToFrame()

		// add the frames to the response
		response.Frames = append(response.Frames, frame)
	}

	return response
}

package archiverappliance

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/grafana/grafana-plugin-sdk-go/backend/log"
	"github.com/sasaki77/archiverappliance-datasource/pkg/models"
)

type client interface {
	FetchRegexTargetPVs(regex string, limit int) ([]string, error)
	ExecuteSingleQuery(target string, qm models.ArchiverQueryModel) (models.SingleData, error)
}

type AAclient struct {
	baseURL string
}

func NewAAClient(ctx context.Context, config models.DatasourceSettings) (*AAclient, error) {
	return &AAclient{
		baseURL: config.URL,
	}, nil
}

func (client AAclient) FetchRegexTargetPVs(regex string, limit int) ([]string, error) {
	regexUrl := buildRegexUrl(regex, client.baseURL, limit)
	regexQueryResponse, _ := archiverRegexQuery(regexUrl)
	pvList, _ := archiverRegexQueryParser(regexQueryResponse)

	return pvList, nil
}

func (client AAclient) ExecuteSingleQuery(target string, qm models.ArchiverQueryModel) (models.SingleData, error) {
	// wrap together the individual operations build a query, execute the query, and compile the data into a singleData structure
	// target: This is the PV to be queried for. As the "query" argument may be a regular expression, the specific PV desired must be specified
	queryUrl := buildQueryUrl(target, client.baseURL, qm)
	queryResponse, _ := archiverSingleQuery(queryUrl)
	parsedResponse, _ := archiverSingleQueryParser(queryResponse)
	return parsedResponse, nil
}

func buildQueryUrl(target string, baseURL string, qm models.ArchiverQueryModel) string {
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
	opQuery, opErr := createOperatorQuery(qm)
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

	// from should be same as to in last operator mode
	var from string
	if qm.Operator == "last" {
		from = qm.TimeRange.To.Format(TIME_FORMAT)
	} else {
		from = qm.TimeRange.From.Format(TIME_FORMAT)
	}

	// assemble the query of the URL and attach it to u
	query_vals := make(url.Values)
	query_vals["pv"] = []string{targetPv}
	query_vals["from"] = []string{from}
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

func archiverSingleQueryParser(jsonAsBytes []byte) (models.SingleData, error) {
	// Convert received data to JSON
	var sD models.SingleData
	var response []models.ArchiverResponseModel

	jsonErr := json.Unmarshal(jsonAsBytes, &response)
	if jsonErr != nil {
		log.DefaultLogger.Warn("Conversion of incoming data to JSON has failed", "Error", jsonErr)
		return sD, jsonErr
	}

	var d models.DataResponse
	if response[0].Meta.Waveform {
		var data models.ArrayResponseModel
		jsonErr = json.Unmarshal(response[0].Data, &data)
		if jsonErr != nil {
			log.DefaultLogger.Warn("Conversion of incoming data to JSON has failed", "Error", jsonErr)
			return sD, jsonErr
		}
		d = data
	} else {
		var data models.ScalarResponseModel
		jsonErr = json.Unmarshal(response[0].Data, &data)
		if jsonErr != nil {
			log.DefaultLogger.Warn("Conversion of incoming data to JSON has failed", "Error", jsonErr)
			return sD, jsonErr
		}
		d = data
	}

	// Obtain PV name
	sD.Name = response[0].Meta.Name
	sD.PVname = response[0].Meta.Name
	sD.Values, _ = d.ToSingleDataValues()

	return sD, nil
}

func buildRegexUrl(regex string, baseURL string, limit int) string {
	// Construct the request URL for the regex search of PVs and return it as a string
	const REGEX_URL = "bpl/getMatchingPVs"

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
	query_vals["limit"] = []string{strconv.Itoa(limit)}
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

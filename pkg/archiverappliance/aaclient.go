package archiverappliance

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/grafana/grafana-plugin-sdk-go/backend/log"
	"github.com/sasaki77/archiverappliance-datasource/pkg/models"
)

type Client interface {
	FetchRegexTargetPVs(regex string, limit int) ([]string, error)
	ExecuteSingleQuery(target string, qm models.ArchiverQueryModel) (models.SingleData, error)
}

type AAclient struct {
	baseURL string
}

func NewAAClient(ctx context.Context, url string) (*AAclient, error) {
	return &AAclient{
		baseURL: url,
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

	// For liveOnly response
	if qm.LiveOnly {
		var sD models.SingleData
		sD.Name = target
		sD.PVname = target
		sD.Values = &models.Scalars{}
		return sD, nil
	}

	queryUrl := buildQueryUrl(target, client.baseURL, qm)
	queryResponse, err := archiverSingleQuery(queryUrl)

	if err != nil {
		err = fmt.Errorf("url = %q: %w", queryUrl, err)
		return models.SingleData{}, err
	}

	defer queryResponse.Close()

	parsedResponse, err := archiverPBSingleQueryParser(queryResponse, models.FieldName(qm.FieldName), qm.MaxDataPoints, qm.HideInvalid)
	if err != nil {
		err = fmt.Errorf("target = %q: %w", target, err)
		return parsedResponse, err
	}

	parsedResponse.Name = target
	parsedResponse.PVname = target

	return parsedResponse, err
}

func buildQueryUrl(target string, baseURL string, qm models.ArchiverQueryModel) string {
	// Build the URL to query the archiver built from Grafana's configuration
	// Set some constants
	const TIME_FORMAT = "2006-01-02T15:04:05.000-07:00"
	const RAW_DATA_URL = "data/getData.raw"

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
	pathBuilder.WriteString(RAW_DATA_URL)
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

func archiverSingleQuery(queryUrl string) (io.ReadCloser, error) {
	// Make the GET request
	httpResponse, getErr := http.Get(queryUrl)

	if getErr != nil {
		log.DefaultLogger.Warn("Get request has failed", "Error", getErr)
		return nil, getErr
	}

	// Return if response is valid
	// httpResponse must be closed by caller
	if httpResponse.StatusCode == http.StatusOK {
		return httpResponse.Body, nil
	}

	// Error handling
	defer httpResponse.Body.Close()

	err := fmt.Errorf("required=200, received=%d: %w", httpResponse.StatusCode, errResponseStatusCode)
	return nil, err
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
	defer httpResponse.Body.Close()

	// Convert get request response to variable and close the file
	jsonAsBytes, ioErr := io.ReadAll(httpResponse.Body)
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

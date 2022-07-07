package main

import (
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/grafana/grafana-plugin-sdk-go/backend/log"
)

func OperatorValidator(input string) bool {
	// return true if the operator given by the user is a valid, recognized operator

	// copied from the types.ts specification
	RECOGNIZED_OPERATORS := []string{
		"",
		"firstSample",
		"lastSample",
		"firstFill",
		"lastFill",
		"mean",
		"min",
		"max",
		"count",
		"ncount",
		"nth",
		"median",
		"std",
		"jitter",
		"ignoreflyers",
		"flyers",
		"variance",
		"popvariance",
		"kurtosis",
		"skewness",
		"raw",
		"last",
	}
	for _, entry := range RECOGNIZED_OPERATORS {
		if entry == input {
			return true
		}
	}
	return false
}

func CreateOperatorQuery(qm ArchiverQueryModel) (string, error) {
	// Create the Prefix in the query to specify the operator seeking the binInterval option if necessary
	// See Datasource.ts buildURL for the matching .ts implementation

	// Skip any unrecognized operators
	if !OperatorValidator(qm.Operator) {
		errMsg := fmt.Sprintf("%v is not a recognized operator", qm.Operator)
		log.DefaultLogger.Debug("Error parsing query", "message", errMsg)
		return "", errors.New(errMsg)
	}

	// No operators are necessary in this case
	if (qm.Operator == "" && qm.IntervalMs == nil) || qm.Operator == "raw" || qm.Operator == "last" {
		return "", nil
	}

	// Load interval from query model
	binInterval, err := loadInterval(qm)
	if err != nil {
		return "", err
	}

	// interval is less than 1 second or interval is not updated from "zero value"
	if binInterval < 1 {
		return "", nil
	}

	opr := qm.Operator
	if opr == "" {
		opr = "mean"
	}

	var opBuilder strings.Builder
	opBuilder.WriteString(opr)
	opBuilder.WriteString("_")
	opBuilder.WriteString(strconv.Itoa(binInterval))

	return opBuilder.String(), nil
}

func loadInterval(qm ArchiverQueryModel) (int, error) {
	var interval int
	intervals := qm.IdentifyFunctionsByName("binInterval")

	// Determine the bin interval size given by the user and detect issues
	if len(intervals) >= 1 {
		if len(intervals) > 1 {
			log.DefaultLogger.Warn(fmt.Sprintf("more than one binInterval has been provided: %v", intervals))
		}

		val, paramErr := intervals[0].GetParametersByName("interval")
		if paramErr != nil {
			log.DefaultLogger.Warn("Conversion of binInterval argument has failed", "Error", paramErr)
			return 0, paramErr
		}

		var atoiErr error
		interval, atoiErr = strconv.Atoi(val)
		if atoiErr != nil {
			log.DefaultLogger.Warn("Failed to convert parameter string to integer", "Error", atoiErr)
			return 0, atoiErr
		}
	} else if len(intervals) == 0 && qm.IntervalMs != nil {
		// interval is not given by user, so interval is determined by IntervalMs
		intervalMs := float64(*qm.IntervalMs)
		interval = int(math.Floor(intervalMs / 1000)) // convert to seconds
	}

	return interval, nil
}

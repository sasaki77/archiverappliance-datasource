package main

import (
    "errors"
    "fmt"
    "strings"
    "strconv"

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
    if ! OperatorValidator(qm.Operator) {
        errMsg := fmt.Sprintf("%v is not a recognized operator", qm.Operator)
        log.DefaultLogger.Debug("Error parsing query", "message", errMsg)
        return "", errors.New(errMsg)
    }

    // No operators are necessary in this case
    if ( qm.Operator == "" && qm.IntervalMs == nil ) || qm.Operator == "raw" || qm.Operator == "last" {
        return "", nil
    }

    // Automatically reduce data size for large querires
    // The IntervalMs should only be non-null when the query is made from the frontend, not the alerting system
    if qm.Operator == "" && qm.IntervalMs != nil {
        interval := float64(*qm.IntervalMs)
        interval = interval/1000 // convert to seconds
        if interval <= 1 {
            return "", nil
        }
        var autoOpBuilder strings.Builder
        autoOpBuilder.WriteString("mean")
        autoOpBuilder.WriteString("_")
        autoOpBuilder.WriteString(strconv.Itoa(int(interval)))
        return autoOpBuilder.String(), nil
    }


    // Determine the bin interval size given by the user and detect issues
    var binInterval *int
    binInterval = nil
    intervals := qm.IdentifyFunctionsByName("binInterval")
    if len(intervals) >= 1 {
        if len(intervals) > 1 {
            log.DefaultLogger.Warn(fmt.Sprintf("more than one binInterval has been provided: %v", intervals))
        }

        val, paramErr := intervals[0].GetParametersByName("interval")
        if paramErr != nil {
            log.DefaultLogger.Warn("Conversion of binInterval argument has failed", "Error", paramErr)
            return "", paramErr
        }
        binInterval = new(int)
        var atoiErr error
        *binInterval, atoiErr = strconv.Atoi(val)
        if atoiErr != nil { 
            log.DefaultLogger.Warn("Failed to convert parameter string to integer", "Error", atoiErr)
        }
    } else if len(intervals) == 0 {
        // use a default value of 1
        binInterval = new(int)
        *binInterval = 1
    }
    var opBuilder strings.Builder
    opBuilder.WriteString(qm.Operator)
    opBuilder.WriteString("_")
    opBuilder.WriteString(strconv.Itoa(*binInterval))

    return opBuilder.String(), nil
}

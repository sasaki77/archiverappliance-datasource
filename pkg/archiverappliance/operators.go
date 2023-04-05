package archiverappliance

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/grafana/grafana-plugin-sdk-go/backend/log"
	"github.com/sasaki77/archiverappliance-datasource/pkg/models"
)

func operatorValidator(input string) bool {
	// return true if the operator given by the user is a valid, recognized operator

	// copied from the types.ts specification
	RECOGNIZED_OPERATORS := []string{
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

func createOperatorQuery(qm models.ArchiverQueryModel) (string, error) {
	// Create the Prefix in the query to specify the operator seeking the binInterval option if necessary
	// See Datasource.ts buildURL for the matching .ts implementation

	// Skip any unrecognized operators
	if !operatorValidator(qm.Operator) {
		errMsg := fmt.Sprintf("%v is not a recognized operator", qm.Operator)
		log.DefaultLogger.Debug("Error parsing query", "message", errMsg)
		return "", errors.New(errMsg)
	}

	// No operators are necessary in this case
	if qm.Operator == "raw" || qm.Operator == "last" {
		return "", nil
	}

	binInterval := qm.Interval

	// interval is less than 1 second or interval is not updated from "zero value"
	if binInterval < 1 {
		if !qm.DisableAutoRaw {
			return "", nil
		}

		// if DisableAutoRaw is enabled, binInterval should be 1 second
		binInterval = 1
	}

	var opBuilder strings.Builder
	opBuilder.WriteString(qm.Operator)
	opBuilder.WriteString("_")
	opBuilder.WriteString(strconv.Itoa(binInterval))

	return opBuilder.String(), nil
}

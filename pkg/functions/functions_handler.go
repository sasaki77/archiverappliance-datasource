package functions

import (
	"errors"
	"fmt"

	"github.com/grafana/grafana-plugin-sdk-go/backend/log"
	"github.com/sasaki77/archiverappliance-datasource/pkg/models"
)

func ApplyFunctions(responseData []*models.SingleData, qm models.ArchiverQueryModel) ([]*models.SingleData, error) {
	// iterate through the list of functions
	// This should be applied to the entirety of the single query as some functions need knowldege off the data series in order to work

	// Apply "Array to Scalar" functions first
	newData := responseData
	newData = applyArrayFunctions(newData, qm)

	// Apply normal functions: Transform, Filter, Sort
	newData = applyScalarFunctions(newData, qm)

	return newData, nil
}

func applyArrayFunctions(responseData []*models.SingleData, qm models.ArchiverQueryModel) []*models.SingleData {
	functions := qm.PickFuncsByCategories([]models.FunctionCategory{models.FUNC_CATEGORY_TOSCALAR})

	if len(functions) == 0 {
		return responseData
	}

	newData := make([]*models.SingleData, 0, len(functions)*len(responseData))
	for _, fdqm := range functions {
		d, err := arrayFunctionSelector(responseData, fdqm)
		if err != nil {
			continue
		}
		newData = append(newData, d...)
	}
	return newData
}

func applyScalarFunctions(responseData []*models.SingleData, qm models.ArchiverQueryModel) []*models.SingleData {
	functions := qm.PickFuncsByCategories([]models.FunctionCategory{models.FUNC_CATEGORY_TRANSFORM, models.FUNC_CATEGORY_FILTER, models.FUNC_CATEGORY_SORT})
	newData := responseData

	for _, fdqm := range functions {
		var err error
		newData, err = functionSelector(newData, fdqm)
		if err != nil {
			errMsg := fmt.Sprintf("Function %v has failed", fdqm.Def.Name)
			log.DefaultLogger.Warn(errMsg)
		}
	}

	return newData
}

func arrayFunctionSelector(responseData []*models.SingleData, fdqm models.FunctionDescriptorQueryModel) ([]*models.SingleData, error) {
	name := fdqm.Def.Name
	var f func(values []float64) float64
	var fname string

	switch name {
	case "toScalarByAvg":
		f = mean
		fname = "avg"
	case "toScalarByMax":
		f = maximum
		fname = "max"
	case "toScalarByMin":
		f = minimum
		fname = "min"
	case "toScalarBySum":
		f = sum
		fname = "sum"
	case "toScalarByMed":
		// One buffer for the whole series: medianOver sorts into it rather
		// than copying each waveform.
		var scratch []float64
		f = func(values []float64) float64 {
			var m float64
			m, scratch = medianOver(values, scratch)
			return m
		}
		fname = "median"
	case "toScalarByStd":
		f = standardDeviation
		fname = "std"
	default:
		errMsg := fmt.Sprintf("Function %v is not a recognized array function", name)
		log.DefaultLogger.Warn(errMsg)
		return []*models.SingleData{}, errors.New(errMsg)
	}

	newData := make([]*models.SingleData, 0, len(responseData))
	for _, oneData := range responseData {
		values, ok := oneData.Values.(*models.Arrays)
		if !ok {
			continue
		}

		// Arrays.Append keeps the two in step, so this only catches a container
		// assembled by hand.
		if len(values.Times) != len(values.Values) {
			errMsg := fmt.Sprintf("%v: %v timestamps for %v waveforms", fname, len(values.Times), len(values.Values))
			log.DefaultLogger.Warn(errMsg)
			return []*models.SingleData{}, errors.New(errMsg)
		}

		// Build through the container so the values come from its block storage
		// rather than one allocation per waveform.
		newValues := models.NewSclars(len(values.Values))
		for idx, val := range values.Values {
			newValues.AppendConcrete(f(val), values.Times[idx])
		}

		var d models.SingleData
		d.PVname = oneData.PVname
		d.Values = newValues
		d.Name = fmt.Sprintf("%s(%s)", oneData.Name, fname)
		newData = append(newData, &d)
	}

	return newData, nil
}

func functionSelector(responseData []*models.SingleData, fdqm models.FunctionDescriptorQueryModel) ([]*models.SingleData, error) {
	// Based on the name (as a string) of the function, select the actual function to be used
	// If the function fails to apply, the data will be returned unaltered
	name := fdqm.Def.Name
	// category := fdqm.Def.Category
	switch name {
	case "scale":
		factor, factorErr := fdqm.ExtractParamFloat64("factor")
		if factorErr != nil {
			return responseData, factorErr
		}
		newData := scale(responseData, factor)
		return newData, nil
	case "offset":
		delta, deltaErr := fdqm.ExtractParamFloat64("delta")
		if deltaErr != nil {
			return responseData, deltaErr
		}
		newData := offset(responseData, delta)
		return newData, nil
	case "delta":
		newData := delta(responseData)
		return newData, nil
	case "fluctuation":
		newData := fluctuation(responseData)
		return newData, nil
	case "movingAverage":
		windowSize, windowSizeErr := fdqm.ExtractParamInt("windowSize")
		if windowSizeErr != nil {
			return responseData, windowSizeErr
		}
		if windowSize < 1 {
			errMsg := fmt.Sprintf("movingAverage: window size must be at least 1, got %v", windowSize)
			log.DefaultLogger.Warn(errMsg)
			return responseData, errors.New(errMsg)
		}
		newData := movingAverage(responseData, windowSize)
		return newData, nil
	case "top":
		number, numberErr := fdqm.ExtractParamInt("number")
		if numberErr != nil {
			return responseData, numberErr
		}
		value, valueErr := fdqm.ExtractParamString("value")
		if valueErr != nil {
			return responseData, valueErr
		}
		newData, err := top(responseData, number, value)
		if err != nil {
			return responseData, err
		}
		return newData, nil
	case "bottom":
		number, numberErr := fdqm.ExtractParamInt("number")
		if numberErr != nil {
			return responseData, numberErr
		}
		value, valueErr := fdqm.ExtractParamString("value")
		if valueErr != nil {
			return responseData, valueErr
		}
		newData, err := bottom(responseData, number, value)
		if err != nil {
			return responseData, err
		}
		return newData, nil
	case "exclude":
		pattern, patternErr := fdqm.ExtractParamString("pattern")
		if patternErr != nil {
			return responseData, patternErr
		}
		newData, err := exclude(responseData, pattern)
		if err != nil {
			return responseData, err
		}
		return newData, nil
	case "sortByAvg":
		order, orderErr := fdqm.ExtractParamString("order")
		if orderErr != nil {
			return responseData, orderErr
		}
		newData, err := sortByAvg(responseData, order)
		if err != nil {
			return responseData, err
		}
		return newData, nil
	case "sortByMax":
		order, orderErr := fdqm.ExtractParamString("order")
		if orderErr != nil {
			return responseData, orderErr
		}
		newData, err := sortByMax(responseData, order)
		if err != nil {
			return responseData, err
		}
		return newData, nil
	case "sortByMin":
		order, orderErr := fdqm.ExtractParamString("order")
		if orderErr != nil {
			return responseData, orderErr
		}
		newData, err := sortByMin(responseData, order)
		if err != nil {
			return responseData, err
		}
		return newData, nil
	case "sortBySum":
		order, orderErr := fdqm.ExtractParamString("order")
		if orderErr != nil {
			return responseData, orderErr
		}
		newData, err := sortBySum(responseData, order)
		if err != nil {
			return responseData, err
		}
		return newData, nil
	case "sortByAbsMax":
		order, orderErr := fdqm.ExtractParamString("order")
		if orderErr != nil {
			return responseData, orderErr
		}
		newData, err := sortByAbsMax(responseData, order)
		if err != nil {
			return responseData, err
		}
		return newData, nil
	case "sortByAbsMin":
		order, orderErr := fdqm.ExtractParamString("order")
		if orderErr != nil {
			return responseData, orderErr
		}
		newData, err := sortByAbsMin(responseData, order)
		if err != nil {
			return responseData, err
		}
		return newData, nil
	default:
		errMsg := fmt.Sprintf("Function %v is not a recognized function", name)
		log.DefaultLogger.Warn(errMsg)
		return responseData, errors.New(errMsg)
	}

	// this should never be reached
	// return responseData, nil
}

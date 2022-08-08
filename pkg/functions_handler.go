package main

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/grafana/grafana-plugin-sdk-go/backend/log"
	"github.com/montanaflynn/stats"
)

type Category string

const (
	Transform = Category("Transform")
	ToScalar  = Category("Array to Scalar")
	Filter    = Category("Filter Series")
	Sort      = Category("Sort")
	Options   = Category("Options")
)

type OptionName string

const (
	MaxNumPVs       = Category("maxNumPVs")
	DisableAutoRaw  = Category("disableAutoRaw")
	DisableExtrapol = Category("disableExtrapol")
	BinInterval     = Category("binInterval")
)

func (qm ArchiverQueryModel) PickFuncsByCategories(categories []Category) []FunctionDescriptorQueryModel {
	response := make([]FunctionDescriptorQueryModel, 0)
	for _, entry := range qm.Functions {
		for _, category := range categories {
			if entry.Def.Category == category {
				response = append(response, entry)
			}
		}
	}
	return response
}

func (qm ArchiverQueryModel) IdentifyFunctionsByName(targetName string) []FunctionDescriptorQueryModel {
	// create a slice of the the FunctionDescrporQueryModels that have the type of name targetName in order
	response := make([]FunctionDescriptorQueryModel, 0)
	for _, entry := range qm.Functions {
		if entry.Def.Name == targetName {
			response = append(response, entry)
		}
	}
	return response
}

func (qm ArchiverQueryModel) LoadIntOption(name OptionName, defaultv int) (int, error) {
	functions := qm.IdentifyFunctionsByName(string(name))
	if len(functions) >= 1 {
		if len(functions) > 1 {
			log.DefaultLogger.Warn(fmt.Sprintf("more than one %s has been provided: %v", name, functions))
		}

		val, paramErr := functions[0].ExtractParamInt(functions[0].Def.Params[0].Name)
		if paramErr != nil {
			log.DefaultLogger.Warn("Conversion of int argument has failed", "Error", paramErr)
			return 0, paramErr
		}
		return val, nil
	} else {
		return defaultv, nil
	}
}

func (qm ArchiverQueryModel) LoadBooleanOption(name OptionName, defaultv bool) (bool, error) {
	functions := qm.IdentifyFunctionsByName(string(name))
	if len(functions) >= 1 {
		if len(functions) > 1 {
			log.DefaultLogger.Warn(fmt.Sprintf("more than one %s has been provided: %v", name, functions))
		}

		val, paramErr := functions[0].ExtractParamBoolean(functions[0].Def.Params[0].Name)
		if paramErr != nil {
			log.DefaultLogger.Warn("Conversion of boolean argument has failed", "Error", paramErr)
			return false, paramErr
		}
		return val, nil
	} else {
		return defaultv, nil
	}
}

func (fdqm FunctionDescriptorQueryModel) GetParametersByName(target string) (string, error) {
	// Provide the argument value for the function given its name.
	//  If multiple are received, only return the first. This should never happen.
	if len(fdqm.Params) < len(fdqm.Def.Params) {
		errMsgLen := fmt.Sprintf("List of arguments exceeded the number of arguments provided (got %v wanted %v)", len(fdqm.Params), len(fdqm.Def.Params))
		return "", errors.New(errMsgLen)
	}
	for idx, def := range fdqm.Def.Params {
		if def.Name == target {
			return fdqm.Params[idx], nil
		}
	}
	errMsg := fmt.Sprintf("Not able to identify argument %v in function %v", target, fdqm.Def.Name)
	return "", errors.New(errMsg)
}

func (fdqm FunctionDescriptorQueryModel) GetParamTypeByName(target string) (string, error) {
	// Provide the argument value for the function given its name.
	//  If multiple are received, only return the first. This should never happen.
	if len(fdqm.Params) < len(fdqm.Def.Params) {
		errMsgLen := fmt.Sprintf("List of arguments exceeded the number of arguments provided (got %v wanted %v)", len(fdqm.Params), len(fdqm.Def.Params))
		return "", errors.New(errMsgLen)
	}
	for idx, def := range fdqm.Def.Params {
		if def.Name == target {
			return fdqm.Def.Params[idx].Type, nil
		}
	}
	errMsg := fmt.Sprintf("Not able to identify type of %v in function %v", target, fdqm.Def.Name)
	return "", errors.New(errMsg)
}

func (fdqm FunctionDescriptorQueryModel) ExtractorBase(target string, targetType string) (string, error) {
	// confirm parameter type
	typeStr, getErr := fdqm.GetParamTypeByName(target)
	badReturn := ""
	if getErr != nil {
		errMsg := fmt.Sprintf("Failed to obtain parameter type for %v", target)
		log.DefaultLogger.Warn(errMsg)
		return badReturn, errors.New(errMsg)
	}
	if typeStr != targetType {
		// Warn but continue
		errMsg := fmt.Sprintf("Type %v not expected", typeStr)
		log.DefaultLogger.Warn(errMsg)
	}

	// attempt to locate and return the function's argument
	valueStr, getErr := fdqm.GetParametersByName(target)
	if getErr != nil {
		errMsg := fmt.Sprintf("Failed to obtain parameter %v", target)
		log.DefaultLogger.Warn(errMsg)
		return badReturn, errors.New(errMsg)
	}
	return valueStr, nil
}

func (fdqm FunctionDescriptorQueryModel) ExtractParamInt(target string) (int, error) {
	var result int

	// get string for argument's value and check type
	response, extractErr := fdqm.ExtractorBase(target, "int")
	if extractErr != nil {
		return result, extractErr
	}

	// convert to int
	result, conversionErr := strconv.Atoi(response)
	if conversionErr != nil {
		errMsg := fmt.Sprintf("Failed to convert %v to int", target)
		log.DefaultLogger.Warn(errMsg)
		return result, errors.New(errMsg)
	}

	return result, nil
}

func (fdqm FunctionDescriptorQueryModel) ExtractParamFloat64(target string) (float64, error) {
	var result float64

	// get string for argument's value and check type
	response, extractErr := fdqm.ExtractorBase(target, "float")
	if extractErr != nil {
		return result, extractErr
	}

	// convert to int
	result, conversionErr := strconv.ParseFloat(response, 64)
	if conversionErr != nil {
		errMsg := fmt.Sprintf("Failed to convert %v to float64", target)
		log.DefaultLogger.Warn(errMsg)
		return result, errors.New(errMsg)
	}

	return result, nil
}

func (fdqm FunctionDescriptorQueryModel) ExtractParamString(target string) (string, error) {
	var result string

	// get string for argument's value and check type
	response, extractErr := fdqm.ExtractorBase(target, "string")
	if extractErr != nil {
		return result, extractErr
	}

	return response, nil
}

func (fdqm FunctionDescriptorQueryModel) ExtractParamBoolean(target string) (bool, error) {
	var result bool

	// get string for argument's value and check type
	response, extractErr := fdqm.ExtractorBase(target, "string")
	if extractErr != nil {
		return result, extractErr
	}

	if response == "true" {
		return true, nil
	} else if response == "false" {
		return false, nil
	}

	return false, errors.New("string not recognized as boolean")
}

func ApplyFunctions(responseData []*SingleData, qm ArchiverQueryModel) ([]*SingleData, error) {
	// iterate through the list of functions
	// This should be applied to the entirety of the single query as some functions need knowldege off the data series in order to work

	// Apply "Array to Scalar" functions first
	newData := responseData
	newData = applyArrayFunctions(newData, qm)

	// Apply normal functions: Transform, Filter, Sort
	newData = applyScalarFunctions(newData, qm)

	return newData, nil
}

func applyArrayFunctions(responseData []*SingleData, qm ArchiverQueryModel) []*SingleData {
	functions := qm.PickFuncsByCategories([]Category{ToScalar})

	if len(functions) == 0 {
		return responseData
	}

	var newData []*SingleData
	for _, fdqm := range functions {
		d, err := arrayFunctionSelector(responseData, fdqm)
		if err != nil {
			continue
		}
		newData = append(newData, d...)
	}
	return newData
}

func applyScalarFunctions(responseData []*SingleData, qm ArchiverQueryModel) []*SingleData {
	functions := qm.PickFuncsByCategories([]Category{Transform, Filter, Sort})
	newData := responseData

	for _, fdqm := range functions {
		var err error
		newData, err = FunctionSelector(newData, fdqm)
		if err != nil {
			errMsg := fmt.Sprintf("Function %v has failed", fdqm.Def.Name)
			log.DefaultLogger.Warn(errMsg)
		}
	}

	return newData
}

func arrayFunctionSelector(responseData []*SingleData, fdqm FunctionDescriptorQueryModel) ([]*SingleData, error) {
	name := fdqm.Def.Name
	var f func(values stats.Float64Data) (float64, error)
	var fname string

	switch name {
	case "toScalarByAvg":
		f = stats.Mean
		fname = "avg"
	case "toScalarByMax":
		f = stats.Max
		fname = "max"
	case "toScalarByMin":
		f = stats.Min
		fname = "min"
	case "toScalarBySum":
		f = stats.Sum
		fname = "sum"
	case "toScalarByMed":
		f = stats.Median
		fname = "median"
	case "toScalarByStd":
		f = stats.StandardDeviation
		fname = "std"
	default:
		errMsg := fmt.Sprintf("Function %v is not a recognized array function", name)
		log.DefaultLogger.Warn(errMsg)
		return []*SingleData{}, errors.New(errMsg)
	}

	var newData []*SingleData
	for _, oneData := range responseData {
		values, ok := oneData.Values.(*Arrays)
		if !ok {
			continue
		}

		var vs []float64
		for _, val := range values.Values {
			v, _ := f(val)
			vs = append(vs, v)
		}

		newValues := &Scalars{Times: values.Times, Values: vs}

		var d SingleData
		d.PVname = oneData.PVname
		d.Values = newValues
		d.Name = fmt.Sprintf("%s(%s)", oneData.Name, fname)
		newData = append(newData, &d)
	}

	return newData, nil
}

func FunctionSelector(responseData []*SingleData, fdqm FunctionDescriptorQueryModel) ([]*SingleData, error) {
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
		newData := Scale(responseData, factor)
		return newData, nil
	case "offset":
		delta, deltaErr := fdqm.ExtractParamFloat64("delta")
		if deltaErr != nil {
			return responseData, deltaErr
		}
		newData := Offset(responseData, delta)
		return newData, nil
	case "delta":
		newData := Delta(responseData)
		return newData, nil
	case "fluctuation":
		newData := Fluctuation(responseData)
		return newData, nil
	case "movingAverage":
		windowSize, windowSizeErr := fdqm.ExtractParamInt("windowSize")
		if windowSizeErr != nil {
			return responseData, windowSizeErr
		}
		newData := MovingAverage(responseData, windowSize)
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
		newData, err := Top(responseData, number, value)
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
		newData, err := Bottom(responseData, number, value)
		if err != nil {
			return responseData, err
		}
		return newData, nil
	case "exclude":
		pattern, patternErr := fdqm.ExtractParamString("pattern")
		if patternErr != nil {
			return responseData, patternErr
		}
		newData, err := Exclude(responseData, pattern)
		if err != nil {
			return responseData, err
		}
		return newData, nil
	case "sortByAvg":
		order, orderErr := fdqm.ExtractParamString("order")
		if orderErr != nil {
			return responseData, orderErr
		}
		newData, err := SortByAvg(responseData, order)
		if err != nil {
			return responseData, err
		}
		return newData, nil
	case "sortByMax":
		order, orderErr := fdqm.ExtractParamString("order")
		if orderErr != nil {
			return responseData, orderErr
		}
		newData, err := SortByMax(responseData, order)
		if err != nil {
			return responseData, err
		}
		return newData, nil
	case "sortByMin":
		order, orderErr := fdqm.ExtractParamString("order")
		if orderErr != nil {
			return responseData, orderErr
		}
		newData, err := SortByMin(responseData, order)
		if err != nil {
			return responseData, err
		}
		return newData, nil
	case "sortBySum":
		order, orderErr := fdqm.ExtractParamString("order")
		if orderErr != nil {
			return responseData, orderErr
		}
		newData, err := SortBySum(responseData, order)
		if err != nil {
			return responseData, err
		}
		return newData, nil
	case "sortByAbsMax":
		order, orderErr := fdqm.ExtractParamString("order")
		if orderErr != nil {
			return responseData, orderErr
		}
		newData, err := SortByAbsMax(responseData, order)
		if err != nil {
			return responseData, err
		}
		return newData, nil
	case "sortByAbsMin":
		order, orderErr := fdqm.ExtractParamString("order")
		if orderErr != nil {
			return responseData, orderErr
		}
		newData, err := SortByAbsMin(responseData, order)
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

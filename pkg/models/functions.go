package models

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/grafana/grafana-plugin-sdk-go/backend/log"
)

type FunctionCategory string

const (
	FUNC_CATEGORY_TRANSFORM = FunctionCategory("Transform")
	FUNC_CATEGORY_TOSCALAR  = FunctionCategory("Array to Scalar")
	FUNC_CATEGORY_FILTER    = FunctionCategory("Filter Series")
	FUNC_CATEGORY_SORT      = FunctionCategory("Sort")
	FUNC_CATEGORY_OPTIONS   = FunctionCategory("Options")
)

type FunctionOption string

const (
	FUNC_OPTION_FIELDNAME       = FunctionOption("fieldName")
	FUNC_OPTION_LIVEONLY        = FunctionOption("liveOnly")
	FUNC_OPTION_MAXNUMPVS       = FunctionOption("maxNumPVs")
	FUNC_OPTION_DISABLEAUTORAW  = FunctionOption("disableAutoRaw")
	FUNC_OPTION_DISABLEEXTRAPOL = FunctionOption("disableExtrapol")
	FUNC_OPTION_BININTERVAL     = FunctionOption("binInterval")
	FUNC_OPTION_ARRAY_FORMAT    = FunctionOption("arrayFormat")
	FUNC_OPTION_IGNOREEMPTYERR  = FunctionOption("ignoreEmptyErr")
	FUNC_OPTION_HIDEINVALID     = FunctionOption("hideInvalid")
)

type FieldName string

const (
	FIELD_NAME_VAL  FieldName = "VAL"
	FIELD_NAME_SEVR FieldName = "SEVR"
	FIELD_NAME_STAT FieldName = "STAT"
)

func (qm ArchiverQueryModel) PickFuncsByCategories(categories []FunctionCategory) []FunctionDescriptorQueryModel {
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

func (qm ArchiverQueryModel) LoadIntOption(name FunctionOption, defaultv int) (int, error) {
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

func (qm ArchiverQueryModel) LoadBooleanOption(name FunctionOption, defaultv bool) (bool, error) {
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

func (qm ArchiverQueryModel) LoadStrOption(name FunctionOption, defaultv string) (string, error) {
	functions := qm.IdentifyFunctionsByName(string(name))
	if len(functions) >= 1 {
		if len(functions) > 1 {
			log.DefaultLogger.Warn(fmt.Sprintf("more than one %s has been provided: %v", name, functions))
		}

		val, paramErr := functions[0].ExtractParamString(functions[0].Def.Params[0].Name)
		if paramErr != nil {
			log.DefaultLogger.Warn("Conversion of int argument has failed", "Error", paramErr)
			return "", paramErr
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

func (fdqm FunctionDescriptorQueryModel) extractorBase(target string, targetType string) (string, error) {
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
	response, extractErr := fdqm.extractorBase(target, "int")
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
	response, extractErr := fdqm.extractorBase(target, "float")
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
	response, extractErr := fdqm.extractorBase(target, "string")
	if extractErr != nil {
		return result, extractErr
	}

	return response, nil
}

func (fdqm FunctionDescriptorQueryModel) ExtractParamBoolean(target string) (bool, error) {
	var result bool

	// get string for argument's value and check type
	response, extractErr := fdqm.extractorBase(target, "string")
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

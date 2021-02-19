package main

import (
    "errors"
    "fmt"
    "strconv"
	"github.com/grafana/grafana-plugin-sdk-go/backend/log"
)

func (qm ArchiverQueryModel) IdentifyFunctionsByName(targetName string) []FunctionDescriptorQueryModel {
    // create a slice of the the FunctionDescrporQueryModels that have the type of name targetName in order
    response := make([]FunctionDescriptorQueryModel, 0, 0) 
    for _, entry := range qm.Functions {
        if entry.Def.Name == targetName {
            response = append(response, entry)
        }
    }
    return response
}

func (fdqm FunctionDescriptorQueryModel) GetParametersByName (target string) (string, error) {
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

func (fdqm FunctionDescriptorQueryModel) GetParamTypeByName (target string) (string, error) {
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

func (fdqm FunctionDescriptorQueryModel) ExtractorBase (target string, targetType string ) (string, error) {
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

func (fdqm FunctionDescriptorQueryModel) ExtractParamInt (target string) (int, error) {
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

func (fdqm FunctionDescriptorQueryModel) ExtractParamFloat64 (target string) (float64, error) {
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

func (fdqm FunctionDescriptorQueryModel) ExtractParamString (target string) (string, error) {
    var result string

    // get string for argument's value and check type
    response, extractErr := fdqm.ExtractorBase(target, "string")
    if extractErr != nil {
        return result, extractErr
    }

    return response, nil
}

func ApplyFunctions(responseData []SingleData, qm ArchiverQueryModel) ([]SingleData, error) {
    // iterate through the list of functions
    newData := responseData
    for _, fdqm := range qm.Functions {
        var err error
        newData, err = FunctionSelector(newData, fdqm)
        if err != nil {
            errMsg := fmt.Sprintf("Function %v has failed", fdqm.Def.Name)
            log.DefaultLogger.Warn(errMsg)
        }
    }
    return newData, nil
}

func FunctionSelector(responseData []SingleData, fdqm FunctionDescriptorQueryModel) ([]SingleData, error) {
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
    return responseData, nil
}

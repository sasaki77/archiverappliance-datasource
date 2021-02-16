package main

import (
    "fmt"
    "math"
    "errors"
    "sort"
    "time"
    "regexp"
	"github.com/grafana/grafana-plugin-sdk-go/backend/log"
)


// Utilities 

type SingleDataOrder struct {
    sD SingleData
    rank float64
}

func FilterIndexer(allData []SingleData, value string) ([]float64, error) {
    rank := make([]float64, len(allData))
    for idx, sData := range allData {
        data := sData.Values
        switch value {
            case "avg":
                var total float64
                for _, val := range data {
                    total += val
                }
                rank[idx] = total/float64(len(data))
            case "min":
                var low_cache float64
                first_run := true
                for _, val := range data {
                    if first_run {
                        low_cache = val
                        first_run = false
                    }
                    if low_cache > val {
                        low_cache = val
                    }
                }
                rank[idx] = low_cache
            case "max":
                var high_cache float64
                first_run := true
                for _, val := range data {
                    if first_run {
                        high_cache = val
                        first_run = false
                    }
                    if high_cache < val {
                        high_cache = val
                    }
                }
                rank[idx] = high_cache
            case "absoluteMin":
                var low_cache float64
                first_run := true
                for _, originalVal := range data {
                    val := math.Abs(originalVal)
                    if first_run {
                        low_cache = val
                        first_run = false
                    }
                    if low_cache > val {
                        low_cache = val
                    }
                }
                rank[idx] = low_cache
            case "absoluteMax":
                var high_cache float64
                first_run := true
                for _, originalVal := range data {
                    val := math.Abs(originalVal)
                    if first_run {
                        high_cache = val
                        first_run = false
                    }
                    if high_cache < val {
                        high_cache = val
                    }
                }
                rank[idx] = high_cache
            case "sum":
                var total float64
                for _, val := range data {
                    total += val
                }
                rank[idx] = total
            default:
                errMsg := fmt.Sprintf("Value %v not recognized", value)
                return rank, errors.New(errMsg)
        }
    }
    return rank, nil
}

func SortCore(allData []SingleData, value string, order string) ([]SingleData, error) {
    newData := make([]SingleData, 0, len(allData))
    rank, idxErr  := FilterIndexer(allData, value)
    if idxErr != nil {
        return allData, idxErr
    }
    if len(rank) != len(allData) {
        errMsg := fmt.Sprintf("Length of data (%v) and indexes (%v)differ", len(allData), len(rank))
        return allData, errors.New(errMsg)
    }
    ordered := make([]SingleDataOrder, len(allData))
    for idx, _ := range allData {
        ordered[idx] = SingleDataOrder{
            sD: allData[idx],
            rank: rank[idx],
        }
    }
    if order == "asc" {
        sort.SliceStable(ordered, func(i, j int) bool{
            return ordered[i].rank < ordered[j].rank
        })
    } else if order == "desc" {
        sort.SliceStable(ordered, func(i, j int) bool{
            return ordered[i].rank > ordered[j].rank
        })
    } else {
        errMsg := fmt.Sprintf("Order %v not recognized", order)
        log.DefaultLogger.Warn(errMsg)
        return allData, errors.New(errMsg)
    }

    for idx, _ := range ordered {
        newData = append(newData, ordered[idx].sD)
    }

    return newData, nil
}


// Transform functions

func Scale(allData []SingleData, factor float64) []SingleData {
    newData := make([]SingleData, len(allData))
    for ddx, oneData := range allData {
        newValues := make([]float64, len(oneData.Values))
        for idx, val := range oneData.Values {
            newValues[idx] = val * factor
        }
        newSd := SingleData{
            Times: oneData.Times,
            Values: newValues,
        }
        newData[ddx] = newSd
    }
    return newData
}

func Offset(allData []SingleData, delta float64) []SingleData {
    newData := make([]SingleData, len(allData))
    for ddx, oneData := range allData {
        newValues := make([]float64, len(oneData.Values))
        for idx, val := range oneData.Values {
            newValues[idx] = val + delta
        }
        newSd := SingleData{
            Times: oneData.Times,
            Values: newValues,
        }
        newData[ddx] = newSd
    }
    return newData
}

func Delta(allData []SingleData) []SingleData {
    newData := make([]SingleData, len(allData))
    for ddx, oneData := range allData {
        newValues := make([]float64, 0, len(oneData.Values))
        newTimes := make([]time.Time, 0,len(oneData.Times))
        for idx, _ := range oneData.Values {
            if idx == 0 {continue}
            newValues = append(newValues, oneData.Values[idx] - oneData.Values[idx-1])
            newTimes = append(newTimes, oneData.Times[idx])
        }
        newSd := SingleData{
            Name: oneData.Name,
            Values: newValues,
            Times: newTimes,
        }
        newData[ddx] = newSd
    }
    return newData
}

func Fluctuation(allData []SingleData) []SingleData {
    newData := make([]SingleData, len(allData))
    for ddx, oneData := range allData {
        newValues := make([]float64, len(oneData.Values))
        var startingValue float64
        for idx, val := range oneData.Values {
            if idx == 0 { startingValue = val}
            newValues[idx] = val - startingValue
        }
        newSd := SingleData{
            Name: oneData.Name,
            Values: newValues,
            Times: oneData.Times,
        }
        newData[ddx] = newSd
    }
    return newData
}

func MovingAverage(allData []SingleData, windowSize int) []SingleData {
    newData := make([]SingleData, len(allData))
    for ddx, oneData := range allData {
        newValues := make([]float64, len(oneData.Values))

        for idx, _ := range oneData.Values {
            var total float64
            total = 0
            var size float64
            size = 0
            for i := 0; i < windowSize; i++ {
                if (idx-i) < 0 {break}
                size = size + 1
                total = total + oneData.Values[idx-i]
            }
            newValues[idx] = total/size
        }

        newSd := SingleData{
            Name: oneData.Name,
            Values: newValues,
            Times: oneData.Times,
        }
        newData[ddx] = newSd
    }
    return newData
}

// Array to Scalar Functions

// Filter Series Functions

func Top(allData []SingleData, number int, value string) ([]SingleData, error) {
    result, sortErr := SortCore(allData, value, "desc")
    if sortErr != nil {
        return allData, sortErr
    }
    if len(result) > number {
        return result[:number], nil
    }
    return result, nil
}

func Bottom(allData []SingleData, number int, value string) ([]SingleData, error) {
    result, sortErr := SortCore(allData, value, "asc")
    if sortErr != nil {
        return allData, sortErr
    }
    if len(result) > number {
        return result[:number], nil
    }
    return result, nil
}

func Exclude(allData []SingleData, pattern string) ([]SingleData, error){
    var newData []SingleData
    var err error

    // in preparation for regexp.Compile in case it panics
    defer func() {
        if recoveryState := recover(); recoveryState != nil {
        switch x := recoveryState.(type) {
            case string:
                err = errors.New(x)
            case error:
                err = x
            default:
                err = errors.New("Unknown panic")
            }

        }
        newData = allData
    }()

    finder, compileErr := regexp.Compile(pattern)

    if compileErr != nil {
        return allData, compileErr
    }

    for _, data := range allData {
        if !finder.MatchString(data.Name) {
            newData = append(newData, data)
        }
    }

    return newData, err
}


// Sort Functions

func SortByAvg(allData []SingleData, order string) ([]SingleData, error) {
    result, sortErr := SortCore(allData, "avg", order)
    if sortErr != nil {
        return allData, sortErr
    }
    return result, nil
}

func SortByMax(allData []SingleData, order string) ([]SingleData, error) {
    result, sortErr := SortCore(allData, "max", order)
    if sortErr != nil {
        return allData, sortErr
    }
    return result, nil
}

func SortByMin(allData []SingleData, order string) ([]SingleData, error) {
    result, sortErr := SortCore(allData, "min", order)
    if sortErr != nil {
        return allData, sortErr
    }
    return result, nil
}

func SortBySum(allData []SingleData, order string) ([]SingleData, error) {
    result, sortErr := SortCore(allData, "sum", order)
    if sortErr != nil {
        return allData, sortErr
    }
    return result, nil
}

func SortByAbsMax(allData []SingleData, order string) ([]SingleData, error) {
    result, sortErr := SortCore(allData, "absoluteMax", order)
    if sortErr != nil {
        return allData, sortErr
    }
    return result, nil
}

func SortByAbsMin(allData []SingleData, order string) ([]SingleData, error) {
    result, sortErr := SortCore(allData, "absoluteMin", order)
    if sortErr != nil {
        return allData, sortErr
    }
    return result, nil
}








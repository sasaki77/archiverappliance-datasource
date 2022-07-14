package main

import (
	"errors"
	"fmt"
	"regexp"
	"sort"

	"github.com/grafana/grafana-plugin-sdk-go/backend/log"
)

// Utilities

type SingleDataOrder struct {
	sD   *SingleData
	rank float64
}

func FilterIndexer(allData []*SingleData, value string) ([]float64, error) {
	// determine a single value for each SingleData. Useful for sorting or ranking SingleData
	rank := make([]float64, len(allData))
	for idx, sData := range allData {

		values, ok := sData.Values.(*Scalars)
		if !ok {
			continue
		}

		var v float64
		var err error

		switch value {
		case "avg":
			v, err = values.Rank(Avg)
		case "min":
			v, err = values.Rank(Min)
		case "max":
			v, err = values.Rank(Max)
		case "absoluteMin":
			v, err = values.Rank(AbsoluteMin)
		case "absoluteMax":
			v, err = values.Rank(AbsoluteMax)
		case "sum":
			v, err = values.Rank(Sum)
		default:
			errMsg := fmt.Sprintf("Value %v not recognized", value)
			return rank, errors.New(errMsg)
		}

		if err != nil {
			continue
		}
		rank[idx] = v
	}
	return rank, nil
}

func SortCore(allData []*SingleData, value string, order string) ([]*SingleData, error) {
	// Sort allData
	// The order parameter chooses whether the order of the sort is ascending or descending
	// The value parameter determines how the rank of each SingleData entry is measured
	newData := make([]*SingleData, 0, len(allData))
	rank, idxErr := FilterIndexer(allData, value)
	if idxErr != nil {
		return allData, idxErr
	}
	if len(rank) != len(allData) {
		errMsg := fmt.Sprintf("Length of data (%v) and indexes (%v)differ", len(allData), len(rank))
		return allData, errors.New(errMsg)
	}
	ordered := make([]SingleDataOrder, len(allData))
	for idx := range allData {
		ordered[idx] = SingleDataOrder{
			sD:   allData[idx],
			rank: rank[idx],
		}
	}
	if order == "asc" {
		sort.SliceStable(ordered, func(i, j int) bool {
			return ordered[i].rank < ordered[j].rank
		})
	} else if order == "desc" {
		sort.SliceStable(ordered, func(i, j int) bool {
			return ordered[i].rank > ordered[j].rank
		})
	} else {
		errMsg := fmt.Sprintf("Order %v not recognized", order)
		log.DefaultLogger.Warn(errMsg)
		return allData, errors.New(errMsg)
	}

	for idx := range ordered {
		newData = append(newData, ordered[idx].sD)
	}

	return newData, nil
}

// Transform functions

func Scale(allData []*SingleData, factor float64) []*SingleData {
	for _, oneData := range allData {
		values, ok := oneData.Values.(*Scalars)
		if !ok {
			continue
		}
		values.Scale(factor)
	}
	return allData
}

func Offset(allData []*SingleData, delta float64) []*SingleData {
	for _, oneData := range allData {
		values, ok := oneData.Values.(*Scalars)
		if !ok {
			continue
		}
		values.Offset(delta)
	}
	return allData
}

func Delta(allData []*SingleData) []*SingleData {
	for _, oneData := range allData {
		values, ok := oneData.Values.(*Scalars)
		if !ok {
			continue
		}
		values.Delta()
	}
	return allData
}

func Fluctuation(allData []*SingleData) []*SingleData {
	for _, oneData := range allData {
		values, ok := oneData.Values.(*Scalars)
		if !ok {
			continue
		}
		values.Fluctuation()
	}
	return allData
}

func MovingAverage(allData []*SingleData, windowSize int) []*SingleData {
	for _, oneData := range allData {
		values, ok := oneData.Values.(*Scalars)
		if !ok {
			continue
		}
		values.MovingAverage(windowSize)
	}
	return allData
}

// Array to Scalar Functions

// Filter Series Functions

func Top(allData []*SingleData, number int, value string) ([]*SingleData, error) {
	result, sortErr := SortCore(allData, value, "desc")
	if sortErr != nil {
		return allData, sortErr
	}
	if len(result) > number {
		return result[:number], nil
	}
	return result, nil
}

func Bottom(allData []*SingleData, number int, value string) ([]*SingleData, error) {
	result, sortErr := SortCore(allData, value, "asc")
	if sortErr != nil {
		return allData, sortErr
	}
	if len(result) > number {
		return result[:number], nil
	}
	return result, nil
}

func Exclude(allData []*SingleData, pattern string) ([]*SingleData, error) {
	var newData []*SingleData
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
				err = errors.New("unknown panic")
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

func SortByAvg(allData []*SingleData, order string) ([]*SingleData, error) {
	result, sortErr := SortCore(allData, "avg", order)
	if sortErr != nil {
		return allData, sortErr
	}
	return result, nil
}

func SortByMax(allData []*SingleData, order string) ([]*SingleData, error) {
	result, sortErr := SortCore(allData, "max", order)
	if sortErr != nil {
		return allData, sortErr
	}
	return result, nil
}

func SortByMin(allData []*SingleData, order string) ([]*SingleData, error) {
	result, sortErr := SortCore(allData, "min", order)
	if sortErr != nil {
		return allData, sortErr
	}
	return result, nil
}

func SortBySum(allData []*SingleData, order string) ([]*SingleData, error) {
	result, sortErr := SortCore(allData, "sum", order)
	if sortErr != nil {
		return allData, sortErr
	}
	return result, nil
}

func SortByAbsMax(allData []*SingleData, order string) ([]*SingleData, error) {
	result, sortErr := SortCore(allData, "absoluteMax", order)
	if sortErr != nil {
		return allData, sortErr
	}
	return result, nil
}

func SortByAbsMin(allData []*SingleData, order string) ([]*SingleData, error) {
	result, sortErr := SortCore(allData, "absoluteMin", order)
	if sortErr != nil {
		return allData, sortErr
	}
	return result, nil
}

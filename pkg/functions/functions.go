package functions

import (
	"errors"
	"fmt"
	"regexp"
	"sort"

	"github.com/grafana/grafana-plugin-sdk-go/backend/log"
	"github.com/sasaki77/archiverappliance-datasource/pkg/models"
)

// Utilities

type singleDataOrder struct {
	sD   *models.SingleData
	rank float64
}

func filterIndexer(allData []*models.SingleData, value string) ([]float64, error) {
	// determine a single value for each SingleData. Useful for sorting or ranking SingleData
	rank := make([]float64, len(allData))
	for idx, sData := range allData {

		values, ok := sData.Values.(*models.Scalars)
		if !ok {
			continue
		}

		var v float64
		var err error

		switch value {
		case "avg":
			v, err = values.Rank(models.RANKTYPE_AVG)
		case "min":
			v, err = values.Rank(models.RANKTYPE_MIN)
		case "max":
			v, err = values.Rank(models.RANKTYPE_MAX)
		case "absoluteMin":
			v, err = values.Rank(models.RANKTYPE_ABSOLUTEMIN)
		case "absoluteMax":
			v, err = values.Rank(models.RANKTYPE_ABSOLUTEMAX)
		case "sum":
			v, err = values.Rank(models.RANKTYPE_SUM)
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

func sortCore(allData []*models.SingleData, value string, order string) ([]*models.SingleData, error) {
	// Sort allData
	// The order parameter chooses whether the order of the sort is ascending or descending
	// The value parameter determines how the rank of each SingleData entry is measured
	newData := make([]*models.SingleData, 0, len(allData))
	rank, idxErr := filterIndexer(allData, value)
	if idxErr != nil {
		return allData, idxErr
	}
	if len(rank) != len(allData) {
		errMsg := fmt.Sprintf("Length of data (%v) and indexes (%v)differ", len(allData), len(rank))
		return allData, errors.New(errMsg)
	}
	ordered := make([]singleDataOrder, len(allData))
	for idx := range allData {
		ordered[idx] = singleDataOrder{
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

func scale(allData []*models.SingleData, factor float64) []*models.SingleData {
	for _, oneData := range allData {
		values, ok := oneData.Values.(*models.Scalars)
		if !ok {
			continue
		}
		values.Scale(factor)
	}
	return allData
}

func offset(allData []*models.SingleData, delta float64) []*models.SingleData {
	for _, oneData := range allData {
		values, ok := oneData.Values.(*models.Scalars)
		if !ok {
			continue
		}
		values.Offset(delta)
	}
	return allData
}

func delta(allData []*models.SingleData) []*models.SingleData {
	for _, oneData := range allData {
		values, ok := oneData.Values.(*models.Scalars)
		if !ok {
			continue
		}
		values.Delta()
	}
	return allData
}

func fluctuation(allData []*models.SingleData) []*models.SingleData {
	for _, oneData := range allData {
		values, ok := oneData.Values.(*models.Scalars)
		if !ok {
			continue
		}
		values.Fluctuation()
	}
	return allData
}

func movingAverage(allData []*models.SingleData, windowSize int) []*models.SingleData {
	for _, oneData := range allData {
		values, ok := oneData.Values.(*models.Scalars)
		if !ok {
			continue
		}
		values.MovingAverage(windowSize)
	}
	return allData
}

// Array to Scalar Functions

// Filter Series Functions

func top(allData []*models.SingleData, number int, value string) ([]*models.SingleData, error) {
	result, sortErr := sortCore(allData, value, "desc")
	if sortErr != nil {
		return allData, sortErr
	}
	if len(result) > number {
		return result[:number], nil
	}
	return result, nil
}

func bottom(allData []*models.SingleData, number int, value string) ([]*models.SingleData, error) {
	result, sortErr := sortCore(allData, value, "asc")
	if sortErr != nil {
		return allData, sortErr
	}
	if len(result) > number {
		return result[:number], nil
	}
	return result, nil
}

func exclude(allData []*models.SingleData, pattern string) ([]*models.SingleData, error) {
	var newData []*models.SingleData
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

func sortByAvg(allData []*models.SingleData, order string) ([]*models.SingleData, error) {
	result, sortErr := sortCore(allData, "avg", order)
	if sortErr != nil {
		return allData, sortErr
	}
	return result, nil
}

func sortByMax(allData []*models.SingleData, order string) ([]*models.SingleData, error) {
	result, sortErr := sortCore(allData, "max", order)
	if sortErr != nil {
		return allData, sortErr
	}
	return result, nil
}

func sortByMin(allData []*models.SingleData, order string) ([]*models.SingleData, error) {
	result, sortErr := sortCore(allData, "min", order)
	if sortErr != nil {
		return allData, sortErr
	}
	return result, nil
}

func sortBySum(allData []*models.SingleData, order string) ([]*models.SingleData, error) {
	result, sortErr := sortCore(allData, "sum", order)
	if sortErr != nil {
		return allData, sortErr
	}
	return result, nil
}

func sortByAbsMax(allData []*models.SingleData, order string) ([]*models.SingleData, error) {
	result, sortErr := sortCore(allData, "absoluteMax", order)
	if sortErr != nil {
		return allData, sortErr
	}
	return result, nil
}

func sortByAbsMin(allData []*models.SingleData, order string) ([]*models.SingleData, error) {
	result, sortErr := sortCore(allData, "absoluteMin", order)
	if sortErr != nil {
		return allData, sortErr
	}
	return result, nil
}

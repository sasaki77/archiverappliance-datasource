package main

import (
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/grafana/grafana-plugin-sdk-go/data"
)

type Scalars struct {
	Times  []time.Time
	Values []float64
}

func (v *Scalars) ToFields(pvname string, name string) []*data.Field {
	var fields []*data.Field
	//add the time dimension
	fields = append(fields, data.NewField("time", nil, v.Times))

	// add values
	labels := make(data.Labels, 1)
	labels["pvname"] = pvname

	valueField := data.NewField(name, labels, v.Values)
	valueField.Config = &data.FieldConfig{DisplayName: name}
	fields = append(fields, valueField)

	return fields
}

func (v *Scalars) Extrapolation(t time.Time) {
	v.Values = append(v.Values, v.Values[len(v.Values)-1])
	v.Times = append(v.Times, t)
}

func (v *Scalars) Scale(factor float64) {
	for idx, val := range v.Values {
		v.Values[idx] = val * factor
	}
}

func (v *Scalars) Offset(delta float64) {
	for idx, val := range v.Values {
		v.Values[idx] = val + delta
	}
}

func (v *Scalars) Delta() {
	newValues := make([]float64, 0, len(v.Values))
	newTimes := make([]time.Time, 0, len(v.Times))
	for idx := range v.Values {
		if idx == 0 {
			continue
		}
		newValues = append(newValues, v.Values[idx]-v.Values[idx-1])
		newTimes = append(newTimes, v.Times[idx])
	}
	if len(newValues) == 0 {
		// handle 1-length data
		newValues = append(newValues, 0)
		newTimes = append(newTimes, v.Times[0])
	}
	v.Times = newTimes
	v.Values = newValues
}

func (v *Scalars) Fluctuation() {
	var startingValue float64
	for idx, val := range v.Values {
		if idx == 0 {
			startingValue = val
		}
		v.Values[idx] = val - startingValue
	}
}

func (v *Scalars) MovingAverage(windowSize int) {
	newValues := make([]float64, len(v.Values))

	for idx := range v.Values {
		var total float64
		total = 0
		var size float64
		size = 0
		for i := 0; i < windowSize; i++ {
			if (idx - i) < 0 {
				break
			}
			size = size + 1
			total = total + v.Values[idx-i]
		}
		newValues[idx] = total / size
	}

	v.Values = newValues
}

type RankType string

const (
	Avg         = RankType("Avg")
	Min         = RankType("Min")
	Max         = RankType("Max")
	AbsoluteMin = RankType("AbsoluteMin")
	AbsoluteMax = RankType("AbsoluteMax")
	Sum         = RankType("Sum")
)

func (v *Scalars) Rank(rankType RankType) (float64, error) {
	data := v.Values
	switch rankType {
	case Avg:
		var total float64
		for _, val := range data {
			total += val
		}
		return total / float64(len(data)), nil
	case Min:
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
		return low_cache, nil
	case Max:
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
		return high_cache, nil
	case AbsoluteMin:
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
		return low_cache, nil
	case AbsoluteMax:
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
		return high_cache, nil
	case Sum:
		var total float64
		for _, val := range data {
			total += val
		}
		return total, nil
	default:
		errMsg := fmt.Sprintf("Value %s not recognized", rankType)
		return 0, errors.New(errMsg)
	}
}

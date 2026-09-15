package models

import (
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/grafana/grafana-plugin-sdk-go/data"
)

type Scalars struct {
	Times  []time.Time
	Values []*float64
}

func NewSclars(length int) *Scalars {
	return &Scalars{
		Times:  make([]time.Time, 0, length),
		Values: make([]*float64, 0, length),
	}
}

func NewSclarsWithValues(t []time.Time, v []*float64) *Scalars {
	return &Scalars{Times: t, Values: v}
}

func (v *Scalars) Append(val *float64, t time.Time) {
	v.Values = append(v.Values, val)
	v.Times = append(v.Times, t)
}

func (v *Scalars) AppendConcrete(val float64, t time.Time) {
	v.Values = append(v.Values, &val)
	v.Times = append(v.Times, t)
}

// SetValConcrete overwrites a value in place. No two entries share a pointer, so
// this cannot touch a neighbour, and ToFrame runs last -- a frame shares these
// floats rather than copying them, so no transform may run after one is built.
func (v *Scalars) SetValConcrete(idx int, val float64) {
	if p := v.Values[idx]; p != nil {
		*p = val
		return
	}

	// Addressing the parameter would make it escape on every call, including the
	// path above that does not need it.
	nv := val
	v.Values[idx] = &nv
}

// defaultValueBlock is the block size used when the count is not known up front.
const defaultValueBlock = 4096

// valueBlock hands out pointers into slices it owns. A full block is replaced
// rather than grown, which is what keeps the pointers already handed out valid;
// nothing frees a block, it lives as long as a pointer into it. The zero value
// is usable.
type valueBlock struct {
	block []float64

	// size is the next block's capacity, 0 for defaultValueBlock. A caller that
	// knows its length passes it, so the whole series fits one block.
	size int
}

func newValueBlock(size int) *valueBlock {
	return &valueBlock{size: size}
}

func (b *valueBlock) add(val float64) *float64 {
	if len(b.block) == cap(b.block) {
		size := b.size
		if size <= 0 {
			size = defaultValueBlock
		}
		b.block = make([]float64, 0, size)
	}
	b.block = append(b.block, val)
	return &b.block[len(b.block)-1]
}

func (v *Scalars) ToFields(pvname string, name string, format FormatOption) []*data.Field {
	// ToFields doesn't use FormatOption in Scalars for now

	var fields []*data.Field

	//add the time dimension
	fields = append(fields, data.NewField("time", nil, v.Times))

	// add values
	labels := make(data.Labels, 1)
	labels["pvname"] = pvname

	valueField := data.NewField(name, labels, v.Values)
	valueField.Config = &data.FieldConfig{DisplayNameFromDS: name}
	fields = append(fields, valueField)

	return fields
}

func (v *Scalars) Extrapolation(t time.Time) {
	if len(v.Values) == 0 {
		return
	}

	var val *float64
	for i := len(v.Values) - 1; i >= 0; i-- {
		vi := v.Values[i]
		if vi != nil {
			val = vi
			break
		}
	}

	if val == nil {
		return
	}

	// Copy rather than repeat the pointer: transforms write through these, so two
	// entries sharing one would alias.
	v.AppendConcrete(*val, t)
}

func (v *Scalars) Scale(factor float64) {
	for idx, val := range v.Values {
		if val == nil {
			continue
		}
		v.SetValConcrete(idx, *val*factor)
	}
}

func (v *Scalars) Offset(delta float64) {
	for idx, val := range v.Values {
		if val == nil {
			continue
		}
		v.SetValConcrete(idx, *val+delta)
	}
}

func (v *Scalars) Delta() {
	newValues := make([]*float64, 0, len(v.Values))
	newTimes := make([]time.Time, 0, len(v.Times))
	// Sized for the worst case, so one block holds the whole result.
	block := newValueBlock(len(v.Values))
	for idx, val := range v.Values {
		if idx == 0 {
			continue
		}

		if val == nil || v.Values[idx-1] == nil {
			continue
		}

		newValues = append(newValues, block.add(*v.Values[idx]-*v.Values[idx-1]))
		newTimes = append(newTimes, v.Times[idx])
	}
	if len(newValues) == 0 {
		// handle 1-length data
		newValues = append(newValues, block.add(0))
		newTimes = append(newTimes, v.Times[0])
	}
	v.Times = newTimes
	v.Values = newValues
}

func (v *Scalars) Fluctuation() {
	var startingValue float64
	isInited := false
	for idx, val := range v.Values {
		if val == nil {
			continue
		}

		if !isInited {
			startingValue = *val
			isInited = true
		}

		v.SetValConcrete(idx, *val-startingValue)
	}
}

func (v *Scalars) MovingAverage(windowSize int) {
	newValues := make([]*float64, len(v.Values))
	block := newValueBlock(len(v.Values))

	for idx := range v.Values {
		if v.Values[idx] == nil {
			continue
		}

		var total float64
		total = 0
		var size float64
		size = 0

		for i := 0; i < windowSize; i++ {
			if (idx - i) < 0 {
				break
			}
			if v.Values[idx-i] == nil {
				continue
			}
			size = size + 1
			total = total + *v.Values[idx-i]
		}
		newValues[idx] = block.add(total / size)
	}

	v.Values = newValues
}

type RankType string

const (
	RANKTYPE_AVG         = RankType("Avg")
	RANKTYPE_MIN         = RankType("Min")
	RANKTYPE_MAX         = RankType("Max")
	RANKTYPE_ABSOLUTEMIN = RankType("AbsoluteMin")
	RANKTYPE_ABSOLUTEMAX = RankType("AbsoluteMax")
	RANKTYPE_SUM         = RankType("Sum")
)

func (v *Scalars) Rank(rankType RankType) (float64, error) {
	data := v.Values
	switch rankType {
	case RANKTYPE_AVG:
		var total float64
		var l int = 0
		for _, val := range data {
			if val == nil {
				continue
			}
			l++
			total += *val
		}
		return total / float64(l), nil
	case RANKTYPE_MIN:
		var low_cache float64
		first_run := true
		for _, val := range data {
			if val == nil {
				continue
			}
			if first_run {
				low_cache = *val
				first_run = false
			}
			if low_cache > *val {
				low_cache = *val
			}
		}
		return low_cache, nil
	case RANKTYPE_MAX:
		var high_cache float64
		first_run := true
		for _, val := range data {
			if val == nil {
				continue
			}
			if first_run {
				high_cache = *val
				first_run = false
			}
			if high_cache < *val {
				high_cache = *val
			}
		}
		return high_cache, nil
	case RANKTYPE_ABSOLUTEMIN:
		var low_cache float64
		first_run := true
		for _, originalVal := range data {
			if originalVal == nil {
				continue
			}
			val := math.Abs(*originalVal)
			if first_run {
				low_cache = val
				first_run = false
			}
			if low_cache > val {
				low_cache = val
			}
		}
		return low_cache, nil
	case RANKTYPE_ABSOLUTEMAX:
		var high_cache float64
		first_run := true
		for _, originalVal := range data {
			if originalVal == nil {
				continue
			}
			val := math.Abs(*originalVal)
			if first_run {
				high_cache = val
				first_run = false
			}
			if high_cache < val {
				high_cache = val
			}
		}
		return high_cache, nil
	case RANKTYPE_SUM:
		var total float64
		for _, val := range data {
			if val == nil {
				continue
			}
			total += *val
		}
		return total, nil
	default:
		errMsg := fmt.Sprintf("Value %s not recognized", rankType)
		return 0, errors.New(errMsg)
	}
}

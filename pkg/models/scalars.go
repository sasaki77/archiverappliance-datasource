package models

import (
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/grafana/grafana-plugin-sdk-go/data"
)

// Scalars holds one PV's scalar samples. Values carries a pointer per sample
// because that is what the Grafana SDK takes for a nullable field, where a nil
// draws a gap, so the type cannot simply become []float64. The arena supplies
// what those pointers point at, a block at a time, which brings a raw query down
// to one allocation per 4096 samples without changing the type or any caller.
//
// Three things follow from splitting the pointers from the storage:
//
//   - The arena backs only the values this container made itself, through
//     AppendConcrete. Entries that arrived already as pointers, from Append or
//     NewSclarsWithValues, point somewhere else entirely.
//   - Values is what keeps the blocks alive. The arena only remembers the block
//     it is currently filling; earlier ones survive because Values still points
//     into them, and are collected once it no longer does. Nothing frees them.
//   - The transforms write through the pointers in Values, straight into that
//     storage, which is how they rewrite a series without allocating.
type Scalars struct {
	Times  []time.Time
	Values []*float64
	arena  valueBlock
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
	v.Values = append(v.Values, v.arena.add(val))
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
	if len(v.Values) == 0 || len(v.Times) == 0 {
		return
	}

	// A single sample has nothing to be differenced against.
	if len(v.Values) < 2 {
		v.SetValConcrete(0, 0)
		v.Values = v.Values[:1]
		v.Times = v.Times[:1]
		return
	}

	var previous float64
	previousHasValue := v.Values[0] != nil
	if previousHasValue {
		previous = *v.Values[0]
	}

	for idx := 1; idx < len(v.Values); idx++ {
		p := v.Values[idx]

		if p == nil || !previousHasValue {
			v.Values[idx-1] = nil
		} else {
			v.SetValConcrete(idx-1, *p-previous)
		}
		v.Times[idx-1] = v.Times[idx]

		previousHasValue = p != nil
		if previousHasValue {
			previous = *p
		}
	}

	v.Values = v.Values[:len(v.Values)-1]
	v.Times = v.Times[:len(v.Times)-1]
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

type windowSample struct {
	value   float64
	present bool
}

type sampleRing struct {
	samples []windowSample
	next    int
}

// push stores a sample and returns the one it displaced, which is the sample
// leaving the window.
func (r *sampleRing) push(p *float64) windowSample {
	leaving := r.samples[r.next]

	if p == nil {
		r.samples[r.next] = windowSample{}
	} else {
		r.samples[r.next] = windowSample{value: *p, present: true}
	}

	r.next++
	if r.next == len(r.samples) {
		r.next = 0
	}

	return leaving
}

func (v *Scalars) MovingAverage(windowSize int) {
	if windowSize < 2 || len(v.Values) == 0 {
		return
	}

	// total and count stay here rather than in the ring so they can live in
	// registers: holding them as struct fields measured 30% slower.
	//
	// Carrying a running total also rounds differently from summing each window
	// on its own, so results can differ in the last digits of a float64.
	ring := sampleRing{samples: make([]windowSample, min(windowSize, len(v.Values)))}
	var total float64
	count := 0

	for _, p := range v.Values {
		leaving := ring.push(p) // reads the sample before the loop overwrites it
		if leaving.present {
			total -= leaving.value
			count--
		}

		if p == nil {
			continue
		}

		total += *p
		count++

		*p = total / float64(count) // count covers at least the sample just pushed
	}
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

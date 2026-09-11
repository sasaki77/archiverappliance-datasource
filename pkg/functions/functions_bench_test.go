package functions

import (
	"fmt"
	"testing"
	"time"

	"github.com/sasaki77/archiverappliance-datasource/pkg/models"
)

// Benchmarks for the processing-function pipeline.
//
// Each function in a chain makes an independent full pass over every series,
// so both the number of functions and the number of points matter. The sort
// and top/bottom functions additionally rank every point of every series.

const (
	benchSeriesPoints = 50000
	benchSeriesCount  = 200
)

func makeBenchSingleData(series int, points int) []*models.SingleData {
	base := time.Date(2021, time.January, 10, 0, 0, 0, 0, time.UTC)
	data := make([]*models.SingleData, series)
	for s := range data {
		v := models.NewSclars(points)
		for i := 0; i < points; i++ {
			v.AppendConcrete(float64((s*7+i)%1000), base.Add(time.Duration(i)*time.Millisecond))
		}
		name := fmt.Sprintf("TEST:PV:%03d", s)
		data[s] = &models.SingleData{Name: name, PVname: name, Values: v}
	}
	return data
}

func transformFunc(name string, paramName string, paramType string, value string) models.FunctionDescriptorQueryModel {
	return models.FunctionDescriptorQueryModel{
		Def: models.FuncDefQueryModel{
			Category: models.FUNC_CATEGORY_TRANSFORM,
			Name:     name,
			Params:   []models.FuncDefParamQueryModel{{Name: paramName, Type: paramType}},
		},
		Params: []string{value},
	}
}

func sortFunc(name string, number string, value string) models.FunctionDescriptorQueryModel {
	return models.FunctionDescriptorQueryModel{
		Def: models.FuncDefQueryModel{
			Category: models.FUNC_CATEGORY_FILTER,
			Name:     name,
			Params: []models.FuncDefParamQueryModel{
				{Name: "number", Type: "int"},
				{Name: "value", Type: "string"},
			},
		},
		Params: []string{number, value},
	}
}

// A single transform over a single long series: the baseline per-point cost.
func BenchmarkApplyFunctionsScale(b *testing.B) {
	qm := models.ArchiverQueryModel{
		Functions: []models.FunctionDescriptorQueryModel{
			transformFunc("scale", "factor", "float", "1"),
		},
	}
	data := makeBenchSingleData(1, benchSeriesPoints*4)

	b.ReportAllocs()
	for b.Loop() {
		if _, err := ApplyFunctions(data, qm); err != nil {
			b.Fatal(err)
		}
	}
}

// A chain of three transforms: three independent full passes over the data.
func BenchmarkApplyFunctionsChain(b *testing.B) {
	qm := models.ArchiverQueryModel{
		Functions: []models.FunctionDescriptorQueryModel{
			transformFunc("scale", "factor", "float", "1"),
			transformFunc("offset", "delta", "float", "0"),
			transformFunc("movingAverage", "windowSize", "int", "10"),
		},
	}
	data := makeBenchSingleData(1, benchSeriesPoints*4)

	b.ReportAllocs()
	for b.Loop() {
		if _, err := ApplyFunctions(data, qm); err != nil {
			b.Fatal(err)
		}
	}
}

// top ranks every point of every series and then sorts all of them, even
// though only `number` series survive.
func BenchmarkApplyFunctionsTop(b *testing.B) {
	qm := models.ArchiverQueryModel{
		Functions: []models.FunctionDescriptorQueryModel{
			sortFunc("top", "5", "avg"),
		},
	}
	data := makeBenchSingleData(benchSeriesCount, benchSeriesPoints)

	b.ReportAllocs()
	for b.Loop() {
		if _, err := ApplyFunctions(data, qm); err != nil {
			b.Fatal(err)
		}
	}
}

// sortByAvg followed by top: the ranking pass runs twice over the same data.
func BenchmarkApplyFunctionsSortThenTop(b *testing.B) {
	qm := models.ArchiverQueryModel{
		Functions: []models.FunctionDescriptorQueryModel{
			{
				Def: models.FuncDefQueryModel{
					Category: models.FUNC_CATEGORY_SORT,
					Name:     "sortByAvg",
					Params:   []models.FuncDefParamQueryModel{{Name: "order", Type: "string"}},
				},
				Params: []string{"desc"},
			},
			sortFunc("top", "5", "avg"),
		},
	}
	data := makeBenchSingleData(benchSeriesCount, benchSeriesPoints)

	b.ReportAllocs()
	for b.Loop() {
		if _, err := ApplyFunctions(data, qm); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkSortCore(b *testing.B) {
	data := makeBenchSingleData(benchSeriesCount, benchSeriesPoints)

	b.ReportAllocs()
	for b.Loop() {
		if _, err := sortCore(data, "avg", "desc"); err != nil {
			b.Fatal(err)
		}
	}
}

// toScalarByMed copies and sorts the whole waveform for every timestamp.
func BenchmarkArrayToScalar(b *testing.B) {
	base := time.Date(2021, time.January, 10, 0, 0, 0, 0, time.UTC)
	for _, name := range []string{"toScalarByAvg", "toScalarByMed", "toScalarByMax"} {
		b.Run(name, func(b *testing.B) {
			v := models.NewArrays(2000)
			for i := 0; i < 2000; i++ {
				row := make([]float64, 100)
				for j := range row {
					row[j] = float64((i + j) % 1000)
				}
				v.Append(row, base.Add(time.Duration(i)*time.Millisecond))
			}
			data := []*models.SingleData{{Name: "TEST:PV:NAME", PVname: "TEST:PV:NAME", Values: v}}
			qm := models.ArchiverQueryModel{
				Functions: []models.FunctionDescriptorQueryModel{
					{
						Def: models.FuncDefQueryModel{
							Category: models.FUNC_CATEGORY_TOSCALAR,
							Name:     name,
						},
					},
				},
			}

			b.ReportAllocs()
			for b.Loop() {
				if _, err := ApplyFunctions(data, qm); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

package functions

import (
	"fmt"
	"testing"
	"time"

	"github.com/sasaki77/archiverappliance-datasource/pkg/models"
	"github.com/sasaki77/archiverappliance-datasource/pkg/testhelper"
)

// Utilites

// Transform funcitons

func TestScale(t *testing.T) {
	var tests = []struct {
		inputSd []*models.SingleData
		delta   float64
		output  []*models.SingleData
	}{
		{
			inputSd: []*models.SingleData{
				{
					Name: "TEST:PV:NAME",
					Values: &models.Scalars{
						Times:  testhelper.TimeArrayHelper(0, 6),
						Values: append([]*float64{nil}, testhelper.InitFloat64SlicePointer([]float64{1})...),
					},
				},
			},
			delta: 2,
			output: []*models.SingleData{
				{
					Name: "TEST:PV:NAME",
					Values: &models.Scalars{
						Times:  testhelper.TimeArrayHelper(0, 6),
						Values: append([]*float64{nil}, testhelper.InitFloat64SlicePointer([]float64{2})...),
					},
				},
			},
		},
	}
	for tdx, testCase := range tests {
		testName := fmt.Sprintf("case %d: %v", tdx, testCase.output)
		t.Run(testName, func(t *testing.T) {
			result := scale(testCase.inputSd, testCase.delta)
			models.SingleDataCompareHelper(result, testCase.output, t)
		})
	}
}

func TestOffset(t *testing.T) {
	var tests = []struct {
		inputSd []*models.SingleData
		delta   float64
		output  []*models.SingleData
	}{
		{
			inputSd: []*models.SingleData{
				{
					Name: "TEST:PV:NAME",
					Values: &models.Scalars{
						Times:  testhelper.TimeArrayHelper(0, 6),
						Values: append([]*float64{nil}, testhelper.InitFloat64SlicePointer([]float64{1})...),
					},
				},
			},
			delta: 2,
			output: []*models.SingleData{
				{
					Name: "TEST:PV:NAME",
					Values: &models.Scalars{
						Times:  testhelper.TimeArrayHelper(0, 6),
						Values: append([]*float64{nil}, testhelper.InitFloat64SlicePointer([]float64{3})...),
					},
				},
			},
		},
	}
	for tdx, testCase := range tests {
		testName := fmt.Sprintf("case %d: %v", tdx, testCase.output)
		t.Run(testName, func(t *testing.T) {
			result := offset(testCase.inputSd, testCase.delta)
			models.SingleDataCompareHelper(result, testCase.output, t)
		})
	}
}

func TestDelta(t *testing.T) {
	var tests = []struct {
		inputSd []*models.SingleData
		output  []*models.SingleData
	}{
		// A nil leaves the pair it belongs to without a difference, but the
		// position stays and carries a nil, so the panel draws a gap there
		// rather than a line straight across it.
		{
			inputSd: []*models.SingleData{
				{
					Name: "TEST:PV:NAME",
					Values: &models.Scalars{
						Times:  []time.Time{testhelper.TimeHelper(0), testhelper.TimeHelper(1), testhelper.TimeHelper(2)},
						Values: append([]*float64{nil}, testhelper.InitFloat64SlicePointer([]float64{1, 1})...),
					},
				},
			},
			output: []*models.SingleData{
				{
					Name: "TEST:PV:NAME",
					Values: &models.Scalars{
						Times:  []time.Time{testhelper.TimeHelper(1), testhelper.TimeHelper(2)},
						Values: append([]*float64{nil}, testhelper.InitFloat64SlicePointer([]float64{0})...),
					},
				},
			},
		},
	}
	for tdx, testCase := range tests {
		testName := fmt.Sprintf("case: %d", tdx)
		t.Run(testName, func(t *testing.T) {
			result := delta(testCase.inputSd)
			models.SingleDataCompareHelper(result, testCase.output, t)
		})
	}
}

func TestFluctuation(t *testing.T) {
	var tests = []struct {
		inputSd []*models.SingleData
		output  []*models.SingleData
	}{
		{
			inputSd: []*models.SingleData{
				{
					Name: "TEST:PV:NAME",
					Values: &models.Scalars{
						Times:  testhelper.TimeArrayHelper(0, 2),
						Values: append([]*float64{nil}, testhelper.InitFloat64SlicePointer([]float64{1})...),
					},
				},
			},
			output: []*models.SingleData{
				{
					Name: "TEST:PV:NAME",
					Values: &models.Scalars{
						Times:  testhelper.TimeArrayHelper(0, 2),
						Values: append([]*float64{nil}, testhelper.InitFloat64SlicePointer([]float64{0})...),
					},
				},
			},
		},
	}
	for tdx, testCase := range tests {
		testName := fmt.Sprintf("case: %d", tdx)
		t.Run(testName, func(t *testing.T) {
			result := fluctuation(testCase.inputSd)
			models.SingleDataCompareHelper(result, testCase.output, t)
		})
	}
}

func TestMovingAverage(t *testing.T) {
	var tests = []struct {
		inputSd    []*models.SingleData
		windowSize int
		output     []*models.SingleData
	}{
		{
			inputSd: []*models.SingleData{
				{
					Name: "TEST:PV:NAME",
					Values: &models.Scalars{
						Times:  testhelper.TimeArrayHelper(0, 7),
						Values: append([]*float64{nil}, testhelper.InitFloat64SlicePointer([]float64{3, 9, 6, 3, 3, 6})...),
					},
				},
			},
			windowSize: 3,
			output: []*models.SingleData{
				{
					Name: "TEST:PV:NAME",
					Values: &models.Scalars{
						Times:  testhelper.TimeArrayHelper(0, 7),
						Values: append([]*float64{nil}, testhelper.InitFloat64SlicePointer([]float64{3, 6, 6, 6, 4, 4})...),
					},
				},
			},
		},
	}
	for tdx, testCase := range tests {
		testName := fmt.Sprintf("case: %d", tdx)
		t.Run(testName, func(t *testing.T) {
			result := movingAverage(testCase.inputSd, testCase.windowSize)
			models.SingleDataCompareHelper(result, testCase.output, t)
		})
	}
}

// The series coming back untouched is covered by testdata/functions.
func TestExcludeReportsAPatternThatDoesNotCompile(t *testing.T) {
	inputSd := []*models.SingleData{{Name: "PV:A", Values: &models.Scalars{}}}

	if _, err := exclude(inputSd, "***Hello"); err == nil {
		t.Error("expected an error")
	}
}

// Sort Functions

func TestSortByAvg(t *testing.T) {
	var tests = []struct {
		inputSd []*models.SingleData
		order   string
		output  []*models.SingleData
	}{
		{
			inputSd: []*models.SingleData{
				{
					Values: &models.Scalars{
						Times: testhelper.TimeArrayHelper(0, 7),

						Values: append([]*float64{nil}, testhelper.InitFloat64SlicePointer([]float64{10, 10, 20, 30, 50, 80})...),
					},
				},
				{
					Values: &models.Scalars{
						Times:  testhelper.TimeArrayHelper(0, 7),
						Values: append([]*float64{nil}, testhelper.InitFloat64SlicePointer([]float64{1, 1, 2, 3, 5, 8})...),
					},
				},
			},
			order: "asc",
			output: []*models.SingleData{
				{
					Values: &models.Scalars{
						Times:  testhelper.TimeArrayHelper(0, 7),
						Values: append([]*float64{nil}, testhelper.InitFloat64SlicePointer([]float64{1, 1, 2, 3, 5, 8})...),
					},
				},
				{
					Values: &models.Scalars{
						Times:  testhelper.TimeArrayHelper(0, 7),
						Values: append([]*float64{nil}, testhelper.InitFloat64SlicePointer([]float64{10, 10, 20, 30, 50, 80})...),
					},
				},
			},
		},
	}
	for tdx, testCase := range tests {
		testName := fmt.Sprintf("case %d: %v", tdx, testCase.order)
		t.Run(testName, func(t *testing.T) {
			result, err := sortByAvg(testCase.inputSd, testCase.order)
			if err != nil {
				t.Errorf("Error not expected %v", err)
			}
			models.SingleDataCompareHelper(result, testCase.output, t)
		})
	}
}

func TestSortByMax(t *testing.T) {
	var tests = []struct {
		inputSd []*models.SingleData
		order   string
		output  []*models.SingleData
	}{
		{
			inputSd: []*models.SingleData{
				{
					Values: &models.Scalars{
						Times:  testhelper.TimeArrayHelper(0, 7),
						Values: append([]*float64{nil}, testhelper.InitFloat64SlicePointer([]float64{10, 10, 20, 30, 50, 80})...),
					},
				},
				{
					Values: &models.Scalars{
						Times:  testhelper.TimeArrayHelper(0, 7),
						Values: append([]*float64{nil}, testhelper.InitFloat64SlicePointer([]float64{1, 1, 2, 3, 5, 81})...),
					},
				},
			},
			order: "desc",
			output: []*models.SingleData{
				{
					Values: &models.Scalars{
						Times:  testhelper.TimeArrayHelper(0, 7),
						Values: append([]*float64{nil}, testhelper.InitFloat64SlicePointer([]float64{1, 1, 2, 3, 5, 81})...),
					},
				},
				{
					Values: &models.Scalars{
						Times:  testhelper.TimeArrayHelper(0, 7),
						Values: append([]*float64{nil}, testhelper.InitFloat64SlicePointer([]float64{10, 10, 20, 30, 50, 80})...),
					},
				},
			},
		},
	}
	for tdx, testCase := range tests {
		testName := fmt.Sprintf("case %d: %v", tdx, testCase.order)
		t.Run(testName, func(t *testing.T) {
			result, err := sortByMax(testCase.inputSd, testCase.order)
			if err != nil {
				t.Errorf("Error not expected %v", err)
			}
			models.SingleDataCompareHelper(result, testCase.output, t)
		})
	}
}

func TestSortByMin(t *testing.T) {
	var tests = []struct {
		inputSd []*models.SingleData
		order   string
		output  []*models.SingleData
	}{
		{
			inputSd: []*models.SingleData{
				{
					Values: &models.Scalars{
						Times:  testhelper.TimeArrayHelper(0, 7),
						Values: append([]*float64{nil}, testhelper.InitFloat64SlicePointer([]float64{10, 10, 20, 30, 50, 80})...),
					},
				},
				{
					Values: &models.Scalars{
						Times:  testhelper.TimeArrayHelper(0, 7),
						Values: append([]*float64{nil}, testhelper.InitFloat64SlicePointer([]float64{-100, 1, 2, 3, 5, 8})...),
					},
				},
			},
			order: "asc",
			output: []*models.SingleData{
				{
					Values: &models.Scalars{
						Times:  testhelper.TimeArrayHelper(0, 7),
						Values: append([]*float64{nil}, testhelper.InitFloat64SlicePointer([]float64{-100, 1, 2, 3, 5, 8})...),
					},
				},
				{
					Values: &models.Scalars{
						Times:  testhelper.TimeArrayHelper(0, 7),
						Values: append([]*float64{nil}, testhelper.InitFloat64SlicePointer([]float64{10, 10, 20, 30, 50, 80})...),
					},
				},
			},
		},
	}
	for tdx, testCase := range tests {
		testName := fmt.Sprintf("case %d: %v", tdx, testCase.order)
		t.Run(testName, func(t *testing.T) {
			result, err := sortByMin(testCase.inputSd, testCase.order)
			if err != nil {
				t.Errorf("Error not expected %v", err)
			}
			models.SingleDataCompareHelper(result, testCase.output, t)
		})
	}
}

func TestSortBySum(t *testing.T) {
	var tests = []struct {
		inputSd []*models.SingleData
		order   string
		output  []*models.SingleData
	}{
		{
			inputSd: []*models.SingleData{
				{
					Values: &models.Scalars{
						Times:  testhelper.TimeArrayHelper(0, 7),
						Values: append([]*float64{nil}, testhelper.InitFloat64SlicePointer([]float64{10, 10, 20, 30, 50, 80})...),
					},
				},
				{
					Values: &models.Scalars{
						Times:  testhelper.TimeArrayHelper(0, 7),
						Values: append([]*float64{nil}, testhelper.InitFloat64SlicePointer([]float64{1, 1, 2, 3, 5, 8000})...),
					},
				},
			},
			order: "desc",
			output: []*models.SingleData{
				{
					Values: &models.Scalars{
						Times:  testhelper.TimeArrayHelper(0, 7),
						Values: append([]*float64{nil}, testhelper.InitFloat64SlicePointer([]float64{1, 1, 2, 3, 5, 8000})...),
					},
				},
				{
					Values: &models.Scalars{
						Times:  testhelper.TimeArrayHelper(0, 7),
						Values: append([]*float64{nil}, testhelper.InitFloat64SlicePointer([]float64{10, 10, 20, 30, 50, 80})...),
					},
				},
			},
		},
	}
	for tdx, testCase := range tests {
		testName := fmt.Sprintf("case %d: %v", tdx, testCase.order)
		t.Run(testName, func(t *testing.T) {
			result, err := sortBySum(testCase.inputSd, testCase.order)
			if err != nil {
				t.Errorf("Error not expected %v", err)
			}
			models.SingleDataCompareHelper(result, testCase.output, t)
		})
	}
}

func TestSortByAbsMax(t *testing.T) {
	var tests = []struct {
		inputSd []*models.SingleData
		order   string
		output  []*models.SingleData
	}{
		{
			inputSd: []*models.SingleData{
				{
					Values: &models.Scalars{
						Times:  testhelper.TimeArrayHelper(0, 7),
						Values: append([]*float64{nil}, testhelper.InitFloat64SlicePointer([]float64{-10, -10, -20, -30, -50, -80})...),
					},
				},
				{
					Values: &models.Scalars{
						Times:  testhelper.TimeArrayHelper(0, 7),
						Values: append([]*float64{nil}, testhelper.InitFloat64SlicePointer([]float64{1, 1, 2, 3, 5, 8})...),
					},
				},
			},
			order: "asc",
			output: []*models.SingleData{
				{
					Values: &models.Scalars{
						Times:  testhelper.TimeArrayHelper(0, 7),
						Values: append([]*float64{nil}, testhelper.InitFloat64SlicePointer([]float64{1, 1, 2, 3, 5, 8})...),
					},
				},
				{
					Values: &models.Scalars{
						Times:  testhelper.TimeArrayHelper(0, 7),
						Values: append([]*float64{nil}, testhelper.InitFloat64SlicePointer([]float64{-10, -10, -20, -30, -50, -80})...),
					},
				},
			},
		},
	}
	for tdx, testCase := range tests {
		testName := fmt.Sprintf("case %d: %v", tdx, testCase.order)
		t.Run(testName, func(t *testing.T) {
			result, err := sortByAbsMax(testCase.inputSd, testCase.order)
			if err != nil {
				t.Errorf("Error not expected %v", err)
			}
			models.SingleDataCompareHelper(result, testCase.output, t)
		})
	}
}

func TestSorByAbsMin(t *testing.T) {
	var tests = []struct {
		inputSd []*models.SingleData
		order   string
		output  []*models.SingleData
	}{
		{
			inputSd: []*models.SingleData{
				{
					Values: &models.Scalars{
						Times:  testhelper.TimeArrayHelper(0, 7),
						Values: append([]*float64{nil}, testhelper.InitFloat64SlicePointer([]float64{-10, 10, 20, 30, 50, 80})...),
					},
				},
				{
					Values: &models.Scalars{
						Times:  testhelper.TimeArrayHelper(0, 7),
						Values: append([]*float64{nil}, testhelper.InitFloat64SlicePointer([]float64{1, 1, 2, 3, 5, 8})...),
					},
				},
			},
			order: "asc",
			output: []*models.SingleData{
				{
					Values: &models.Scalars{
						Times:  testhelper.TimeArrayHelper(0, 7),
						Values: append([]*float64{nil}, testhelper.InitFloat64SlicePointer([]float64{1, 1, 2, 3, 5, 8})...),
					},
				},
				{
					Values: &models.Scalars{
						Times:  testhelper.TimeArrayHelper(0, 7),
						Values: append([]*float64{nil}, testhelper.InitFloat64SlicePointer([]float64{-10, 10, 20, 30, 50, 80})...),
					},
				},
			},
		},
	}
	for tdx, testCase := range tests {
		testName := fmt.Sprintf("case %d: %v", tdx, testCase.order)
		t.Run(testName, func(t *testing.T) {
			result, err := sortByAbsMin(testCase.inputSd, testCase.order)
			if err != nil {
				t.Errorf("Error not expected %v", err)
			}
			models.SingleDataCompareHelper(result, testCase.output, t)
		})
	}
}

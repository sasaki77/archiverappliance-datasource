package functions

import (
	"fmt"
	"testing"

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
						Values: []float64{1, 1, 2, 3, 5, 8},
					},
				},
			},
			delta: 2,
			output: []*models.SingleData{
				{
					Name: "TEST:PV:NAME",
					Values: &models.Scalars{
						Times:  testhelper.TimeArrayHelper(0, 6),
						Values: []float64{2, 2, 4, 6, 10, 16},
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
						Values: []float64{1, 1, 2, 3, 5, 8},
					},
				},
			},
			delta: 2,
			output: []*models.SingleData{
				{
					Name: "TEST:PV:NAME",
					Values: &models.Scalars{
						Times:  testhelper.TimeArrayHelper(0, 6),
						Values: []float64{3, 3, 4, 5, 7, 10},
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
		{
			inputSd: []*models.SingleData{
				{
					Name: "TEST:PV:NAME",
					Values: &models.Scalars{
						Times:  testhelper.TimeArrayHelper(0, 6),
						Values: []float64{1, 1, 2, 3, 5, 8},
					},
				},
			},
			output: []*models.SingleData{
				{
					Name: "TEST:PV:NAME",
					Values: &models.Scalars{
						Times:  testhelper.TimeArrayHelper(1, 6),
						Values: []float64{0, 1, 1, 2, 3},
					},
				},
			},
		},
		{
			inputSd: []*models.SingleData{
				{
					Name: "TEST:PV:NAME",
					Values: &models.Scalars{
						Times:  testhelper.TimeArrayHelper(0, 1),
						Values: []float64{2},
					},
				},
			},
			output: []*models.SingleData{
				{
					Name: "TEST:PV:NAME",
					Values: &models.Scalars{
						Times:  testhelper.TimeArrayHelper(0, 1),
						Values: []float64{0},
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
						Times:  testhelper.TimeArrayHelper(0, 6),
						Values: []float64{1, 1, 2, 3, 5, 8},
					},
				},
			},
			output: []*models.SingleData{
				{
					Name: "TEST:PV:NAME",
					Values: &models.Scalars{
						Times:  testhelper.TimeArrayHelper(0, 6),
						Values: []float64{0, 0, 1, 2, 4, 7},
					},
				},
			},
		},
		{
			inputSd: []*models.SingleData{
				{
					Name: "TEST:PV:NAME",
					Values: &models.Scalars{
						Times:  testhelper.TimeArrayHelper(0, 1),
						Values: []float64{1},
					},
				},
			},
			output: []*models.SingleData{
				{
					Name: "TEST:PV:NAME",
					Values: &models.Scalars{
						Times:  testhelper.TimeArrayHelper(0, 1),
						Values: []float64{0},
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
						Times:  testhelper.TimeArrayHelper(0, 6),
						Values: []float64{3, 9, 6, 3, 3, 6},
					},
				},
			},
			windowSize: 3,
			output: []*models.SingleData{
				{
					Name: "TEST:PV:NAME",
					Values: &models.Scalars{
						Times:  testhelper.TimeArrayHelper(0, 6),
						Values: []float64{3, 6, 6, 6, 4, 4},
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

// Array to Scalar Functions

func TestToScalarByAvg(t *testing.T) {
	t.Skipf("Not Implemeneted")
}

func TestToScalarByMax(t *testing.T) {
	t.Skipf("Not Implemeneted")
}

func TestToScalarByMin(t *testing.T) {
	t.Skipf("Not Implemeneted")
}

func TestToScalarBySum(t *testing.T) {
	t.Skipf("Not Implemeneted")
}

func TestToScalarByMed(t *testing.T) {
	t.Skipf("Not Implemeneted")
}

func TestToScalarByStd(t *testing.T) {
	t.Skipf("Not Implemeneted")
}

// Filter Series Functions

func TestTop(t *testing.T) {
	var tests = []struct {
		inputSd []*models.SingleData
		number  int
		value   string
		output  []*models.SingleData
	}{
		{
			inputSd: []*models.SingleData{
				{
					Name: "TEST:PV:NAME1",
					Values: &models.Scalars{
						Times:  testhelper.TimeArrayHelper(0, 6),
						Values: []float64{1, 1, 2, 3, 5, 8},
					},
				},
				{
					Name: "TEST:PV:NAME2",
					Values: &models.Scalars{
						Times:  testhelper.TimeArrayHelper(0, 6),
						Values: []float64{10, 10, 20, 30, 50, 80},
					},
				},
			},
			number: 1,
			value:  "avg",
			output: []*models.SingleData{
				{
					Name: "TEST:PV:NAME2",
					Values: &models.Scalars{
						Times:  testhelper.TimeArrayHelper(0, 6),
						Values: []float64{10, 10, 20, 30, 50, 80},
					},
				},
			},
		},
		{
			inputSd: []*models.SingleData{
				{
					Name: "TEST:PV:NAME1",
					Values: &models.Scalars{
						Times:  testhelper.TimeArrayHelper(0, 6),
						Values: []float64{1, 1, 2, 3, 5, 8},
					},
				},
				{
					Name: "TEST:PV:NAME2",
					Values: &models.Scalars{
						Times:  testhelper.TimeArrayHelper(0, 6),
						Values: []float64{10, 10, 20, 30, 50, 80},
					},
				},
			},
			number: 1,
			value:  "min",
			output: []*models.SingleData{
				{
					Name: "TEST:PV:NAME2",
					Values: &models.Scalars{
						Times:  testhelper.TimeArrayHelper(0, 6),
						Values: []float64{10, 10, 20, 30, 50, 80},
					},
				},
			},
		},
		{
			inputSd: []*models.SingleData{
				{
					Name: "TEST:PV:NAME1",
					Values: &models.Scalars{
						Times:  testhelper.TimeArrayHelper(0, 6),
						Values: []float64{1, 1, 2, 3, 5, 81},
					},
				},
				{
					Name: "TEST:PV:NAME2",
					Values: &models.Scalars{
						Times:  testhelper.TimeArrayHelper(0, 6),
						Values: []float64{10, 10, 20, 30, 50, 80},
					},
				},
			},
			number: 1,
			value:  "max",
			output: []*models.SingleData{
				{
					Name: "TEST:PV:NAME1",
					Values: &models.Scalars{
						Times:  testhelper.TimeArrayHelper(0, 6),
						Values: []float64{1, 1, 2, 3, 5, 81},
					},
				},
			},
		},
		{
			inputSd: []*models.SingleData{
				{
					Name: "TEST:PV:NAME1",
					Values: &models.Scalars{
						Times:  testhelper.TimeArrayHelper(0, 6),
						Values: []float64{1, 1, 2, 3, 5, 8},
					},
				},
				{
					Name: "TEST:PV:NAME2",
					Values: &models.Scalars{
						Times:  testhelper.TimeArrayHelper(0, 6),
						Values: []float64{0, 10, 20, 30, 50, 80},
					},
				},
			},
			number: 1,
			value:  "absoluteMin",
			output: []*models.SingleData{
				{
					Name: "TEST:PV:NAME1",
					Values: &models.Scalars{
						Times:  testhelper.TimeArrayHelper(0, 6),
						Values: []float64{1, 1, 2, 3, 5, 8},
					},
				},
			},
		},
		{
			inputSd: []*models.SingleData{
				{
					Name: "TEST:PV:NAME1",
					Values: &models.Scalars{
						Times:  testhelper.TimeArrayHelper(0, 6),
						Values: []float64{1, 1, 2, 3, 5, 8},
					},
				},
				{
					Name: "TEST:PV:NAME2",
					Values: &models.Scalars{
						Times:  testhelper.TimeArrayHelper(0, 6),
						Values: []float64{10, 10, 20, 30, 50, 80},
					},
				},
			},
			number: 1,
			value:  "absoluteMax",
			output: []*models.SingleData{
				{
					Name: "TEST:PV:NAME2",
					Values: &models.Scalars{
						Times:  testhelper.TimeArrayHelper(0, 6),
						Values: []float64{10, 10, 20, 30, 50, 80},
					},
				},
			},
		},
		{
			inputSd: []*models.SingleData{
				{
					Name: "TEST:PV:NAME1",
					Values: &models.Scalars{
						Times:  testhelper.TimeArrayHelper(0, 6),
						Values: []float64{1, 1, 2, 3, 5, 8},
					},
				},
				{
					Name: "TEST:PV:NAME2",
					Values: &models.Scalars{
						Times:  testhelper.TimeArrayHelper(0, 6),
						Values: []float64{10},
					},
				},
			},
			number: 1,
			value:  "sum",
			output: []*models.SingleData{
				{
					Name: "TEST:PV:NAME1",
					Values: &models.Scalars{
						Times:  testhelper.TimeArrayHelper(0, 6),
						Values: []float64{1, 1, 2, 3, 5, 8},
					},
				},
			},
		},
	}
	for tdx, testCase := range tests {
		testName := fmt.Sprintf("case %d: %v", tdx, testCase.value)
		t.Run(testName, func(t *testing.T) {
			result, err := top(testCase.inputSd, testCase.number, testCase.value)
			if err != nil {
				t.Errorf("Error not expected %v", err)
			}
			models.SingleDataCompareHelper(result, testCase.output, t)
		})
	}
}

func TestBottom(t *testing.T) {
	var tests = []struct {
		inputSd []*models.SingleData
		number  int
		value   string
		output  []*models.SingleData
	}{
		{
			inputSd: []*models.SingleData{
				{
					Name: "TEST:PV:NAME1",
					Values: &models.Scalars{
						Times:  testhelper.TimeArrayHelper(0, 6),
						Values: []float64{1, 1, 2, 3, 5, 8},
					},
				},
				{
					Name: "TEST:PV:NAME2",
					Values: &models.Scalars{
						Times:  testhelper.TimeArrayHelper(0, 6),
						Values: []float64{10, 10, 20, 30, 50, 80},
					},
				},
			},
			number: 1,
			value:  "avg",
			output: []*models.SingleData{
				{
					Name: "TEST:PV:NAME1",
					Values: &models.Scalars{
						Times:  testhelper.TimeArrayHelper(0, 6),
						Values: []float64{1, 1, 2, 3, 5, 8},
					},
				},
			},
		},
		{
			inputSd: []*models.SingleData{
				{
					Name: "TEST:PV:NAME1",
					Values: &models.Scalars{
						Times:  testhelper.TimeArrayHelper(0, 6),
						Values: []float64{1, 1, 2, 3, 5, 8},
					},
				},
				{
					Name: "TEST:PV:NAME2",
					Values: &models.Scalars{
						Times:  testhelper.TimeArrayHelper(0, 6),
						Values: []float64{10, 10, 20, 30, 50, 80},
					},
				},
			},
			number: 1,
			value:  "min",
			output: []*models.SingleData{
				{
					Name: "TEST:PV:NAME1",
					Values: &models.Scalars{
						Times:  testhelper.TimeArrayHelper(0, 6),
						Values: []float64{1, 1, 2, 3, 5, 8},
					},
				},
			},
		},
		{
			inputSd: []*models.SingleData{
				{
					Name: "TEST:PV:NAME1",
					Values: &models.Scalars{
						Times:  testhelper.TimeArrayHelper(0, 6),
						Values: []float64{1, 1, 2, 3, 5, 8},
					},
				},
				{
					Name: "TEST:PV:NAME2",
					Values: &models.Scalars{
						Times:  testhelper.TimeArrayHelper(0, 6),
						Values: []float64{10, 10, 20, 30, 50, 80},
					},
				},
			},
			number: 1,
			value:  "max",
			output: []*models.SingleData{
				{
					Name: "TEST:PV:NAME1",
					Values: &models.Scalars{
						Times:  testhelper.TimeArrayHelper(0, 6),
						Values: []float64{1, 1, 2, 3, 5, 8},
					},
				},
			},
		},
		{
			inputSd: []*models.SingleData{
				{
					Name: "TEST:PV:NAME1",
					Values: &models.Scalars{
						Times:  testhelper.TimeArrayHelper(0, 6),
						Values: []float64{1, 1, 2, 3, 5, 8},
					},
				},
				{
					Name: "TEST:PV:NAME2",
					Values: &models.Scalars{
						Times:  testhelper.TimeArrayHelper(0, 6),
						Values: []float64{0, 10, 20, 30, 50, 80},
					},
				},
			},
			number: 1,
			value:  "absoluteMin",
			output: []*models.SingleData{
				{
					Name: "TEST:PV:NAME2",
					Values: &models.Scalars{
						Times:  testhelper.TimeArrayHelper(0, 6),
						Values: []float64{0, 10, 20, 30, 50, 80},
					},
				},
			},
		},
		{
			inputSd: []*models.SingleData{
				{
					Name: "TEST:PV:NAME1",
					Values: &models.Scalars{
						Times:  testhelper.TimeArrayHelper(0, 6),
						Values: []float64{1, 1, 2, 3, 5, 8},
					},
				},
				{
					Name: "TEST:PV:NAME2",
					Values: &models.Scalars{
						Times:  testhelper.TimeArrayHelper(0, 6),
						Values: []float64{10, 10, 20, 30, 50, 80},
					},
				},
			},
			number: 1,
			value:  "absoluteMax",
			output: []*models.SingleData{
				{
					Name: "TEST:PV:NAME1",
					Values: &models.Scalars{
						Times:  testhelper.TimeArrayHelper(0, 6),
						Values: []float64{1, 1, 2, 3, 5, 8},
					},
				},
			},
		},
		{
			inputSd: []*models.SingleData{
				{
					Name: "TEST:PV:NAME1",
					Values: &models.Scalars{
						Times:  testhelper.TimeArrayHelper(0, 6),
						Values: []float64{1, 1, 2, 3, 5, 8},
					},
				},
				{
					Name: "TEST:PV:NAME2",
					Values: &models.Scalars{
						Times:  testhelper.TimeArrayHelper(0, 6),
						Values: []float64{10},
					},
				},
			},
			number: 1,
			value:  "sum",
			output: []*models.SingleData{
				{
					Name: "TEST:PV:NAME2",
					Values: &models.Scalars{
						Times:  testhelper.TimeArrayHelper(0, 6),
						Values: []float64{10},
					},
				},
			},
		},
	}
	for tdx, testCase := range tests {
		testName := fmt.Sprintf("case %d: %v", tdx, testCase.value)
		t.Run(testName, func(t *testing.T) {
			result, err := bottom(testCase.inputSd, testCase.number, testCase.value)
			if err != nil {
				t.Errorf("Error not expected %v", err)
			}
			models.SingleDataCompareHelper(result, testCase.output, t)
		})
	}
}

func TestExclude(t *testing.T) {
	var tests = []struct {
		inputSd []*models.SingleData
		pattern string
		err     bool
		output  []*models.SingleData
	}{
		{
			inputSd: []*models.SingleData{
				{
					Name:   "Hello there",
					Values: &models.Scalars{},
				},
				{
					Name:   "General Kenobi!",
					Values: &models.Scalars{},
				},
			},
			pattern: "Hello",
			err:     false,
			output: []*models.SingleData{
				{
					Name:   "General Kenobi!",
					Values: &models.Scalars{},
				},
			},
		},
		{
			inputSd: []*models.SingleData{
				{
					Name:   "this is a phrase",
					Values: &models.Scalars{},
				},
				{
					Name:   "a phrase containing this",
					Values: &models.Scalars{},
				},
			},
			pattern: "^this",
			err:     false,
			output: []*models.SingleData{
				{
					Name:   "a phrase containing this",
					Values: &models.Scalars{},
				},
			},
		},
		{
			inputSd: []*models.SingleData{
				{
					Name:   "abcdef12ghi",
					Values: &models.Scalars{},
				},
				{
					Name:   "nonumbers",
					Values: &models.Scalars{},
				},
			},
			pattern: "f[1-9]*",
			err:     false,
			output: []*models.SingleData{
				{
					Name:   "nonumbers",
					Values: &models.Scalars{},
				},
			},
		},
		{
			inputSd: []*models.SingleData{
				{
					Name:   "abcdef12ghi",
					Values: &models.Scalars{},
				},
				{
					Name:   "nonumbers",
					Values: &models.Scalars{},
				},
			},
			pattern: "***Hello",
			err:     true,
			output: []*models.SingleData{
				{
					Name:   "abcdef12ghi",
					Values: &models.Scalars{},
				},
				{
					Name:   "nonumbers",
					Values: &models.Scalars{},
				},
			},
		},
	}
	for tdx, testCase := range tests {
		testName := fmt.Sprintf("case %d: %v", tdx, testCase.pattern)
		t.Run(testName, func(t *testing.T) {
			result, err := exclude(testCase.inputSd, testCase.pattern)
			if testCase.err {
				if err == nil {
					t.Errorf("Error expected but not received %v", err)
				}
			} else {
				if err != nil {
					t.Errorf("Error not expected %v", err)
				}
			}
			models.SingleDataCompareHelper(result, testCase.output, t)
		})
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
						Times:  testhelper.TimeArrayHelper(0, 6),
						Values: []float64{10, 10, 20, 30, 50, 80},
					},
				},
				{
					Values: &models.Scalars{
						Times:  testhelper.TimeArrayHelper(0, 6),
						Values: []float64{1, 1, 2, 3, 5, 8},
					},
				},
			},
			order: "asc",
			output: []*models.SingleData{
				{
					Values: &models.Scalars{
						Times:  testhelper.TimeArrayHelper(0, 6),
						Values: []float64{1, 1, 2, 3, 5, 8},
					},
				},
				{
					Values: &models.Scalars{
						Times:  testhelper.TimeArrayHelper(0, 6),
						Values: []float64{10, 10, 20, 30, 50, 80},
					},
				},
			},
		},
		{
			inputSd: []*models.SingleData{
				{
					Values: &models.Scalars{
						Times:  testhelper.TimeArrayHelper(0, 6),
						Values: []float64{1, 1, 2, 3, 5, 8},
					},
				},
				{
					Values: &models.Scalars{
						Times:  testhelper.TimeArrayHelper(0, 6),
						Values: []float64{10, 10, 20, 30, 50, 80},
					},
				},
			},
			order: "desc",
			output: []*models.SingleData{
				{
					Values: &models.Scalars{
						Times:  testhelper.TimeArrayHelper(0, 6),
						Values: []float64{10, 10, 20, 30, 50, 80},
					},
				},
				{
					Values: &models.Scalars{
						Times:  testhelper.TimeArrayHelper(0, 6),
						Values: []float64{1, 1, 2, 3, 5, 8},
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
						Times:  testhelper.TimeArrayHelper(0, 6),
						Values: []float64{10, 10, 20, 30, 50, 80},
					},
				},
				{
					Values: &models.Scalars{
						Times:  testhelper.TimeArrayHelper(0, 6),
						Values: []float64{1, 1, 2, 3, 5, 81},
					},
				},
			},
			order: "desc",
			output: []*models.SingleData{
				{
					Values: &models.Scalars{
						Times:  testhelper.TimeArrayHelper(0, 6),
						Values: []float64{1, 1, 2, 3, 5, 81},
					},
				},
				{
					Values: &models.Scalars{
						Times:  testhelper.TimeArrayHelper(0, 6),
						Values: []float64{10, 10, 20, 30, 50, 80},
					},
				},
			},
		},
		{
			inputSd: []*models.SingleData{
				{
					Values: &models.Scalars{
						Times:  testhelper.TimeArrayHelper(0, 6),
						Values: []float64{1, 1, 2, 3, 5, 81},
					},
				},
				{
					Values: &models.Scalars{
						Times:  testhelper.TimeArrayHelper(0, 6),
						Values: []float64{10, 10, 20, 30, 50, 80},
					},
				},
			},
			order: "asc",
			output: []*models.SingleData{
				{
					Values: &models.Scalars{
						Times:  testhelper.TimeArrayHelper(0, 6),
						Values: []float64{10, 10, 20, 30, 50, 80},
					},
				},
				{
					Values: &models.Scalars{
						Times:  testhelper.TimeArrayHelper(0, 6),
						Values: []float64{1, 1, 2, 3, 5, 81},
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
						Times:  testhelper.TimeArrayHelper(0, 6),
						Values: []float64{10, 10, 20, 30, 50, 80},
					},
				},
				{
					Values: &models.Scalars{
						Times:  testhelper.TimeArrayHelper(0, 6),
						Values: []float64{-100, 1, 2, 3, 5, 8},
					},
				},
			},
			order: "asc",
			output: []*models.SingleData{
				{
					Values: &models.Scalars{
						Times:  testhelper.TimeArrayHelper(0, 6),
						Values: []float64{-100, 1, 2, 3, 5, 8},
					},
				},
				{
					Values: &models.Scalars{
						Times:  testhelper.TimeArrayHelper(0, 6),
						Values: []float64{10, 10, 20, 30, 50, 80},
					},
				},
			},
		},
		{
			inputSd: []*models.SingleData{
				{
					Values: &models.Scalars{
						Times:  testhelper.TimeArrayHelper(0, 6),
						Values: []float64{-100, 1, 2, 3, 5, 8},
					},
				},
				{
					Values: &models.Scalars{
						Times:  testhelper.TimeArrayHelper(0, 6),
						Values: []float64{10, 10, 20, 30, 50, 80},
					},
				},
			},
			order: "desc",
			output: []*models.SingleData{
				{
					Values: &models.Scalars{
						Times:  testhelper.TimeArrayHelper(0, 6),
						Values: []float64{10, 10, 20, 30, 50, 80},
					},
				},
				{
					Values: &models.Scalars{
						Times:  testhelper.TimeArrayHelper(0, 6),
						Values: []float64{-100, 1, 2, 3, 5, 8},
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
						Times:  testhelper.TimeArrayHelper(0, 6),
						Values: []float64{10, 10, 20, 30, 50, 80},
					},
				},
				{
					Values: &models.Scalars{
						Times:  testhelper.TimeArrayHelper(0, 6),
						Values: []float64{1, 1, 2, 3, 5, 8000},
					},
				},
			},
			order: "desc",
			output: []*models.SingleData{
				{
					Values: &models.Scalars{
						Times:  testhelper.TimeArrayHelper(0, 6),
						Values: []float64{1, 1, 2, 3, 5, 8000},
					},
				},
				{
					Values: &models.Scalars{
						Times:  testhelper.TimeArrayHelper(0, 6),
						Values: []float64{10, 10, 20, 30, 50, 80},
					},
				},
			},
		},
		{
			inputSd: []*models.SingleData{
				{
					Values: &models.Scalars{
						Times:  testhelper.TimeArrayHelper(0, 6),
						Values: []float64{1, 1, 2, 3, 5, 8000},
					},
				},
				{
					Values: &models.Scalars{
						Times:  testhelper.TimeArrayHelper(0, 6),
						Values: []float64{10, 10, 20, 30, 50, 80},
					},
				},
			},
			order: "asc",
			output: []*models.SingleData{
				{
					Values: &models.Scalars{
						Times:  testhelper.TimeArrayHelper(0, 6),
						Values: []float64{10, 10, 20, 30, 50, 80},
					},
				},
				{
					Values: &models.Scalars{
						Times:  testhelper.TimeArrayHelper(0, 6),
						Values: []float64{1, 1, 2, 3, 5, 8000},
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
						Times:  testhelper.TimeArrayHelper(0, 6),
						Values: []float64{-10, -10, -20, -30, -50, -80},
					},
				},
				{
					Values: &models.Scalars{
						Times:  testhelper.TimeArrayHelper(0, 6),
						Values: []float64{1, 1, 2, 3, 5, 8},
					},
				},
			},
			order: "asc",
			output: []*models.SingleData{
				{
					Values: &models.Scalars{
						Times:  testhelper.TimeArrayHelper(0, 6),
						Values: []float64{1, 1, 2, 3, 5, 8},
					},
				},
				{
					Values: &models.Scalars{
						Times:  testhelper.TimeArrayHelper(0, 6),
						Values: []float64{-10, -10, -20, -30, -50, -80},
					},
				},
			},
		},
		{
			inputSd: []*models.SingleData{
				{
					Values: &models.Scalars{
						Times:  testhelper.TimeArrayHelper(0, 6),
						Values: []float64{1, 1, 2, 3, 5, 8},
					},
				},
				{
					Values: &models.Scalars{
						Times:  testhelper.TimeArrayHelper(0, 6),
						Values: []float64{-10, -10, -20, -30, -50, -80},
					},
				},
			},
			order: "desc",
			output: []*models.SingleData{
				{
					Values: &models.Scalars{
						Times:  testhelper.TimeArrayHelper(0, 6),
						Values: []float64{-10, -10, -20, -30, -50, -80},
					},
				},
				{
					Values: &models.Scalars{
						Times:  testhelper.TimeArrayHelper(0, 6),
						Values: []float64{1, 1, 2, 3, 5, 8},
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
						Times:  testhelper.TimeArrayHelper(0, 6),
						Values: []float64{-10, 10, 20, 30, 50, 80},
					},
				},
				{
					Values: &models.Scalars{
						Times:  testhelper.TimeArrayHelper(0, 6),
						Values: []float64{1, 1, 2, 3, 5, 8},
					},
				},
			},
			order: "asc",
			output: []*models.SingleData{
				{
					Values: &models.Scalars{
						Times:  testhelper.TimeArrayHelper(0, 6),
						Values: []float64{1, 1, 2, 3, 5, 8},
					},
				},
				{
					Values: &models.Scalars{
						Times:  testhelper.TimeArrayHelper(0, 6),
						Values: []float64{-10, 10, 20, 30, 50, 80},
					},
				},
			},
		},
		{
			inputSd: []*models.SingleData{
				{
					Values: &models.Scalars{
						Times:  testhelper.TimeArrayHelper(0, 6),
						Values: []float64{1, 1, 2, 3, 5, 8},
					},
				},
				{
					Values: &models.Scalars{
						Times:  testhelper.TimeArrayHelper(0, 6),
						Values: []float64{-10, 10, 20, 30, 50, 80},
					},
				},
			},
			order: "desc",
			output: []*models.SingleData{
				{
					Values: &models.Scalars{
						Times:  testhelper.TimeArrayHelper(0, 6),
						Values: []float64{-10, 10, 20, 30, 50, 80},
					},
				},
				{
					Values: &models.Scalars{
						Times:  testhelper.TimeArrayHelper(0, 6),
						Values: []float64{1, 1, 2, 3, 5, 8},
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

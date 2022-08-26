package functions

import (
	"fmt"
	"testing"

	"github.com/sasaki77/archiverappliance-datasource/pkg/models"
	"github.com/sasaki77/archiverappliance-datasource/pkg/testhelper"
)

// Tests

func TestApplyFunctions(t *testing.T) {
	var tests = []struct {
		name     string
		inputSd  []*models.SingleData
		inputAqm models.ArchiverQueryModel
		output   []*models.SingleData
	}{
		{
			name: "Offset test",
			inputSd: []*models.SingleData{
				{
					Values: &models.Scalars{
						Times:  testhelper.TimeArrayHelper(0, 6),
						Values: []float64{1, 1, 2, 3, 5, 8},
					},
				},
			},
			inputAqm: models.ArchiverQueryModel{
				Functions: []models.FunctionDescriptorQueryModel{
					{
						Def: models.FuncDefQueryModel{
							Category: "Transform",
							Name:     "offset",
							Params: []models.FuncDefParamQueryModel{
								{
									Name: "delta",
									Type: "float",
								},
							},
						},
						Params: []string{"2"},
					},
				},
			},
			output: []*models.SingleData{
				{
					Values: &models.Scalars{
						Times:  testhelper.TimeArrayHelper(0, 6),
						Values: []float64{3, 3, 4, 5, 7, 10},
					},
				},
			},
		},
		{
			name: "Offset and Scale test",
			inputSd: []*models.SingleData{
				{
					Values: &models.Scalars{
						Times:  testhelper.TimeArrayHelper(0, 6),
						Values: []float64{1, 1, 2, 3, 5, 8},
					},
				},
			},
			inputAqm: models.ArchiverQueryModel{
				Functions: []models.FunctionDescriptorQueryModel{
					{
						Def: models.FuncDefQueryModel{
							Category: "Transform",
							Name:     "offset",
							Params: []models.FuncDefParamQueryModel{
								{
									Name: "delta",
									Type: "float",
								},
							},
						},
						Params: []string{"2"},
					},
					{
						Def: models.FuncDefQueryModel{
							Category: "Transform",
							Name:     "scale",
							Params: []models.FuncDefParamQueryModel{
								{
									Name: "factor",
									Type: "float",
								},
							},
						},
						Params: []string{"3"},
					},
				},
			},
			output: []*models.SingleData{
				{
					Values: &models.Scalars{
						Times:  testhelper.TimeArrayHelper(0, 6),
						Values: []float64{9, 9, 12, 15, 21, 30},
					},
				},
			},
		},
		{
			name: "Offset, Scale, and Filter test",
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
						Values: []float64{100, 100, 100, 100, 100, 100},
					},
				},
			},
			inputAqm: models.ArchiverQueryModel{
				Functions: []models.FunctionDescriptorQueryModel{
					{
						Def: models.FuncDefQueryModel{
							Category: "Transform",
							Name:     "offset",
							Params: []models.FuncDefParamQueryModel{
								{
									Name: "delta",
									Type: "float",
								},
							},
						},
						Params: []string{"2"},
					},
					{
						Def: models.FuncDefQueryModel{
							Category: "Transform",
							Name:     "scale",
							Params: []models.FuncDefParamQueryModel{
								{
									Name: "factor",
									Type: "float",
								},
							},
						},
						Params: []string{"3"},
					},
					{
						Def: models.FuncDefQueryModel{
							Fake:          nil,
							Category:      "Filter Series",
							DefaultParams: testhelper.InitRawMsg(`[1 avg]`),
							Name:          "bottom",
							Params: []models.FuncDefParamQueryModel{
								{
									Name:    "number",
									Options: nil,
									Type:    "int",
								},
								{
									Name:    "value",
									Options: &[]string{"avg", "min", "max", "absoluteMin", "absoluteMax", "sum"},
									Type:    "string",
								},
							},
						},
						Params: []string{"1", "avg"},
					},
				},
			},
			output: []*models.SingleData{
				{
					Values: &models.Scalars{
						Times:  testhelper.TimeArrayHelper(0, 6),
						Values: []float64{9, 9, 12, 15, 21, 30},
					},
				},
			},
		},
		{
			name: "Array to Scalar and Offset test",
			inputSd: []*models.SingleData{
				{
					Values: &models.Arrays{
						Times:  testhelper.TimeArrayHelper(0, 3),
						Values: [][]float64{{1, 2, 3}, {4, 5, 6}, {7, 8, 9, 10}},
					},
				},
			},
			inputAqm: models.ArchiverQueryModel{
				Functions: []models.FunctionDescriptorQueryModel{
					{
						Def: models.FuncDefQueryModel{
							Category: models.FunctionCategory("Array to Scalar"),
							Name:     "toScalarByAvg",
						},
					},
					{
						Def: models.FuncDefQueryModel{
							Category: models.FunctionCategory("Array to Scalar"),
							Name:     "toScalarByMax",
						},
					},
					{
						Def: models.FuncDefQueryModel{
							Category: "Transform",
							Name:     "scale",
							Params: []models.FuncDefParamQueryModel{
								{
									Name: "factor",
									Type: "float",
								},
							},
						},
						Params: []string{"2"},
					},
				},
			},
			output: []*models.SingleData{
				{
					Name: "(avg)",
					Values: &models.Scalars{
						Times:  testhelper.TimeArrayHelper(0, 3),
						Values: []float64{4, 10, 17},
					},
				},
				{
					Name: "(max)",
					Values: &models.Scalars{
						Times:  testhelper.TimeArrayHelper(0, 3),
						Values: []float64{6, 12, 20},
					},
				},
			},
		},
	}

	for tdx, testCase := range tests {
		testName := fmt.Sprintf("case %d: %v", tdx, testCase.output)
		t.Run(testName, func(t *testing.T) {
			result, err := ApplyFunctions(tests[tdx].inputSd, testCase.inputAqm)
			if err != nil {
				t.Errorf("An error has been generated")
			}
			models.SingleDataCompareHelper(result, testCase.output, t)
		})
	}
}

func TestArrayFunctionSelector(t *testing.T) {
	var tests = []struct {
		name      string
		inputSd   []*models.SingleData
		inputFdqm models.FunctionDescriptorQueryModel
		output    []*models.SingleData
	}{
		{
			name: "toScalarByAvg test",
			inputSd: []*models.SingleData{
				{
					Values: &models.Arrays{
						Times:  testhelper.TimeArrayHelper(0, 3),
						Values: [][]float64{{1, 2, 3}, {4, 5, 6}, {7, 8, 9, 10}},
					},
				},
			},
			inputFdqm: models.FunctionDescriptorQueryModel{
				Def: models.FuncDefQueryModel{
					Category: models.FunctionCategory("Array to Scalar"),
					Name:     "toScalarByAvg",
				},
			},
			output: []*models.SingleData{
				{
					Name: "(avg)",
					Values: &models.Scalars{
						Times:  testhelper.TimeArrayHelper(0, 3),
						Values: []float64{2, 5, 8.5},
					},
				},
			},
		},
		{
			name: "toScalarByMax test",
			inputSd: []*models.SingleData{
				{
					Values: &models.Arrays{
						Times:  testhelper.TimeArrayHelper(0, 3),
						Values: [][]float64{{1, 2, 3}, {4, 5, 6}, {7, 8, 9, 10}},
					},
				},
			},
			inputFdqm: models.FunctionDescriptorQueryModel{
				Def: models.FuncDefQueryModel{
					Category: models.FunctionCategory("Array to Scalar"),
					Name:     "toScalarByMax",
				},
			},
			output: []*models.SingleData{
				{
					Name: "(max)",
					Values: &models.Scalars{
						Times:  testhelper.TimeArrayHelper(0, 3),
						Values: []float64{3, 6, 10},
					},
				},
			},
		},
		{
			name: "toScalarByMin test",
			inputSd: []*models.SingleData{
				{
					Values: &models.Arrays{
						Times:  testhelper.TimeArrayHelper(0, 3),
						Values: [][]float64{{1, 2, 3}, {4, 5, 6}, {7, 8, 9, 10}},
					},
				},
			},
			inputFdqm: models.FunctionDescriptorQueryModel{
				Def: models.FuncDefQueryModel{
					Category: models.FunctionCategory("Array to Scalar"),
					Name:     "toScalarByMin",
				},
			},
			output: []*models.SingleData{
				{
					Name: "(min)",
					Values: &models.Scalars{
						Times:  testhelper.TimeArrayHelper(0, 3),
						Values: []float64{1, 4, 7},
					},
				},
			},
		},
		{
			name: "toScalarBySum test",
			inputSd: []*models.SingleData{
				{
					Values: &models.Arrays{
						Times:  testhelper.TimeArrayHelper(0, 3),
						Values: [][]float64{{1, 2, 3}, {4, 5, 6}, {7, 8, 9, 10}},
					},
				},
			},
			inputFdqm: models.FunctionDescriptorQueryModel{
				Def: models.FuncDefQueryModel{
					Category: models.FunctionCategory("Array to Scalar"),
					Name:     "toScalarBySum",
				},
			},
			output: []*models.SingleData{
				{
					Name: "(sum)",
					Values: &models.Scalars{
						Times:  testhelper.TimeArrayHelper(0, 3),
						Values: []float64{6, 15, 34},
					},
				},
			},
		},
		{
			name: "toScalarByMed test",
			inputSd: []*models.SingleData{
				{
					Values: &models.Arrays{
						Times:  testhelper.TimeArrayHelper(0, 3),
						Values: [][]float64{{1, 2, 6}, {4, 5, 10}, {7, 8, 9, 10}},
					},
				},
			},
			inputFdqm: models.FunctionDescriptorQueryModel{
				Def: models.FuncDefQueryModel{
					Category: models.FunctionCategory("Array to Scalar"),
					Name:     "toScalarByMed",
				},
			},
			output: []*models.SingleData{
				{
					Name: "(median)",
					Values: &models.Scalars{
						Times:  testhelper.TimeArrayHelper(0, 3),
						Values: []float64{2, 5, 8.5},
					},
				},
			},
		},
		{
			name: "toScalarByStd test",
			inputSd: []*models.SingleData{
				{
					Values: &models.Arrays{
						Times:  testhelper.TimeArrayHelper(0, 3),
						Values: [][]float64{{10, 10, 20, 20}, {5, 5, 10, 10}},
					},
				},
			},
			inputFdqm: models.FunctionDescriptorQueryModel{
				Def: models.FuncDefQueryModel{
					Category: models.FunctionCategory("Array to Scalar"),
					Name:     "toScalarByStd",
				},
			},
			output: []*models.SingleData{
				{
					Name: "(std)",
					Values: &models.Scalars{
						Times:  testhelper.TimeArrayHelper(0, 3),
						Values: []float64{5, 2.5},
					},
				},
			},
		},
	}

	for tdx, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			result, err := arrayFunctionSelector(tests[tdx].inputSd, testCase.inputFdqm)
			if err != nil {
				t.Errorf("An error has been generated")
			}
			models.SingleDataCompareHelper(result, testCase.output, t)
		})
	}
}

func TestFunctionSelector(t *testing.T) {
	var tests = []struct {
		inputSd   []*models.SingleData
		inputFdqm models.FunctionDescriptorQueryModel
		output    []*models.SingleData
	}{
		{
			inputSd: []*models.SingleData{
				{
					Values: &models.Scalars{
						Times:  testhelper.TimeArrayHelper(0, 6),
						Values: []float64{1, 1, 2, 3, 5, 8},
					},
				},
			},
			inputFdqm: models.FunctionDescriptorQueryModel{
				Def: models.FuncDefQueryModel{
					Category: "Transform",
					Name:     "offset",
					Params: []models.FuncDefParamQueryModel{
						{
							Name: "delta",
							Type: "float",
						},
					},
				},
				Params: []string{"2"},
			},
			output: []*models.SingleData{
				{
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
			result, err := functionSelector(tests[tdx].inputSd, testCase.inputFdqm)
			if err != nil {
				t.Errorf("An error has been generated")
			}
			models.SingleDataCompareHelper(result, testCase.output, t)
		})
	}
}

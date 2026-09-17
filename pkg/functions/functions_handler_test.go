package functions

import (
	"testing"

	"github.com/sasaki77/archiverappliance-datasource/pkg/models"
	"github.com/sasaki77/archiverappliance-datasource/pkg/testhelper"
)

// Tests

// Arrays.Append keeps Times and Values in step, so a mismatch means the
// container was built by hand. Report it rather than indexing past the end.
func TestArrayFunctionSelectorMismatchedLengths(t *testing.T) {
	inputSd := []*models.SingleData{
		{
			Values: &models.Arrays{
				Times:  testhelper.TimeArrayHelper(0, 1),
				Values: [][]float64{{1, 2, 3}, {4, 5, 6}},
			},
		},
	}
	inputFdqm := models.FunctionDescriptorQueryModel{
		Def: models.FuncDefQueryModel{
			Category: models.FunctionCategory("Array to Scalar"),
			Name:     "toScalarByAvg",
		},
	}

	if _, err := arrayFunctionSelector(inputSd, inputFdqm); err == nil {
		t.Error("Expected an error for a container with fewer times than waveforms")
	}
}

// A window below 1 has no meaning. Reporting it tells the user what is wrong,
// where the series it would otherwise produce would not.
func TestFunctionSelectorRejectsWindowBelowOne(t *testing.T) {
	inputSd := []*models.SingleData{
		{
			Values: &models.Scalars{
				Times:  testhelper.TimeArrayHelper(0, 3),
				Values: testhelper.InitFloat64SlicePointer([]float64{1, 2, 3}),
			},
		},
	}

	for _, windowSize := range []string{"0", "-1"} {
		inputFdqm := models.FunctionDescriptorQueryModel{
			Def: models.FuncDefQueryModel{
				Category: "Transform",
				Name:     "movingAverage",
				Params:   []models.FuncDefParamQueryModel{{Name: "windowSize", Type: "int"}},
			},
			Params: []string{windowSize},
		}

		result, err := functionSelector(inputSd, inputFdqm)
		if err == nil {
			t.Errorf("window %s: expected an error", windowSize)
		}
		if v := result[0].Values.(*models.Scalars); *v.Values[0] != 1 {
			t.Errorf("window %s: the data should come back untouched, got %v", windowSize, *v.Values[0])
		}
	}
}

func TestFunctionSelectorRejectsNegativeCount(t *testing.T) {
	for _, name := range []string{"top", "bottom"} {
		inputSd := []*models.SingleData{
			{
				Name: "PV:A",
				Values: &models.Scalars{
					Times:  testhelper.TimeArrayHelper(0, 1),
					Values: testhelper.InitFloat64SlicePointer([]float64{1}),
				},
			},
			{
				Name: "PV:B",
				Values: &models.Scalars{
					Times:  testhelper.TimeArrayHelper(0, 1),
					Values: testhelper.InitFloat64SlicePointer([]float64{2}),
				},
			},
		}

		inputFdqm := models.FunctionDescriptorQueryModel{
			Def: models.FuncDefQueryModel{
				Category: "Filter Series",
				Name:     name,
				Params: []models.FuncDefParamQueryModel{
					{Name: "number", Type: "int"},
					{Name: "value", Type: "string"},
				},
			},
			Params: []string{"-1", "avg"},
		}

		result, err := functionSelector(inputSd, inputFdqm)
		if err == nil {
			t.Errorf("%s: expected an error", name)
		}
		if len(result) != 2 || result[0].Name != "PV:A" || result[1].Name != "PV:B" {
			t.Errorf("%s: the data should come back untouched, got %d series", name, len(result))
		}
	}
}

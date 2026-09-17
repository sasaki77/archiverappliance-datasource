package functions

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/sasaki77/archiverappliance-datasource/pkg/models"
)

// The vectors in testdata/functions are shared with src/specs/vectors.test.ts,
// so that both query paths are held to the same results.
const vectorDir = "../../testdata/functions"

type vectorSeries struct {
	Name   string            `json:"name"`
	Times  []int64           `json:"times"`
	Values []json.RawMessage `json:"values"`
}

type vectorFunction struct {
	Name   string   `json:"name"`
	Params []string `json:"params"`
}

type vectorCase struct {
	Name      string           `json:"name"`
	Input     []vectorSeries   `json:"input"`
	Functions []vectorFunction `json:"functions"`
	Output    []vectorSeries   `json:"output"`
}

func loadVectorDefs(t *testing.T) map[string]models.FuncDefQueryModel {
	t.Helper()

	raw, err := os.ReadFile(filepath.Join(vectorDir, "defs.json"))
	if err != nil {
		t.Fatal(err)
	}

	var defs []models.FuncDefQueryModel
	if err := json.Unmarshal(raw, &defs); err != nil {
		t.Fatal(err)
	}

	index := make(map[string]models.FuncDefQueryModel, len(defs))
	for _, def := range defs {
		index[def.Name] = def
	}
	return index
}

func decodeVectorValue(t *testing.T, raw json.RawMessage) (*float64, []float64, bool) {
	t.Helper()

	switch s := strings.TrimSpace(string(raw)); {
	case s == "null":
		return nil, nil, false
	case s == `"NaN"`:
		nan := math.NaN()
		return &nan, nil, false
	case strings.HasPrefix(s, "["):
		var wf []float64
		if err := json.Unmarshal(raw, &wf); err != nil {
			t.Fatal(err)
		}
		return nil, wf, true
	default:
		var v float64
		if err := json.Unmarshal(raw, &v); err != nil {
			t.Fatal(err)
		}
		return &v, nil, false
	}
}

func vectorToSingleData(t *testing.T, s vectorSeries) *models.SingleData {
	t.Helper()

	isWaveform := false
	for _, raw := range s.Values {
		if _, _, wf := decodeVectorValue(t, raw); wf {
			isWaveform = true
		}
	}

	d := &models.SingleData{Name: s.Name, PVname: s.Name}
	if isWaveform {
		arrays := models.NewArrays(len(s.Values))
		for i, raw := range s.Values {
			_, wf, _ := decodeVectorValue(t, raw)
			arrays.Append(wf, time.UnixMilli(s.Times[i]))
		}
		d.Values = arrays
		return d
	}

	scalars := models.NewSclars(len(s.Values))
	for i, raw := range s.Values {
		v, _, _ := decodeVectorValue(t, raw)
		scalars.Append(v, time.UnixMilli(s.Times[i]))
	}
	d.Values = scalars
	return d
}

func sameVectorFloat(want, got float64) bool {
	if math.IsNaN(want) || math.IsNaN(got) {
		return math.IsNaN(want) && math.IsNaN(got)
	}
	// movingAverage carries a running total, which rounds in the last digits.
	return math.Abs(want-got) <= 1e-9*math.Max(1, math.Abs(want))
}

func compareVectorSeries(t *testing.T, want vectorSeries, got *models.SingleData) {
	t.Helper()

	if got.Name != want.Name {
		t.Errorf("name: want %q, got %q", want.Name, got.Name)
	}

	scalars, ok := got.Values.(*models.Scalars)
	if !ok {
		t.Fatalf("%s: want scalars, got %T", want.Name, got.Values)
	}

	if len(scalars.Times) != len(want.Times) || len(scalars.Values) != len(want.Values) {
		t.Fatalf("%s: want %d samples, got %d times and %d values", want.Name, len(want.Times), len(scalars.Times), len(scalars.Values))
	}

	for i := range want.Times {
		if ms := scalars.Times[i].UnixMilli(); ms != want.Times[i] {
			t.Errorf("%s: time[%d]: want %d, got %d", want.Name, i, want.Times[i], ms)
		}

		wantV, _, _ := decodeVectorValue(t, want.Values[i])
		gotV := scalars.Values[i]
		switch {
		case wantV == nil && gotV == nil:
		case wantV == nil || gotV == nil:
			t.Errorf("%s: value[%d]: want %v, got %v", want.Name, i, fmtVectorPtr(wantV), fmtVectorPtr(gotV))
		case !sameVectorFloat(*wantV, *gotV):
			t.Errorf("%s: value[%d]: want %v, got %v", want.Name, i, *wantV, *gotV)
		}
	}
}

func fmtVectorPtr(p *float64) string {
	if p == nil {
		return "null"
	}
	return fmt.Sprint(*p)
}

func TestFunctionVectors(t *testing.T) {
	defs := loadVectorDefs(t)

	files, err := filepath.Glob(filepath.Join(vectorDir, "*.json"))
	if err != nil {
		t.Fatal(err)
	}

	for _, file := range files {
		if filepath.Base(file) == "defs.json" {
			continue
		}

		raw, err := os.ReadFile(file)
		if err != nil {
			t.Fatal(err)
		}

		var cases []vectorCase
		if err := json.Unmarshal(raw, &cases); err != nil {
			t.Fatalf("%s: %v", file, err)
		}

		for _, c := range cases {
			t.Run(filepath.Base(file)+"/"+c.Name, func(t *testing.T) {
				var qm models.ArchiverQueryModel
				for _, f := range c.Functions {
					def, ok := defs[f.Name]
					if !ok {
						t.Fatalf("no definition for %s", f.Name)
					}
					qm.Functions = append(qm.Functions, models.FunctionDescriptorQueryModel{Def: def, Params: f.Params})
				}

				input := make([]*models.SingleData, 0, len(c.Input))
				for _, s := range c.Input {
					input = append(input, vectorToSingleData(t, s))
				}

				got, err := ApplyFunctions(input, qm)
				if err != nil {
					t.Fatal(err)
				}

				if len(got) != len(c.Output) {
					names := make([]string, len(got))
					for i, d := range got {
						names[i] = d.Name
					}
					t.Fatalf("want %d series, got %d: %v", len(c.Output), len(got), names)
				}

				for i, want := range c.Output {
					compareVectorSeries(t, want, got[i])
				}
			})
		}
	}
}

// A function the editor offers but the dispatch does not know is dropped from
// the query without a visible error.
func TestFunctionVectorDefsAreDispatched(t *testing.T) {
	for name, def := range loadVectorDefs(t) {
		var defaults []string
		if def.DefaultParams != nil {
			if err := json.Unmarshal(*def.DefaultParams, &defaults); err != nil {
				t.Fatal(err)
			}
		}
		fdqm := models.FunctionDescriptorQueryModel{Def: def, Params: defaults}

		var err error
		switch def.Category {
		case models.FUNC_CATEGORY_TOSCALAR:
			_, err = arrayFunctionSelector(nil, fdqm)
		case models.FUNC_CATEGORY_TRANSFORM, models.FUNC_CATEGORY_FILTER, models.FUNC_CATEGORY_SORT:
			_, err = functionSelector(nil, fdqm)
		default:
			continue
		}

		if err != nil && strings.Contains(err.Error(), "not a recognized") {
			t.Errorf("%s: %v", name, err)
		}
	}
}

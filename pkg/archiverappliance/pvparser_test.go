package archiverappliance

import (
	"encoding/json"
	"fmt"
	"os"
	"slices"
	"testing"
)

// The cases are shared with src/specs/pvnames.test.ts, so that both query paths
// expand a target into the same PVs.
func TestIsolateBasicQuery(t *testing.T) {
	raw, err := os.ReadFile("../../testdata/pvnames.json")
	if err != nil {
		t.Fatal(err)
	}

	var tests []struct {
		Input  string   `json:"input"`
		Output []string `json:"output"`
	}
	if err := json.Unmarshal(raw, &tests); err != nil {
		t.Fatal(err)
	}

	for _, testCase := range tests {
		t.Run(testCase.Input, func(t *testing.T) {
			result := isolateBasicQuery(testCase.Input)
			if !slices.Equal(result, testCase.Output) {
				t.Errorf("got %q, want %q", result, testCase.Output)
			}
		})
	}
}

func TestSplitLowestLevelOnly(t *testing.T) {
	var tests = []struct {
		input  string
		output []string
	}{
		{
			input:  "one (two (three)) (four five) six",
			output: []string{"one (two (three)) (four five) six"},
		},
		{
			input:  "one two |three",
			output: []string{"one two ", "three"},
		},
		{
			input:  "one two |three | four",
			output: []string{"one two ", "three ", " four"},
		},
		{
			input:  "one (two |three) | four",
			output: []string{"one (two |three) ", " four"},
		},
		// {
		//     input:   "one (two (three)) (four five) six",
		//     output:  []string{"one (two (three)) (four five) six"},
		// },
	}
	for idx, testCase := range tests {
		testName := fmt.Sprintf("%d: %s", idx, testCase.input)
		t.Run(testName, func(t *testing.T) {
			result := splitLowestLevelOnly(testCase.input)
			if len(result) != len(testCase.output) {
				t.Fatalf("Lengths differ - Wanted: %v Got: %v", testCase.output, result)
			}
			for idx := range testCase.output {
				if testCase.output[idx] != result[idx] {
					t.Errorf("got %v, want %v", result, testCase.output)
				}
			}
		})
	}
}

func TestLocateOuterParen(t *testing.T) {
	var tests = []struct {
		input         string
		outputPhrases []string
		outputIdxs    [][]int
	}{
		{
			input:         "one (two (three)) (four five) six",
			outputPhrases: []string{"(two (three))", "(four five)"},
			outputIdxs:    [][]int{{4, 17}, {18, 29}},
		},
		{
			input:         "one (match) (here)",
			outputPhrases: []string{"(match)", "(here)"},
			outputIdxs:    [][]int{{4, 11}, {12, 18}},
		},
		{
			input:         "one (match) here",
			outputPhrases: []string{"(match)"},
			outputIdxs:    [][]int{{4, 11}},
		},
		{
			input:         "no matches here",
			outputPhrases: []string{},
			outputIdxs:    [][]int{},
		},
		// {
		//     input:           "",
		//     outputPhrases:  []string{},
		//     outputIdxs:     [][]int{},
		// },
	}
	for idx, testCase := range tests {
		testName := fmt.Sprintf("%d: %s, ", idx, testCase.input)
		t.Run(testName, func(t *testing.T) {
			result := locateOuterParen(testCase.input)
			// Check Phrases
			if len(result.Phrases) != len(testCase.outputPhrases) {
				t.Fatalf("Lengths differ - Wanted: %v Got: %v", testCase.outputPhrases, result.Phrases)
			}
			for idx := range testCase.outputPhrases {
				if testCase.outputPhrases[idx] != result.Phrases[idx] {
					t.Errorf("got %v, want %v", result.Phrases, testCase.outputPhrases)
				}
			}

			// Check indices
			if len(result.Idxs) != len(testCase.outputIdxs) {
				t.Fatalf("Lengths differ - Wanted: %v Got: %v", testCase.outputIdxs, result.Idxs)
			}
			for idx := range testCase.outputPhrases {
				for i := 0; i < 2; i++ {
					if testCase.outputIdxs[idx][i] != result.Idxs[idx][i] {
						t.Errorf("got %v, want %v", result.Idxs[idx], testCase.outputIdxs[idx])
					}
				}
			}
		})
	}
}

func TestPermuteQuery(t *testing.T) {
	var tests = []struct {
		input  [][]string
		output [][]string
	}{
		{input: [][]string{{"a", "d"}, {"b", "c"}}, output: [][]string{{"a", "b"}, {"a", "c"}, {"d", "b"}, {"d", "c"}}},
		{input: [][]string{{"a"}, {"b", "c"}}, output: [][]string{{"a", "b"}, {"a", "c"}}},
		{input: [][]string{{"a"}, {"b"}}, output: [][]string{{"a", "b"}}},
		{input: [][]string{{"a"}}, output: [][]string{{"a"}}},
		{input: [][]string{{"a", "b"}}, output: [][]string{{"a"}, {"b"}}},
		{input: [][]string{{"a"}, {"b"}, {"c", "d"}}, output: [][]string{{"a", "b", "c"}, {"a", "b", "d"}}},
		{input: [][]string{{"a"}, {"b"}, {"c"}, {"d", "e"}}, output: [][]string{{"a", "b", "c", "d"}, {"a", "b", "c", "e"}}},
		{input: [][]string{{}}, output: [][]string{}},
		{input: [][]string{}, output: [][]string{{}}},
	}
	for idx, testCase := range tests {
		testName := fmt.Sprintf("%d: %s, %s", idx, testCase.input, testCase.output)
		t.Run(testName, func(t *testing.T) {
			result := permuteQuery(testCase.input)
			if len(result) != len(testCase.output) {
				t.Fatalf("Lengths differ (0th index) - Wanted: %v (%v) Got: %v (%v)", testCase.output, len(testCase.output), result, len(result))
			}
			for idx := range testCase.output {
				if len(result[idx]) != len(testCase.output[idx]) {
					t.Fatalf("Lengths differ (1st index) - Wanted: %v (%v) Got: %v (%v)", testCase.output[idx], len(testCase.output[idx]), result[idx], len(result[idx]))
				}
				for jdx := range testCase.output[idx] {
					if testCase.output[idx][jdx] != result[idx][jdx] {
						t.Errorf("got %v, want %v at [%v][%v]", result, testCase.output, idx, jdx)
					}
				}
			}
		})
	}
}

func TestSelectiveInsert(t *testing.T) {
	var tests = []struct {
		input   string
		idxs    [][]int
		inserts []string
		output  string
	}{
		{
			input:   "hello there",
			idxs:    [][]int{{0, 5}},
			inserts: []string{"be"},
			output:  "be there",
		},
		{
			input:   "hello there",
			idxs:    [][]int{{6, 11}},
			inserts: []string{"friend"},
			output:  "hello friend",
		},
		{
			input:   "hello there",
			idxs:    [][]int{{0, 4}, {6, 11}},
			inserts: []string{"y", "what's up"},
			output:  "yo what's up",
		},
		{
			input:   "This won't work",
			idxs:    [][]int{{0, 4}, {6, 11}},
			inserts: []string{"y"},
			output:  "",
		},
		// {
		//     input: ,
		//     idxs: ,
		//     inserts: ,
		//     outpu:,
		// },
	}
	for idx, testCase := range tests {
		testName := fmt.Sprintf("%d: %s, %s", idx, testCase.input, testCase.output)
		t.Run(testName, func(t *testing.T) {
			result := selectiveInsert(testCase.input, testCase.idxs, testCase.inserts)
			if testCase.output != result {
				t.Errorf("Incorrect output - Wanted: %v Got: %v", testCase.output, result)
			}
		})
	}
}

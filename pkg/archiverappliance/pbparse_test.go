package archiverappliance

import (
	"math"
	"os"
	"testing"
	"time"

	"github.com/sasaki77/archiverappliance-datasource/pkg/models"
)

func TestParseSingleScalarData(t *testing.T) {
	ARCHIVER_FLOAT_PRECISION := 1e-18
	var tests = []struct {
		name     string
		firstVal float64
		lastVal  float64
	}{
		{
			name:     "SCALAR_BYTE_sampledata",
			firstVal: 0.0,
			lastVal:  110,
		},
		{
			name:     "SCALAR_SHORT_sampledata",
			firstVal: -32768,
			lastVal:  -32403,
		},
		{
			name:     "SCALAR_INT_sampledata",
			firstVal: -2147483648,
			lastVal:  -2147483283,
		},
		{
			name:     "SCALAR_ENUM_sampledata",
			firstVal: 0,
			lastVal:  365,
		},
		{
			name:     "SCALAR_FLOAT_sampledata",
			firstVal: 0.0,
			lastVal:  365,
		},
		{
			name:     "SCALAR_DOUBLE_sampledata",
			firstVal: 0.0,
			lastVal:  365,
		},
	}

	length := 366
	firstDate := time.Date(2012, 1, 1, 9, 43, 37, 0, time.UTC)
	lastDate := time.Date(2012, 12, 31, 9, 43, 37, 0, time.UTC)

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			f, err := os.Open("../test_data/pb/" + testCase.name)

			if err != nil {
				t.Fatalf("Failed to load test data: %v", err)
			}

			defer f.Close()

			sD, err := archiverPBSingleQueryParser(f)
			if err != nil {
				t.Fatalf("Failed to parse the data: %v", err)
			}

			if sD.Name != testCase.name {
				t.Fatalf("Names differ - Wanted: %v Got: %v", testCase.name, sD.Name)
			}

			if sD.PVname != testCase.name {
				t.Fatalf("Names differ - Wanted: %v Got: %v", testCase.name, sD.Name)
			}

			switch v := sD.Values.(type) {
			case *models.Scalars:
				resultLength := len(v.Times)
				if resultLength != length {
					t.Fatalf("Lengths differ - Wanted: %v Got: %v", length, resultLength)
				}

				if math.Abs(v.Values[0]-testCase.firstVal) > ARCHIVER_FLOAT_PRECISION {
					t.Fatalf("First values differ - Wanted: %v Got: %v", testCase.firstVal, v.Values[0])
				}

				if !v.Times[0].Equal(firstDate) {
					t.Fatalf("Fisrt date differ - Wanted: %v Got: %v", v.Times[0], firstDate)
				}

				lastIndex := resultLength - 1
				if math.Abs(v.Values[lastIndex]-testCase.lastVal) > ARCHIVER_FLOAT_PRECISION {
					t.Fatalf("last values differ - Wanted: %v Got: %v", testCase.lastVal, v.Values[lastIndex])
				}

				if !v.Times[lastIndex].Equal(lastDate) {
					t.Fatalf("Last date differ - Wanted: %v Got: %v", v.Times[lastIndex], lastDate)
				}
			}

		})
	}

}

func TestParseMultipleScalarData(t *testing.T) {
	ARCHIVER_FLOAT_PRECISION := 1e-18
	var tests = []struct {
		name      string
		length    int
		firstVal  float64
		firstDate time.Time
		lastVal   float64
		lastDate  time.Time
	}{
		{
			name:      "multipleChunksInMultipleYears",
			length:    730000,
			firstVal:  0.0,
			firstDate: time.Date(1970, 1, 1, 9, 43, 37, 0, time.UTC),
			lastVal:   364,
			lastDate:  time.Date(3969, 12, 31, 9, 43, 37, 0, time.UTC),
		},
		{
			name:      "multipleChunksInSameYear",
			length:    366,
			firstVal:  0.0,
			firstDate: time.Date(2012, 1, 1, 9, 43, 37, 0, time.UTC),
			lastVal:   365,
			lastDate:  time.Date(2012, 12, 31, 9, 43, 37, 0, time.UTC),
		},
		{
			name:      "multipleChunksOfRandomSizeInSameYear",
			length:    366,
			firstVal:  0.0,
			firstDate: time.Date(2012, 1, 1, 9, 43, 37, 0, time.UTC),
			lastVal:   365,
			lastDate:  time.Date(2012, 12, 31, 9, 43, 37, 0, time.UTC),
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			f, err := os.Open("../test_data/pb/" + testCase.name)

			if err != nil {
				t.Fatalf("Failed to load test data: %v", err)
			}

			defer f.Close()

			sD, err := archiverPBSingleQueryParser(f)
			if err != nil {
				t.Fatalf("Failed to parse the data: %v", err)
			}

			if sD.Name != testCase.name {
				t.Fatalf("Names differ - Wanted: %v Got: %v", testCase.name, sD.Name)
			}

			if sD.PVname != testCase.name {
				t.Fatalf("Names differ - Wanted: %v Got: %v", testCase.name, sD.Name)
			}

			switch v := sD.Values.(type) {
			case *models.Scalars:
				resultLength := len(v.Times)
				if resultLength != testCase.length {
					t.Fatalf("Lengths differ - Wanted: %v Got: %v", testCase.length, resultLength)
				}

				if math.Abs(v.Values[0]-testCase.firstVal) > ARCHIVER_FLOAT_PRECISION {
					t.Fatalf("First values differ - Wanted: %v Got: %v", testCase.firstVal, v.Values[0])
				}

				if !v.Times[0].Equal(testCase.firstDate) {
					t.Fatalf("Fisrt date differ - Wanted: %v Got: %v", v.Times[0], testCase.firstDate)
				}

				lastIndex := resultLength - 1
				if math.Abs(v.Values[lastIndex]-testCase.lastVal) > ARCHIVER_FLOAT_PRECISION {
					t.Fatalf("last values differ - Wanted: %v Got: %v", testCase.lastVal, v.Values[lastIndex])
				}

				if !v.Times[lastIndex].Equal(testCase.lastDate) {
					t.Fatalf("Last date differ - Wanted: %v Got: %v", v.Times[lastIndex], testCase.lastDate)
				}
			}

		})
	}
}

func TestParseSingleStringData(t *testing.T) {
	var tests = []struct {
		name     string
		firstVal string
		lastVal  string
	}{
		{
			name:     "SCALAR_STRING_sampledata",
			firstVal: "0",
			lastVal:  "365",
		},
	}

	length := 366
	firstDate := time.Date(2012, 1, 1, 9, 43, 37, 0, time.UTC)
	lastDate := time.Date(2012, 12, 31, 9, 43, 37, 0, time.UTC)

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			f, err := os.Open("../test_data/pb/" + testCase.name)

			if err != nil {
				t.Fatalf("Failed to load test data: %v", err)
			}

			defer f.Close()

			sD, err := archiverPBSingleQueryParser(f)
			if err != nil {
				t.Fatalf("Failed to parse the data: %v", err)
			}

			if sD.Name != testCase.name {
				t.Fatalf("Names differ - Wanted: %v Got: %v", testCase.name, sD.Name)
			}

			if sD.PVname != testCase.name {
				t.Fatalf("Names differ - Wanted: %v Got: %v", testCase.name, sD.Name)
			}

			switch v := sD.Values.(type) {
			case *models.Strings:
				resultLength := len(v.Times)
				if resultLength != length {
					t.Fatalf("Lengths differ - Wanted: %v Got: %v", length, resultLength)
				}

				if v.Values[0] != testCase.firstVal {
					t.Fatalf("First values differ - Wanted: %v Got: %v", testCase.firstVal, v.Values[0])
				}

				if !v.Times[0].Equal(firstDate) {
					t.Fatalf("Fisrt date differ - Wanted: %v Got: %v", v.Times[0], firstDate)
				}

				lastIndex := resultLength - 1
				if v.Values[lastIndex] != testCase.lastVal {
					t.Fatalf("last values differ - Wanted: %v Got: %v", testCase.lastVal, v.Values[lastIndex])
				}

				if !v.Times[lastIndex].Equal(lastDate) {
					t.Fatalf("Last date differ - Wanted: %v Got: %v", v.Times[lastIndex], lastDate)
				}
			}

		})
	}
}

func TestParseArrayData(t *testing.T) {
	ARCHIVER_FLOAT_PRECISION := 1e-18
	type outputArray struct {
		firstVal float64
		date     time.Time
		length   int
	}
	var tests = []struct {
		name      string
		length    int
		firstData outputArray
		lastData  outputArray
	}{
		{
			name:      "WAVEFORM_BYTE_sampledata",
			length:    2,
			firstData: outputArray{0, time.Date(2012, 1, 1, 9, 43, 37, 0, time.UTC), 0},
			lastData:  outputArray{1, time.Date(2012, 1, 2, 9, 43, 37, 0, time.UTC), 65536},
		},
		{
			name:      "WAVEFORM_SHORT_sampledata",
			length:    2,
			firstData: outputArray{0, time.Date(2012, 1, 1, 9, 43, 37, 0, time.UTC), 1},
			lastData:  outputArray{1, time.Date(2012, 1, 2, 9, 43, 37, 0, time.UTC), 1},
		},
		{
			name:      "WAVEFORM_INT_sampledata",
			length:    2,
			firstData: outputArray{-1, time.Date(2012, 1, 1, 9, 43, 37, 0, time.UTC), 0},
			lastData:  outputArray{1, time.Date(2012, 1, 2, 9, 43, 37, 0, time.UTC), 1},
		},
		{
			name:      "WAVEFORM_ENUM_sampledata",
			length:    2,
			firstData: outputArray{0, time.Date(2012, 1, 1, 9, 43, 37, 0, time.UTC), 1024},
			lastData:  outputArray{1, time.Date(2012, 1, 2, 9, 43, 37, 0, time.UTC), 1024},
		},
		{
			name:      "WAVEFORM_FLOAT_sampledata",
			length:    2,
			firstData: outputArray{-1, time.Date(2012, 1, 1, 9, 43, 37, 0, time.UTC), 0},
			lastData:  outputArray{0.9999996423721313, time.Date(2012, 1, 2, 9, 43, 37, 0, time.UTC), 1},
		},
		{
			name:      "WAVEFORM_DOUBLE_sampledata",
			length:    2,
			firstData: outputArray{-1, time.Date(2012, 1, 1, 9, 43, 37, 0, time.UTC), 0},
			lastData:  outputArray{0.0008726645152351496, time.Date(2012, 1, 2, 9, 43, 37, 0, time.UTC), 1},
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			f, err := os.Open("../test_data/pb/" + testCase.name)

			if err != nil {
				t.Fatalf("Failed to load test data: %v", err)
			}

			defer f.Close()

			sD, err := archiverPBSingleQueryParser(f)
			if err != nil {
				t.Fatalf("Failed to parse the data: %v", err)
			}

			if sD.Name != testCase.name {
				t.Fatalf("Names differ - Wanted: %v Got: %v", testCase.name, sD.Name)
			}

			if sD.PVname != testCase.name {
				t.Fatalf("Names differ - Wanted: %v Got: %v", testCase.name, sD.Name)
			}

			switch v := sD.Values.(type) {
			case *models.Arrays:
				resultLength := len(v.Times)
				if resultLength != testCase.length {
					t.Fatalf("Lengths differ - Wanted: %v Got: %v", testCase.length, resultLength)
				}

				if len(v.Values[0]) != testCase.firstData.length {
					t.Fatalf("First Data Lengths differ - Wanted: %v Got: %v", testCase.firstData.length, len(v.Values[0]))
				}

				if testCase.firstData.length > 0 && math.Abs(v.Values[0][0]-testCase.firstData.firstVal) > ARCHIVER_FLOAT_PRECISION {
					t.Fatalf("last values differ - Wanted: %v Got: %v", testCase.firstData.firstVal, v.Values[0][0])
				}

				if !v.Times[0].Equal(testCase.firstData.date) {
					t.Fatalf("Fisrt date differ - Wanted: %v Got: %v", v.Times[0], testCase.firstData.date)
				}

				lastIndex := resultLength - 1

				if len(v.Values[lastIndex]) != testCase.lastData.length {
					t.Fatalf("Last Data Lengths differ - Wanted: %v Got: %v", testCase.lastData.length, len(v.Values[lastIndex]))
				}

				if math.Abs(v.Values[lastIndex][0]-testCase.lastData.firstVal) > ARCHIVER_FLOAT_PRECISION {
					t.Fatalf("last values differ - Wanted: %v Got: %v", testCase.lastData.firstVal, v.Values[lastIndex][0])
				}

				if !v.Times[lastIndex].Equal(testCase.lastData.date) {
					t.Fatalf("Last date differ - Wanted: %v Got: %v", v.Times[lastIndex], testCase.lastData.date)
				}
			}

		})
	}

}

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
		fileName string
		pvname   string
		firstVal float64
		lastVal  float64
	}{
		{
			name:     "Scalar byte",
			fileName: "SCALAR_BYTE_sampledata",
			pvname:   "SCALAR_BYTE_sampledata",
			firstVal: 0.0,
			lastVal:  110,
		},
		{
			name:     "Scalar short",
			fileName: "SCALAR_SHORT_sampledata",
			pvname:   "SCALAR_SHORT_sampledata",
			firstVal: -32768,
			lastVal:  -32403,
		},
		{
			name:     "Scalar int",
			fileName: "SCALAR_INT_sampledata",
			pvname:   "SCALAR_INT_sampledata",
			firstVal: -2147483648,
			lastVal:  -2147483283,
		},
		{
			name:     "Scalar enum",
			fileName: "SCALAR_ENUM_sampledata",
			pvname:   "SCALAR_ENUM_sampledata",
			firstVal: 0,
			lastVal:  365,
		},
		{
			name:     "Scalar float",
			fileName: "SCALAR_FLOAT_sampledata",
			pvname:   "SCALAR_FLOAT_sampledata",
			firstVal: 0.0,
			lastVal:  365,
		},
		{
			name:     "Scalar double",
			fileName: "SCALAR_DOUBLE_sampledata",
			pvname:   "SCALAR_DOUBLE_sampledata",
			firstVal: 0.0,
			lastVal:  365,
		},
	}

	length := 366
	firstDate := time.Date(2012, 1, 1, 9, 43, 37, 0, time.UTC)
	lastDate := time.Date(2012, 12, 31, 9, 43, 37, 0, time.UTC)

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			f, err := os.Open("../test_data/pb/" + testCase.fileName)

			if err != nil {
				t.Fatalf("Failed to load test data: %v", err)
			}

			defer f.Close()

			sD, err := archiverPBSingleQueryParser(f)
			if err != nil {
				t.Fatalf("Failed to parse the data: %v", err)
			}

			if sD.Name != testCase.pvname {
				t.Fatalf("Names differ - Wanted: %v Got: %v", testCase.fileName, sD.Name)
			}

			if sD.PVname != testCase.pvname {
				t.Fatalf("Names differ - Wanted: %v Got: %v", testCase.pvname, sD.Name)
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
		fileName  string
		pvname    string
		length    int
		firstVal  float64
		firstDate time.Time
		lastVal   float64
		lastDate  time.Time
	}{
		{
			name:      "multipleChunksInMultipleYears",
			fileName:  "multipleChunksInMultipleYears",
			pvname:    "multipleChunksInMultipleYears",
			length:    730000,
			firstVal:  0.0,
			firstDate: time.Date(1970, 1, 1, 9, 43, 37, 0, time.UTC),
			lastVal:   364,
			lastDate:  time.Date(3969, 12, 31, 9, 43, 37, 0, time.UTC),
		},
		{
			name:      "multipleChunksInSameYear",
			fileName:  "multipleChunksInSameYear",
			pvname:    "multipleChunksInSameYear",
			length:    366,
			firstVal:  0.0,
			firstDate: time.Date(2012, 1, 1, 9, 43, 37, 0, time.UTC),
			lastVal:   365,
			lastDate:  time.Date(2012, 12, 31, 9, 43, 37, 0, time.UTC),
		},
		{
			name:      "multipleChunksOfRandomSizeInSameYear",
			fileName:  "multipleChunksOfRandomSizeInSameYear",
			pvname:    "multipleChunksOfRandomSizeInSameYear",
			length:    366,
			firstVal:  0.0,
			firstDate: time.Date(2012, 1, 1, 9, 43, 37, 0, time.UTC),
			lastVal:   365,
			lastDate:  time.Date(2012, 12, 31, 9, 43, 37, 0, time.UTC),
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			f, err := os.Open("../test_data/pb/" + testCase.fileName)

			if err != nil {
				t.Fatalf("Failed to load test data: %v", err)
			}

			defer f.Close()

			sD, err := archiverPBSingleQueryParser(f)
			if err != nil {
				t.Fatalf("Failed to parse the data: %v", err)
			}

			if sD.Name != testCase.pvname {
				t.Fatalf("Names differ - Wanted: %v Got: %v", testCase.fileName, sD.Name)
			}

			if sD.PVname != testCase.pvname {
				t.Fatalf("Names differ - Wanted: %v Got: %v", testCase.pvname, sD.Name)
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

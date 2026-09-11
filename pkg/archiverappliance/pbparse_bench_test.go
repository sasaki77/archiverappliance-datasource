package archiverappliance

import (
	"bytes"
	"os"
	"testing"

	"github.com/sasaki77/archiverappliance-datasource/pkg/models"
)

// Benchmarks for the PB parsing hot path.
//
// The fixture is read into memory once and replayed through a reused
// bytes.Reader so that neither file I/O nor the reader itself shows up in the
// reported ns/op and allocs/op. allocs/op is the primary metric here: the
// parser runs once per archived sample, so a constant factor per sample
// dominates on raw queries that return hundreds of thousands of points.

func loadPBFixture(b *testing.B, name string) []byte {
	b.Helper()
	payload, err := os.ReadFile("../test_data/pb/" + name)
	if err != nil {
		b.Fatalf("failed to read fixture %q: %v", name, err)
	}
	return payload
}

// valuesLen reports the number of parsed samples so that a benchmark cannot
// silently measure an early-error path instead of a full parse.
func valuesLen(v models.Values) int {
	switch t := v.(type) {
	case *models.Scalars:
		return len(t.Times)
	case *models.Arrays:
		return len(t.Times)
	case *models.Strings:
		return len(t.Times)
	case *models.Enums:
		return len(t.Times)
	default:
		return 0
	}
}

func benchmarkPBparse(b *testing.B, name string, field models.FieldName, capacity int, hideInvalid bool) {
	payload := loadPBFixture(b, name)
	reader := bytes.NewReader(payload)

	// Parse once outside the timed loop to verify the benchmark exercises the
	// real parse path and to report the sample count.
	sD, err := archiverPBSingleQueryParser(reader, field, capacity, hideInvalid)
	if err != nil {
		b.Fatalf("parse of %q failed: %v", name, err)
	}
	points := valuesLen(sD.Values)
	if points == 0 {
		b.Fatalf("parse of %q produced no samples", name)
	}

	b.SetBytes(int64(len(payload)))
	b.ReportAllocs()

	for b.Loop() {
		reader.Reset(payload)
		if _, err := archiverPBSingleQueryParser(reader, field, capacity, hideInvalid); err != nil {
			b.Fatalf("parse of %q failed: %v", name, err)
		}
	}

	b.StopTimer()
	b.ReportMetric(float64(points), "points")
	b.ReportMetric(float64(b.Elapsed().Nanoseconds())/float64(b.N*points), "ns/point")
}

// One day of a scalar double PV (~1.4 MB): the representative raw-query shape.
func BenchmarkPBparseOneday(b *testing.B) {
	benchmarkPBparse(b, "onedaysdbrdouble", models.FIELD_NAME_VAL, 1000, false)
}

// Same data with severity filtering enabled, which adds a check per sample.
func BenchmarkPBparseOnedayHideInvalid(b *testing.B) {
	benchmarkPBparse(b, "onedaysdbrdouble", models.FIELD_NAME_VAL, 1000, true)
}

// Multiple chunks spanning a year boundary (~12.5 MB): exercises the chunk
// header path and calcTime's per-sample time.Date call at scale.
func BenchmarkPBparseMultiYear(b *testing.B) {
	benchmarkPBparse(b, "multipleChunksInMultipleYears", models.FIELD_NAME_VAL, 1000, false)
}

// The meta-field path (getMetaValue) rather than getNumericValue.
func BenchmarkPBparseSevr(b *testing.B) {
	benchmarkPBparse(b, "onedaysdbrdouble", models.FIELD_NAME_SEVR, 1000, false)
}

// Waveform data: lines are large, so this is the case the bufio buffer size
// and the per-sample array conversion affect most.
func BenchmarkPBparseWaveform(b *testing.B) {
	benchmarkPBparse(b, "WAVEFORM_BYTE_sampledata", models.FIELD_NAME_VAL, 1000, false)
}

func BenchmarkPBparseString(b *testing.B) {
	benchmarkPBparse(b, "SCALAR_STRING_sampledata", models.FIELD_NAME_VAL, 1000, false)
}

// unescapeLine runs once per sample. The fast path for lines containing no
// escape sequences (the common case in real archiver data) is measured here.
func BenchmarkUnescapeLine(b *testing.B) {
	plain := bytes.Repeat([]byte{0x41}, 64)
	escaped := append(bytes.Repeat([]byte{0x41}, 32), append([]byte{0x1B, 0x01}, bytes.Repeat([]byte{0x41}, 30)...)...)

	b.Run("NoEscape", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			_ = unescapeLine(plain)
		}
	})
	b.Run("WithEscape", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			_ = unescapeLine(escaped)
		}
	})
}

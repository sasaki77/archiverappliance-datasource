package archiverappliance

import (
	"bytes"
	"fmt"
	"os"
	"testing"

	"github.com/sasaki77/archiverappliance-datasource/pkg/archiverappliance/pb"
	"github.com/sasaki77/archiverappliance-datasource/pkg/models"
	"google.golang.org/protobuf/proto"
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

// Waveform data: lines are large, so this is the case the read buffer and the
// per-sample array conversion affect most.
func BenchmarkPBparseWaveform(b *testing.B) {
	benchmarkPBparse(b, "WAVEFORM_BYTE_sampledata", models.FIELD_NAME_VAL, 1000, false)
}

// buildWaveformPB encodes a chunk of double waveform samples in the archiver's
// line format, so that a realistic number of oversized lines can be measured.
// The checked-in waveform fixture holds two samples, which is too few to show
// what the read path costs per sample.
func buildWaveformPB(b *testing.B, samples int, width int) []byte {
	b.Helper()

	pvname := "TEST:WAVEFORM"
	year := int32(2021)
	elementCount := int32(width)
	payloadType := pb.PayloadType_WAVEFORM_DOUBLE

	appendLine := func(out []byte, m proto.Message) []byte {
		encoded, err := proto.Marshal(m)
		if err != nil {
			b.Fatalf("failed to encode message: %v", err)
		}
		return append(escapeLine(out, encoded), '\n')
	}

	out := appendLine(nil, &pb.PayloadInfo{
		Type:         &payloadType,
		Pvname:       &pvname,
		Year:         &year,
		ElementCount: &elementCount,
	})

	val := make([]float64, width)
	for i := range val {
		val[i] = float64(i)
	}
	for i := 0; i < samples; i++ {
		sec := uint32(i)
		nano := uint32(0)
		out = appendLine(out, &pb.VectorDouble{
			Secondsintoyear: &sec,
			Nano:            &nano,
			Val:             val,
		})
	}

	return out
}

// escapeLine is the inverse of unescapeLine, used only to build test data.
func escapeLine(out []byte, line []byte) []byte {
	for _, c := range line {
		switch EscapeCharType(c) {
		case EscapeCharType_ESCAPE_CHAR:
			out = append(out, byte(EscapeCharType_ESCAPE_CHAR), byte(EscapeCharType_ESCAPE_ESCAPE_CHAR))
		case EscapeCharType_NEWLINE_CHAR:
			out = append(out, byte(EscapeCharType_ESCAPE_CHAR), byte(EscapeCharType_NEWLINE_ESCAPE_CHAR))
		case EscapeCharType_CARRIAGERETURN_CHAR:
			out = append(out, byte(EscapeCharType_ESCAPE_CHAR), byte(EscapeCharType_CARRIAGERETURN_ESCAPE_CHAR))
		default:
			out = append(out, c)
		}
	}
	return out
}

// Widths either side of bufio's 4 KiB default buffer, so that both the direct
// and the fragment-joining read paths are covered.
func BenchmarkPBparseWaveformSynthetic(b *testing.B) {
	for _, width := range []int{100, 2000} {
		b.Run(fmt.Sprintf("Width%d", width), func(b *testing.B) {
			const samples = 1000
			payload := buildWaveformPB(b, samples, width)
			reader := bytes.NewReader(payload)

			sD, err := archiverPBSingleQueryParser(reader, models.FIELD_NAME_VAL, 1000, false)
			if err != nil {
				b.Fatalf("parse failed: %v", err)
			}
			if got := valuesLen(sD.Values); got != samples {
				b.Fatalf("parsed %v samples, want %v", got, samples)
			}

			b.SetBytes(int64(len(payload)))
			b.ReportAllocs()

			for b.Loop() {
				reader.Reset(payload)
				if _, err := archiverPBSingleQueryParser(reader, models.FIELD_NAME_VAL, 1000, false); err != nil {
					b.Fatalf("parse failed: %v", err)
				}
			}
		})
	}
}

func BenchmarkPBparseString(b *testing.B) {
	benchmarkPBparse(b, "SCALAR_STRING_sampledata", models.FIELD_NAME_VAL, 1000, false)
}

// unescapeLine runs once per sample. The fast path for lines containing no
// escape sequences (the common case in real archiver data) is measured here.
//
// unescapeLine unescapes in place, so each iteration restores its input from a
// pristine copy. The restore is a 64-byte copy and is included in both
// measurements, which keeps the two sub-benchmarks comparable with each other.
func BenchmarkUnescapeLine(b *testing.B) {
	plain := bytes.Repeat([]byte{0x41}, 64)
	escaped := append(bytes.Repeat([]byte{0x41}, 32),
		append([]byte{byte(EscapeCharType_ESCAPE_CHAR), byte(EscapeCharType_ESCAPE_ESCAPE_CHAR)},
			bytes.Repeat([]byte{0x41}, 30)...)...)

	for _, tc := range []struct {
		name  string
		input []byte
	}{
		{"NoEscape", plain},
		{"WithEscape", escaped},
	} {
		b.Run(tc.name, func(b *testing.B) {
			line := make([]byte, len(tc.input))
			b.ReportAllocs()
			for b.Loop() {
				line = line[:len(tc.input)]
				copy(line, tc.input)
				_ = unescapeLine(line)
			}
		})
	}
}

package models

import (
	"testing"
	"time"
)

// Frame conversion is the last full traversal of a dataset before it is handed
// to Grafana, and the SDK's data.NewField always copies its input.

func BenchmarkSingleDataToFrameScalar(b *testing.B) {
	sd := &SingleData{
		Name:   "TEST:PV:NAME",
		PVname: "TEST:PV:NAME",
		Values: makeBenchScalars(benchPoints, 0),
	}

	b.ReportAllocs()
	for b.Loop() {
		_ = sd.ToFrame(FormatOption(FORMAT_TIMESERIES))
	}
}

// makeBenchArrays builds a waveform series: rows timestamps, each holding a
// width-element array.
func makeBenchArrays(rows int, width int) *Arrays {
	v := NewArrays(rows)
	base := time.Date(2021, time.January, 10, 0, 0, 0, 0, time.UTC)
	for i := 0; i < rows; i++ {
		row := make([]float64, width)
		for j := range row {
			row[j] = float64((i + j) % 1000)
		}
		v.Append(row, base.Add(time.Duration(i)*time.Millisecond))
	}
	return v
}

func BenchmarkSingleDataToFrameArray(b *testing.B) {
	// The timeseries format transposes the whole dataset before copying it
	// into one field per waveform element.
	for _, format := range []FormatOption{FormatOption(FORMAT_TIMESERIES), FormatOption(FORMAT_INDEX), FormatOption(FORMAT_DTSPACE)} {
		b.Run(string(format), func(b *testing.B) {
			sd := &SingleData{
				Name:   "TEST:PV:NAME",
				PVname: "TEST:PV:NAME",
				Values: makeBenchArrays(2000, 100),
			}
			b.ReportAllocs()
			for b.Loop() {
				_ = sd.ToFrame(format)
			}
		})
	}
}

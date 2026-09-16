package archiverappliance

import (
	"math"
	"math/rand"
	"testing"

	"github.com/sasaki77/archiverappliance-datasource/pkg/archiverappliance/pb"
	"google.golang.org/protobuf/encoding/protowire"
	"google.golang.org/protobuf/proto"
)

func sameF(a, b float64) bool {
	if math.IsNaN(a) && math.IsNaN(b) {
		return true
	}
	return a == b
}

// Tests
//
// The generated code stays in the build for the chunk headers and the waveform
// and string paths, so it serves as the reference these tests decode against.

// buildMessage fills a generated message of the given payload type with random
// values, so that proto.Marshal produces a line the decoder has to agree with.
func buildMessage(r *rand.Rand, dataType pb.PayloadType) proto.Message {
	sec := proto.Uint32(r.Uint32() % 31622400)
	nano := proto.Uint32(r.Uint32() % 1000000000)
	sev := proto.Int32(int32(r.Intn(4)))
	stat := proto.Int32(int32(r.Intn(100)) - 50)

	switch dataType {
	case pb.PayloadType_SCALAR_DOUBLE:
		v := (r.Float64() - 0.5) * math.Pow(10, float64(r.Intn(20)-10))
		return &pb.ScalarDouble{Secondsintoyear: sec, Nano: nano, Val: proto.Float64(v), Severity: sev, Status: stat}
	case pb.PayloadType_SCALAR_FLOAT:
		v := float32((r.Float64() - 0.5) * math.Pow(10, float64(r.Intn(10)-5)))
		return &pb.ScalarFloat{Secondsintoyear: sec, Nano: nano, Val: proto.Float32(v), Severity: sev, Status: stat}
	case pb.PayloadType_SCALAR_INT:
		return &pb.ScalarInt{Secondsintoyear: sec, Nano: nano, Val: proto.Int32(int32(r.Uint32())), Severity: sev, Status: stat}
	case pb.PayloadType_SCALAR_SHORT:
		return &pb.ScalarShort{Secondsintoyear: sec, Nano: nano, Val: proto.Int32(int32(r.Intn(65536) - 32768)), Severity: sev, Status: stat}
	case pb.PayloadType_SCALAR_ENUM:
		return &pb.ScalarEnum{Secondsintoyear: sec, Nano: nano, Val: proto.Int32(int32(r.Intn(65536) - 32768)), Severity: sev, Status: stat}
	case pb.PayloadType_SCALAR_BYTE:
		b := make([]byte, r.Intn(3))
		r.Read(b)
		return &pb.ScalarByte{Secondsintoyear: sec, Nano: nano, Val: b, Severity: sev, Status: stat}
	case pb.PayloadType_SCALAR_STRING:
		return &pb.ScalarString{Secondsintoyear: sec, Nano: nano, Val: proto.String("v"), Severity: sev, Status: stat}
	case pb.PayloadType_WAVEFORM_DOUBLE:
		v := make([]float64, r.Intn(4))
		for i := range v {
			v[i] = r.Float64()
		}
		return &pb.VectorDouble{Secondsintoyear: sec, Nano: nano, Val: v, Severity: sev, Status: stat}
	}

	panic("buildMessage: unhandled payload type")
}

var numericTypes = []pb.PayloadType{
	pb.PayloadType_SCALAR_DOUBLE,
	pb.PayloadType_SCALAR_FLOAT,
	pb.PayloadType_SCALAR_INT,
	pb.PayloadType_SCALAR_SHORT,
	pb.PayloadType_SCALAR_ENUM,
	pb.PayloadType_SCALAR_BYTE,
}

// TestDecodeSampleMatchesUnmarshal compares the value path against the
// generated code over random samples of every scalar type.
func TestDecodeSampleMatchesUnmarshal(t *testing.T) {
	r := rand.New(rand.NewSource(1))

	var s sample
	for n := 0; n < 20000; n++ {
		dataType := numericTypes[r.Intn(len(numericTypes))]
		msg := buildMessage(r, dataType)

		line, err := proto.Marshal(msg)
		if err != nil {
			t.Fatalf("marshal: %v", err)
		}

		want := initPBMessage(dataType)
		if err := proto.Unmarshal(line, want); err != nil {
			t.Fatalf("unmarshal: %v", err)
		}
		ref := want.(pb.NumericSamepleData)

		if err := decodeSample(line, dataType, &s); err != nil {
			t.Fatalf("decodeSample(%v): %v", dataType, err)
		}

		if s.secondsintoyear != ref.GetSecondsintoyear() || s.nano != ref.GetNano() {
			t.Fatalf("%v: time %v.%v, want %v.%v", dataType, s.secondsintoyear, s.nano, ref.GetSecondsintoyear(), ref.GetNano())
		}
		if s.severity != ref.GetSeverity() || s.status != ref.GetStatus() {
			t.Fatalf("%v: sevr/stat %v/%v, want %v/%v", dataType, s.severity, s.status, ref.GetSeverity(), ref.GetStatus())
		}
		if !sameF(s.val, ref.GetValAsFloat64()) {
			t.Fatalf("%v: val %v, want %v", dataType, s.val, ref.GetValAsFloat64())
		}
	}
}

// TestDecodeMetaMatchesUnmarshal covers the meta path, which also runs over
// string and waveform messages.
func TestDecodeMetaMatchesUnmarshal(t *testing.T) {
	r := rand.New(rand.NewSource(2))

	types := append([]pb.PayloadType{pb.PayloadType_SCALAR_STRING, pb.PayloadType_WAVEFORM_DOUBLE}, numericTypes...)

	var s sample
	for n := 0; n < 20000; n++ {
		dataType := types[r.Intn(len(types))]
		msg := buildMessage(r, dataType)

		line, err := proto.Marshal(msg)
		if err != nil {
			t.Fatalf("marshal: %v", err)
		}

		want := initPBMessage(dataType)
		if err := proto.Unmarshal(line, want); err != nil {
			t.Fatalf("unmarshal: %v", err)
		}
		ref := want.(pb.MetaFieldData)

		if err := decodeMeta(line, dataType, &s); err != nil {
			t.Fatalf("decodeMeta(%v): %v", dataType, err)
		}

		if s.secondsintoyear != ref.GetSecondsintoyear() || s.nano != ref.GetNano() {
			t.Fatalf("%v: time %v.%v, want %v.%v", dataType, s.secondsintoyear, s.nano, ref.GetSecondsintoyear(), ref.GetNano())
		}
		if s.severity != ref.GetSeverity() || s.status != ref.GetStatus() {
			t.Fatalf("%v: sevr/stat %v/%v, want %v/%v", dataType, s.severity, s.status, ref.GetSeverity(), ref.GetStatus())
		}
	}
}

// TestDecodeAgreesOnRejection checks that the decoder accepts and rejects the
// same lines the generated code does, including hand-built ones that
// proto.Marshal would not produce.
func TestDecodeAgreesOnRejection(t *testing.T) {
	// field 1 = 5, field 2 = 7, field 3 (fixed64) = 1.0, field 4 = 2
	full := []byte{0x08, 0x05, 0x10, 0x07, 0x19, 0, 0, 0, 0, 0, 0, 0xf0, 0x3f, 0x20, 0x02}

	var tests = []struct {
		name string
		line []byte
	}{
		{name: "in order", line: full},
		{name: "reversed", line: []byte{0x20, 0x02, 0x19, 0, 0, 0, 0, 0, 0, 0xf0, 0x3f, 0x10, 0x07, 0x08, 0x05}},
		{name: "unknown field between", line: []byte{0x08, 0x05, 0x30, 0x63, 0x10, 0x07, 0x19, 0, 0, 0, 0, 0, 0, 0xf0, 0x3f}},
		{name: "missing secondsintoyear", line: full[2:]},
		{name: "missing value", line: []byte{0x08, 0x05, 0x10, 0x07}},
		{name: "truncated value", line: []byte{0x08, 0x05, 0x10, 0x07, 0x19, 0, 0}},
		{name: "truncated tag", line: []byte{0x08, 0x05, 0xff}},
		{name: "empty", line: []byte{}},
		{name: "duplicate value keeps the last", line: append(append([]byte{}, full...), 0x19, 0, 0, 0, 0, 0, 0, 0, 0x40)},
		// A known number carrying the wrong wire type is an unknown field, not
		// an error, so an optional falls back to its default and a required one
		// is missed.
		{name: "optional with the wrong wire type", line: wrongWireType(full[:13], 4, protowire.Fixed32Type)},
		{name: "required with the wrong wire type", line: wrongWireType(full[2:], 1, protowire.Fixed32Type)},
	}

	var s sample
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			msg := initPBMessage(pb.PayloadType_SCALAR_DOUBLE)
			wantErr := proto.Unmarshal(testCase.line, msg) != nil

			gotErr := decodeSample(testCase.line, pb.PayloadType_SCALAR_DOUBLE, &s) != nil

			if gotErr != wantErr {
				t.Fatalf("rejected = %v, proto.Unmarshal rejected = %v", gotErr, wantErr)
			}
			if wantErr {
				return
			}

			ref := msg.(pb.NumericSamepleData)
			if s.secondsintoyear != ref.GetSecondsintoyear() || s.nano != ref.GetNano() ||
				s.severity != ref.GetSeverity() || s.status != ref.GetStatus() ||
				!sameF(s.val, ref.GetValAsFloat64()) {
				t.Fatalf("got %+v, want %v.%v sevr=%v stat=%v val=%v", s,
					ref.GetSecondsintoyear(), ref.GetNano(), ref.GetSeverity(), ref.GetStatus(), ref.GetValAsFloat64())
			}
		})
	}
}

// wrongWireType appends field num to line carrying four bytes under a wire type
// the .proto does not give it.
func wrongWireType(line []byte, num protowire.Number, typ protowire.Type) []byte {
	out := append(append([]byte{}, line...), protowire.AppendTag(nil, num, typ)...)
	return append(out, 1, 0, 0, 0)
}

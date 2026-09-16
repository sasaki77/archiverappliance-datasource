package archiverappliance

import (
	"errors"
	"math"

	"github.com/sasaki77/archiverappliance-datasource/pkg/archiverappliance/pb"
	"google.golang.org/protobuf/encoding/protowire"
)

// Scalar samples decode here by hand rather than through the generated code in
// pb, which the chunk headers and the waveform and string paths still use.
//
// The generated types keep every required field behind a pointer and
// proto.Unmarshal clears the message before each line, so a scalar costs three
// allocations per sample, and a raw query is hundreds of thousands of samples.
// (Merge: true skips the clear, and the allocations with it, but then an absent
// optional keeps the previous sample's value: one alarming sample would mark
// every later one.) Waveforms and strings gain nothing, their value being
// allocated per sample either way.
//
// The cost is that changes to EPICSEvent.proto have to be mirrored here. Only a
// new type for field 3 could do that silently, and
// TestDecodeSampleMatchesUnmarshal decodes both ways to catch it; the .proto
// describes files already written to disk, so its types do not move.
type sample struct {
	secondsintoyear uint32
	nano            uint32
	severity        int32
	status          int32
	val             float64
}

// Every scalar and vector message shares these. Only field 3 differs, and only
// in its type.
const (
	fieldSecondsintoyear = 1
	fieldNano            = 2
	fieldVal             = 3
	fieldSeverity        = 4
	fieldStatus          = 5
)

var errMissingRequiredField = errors.New("sample is missing a required field")

// decodeSample needs the payload type because field 3 is the one field whose
// encoding depends on it.
func decodeSample(line []byte, dataType pb.PayloadType, s *sample) error {
	return decode(line, dataType, true, s)
}

// decodeMeta skips the value, so it serves a waveform as readily as a scalar.
// It has to: severity and status are queryable on any PV.
func decodeMeta(line []byte, dataType pb.PayloadType, s *sample) error {
	return decode(line, dataType, false, s)
}

func decode(line []byte, dataType pb.PayloadType, wantVal bool, s *sample) error {
	*s = sample{}

	var seenTime, seenNano, seenVal bool

	for len(line) > 0 {
		num, typ, n := protowire.ConsumeTag(line)
		if n < 0 {
			return protowire.ParseError(n)
		}
		line = line[n:]

		switch {
		case num == fieldSecondsintoyear && typ == protowire.VarintType:
			v, n := protowire.ConsumeVarint(line)
			if n < 0 {
				return protowire.ParseError(n)
			}
			line, s.secondsintoyear, seenTime = line[n:], uint32(v), true

		case num == fieldNano && typ == protowire.VarintType:
			v, n := protowire.ConsumeVarint(line)
			if n < 0 {
				return protowire.ParseError(n)
			}
			line, s.nano, seenNano = line[n:], uint32(v), true

		case num == fieldSeverity && typ == protowire.VarintType:
			v, n := protowire.ConsumeVarint(line)
			if n < 0 {
				return protowire.ParseError(n)
			}
			line, s.severity = line[n:], int32(v)

		case num == fieldStatus && typ == protowire.VarintType:
			v, n := protowire.ConsumeVarint(line)
			if n < 0 {
				return protowire.ParseError(n)
			}
			line, s.status = line[n:], int32(v)

		case num == fieldVal:
			seenVal = true
			if !wantVal {
				n := protowire.ConsumeFieldValue(num, typ, line)
				if n < 0 {
					return protowire.ParseError(n)
				}
				line = line[n:]
				continue
			}

			n, err := decodeVal(line, dataType, typ, s)
			if err != nil {
				return err
			}
			line = line[n:]

		default:
			n := protowire.ConsumeFieldValue(num, typ, line)
			if n < 0 {
				return protowire.ParseError(n)
			}
			line = line[n:]
		}
	}

	// A vector's value is repeated rather than required, so only a scalar's is
	// demanded. decodeMeta demands it too, as the generated code does.
	if !seenTime || !seenNano || (!seenVal && scalarPayload(dataType)) {
		return errMissingRequiredField
	}

	return nil
}

func scalarPayload(dataType pb.PayloadType) bool {
	switch dataType {
	case pb.PayloadType_SCALAR_STRING,
		pb.PayloadType_SCALAR_BYTE,
		pb.PayloadType_SCALAR_SHORT,
		pb.PayloadType_SCALAR_INT,
		pb.PayloadType_SCALAR_ENUM,
		pb.PayloadType_SCALAR_FLOAT,
		pb.PayloadType_SCALAR_DOUBLE:
		return true
	}

	return false
}

// decodeVal returns how many bytes field 3 took. None of these encodings
// follow from the Go type, so each one is the .proto's word against a guess.
func decodeVal(line []byte, dataType pb.PayloadType, typ protowire.Type, s *sample) (int, error) {
	switch dataType {
	case pb.PayloadType_SCALAR_DOUBLE:
		if typ != protowire.Fixed64Type {
			return 0, errIllegalPayloadType
		}
		v, n := protowire.ConsumeFixed64(line)
		if n < 0 {
			return 0, protowire.ParseError(n)
		}
		s.val = math.Float64frombits(v)
		return n, nil

	case pb.PayloadType_SCALAR_FLOAT:
		if typ != protowire.Fixed32Type {
			return 0, errIllegalPayloadType
		}
		v, n := protowire.ConsumeFixed32(line)
		if n < 0 {
			return 0, protowire.ParseError(n)
		}
		s.val = float64(math.Float32frombits(v))
		return n, nil

	case pb.PayloadType_SCALAR_INT:
		if typ != protowire.Fixed32Type {
			return 0, errIllegalPayloadType
		}
		v, n := protowire.ConsumeFixed32(line)
		if n < 0 {
			return 0, protowire.ParseError(n)
		}
		s.val = float64(int32(v))
		return n, nil

	case pb.PayloadType_SCALAR_SHORT, pb.PayloadType_SCALAR_ENUM:
		if typ != protowire.VarintType {
			return 0, errIllegalPayloadType
		}
		v, n := protowire.ConsumeVarint(line)
		if n < 0 {
			return 0, protowire.ParseError(n)
		}
		s.val = float64(int32(protowire.DecodeZigZag(v)))
		return n, nil

	case pb.PayloadType_SCALAR_BYTE:
		if typ != protowire.BytesType {
			return 0, errIllegalPayloadType
		}
		v, n := protowire.ConsumeBytes(line)
		if n < 0 {
			return 0, protowire.ParseError(n)
		}
		// Only the first byte is the sample, as pb.ScalarByte reads it.
		if len(v) > 0 {
			s.val = float64(v[0])
		}
		return n, nil
	}

	return 0, errIllegalPayloadType
}

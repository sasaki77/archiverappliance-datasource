package archiverappliance

import (
	"bufio"
	"bytes"
	"io"
	"time"

	"github.com/grafana/grafana-plugin-sdk-go/backend/log"
	"github.com/sasaki77/archiverappliance-datasource/pkg/archiverappliance/pb"
	"github.com/sasaki77/archiverappliance-datasource/pkg/models"
	"google.golang.org/protobuf/proto"
)

type EscapeCharType byte

const (
	EscapeCharType_ESCAPE_CHAR                EscapeCharType = 0x1B
	EscapeCharType_ESCAPE_ESCAPE_CHAR         EscapeCharType = 0x01
	EscapeCharType_NEWLINE_CHAR               EscapeCharType = 0x0A
	EscapeCharType_NEWLINE_ESCAPE_CHAR        EscapeCharType = 0x02
	EscapeCharType_CARRIAGERETURN_CHAR        EscapeCharType = 0x0D
	EscapeCharType_CARRIAGERETURN_ESCAPE_CHAR EscapeCharType = 0x03
)

type MessageType int

const (
	MessageType_Numeric MessageType = 0
	MessageType_String  MessageType = 1
	MessageType_Array   MessageType = 2
	MessageType_Enum    MessageType = 3
)

type EPICSSeverity int

const (
	EPICSSeverity_NO_ALARM EPICSSeverity = 0
	EPICSSeverity_MINOR    EPICSSeverity = 1
	EPICSSeverity_MAJOR    EPICSSeverity = 2
	EPICSSeverity_INVALID  EPICSSeverity = 3
)

func archiverPBSingleQueryParser(in io.Reader, field models.FieldName, initialCapacity int, hideInvalid bool) (models.SingleData, error) {
	var sD models.SingleData
	info := &pb.PayloadInfo{}
	inChunk := false
	var dataType pb.PayloadType = -1

	// Sample timestamps are an offset into the chunk's year, which only
	// changes at a chunk boundary.
	var yearStart time.Time

	// bufio.Scanner cannot handle the long lines a waveform PV produces.
	reader := newLineReader(in)

	// Check if response data is valid
	_, err := reader.peek()
	if err != nil {
		// Peek(1) returns EOF error if its size is 0
		if err == io.EOF {
			return sD, errEmptyResponse
		}
		// Other errors should be handled as parse errors
		return sD, errFailedToParsePBFormat
	}

	var values models.Values

	// The payload type is fixed for the whole of a chunk, so one message is
	// allocated per chunk and reused for every sample in it. proto.Unmarshal
	// resets the message before decoding into it.
	var message proto.Message

	for {
		line, err := reader.readLine()
		if err != nil {
			if err != io.EOF {
				log.DefaultLogger.Error("Failed to read pb message", "error", err)
				return sD, errFailedToParsePBFormat
			}
			break
		}

		// length of a line is 0 if chunk is end
		if len(line) <= 0 {
			inChunk = false
			dataType = -1
			continue
		}

		unescapedLine := unescapeLine(line)

		// Find a chunk
		if !inChunk {
			if err := proto.Unmarshal(unescapedLine, info); err != nil {
				log.DefaultLogger.Error("Failed to parse payload info", "error", err)
				return sD, errFailedToParsePBFormat
			}

			inChunk = true
			dataType = info.GetType()
			yearStart = startOfYear(info.GetYear())
			message = initPBMessage(dataType)

			messageType, _ := getMessageType(dataType, field)

			// values is already initialized
			if values != nil {
				continue
			}

			// Inialialize values
			values, err = getInitializedValues(messageType, field, initialCapacity)
			if err != nil {
				return sD, err
			}

			continue
		}

		// Handle chunk data
		switch v := values.(type) {
		case *models.Scalars:
			var value float64
			var ok bool
			var sec, nano uint32
			var err error

			if field == models.FIELD_NAME_VAL {
				value, ok, sec, nano, err = getNumericValue(unescapedLine, message, hideInvalid)
			} else {
				value, ok, sec, nano, err = getMetaValue(unescapedLine, message, field)
			}

			if err != nil {
				return sD, errFailedToParsePBFormat
			}
			t := offsetIntoYear(yearStart, sec, nano)
			if !ok {
				// hideInvalid dropped this sample; the nil draws the gap.
				v.Append(nil, t)
				continue
			}
			v.AppendConcrete(value, t)
		case *models.Arrays:
			value, sec, nano, err := getArrayValue(unescapedLine, message)
			if err != nil {
				return sD, errFailedToParsePBFormat
			}
			t := offsetIntoYear(yearStart, sec, nano)
			v.Append(value, t)
		case *models.Strings:
			strMessage, ok := message.(*pb.ScalarString)
			if !ok {
				return sD, errIllegalPayloadType
			}
			value, sec, nano, err := getStringValue(unescapedLine, strMessage)
			if err != nil {
				return sD, errFailedToParsePBFormat
			}
			t := offsetIntoYear(yearStart, sec, nano)
			v.Append(value, t)
		case *models.Enums:
			value, ok, sec, nano, err := getMetaValue(unescapedLine, message, field)

			if err != nil {
				return sD, errFailedToParsePBFormat
			}
			if !ok {
				continue
			}
			t := offsetIntoYear(yearStart, sec, nano)
			v.Append(int16(value), t)
		default:
			return sD, errIllegalPayloadType
		}
	}

	pvname := info.GetPvname()
	if pvname == "" {
		return sD, errFailedToParsePBFormat
	}

	sD.Name = pvname
	sD.PVname = pvname
	sD.Values = values

	return sD, nil
}

// lineReader hands out one line at a time without allocating per line, which
// matters because the parser reads one line per archived sample.
type lineReader struct {
	reader *bufio.Reader

	// joined holds lines too long for the reader's buffer. A waveform PV
	// produces one such line per sample, so the buffer is kept and regrown
	// rather than allocated each time.
	joined []byte
}

func newLineReader(in io.Reader) *lineReader {
	return &lineReader{reader: bufio.NewReader(in)}
}

func (r *lineReader) peek() ([]byte, error) {
	return r.reader.Peek(1)
}

// readLine returns the next line without its delimiter. The result aliases
// either the reader's buffer or the joined buffer, so it is only valid until
// the next call, which is all the parser needs.
func (r *lineReader) readLine() ([]byte, error) {
	line, err := r.reader.ReadSlice('\n')
	if err == nil {
		return line[:len(line)-1], nil
	}

	if err != bufio.ErrBufferFull {
		// ReadSlice returns what it has along with io.EOF for a final line
		// with no delimiter. The archiver always terminates its lines, so
		// treat a partial tail as the end of the stream.
		return nil, err
	}

	// The line is longer than the reader's buffer, so collect it a bufferful
	// at a time. Each fragment has to be copied out before the next read
	// refills the buffer over it. ReadSlice hands back the bytes it managed
	// to read even when it reports an error, so append before inspecting err.
	r.joined = append(r.joined[:0], line...)
	for err == bufio.ErrBufferFull {
		line, err = r.reader.ReadSlice('\n')
		r.joined = append(r.joined, line...)
	}

	if err != nil {
		return nil, err
	}

	return r.joined[:len(r.joined)-1], nil
}

// unescapeLine reverses the escaping the archiver applies to sample lines.
//
// It runs once per sample, so the common case matters: real data rarely
// contains an escape sequence, and such a line is returned untouched. Every
// rule either drops a byte or turns two into one, so the result is never
// longer than the input and can be written back over line in place. The
// returned slice therefore aliases line, which the caller must not need
// afterwards.
func unescapeLine(line []byte) []byte {
	if bytes.IndexByte(line, byte(EscapeCharType_ESCAPE_CHAR)) < 0 &&
		bytes.IndexByte(line, byte(EscapeCharType_NEWLINE_CHAR)) < 0 {
		return line
	}

	buf := line[:0]
	escaped := false

	for _, b := range line {
		if escaped {
			switch EscapeCharType(b) {
			case EscapeCharType_ESCAPE_ESCAPE_CHAR:
				buf = append(buf, byte(EscapeCharType_ESCAPE_CHAR))
			case EscapeCharType_NEWLINE_ESCAPE_CHAR:
				buf = append(buf, byte(EscapeCharType_NEWLINE_CHAR))
			case EscapeCharType_CARRIAGERETURN_ESCAPE_CHAR:
				buf = append(buf, byte(EscapeCharType_CARRIAGERETURN_CHAR))
			default:
				buf = append(buf, b)
			}

			escaped = false
			continue
		}

		if EscapeCharType(b) == EscapeCharType_NEWLINE_CHAR {
			continue
		}

		if EscapeCharType(b) == EscapeCharType_ESCAPE_CHAR {
			escaped = true
			continue
		}

		buf = append(buf, b)
	}

	return buf
}

// getMetaValue returns a plain float64 for the same reason as getNumericValue.
// ok is always true on success here; it exists so both call sites are identical.
func getMetaValue(line []byte, message proto.Message, field models.FieldName) (val float64, ok bool, sec uint32, nano uint32, err error) {
	if message == nil {
		return 0, false, 0, 0, errIllegalPayloadType
	}

	if err := proto.Unmarshal(line, message); err != nil {
		log.DefaultLogger.Error("Failed to parse payload data", "error", err)
		return 0, false, 0, 0, errIllegalPayloadType
	}

	sample, isMeta := message.(pb.MetaFieldData)

	if !isMeta {
		return 0, false, 0, 0, errIllegalPayloadType
	}

	switch field {
	case models.FIELD_NAME_SEVR,
		models.FIELD_NAME_SEVR_AS_ENUM:
		val = float64(sample.GetSeverity())
	case models.FIELD_NAME_STAT,
		models.FIELD_NAME_STAT_AS_ENUM:
		val = float64(sample.GetStatus())
	default:
		return 0, false, 0, 0, errIllegalFieldName
	}

	sec = sample.GetSecondsintoyear()
	nano = sample.GetNano()

	return val, true, sec, nano, nil
}

// getNumericValue returns a plain float64, not a pointer, so that the container
// can place the value in its own block storage: a pointer here would cost one
// allocation per archived sample. ok is false when hideInvalid drops a sample,
// which the caller records as a gap.
func getNumericValue(line []byte, message proto.Message, hideInvalid bool) (val float64, ok bool, sec uint32, nano uint32, err error) {
	if message == nil {
		return 0, false, 0, 0, errIllegalPayloadType
	}

	if err := proto.Unmarshal(line, message); err != nil {
		log.DefaultLogger.Error("Failed to parse payload data", "error", err)
		return 0, false, 0, 0, errIllegalPayloadType
	}

	sample, isNumeric := message.(pb.NumericSamepleData)

	if !isNumeric {
		return 0, false, 0, 0, errIllegalPayloadType
	}

	sec = sample.GetSecondsintoyear()
	nano = sample.GetNano()

	if hideInvalid {
		sev := EPICSSeverity(sample.GetSeverity())
		if sev == EPICSSeverity_INVALID {
			return 0, false, sec, nano, nil
		}
	}

	return sample.GetValAsFloat64(), true, sec, nano, nil
}

func getStringValue(line []byte, message *pb.ScalarString) (val string, sec uint32, nano uint32, err error) {
	if message == nil {
		return "", 0, 0, errIllegalPayloadType
	}

	if err := proto.Unmarshal(line, message); err != nil {
		log.DefaultLogger.Error("Failed to parse payload data", "error", err)
		return "", 0, 0, errIllegalPayloadType
	}

	val = message.GetVal()
	sec = message.GetSecondsintoyear()
	nano = message.GetNano()

	return val, sec, nano, nil
}

func getArrayValue(line []byte, message proto.Message) (val []float64, sec uint32, nano uint32, err error) {
	if message == nil {
		return []float64{}, 0, 0, errIllegalPayloadType
	}

	if err := proto.Unmarshal(line, message); err != nil {
		log.DefaultLogger.Error("Failed to parse payload data", "error", err)
		return []float64{}, 0, 0, errIllegalPayloadType
	}

	sample, ok := message.(pb.ArraySamepleData)

	if !ok {
		return []float64{}, 0, 0, errIllegalPayloadType
	}

	val = sample.GetValAsFloat64()
	sec = sample.GetSecondsintoyear()
	nano = sample.GetNano()

	return val, sec, nano, nil
}

func initPBMessage(dataType pb.PayloadType) proto.Message {
	var m proto.Message

	switch dataType {
	case pb.PayloadType_SCALAR_STRING:
		m = &pb.ScalarString{}
	case pb.PayloadType_SCALAR_BYTE:
		m = &pb.ScalarByte{}
	case pb.PayloadType_SCALAR_SHORT:
		m = &pb.ScalarShort{}
	case pb.PayloadType_SCALAR_INT:
		m = &pb.ScalarInt{}
	case pb.PayloadType_SCALAR_ENUM:
		m = &pb.ScalarEnum{}
	case pb.PayloadType_SCALAR_FLOAT:
		m = &pb.ScalarFloat{}
	case pb.PayloadType_SCALAR_DOUBLE:
		m = &pb.ScalarDouble{}
	case pb.PayloadType_WAVEFORM_STRING:
		m = &pb.VectorString{}
	case pb.PayloadType_WAVEFORM_BYTE:
		m = &pb.VectorChar{}
	case pb.PayloadType_WAVEFORM_SHORT:
		m = &pb.VectorShort{}
	case pb.PayloadType_WAVEFORM_INT:
		m = &pb.VectorInt{}
	case pb.PayloadType_WAVEFORM_ENUM:
		m = &pb.VectorEnum{}
	case pb.PayloadType_WAVEFORM_FLOAT:
		m = &pb.VectorFloat{}
	case pb.PayloadType_WAVEFORM_DOUBLE:
		m = &pb.VectorDouble{}
	default:
		return nil
	}

	return m
}

func getMessageType(dataType pb.PayloadType, field models.FieldName) (MessageType, error) {
	switch field {
	case models.FIELD_NAME_SEVR_AS_ENUM,
		models.FIELD_NAME_STAT_AS_ENUM:
		return MessageType_Enum, nil
	case models.FIELD_NAME_SEVR,
		models.FIELD_NAME_STAT:
		return MessageType_Numeric, nil
	}

	switch dataType {
	case pb.PayloadType_SCALAR_BYTE,
		pb.PayloadType_SCALAR_SHORT,
		pb.PayloadType_SCALAR_INT,
		pb.PayloadType_SCALAR_ENUM,
		pb.PayloadType_SCALAR_FLOAT,
		pb.PayloadType_SCALAR_DOUBLE:
		{
			return MessageType_Numeric, nil
		}
	case pb.PayloadType_SCALAR_STRING:
		{
			return MessageType_String, nil
		}
	case pb.PayloadType_WAVEFORM_BYTE,
		pb.PayloadType_WAVEFORM_SHORT,
		pb.PayloadType_WAVEFORM_INT,
		pb.PayloadType_WAVEFORM_ENUM,
		pb.PayloadType_WAVEFORM_FLOAT,
		pb.PayloadType_WAVEFORM_DOUBLE:
		{
			return MessageType_Array, nil
		}
	}

	return -1, errIllegalPayloadType
}

func getInitializedValues(mtype MessageType, field models.FieldName, capacity int) (values models.Values, err error) {
	switch mtype {
	case MessageType_Numeric:
		values = models.NewSclars(capacity)
	case MessageType_String:
		values = models.NewStrings(capacity)
	case MessageType_Array:
		values = models.NewArrays(capacity)
	case MessageType_Enum:
		switch field {
		case models.FIELD_NAME_SEVR_AS_ENUM:
			values = models.NewSevirityEnums(capacity)
		case models.FIELD_NAME_STAT_AS_ENUM:
			values = models.NewStatusEnums(capacity)
		default:
			return nil, errIllegalFieldName
		}
	}

	return values, nil
}

func startOfYear(year int32) time.Time {
	return time.Date(int(year), 1, 1, 0, 0, 0, 0, time.UTC)
}

// offsetIntoYear resolves a sample's timestamp. Adding to the start of the
// year keeps time.Date's calendar normalisation out of the per-sample path.
func offsetIntoYear(yearStart time.Time, sec uint32, nano uint32) time.Time {
	return yearStart.Add(time.Duration(sec)*time.Second + time.Duration(nano)*time.Nanosecond)
}

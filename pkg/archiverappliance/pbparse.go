package archiverappliance

import (
	"bufio"
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
)

func archiverPBSingleQueryParser(in io.Reader, field models.FieldName, initialCapacity int) (models.SingleData, error) {
	var sD models.SingleData
	info := &pb.PayloadInfo{}
	inChunk := false
	var dataType pb.PayloadType = -1
	var year int32 = -1

	// Use ReadBytes insetead of bufioc.Scanner to handle large size array
	reader := bufio.NewReader(in)

	var values models.Values
	for {
		lineWithDelim, err := reader.ReadBytes('\n')
		if err != nil {
			if err != io.EOF {
				log.DefaultLogger.Error("Failed to read pb message:", err)
				return sD, errFailedToParsePBFormat
			}
			break
		}
		line := lineWithDelim[:len(lineWithDelim)-1]

		// length of a line is 0 if chunk is end
		if len(line) <= 0 {
			inChunk = false
			dataType = -1
			continue
		}

		escapedLine := unescapeLine(line)

		// Find a chunk
		if !inChunk {
			if err := proto.Unmarshal(escapedLine, info); err != nil {
				log.DefaultLogger.Error("Failed to parse paylod info:", err)
			}

			inChunk = true
			dataType = *info.Type
			year = *info.Year

			messageType, _ := getMessageType(dataType, field)

			// values is already initialized
			if values != nil {
				continue
			}

			// Inialialize values
			switch messageType {
			case MessageType_Numeric:
				values = models.NewSclars(initialCapacity)
			case MessageType_String:
				values = models.NewStrings(initialCapacity)
			case MessageType_Array:
				values = models.NewArrays(initialCapacity)
			}

			continue
		}

		// Handle chunk data
		switch v := values.(type) {
		case *models.Scalars:
			var value float64
			var sec, nano uint32
			var err error

			if field == models.FIELD_NAME_VAL {
				value, sec, nano, err = getNumericValue(escapedLine, dataType)
			} else {
				value, sec, nano, err = getMetaValue(escapedLine, dataType, field)
			}

			if err != nil {
				return sD, errFailedToParsePBFormat
			}
			t := time.Date(int(year), 1, 1, 0, 0, int(sec), int(nano), time.UTC)
			v.AppendConcrete(value, t)
		case *models.Arrays:
			value, sec, nano, err := getArrayValue(escapedLine, dataType)
			if err != nil {
				return sD, errFailedToParsePBFormat
			}
			t := time.Date(int(year), 1, 1, 0, 0, int(sec), int(nano), time.UTC)
			v.Append(value, t)
		case *models.Strings:
			value, sec, nano, err := getStringValue(escapedLine)
			if err != nil {
				return sD, errFailedToParsePBFormat
			}
			t := time.Date(int(year), 1, 1, 0, 0, int(sec), int(nano), time.UTC)
			v.Append(value, t)
		default:
			return sD, errIllegalPayloadType
		}
	}

	sD.Name = *info.Pvname
	sD.PVname = *info.Pvname
	sD.Values = values

	return sD, nil
}

func unescapeLine(line []byte) []byte {
	buf := make([]byte, 0, len(line))
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

func getMetaValue(line []byte, dataType pb.PayloadType, field models.FieldName) (val float64, sec uint32, nano uint32, err error) {
	message := initPBMessage(dataType)

	if message == nil {
		return 0, 0, 0, errIllegalPayloadType
	}

	if err := proto.Unmarshal(line, *message); err != nil {
		log.DefaultLogger.Error("Failed to parse paylod data:", err)
		return 0, 0, 0, errIllegalPayloadType
	}

	sample, ok := (*message).(pb.MetaFieldData)

	if !ok {
		return 0, 0, 0, errIllegalPayloadType
	}

	switch field {
	case models.FIELD_NAME_SEVR:
		val = float64(sample.GetSeverity())
	case models.FIELD_NAME_STAT:
		val = float64(sample.GetStatus())
	default:
		return 0, 0, 0, errIllegalFieldName
	}

	sec = sample.GetSecondsintoyear()
	nano = sample.GetNano()

	return val, sec, nano, nil
}

func getNumericValue(line []byte, dataType pb.PayloadType) (val float64, sec uint32, nano uint32, err error) {
	message := initPBMessage(dataType)

	if message == nil {
		return 0, 0, 0, errIllegalPayloadType
	}

	if err := proto.Unmarshal(line, *message); err != nil {
		log.DefaultLogger.Error("Failed to parse paylod data:", err)
		return 0, 0, 0, errIllegalPayloadType
	}

	sample, ok := (*message).(pb.NumericSamepleData)

	if !ok {
		return 0, 0, 0, errIllegalPayloadType
	}

	val = sample.GetValAsFloat64()
	sec = sample.GetSecondsintoyear()
	nano = sample.GetNano()

	return val, sec, nano, nil
}

func getStringValue(line []byte) (val string, sec uint32, nano uint32, err error) {
	message := &pb.ScalarString{}

	if err := proto.Unmarshal(line, message); err != nil {
		log.DefaultLogger.Error("Failed to parse paylod data:", err)
		return "", 0, 0, errIllegalPayloadType
	}

	val = message.GetVal()
	sec = message.GetSecondsintoyear()
	nano = message.GetNano()

	return val, sec, nano, nil
}

func getArrayValue(line []byte, dataType pb.PayloadType) (val []float64, sec uint32, nano uint32, err error) {
	message := initPBMessage(dataType)

	if message == nil {
		return []float64{}, 0, 0, errIllegalPayloadType
	}

	if err := proto.Unmarshal(line, *message); err != nil {
		log.DefaultLogger.Error("Failed to parse paylod data:", err)
		return []float64{}, 0, 0, errIllegalPayloadType
	}

	sample, ok := (*message).(pb.ArraySamepleData)

	if !ok {
		return []float64{}, 0, 0, errIllegalPayloadType
	}

	val = sample.GetValAsFloat64()
	sec = sample.GetSecondsintoyear()
	nano = sample.GetNano()

	return val, sec, nano, nil
}

func initPBMessage(dataType pb.PayloadType) *proto.Message {
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

	return &m
}

func getMessageType(dataType pb.PayloadType, field models.FieldName) (MessageType, error) {
	if field != models.FIELD_NAME_VAL {
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

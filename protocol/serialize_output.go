package protocol

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"reflect"
)

func Serialize(struc interface{}) ([]byte, error) {
	if reflect.ValueOf(struc).Kind() != reflect.Struct {
		return nil, fmt.Errorf("input is not of type Struct")
	}

	buf := new(bytes.Buffer)

	value := reflect.ValueOf(struc)
	numberOfFields := value.NumField()
	for i := 0; i < numberOfFields; i++ {
		field := value.Field(i)
		if field.CanInterface() {
			err := serializeField(buf, field.Interface())
			if err != nil {
				return nil, err
			}
		} else {
			return nil, fmt.Errorf("enable to interface with field")
		}
	}

	return buf.Bytes(), nil
}

func serializeField(buf *bytes.Buffer, data interface{}) error {
	switch data.(type) {
	case uint8:
		return writeUint8(buf, data.(uint8))
	case uint16:
		return writeUint16(buf, data.(uint16))
	case [4]byte:
		return writeSliceByte4(buf, data.([4]byte))
	case [20]byte:
		return writeSliceByte20(buf, data.([20]byte))
	case []byte:
		return writeSliceByte(buf, data.([]byte))
	case [][20]byte:
		return writeSliceOfSliceByte20(buf, data.([][20]byte))
	case string:
		return writeString(buf, data.(string))
	default:
		return fmt.Errorf("unsupported data type: %T", data)
	}
}

func writeUint8(buf *bytes.Buffer, data uint8) error {
	return binary.Write(buf, binary.BigEndian, data)
}

func writeUint16(buf *bytes.Buffer, data uint16) error {
	return binary.Write(buf, binary.BigEndian, data)
}

func writeString(buf *bytes.Buffer, data string) error {
	stringBytes := []byte(data)
	return binary.Write(buf, binary.BigEndian, stringBytes)
}

func writeSliceByte(buf *bytes.Buffer, data []byte) error {
	return binary.Write(buf, binary.BigEndian, data)
}

func writeSliceByte4(buf *bytes.Buffer, data [4]byte) error {
	return binary.Write(buf, binary.BigEndian, data)
}

func writeSliceByte20(buf *bytes.Buffer, data [20]byte) error {
	return binary.Write(buf, binary.BigEndian, data)
}

func writeSliceOfSliceByte20(buf *bytes.Buffer, data [][20]byte) error {
	for i := 0; i < len(data); i++ {
		if err := binary.Write(buf, binary.BigEndian, data[i]); err != nil {
			return fmt.Errorf("error serializing output of type [][20]byte")
		}
	}
	return nil
}

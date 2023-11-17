package protocol

import (
	"encoding/binary"
	"fmt"
	"io"
	"reflect"
)

func Serialize(writer io.Writer, struc interface{}) error {
	value := reflect.ValueOf(struc)
	if value.Kind() != reflect.Struct {
		return fmt.Errorf("input is not of type struct")
	}

	for i := 0; i < value.NumField(); i++ {
		field := value.Field(i)

		if field.CanInterface() {
			err := serializeField(writer, field.Interface())
			if err != nil {
				return err
			}
		} else {
			return fmt.Errorf("can't interface with field without panicking")
		}
	}

	return nil
}

func serializeField(writer io.Writer, field interface{}) error {
	switch data := field.(type) {
	case uint8, uint16, uint32, uint64, int8, int16, int32, int64:
		return write(writer, data)
	case string:
		return writeString(writer, data)
	case []uint8:
		return writeArray(writer, data)
	case [4]uint8:
		return writeArray(writer, data[:])
	default:
		return fmt.Errorf("serialize unsupported type: %T", data)
	}
}

func writeString(writer io.Writer, data string) error {
	// Write the size of the string
	err := write(writer, len(data))
	if err != nil {
		return err
	}

	// Write the content of the string in bytes
	err = write(writer, []uint8(data))
	if err != nil {
		return err
	}

	return nil
}

func writeArray[T any](writer io.Writer, data []T) error {
	err := write(writer, uint32(len(data)))
	if err != nil {
		return err
	}

	for _, element := range data {
		err := serializeField(writer, element)
		if err != nil {
			return err
		}
	}

	return nil
}

func write(writer io.Writer, data interface{}) error {
	return binary.Write(writer, binary.LittleEndian, data)
}

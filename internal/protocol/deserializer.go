package protocol

import (
	"encoding/binary"
	"fmt"
	"io"
	"reflect"
)

func Deserialize(reader io.Reader) (interface{}, error) {
	structType, err := deserializeStructType(reader)
	if err != nil {
		_ = fmt.Errorf("Error deserializing struct type: %v", err)
		return nil, err
	}

	struc := PacketStructFromType(structType)
	err = deserializeStruct(reader, struc)
	if err != nil {
		_ = fmt.Errorf("Error deserializing struct: %v", err)
		return nil, err
	}

	return struc, nil
}

func deserializeStructType(reader io.Reader) (uint8, error) {
	// First byte is the type of the struct
	var structType uint8
	err := read(reader, &structType)
	if err != nil {
		var zero uint8
		return zero, err
	}

	return structType, nil
}

func deserializeStruct(reader io.Reader, struc interface{}) error {
	value := reflect.ValueOf(struc)
	indirect := reflect.Indirect(value)
	if value.Kind() != reflect.Ptr || indirect.Kind() != reflect.Struct {
		return fmt.Errorf("Input is not a pointer to a struct")
	}

	for i := 0; i < indirect.NumField(); i++ {
		field := indirect.Field(i)

		if field.CanSet() {
			err := deserializeField(reader, field.Addr().Interface(), indirect.Addr().Interface())
			if err != nil {
				return err
			}
		} else {
			return fmt.Errorf("Unable to set field without panicking")
		}
	}

	return nil
}

func deserializeField(reader io.Reader, field interface{}, struc interface{}) error {
	switch data := field.(type) {
	case *uint8, *uint16, *uint32, *uint64, *int8, *int16, *int32, *int64:
		return readFixedSize(reader, data)
	case *[4]byte, *[20]byte:
		return readFixedSize(reader, data)
	case *string:
		return readString(reader, data)
	case *[][20]byte:
		return readArray(reader, data)
	default:
		return fmt.Errorf("Unsupported type: %T", data)
	}
}

func readFixedSize(reader io.Reader, data interface{}) error {
	return read(reader, data)
}

func readString(reader io.Reader, data *string) error {
	// Read the size of the string
	var size int
	err := read(reader, &size)
	if err != nil {
		return err
	}

	// Read the content of the string in bytes
	bytes := make([]byte, size)
	err = read(reader, &bytes)
	if err != nil {
		return err
	}

	*data = string(bytes)

	return nil
}

func readArray[T any](reader io.Reader, data *[]T) error {
	// Read the size of the array
	var size int
	err := read(reader, &size)
	if err != nil {
		return err
	}

	// Read the content of the array
	for i := 0; i < size; i++ {
		var element T
		err = deserializeField(reader, &element, data)
		if err != nil {
			return err
		}

		*data = append(*data, element)
	}

	return nil
}

func read(reader io.Reader, data interface{}) error {
	return binary.Read(reader, binary.LittleEndian, data)
}

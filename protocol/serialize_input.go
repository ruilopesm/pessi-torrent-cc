package protocol

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"reflect"
)

type StringReader interface {
	ReadString(reader *bytes.Reader) error
}

type SliceOfSliceByte20Reader interface {
	ReadSliceOfSliceByte20(reader *bytes.Reader) error
}

type SliceByteReader interface {
	ReadSliceByte(reader *bytes.Reader) error
}

func Deserialize(data []byte, struc interface{}) error {
	if reflect.ValueOf(struc).Kind() != reflect.Ptr || reflect.Indirect(reflect.ValueOf(struc)).Kind() != reflect.Struct {
		return fmt.Errorf("input is not a pointer to a struct")
	}

	reader := bytes.NewReader(data)
	value := reflect.Indirect(reflect.ValueOf(struc))

	for i := 0; i < value.NumField(); i++ {
		field := value.Field(i)
		if field.CanSet() {
			err := deserializeField(reader, field.Addr().Interface(), value.Addr().Interface())

			if err != nil {
				return err
			}
		} else {
			return fmt.Errorf("unable to set value in field")
		}
	}

	return nil
}

func deserializeField(reader *bytes.Reader, data interface{}, struc interface{}) error {
	switch v := data.(type) {
	case *uint8:
		return readUint8(reader, v)
	case *uint16:
		return readUint16(reader, v)
	case *[4]byte:
		return readSliceByte4(reader, v)
	case *[20]byte:
		return readSliceByte20(reader, v)
	case *string:
		return readStructString(reader, struc)
	case *[][20]byte:
		return readStructSliceOfSliceByte20(reader, struc)
	case *[]byte:
		return readStructSliceByte(reader, struc)
	default:
		return fmt.Errorf("unsupported data type: %T", data)
	}
}

// Readers for fixed types

func readUint8(reader *bytes.Reader, data *uint8) error {
	return binary.Read(reader, binary.BigEndian, data)
}

func readUint16(reader *bytes.Reader, data *uint16) error {
	return binary.Read(reader, binary.BigEndian, data)
}

func readSliceByte(reader *bytes.Reader, data *[]byte) error {
	return binary.Read(reader, binary.BigEndian, data)
}

func readSliceByte4(reader *bytes.Reader, data *[4]byte) error {
	return binary.Read(reader, binary.BigEndian, data)
}

func readSliceByte20(reader *bytes.Reader, data *[20]byte) error {
	return binary.Read(reader, binary.BigEndian, data)
}

// Readers for unfixed types

func readStructString(reader *bytes.Reader, struc interface{}) error {
	// Use reflection to check if the struct has a ReadString method
	value := reflect.ValueOf(struc)
	method := value.MethodByName("ReadString")
	if !method.IsValid() {
		return fmt.Errorf("ReadString method not found on struct")
	}

	// Call the ReadString method dynamically
	args := []reflect.Value{reflect.ValueOf(reader)}
	result := method.Call(args)

	// Check the result of the method call
	if len(result) != 1 {
		return fmt.Errorf("ReadString method should return one value")
	}

	// Check if the method returned an error
	errValue := result[0].Interface()
	if errValue != nil {
		return errValue.(error)
	}

	return nil
}

func ReadStringCallback(reader *bytes.Reader, data *string, size int) error {
	stringBytes := make([]byte, size)
	if err := binary.Read(reader, binary.BigEndian, stringBytes); err != nil {
		return err
	}
	*data = string(stringBytes)
	return nil
}

func readStructSliceOfSliceByte20(reader *bytes.Reader, struc interface{}) error {
	// Use reflection to check if the struct has a ReadSliceOfSliceByte20 method
	value := reflect.ValueOf(struc)
	method := value.MethodByName("ReadSliceOfSliceByte20")
	if !method.IsValid() {
		return fmt.Errorf("ReadSliceOfSliceByte20 method not found on struct")
	}

	// Call the ReadSliceOfSliceByte20 method dynamically
	args := []reflect.Value{reflect.ValueOf(reader)}
	result := method.Call(args)

	// Check the result of the method call
	if len(result) != 1 {
		return fmt.Errorf("ReadSliceOfSliceByte20 method should return one value")
	}

	// Check if the method returned an error
	errValue := result[0].Interface()
	if errValue != nil {
		return errValue.(error)
	}

	return nil
}

func ReadSliceOfSliceByte20Callback(reader *bytes.Reader, data *[][20]byte, size int) error {
	for i := 0; i < size; i++ {
		// Create a new array for each chunk
		var newChunkHash [20]byte
		if err := binary.Read(reader, binary.BigEndian, &newChunkHash); err != nil {
			return err
		}
		*data = append(*data, newChunkHash)
	}

	return nil
}

func readStructSliceByte(reader *bytes.Reader, struc interface{}) error {
	// Use reflection to check if the struct has a ReadSliceByte method
	value := reflect.ValueOf(struc)
	method := value.MethodByName("ReadSliceByte")
	if !method.IsValid() {
		return fmt.Errorf("ReadSliceByte method not found on struct")
	}

	// Call the ReadSliceByte method dynamically
	args := []reflect.Value{reflect.ValueOf(reader)}
	result := method.Call(args)

	// Check the result of the method call
	if len(result) != 1 {
		return fmt.Errorf("ReadSliceByte method should return one value")
	}

	// Check if the method returned an error
	errValue := result[0].Interface()
	if errValue != nil {
		return errValue.(error)
	}

	return nil
}

func ReadSliceByteCallback(reader *bytes.Reader, data *[]byte, size int) error {
	*data = make([]byte, size)
	return binary.Read(reader, binary.BigEndian, data)
}

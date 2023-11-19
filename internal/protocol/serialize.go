package protocol

import (
	"encoding/binary"
	"fmt"
	"io"
	"reflect"
)

func SerializePacket(writer io.Writer, packet Packet) error {
	// First byte is the type of the struct
	err := write(writer, packet.GetPacketType())
	if err != nil {
		return err
	}

	err = SerializeStruct(writer, packet)
	if err != nil {
		return err
	}

	return nil
}

func SerializeStruct(writer io.Writer, struc interface{}) error {
	value := reflect.ValueOf(struc)

	if value.Kind() == reflect.Ptr {
		value = value.Elem()
	}

	for i := 0; i < value.NumField(); i++ {
		field := value.Field(i)

		if field.CanInterface() {
			var err error
			if field.Type().Kind() == reflect.Struct {
				err = SerializeStruct(writer, field.Interface())
			} else {
				err = serializeField(writer, field.Interface())
			}
			if err != nil {
				return err
			}
		} else {
			return fmt.Errorf("can't interface with field without panicking")
		}
	}
	return nil
}

func DeserializePacket(reader io.Reader) (Packet, error) {
	// First byte is the type of the struct
	var structType uint8
	err := read(reader, &structType)
	if err != nil {
		return nil, err
	}

	packet := PacketStructFromType(structType)

	err = DeserializeToStruct(reader, packet)
	if err != nil {
		return nil, err
	}

	return packet, nil
}

func DeserializeToStruct(reader io.Reader, struc interface{}) error {
	value := reflect.ValueOf(struc)
	indirect := reflect.Indirect(value)
	if value.Kind() != reflect.Ptr || indirect.Kind() != reflect.Struct {
		return fmt.Errorf("input is not of type struct (type: %v) (struct: %v)", value.Type(), struc)
	}

	for i := 0; i < indirect.NumField(); i++ {
		field := indirect.Field(i)

		if field.CanSet() {
			var err error
			if field.Type().Kind() == reflect.Struct {
				err = DeserializeToStruct(reader, field.Addr().Interface())
			} else {
				err = deserializeToField(reader, field.Addr().Interface())
			}
			if err != nil {
				return fmt.Errorf("error deserializing field %v: %w", field.Interface(), err)
			}
		} else {
			return fmt.Errorf("unable to set field without panicking")
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
	case []uint16:
		return writeArray(writer, data)
	case []uint32:
		return writeArray(writer, data)
	case []uint64:
		return writeArray(writer, data)
	case [4]uint8:
		return writeArray(writer, data[:])
	case [20]uint8:
		return writeArray(writer, data[:])
	case [][20]uint8:
		return writeArray(writer, data)
	case []NodeFileInfo:
		return writeArray(writer, data)
	case NodeFileInfo:
		return SerializeStruct(writer, data)
	default:
		return fmt.Errorf("serialize unsupported type: %T", data)
	}
}

func deserializeToField(reader io.Reader, field any) error {
	switch data := field.(type) {
	case *uint8, *uint16, *uint32, *uint64, *int8, *int16, *int32, *int64:
		return read(reader, data)
	case *string:
		return readString(reader, data)
	case []uint8:
		return readArray(reader, &data)
	case *[]uint8:
		return readArray(reader, data)
	case []uint16:
		return readArray(reader, &data)
	case []uint32:
		return readArray(reader, &data)
	case []uint64:
		return readArray(reader, &data)
	case *[4]uint8:
		var array []uint8
		err := readArray(reader, &array)
		if err != nil {
			return err
		}
		copy(data[:], array)
		return err
	case *[20]uint8:
		var array []uint8
		err := readArray(reader, &array)
		if err != nil {
			return err
		}
		copy(data[:], array)
		return err
	case *[][20]uint8:
		return readArray(reader, data)
	case *[]NodeFileInfo:
		return readArray(reader, data)
	case *NodeFileInfo:
		return DeserializeToStruct(reader, data)
	default:
		return fmt.Errorf("deserialize unsupported type: %T", field)
	}
}

func writeString(writer io.Writer, data string) error {
	// Write the size of the string
	err := write(writer, uint32(len(data)))
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

func readString(reader io.Reader, str *string) error {
	// Read the size of the string
	var size uint32
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

	*str = string(bytes)

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

func readArray[T any](reader io.Reader, array *[]T) error {
	// Read the size of the array
	var size uint32
	err := read(reader, &size)
	if err != nil {
		return err
	}

	// Read the content of the array
	for i := uint32(0); i < size; i++ {
		var element T
		err = deserializeToField(reader, &element)
		if err != nil {
			return err
		}

		*array = append(*array, element)
	}

	return nil
}

func read(reader io.Reader, data interface{}) error {
	return binary.Read(reader, binary.LittleEndian, data)
}

func write(writer io.Writer, data interface{}) error {
	return binary.Write(writer, binary.LittleEndian, data)
}

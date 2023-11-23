package protocol

import (
	"bufio"
	"bytes"
	"fmt"
	"reflect"
	"testing"
)

func testSerializeStruct(struc interface{}, deserialize interface{}, t *testing.T) {
	buffer := bytes.Buffer{}
	writer := bufio.NewWriter(&buffer)

	err := SerializeStruct(writer, struc)
	if err != nil {
		t.Fatalf("error serializing struct: %v", err)
	}

	err = writer.Flush()
	if err != nil {
		t.Fatalf("error serializing struct: %v", err)
	}

	reader := bufio.NewReader(&buffer)

	err = DeserializeToStruct(reader, deserialize)
	if err != nil {
		t.Fatalf("error serializing struct: %v", err)
	}
}

func TestBytes(t *testing.T) {
	buffer := bytes.Buffer{}
	writer := bufio.NewWriter(&buffer)

	_, err := writer.Write([]byte{1, 2, 3, 4, 5})
	if err != nil {
		t.Fatalf("error writing bytes: %v", err)
	}

	err = writer.Flush()
	if err != nil {
		t.Fatalf("error flushing buffer: %v", err)
	}

	reader := bufio.NewReader(&buffer)

	var result []byte
	for i := 0; i < 5; i++ {
		b, err := reader.ReadByte()
		if err != nil {
			t.Fatalf("error reading byte: %v", err)
		}
		result = append(result, b)
	}

	if !bytes.Equal(result, []byte{1, 2, 3, 4, 5}) {
		t.Fatalf("result != []byte{1, 2, 3, 4, 5}")
	}
}

func TestSerialize(t *testing.T) {
	// create dummy PublishFilePacket
	packet := NewPublishFilePacket("test.txt", [20]byte{1, 2, 3, 4, 5}, [][20]byte{{6, 7, 8}, {9, 10, 11}})

	var deserialize PublishFilePacket
	testSerializeStruct(&packet, &deserialize, t)
	checkEquals(packet, deserialize, t)

	// create dummy InitPacket
	initPacket := NewInitPacket([4]byte{1, 2, 3, 4}, 1234)

	var deserializeInit InitPacket
	testSerializeStruct(&initPacket, &deserializeInit, t)
	checkEquals(initPacket, deserializeInit, t)

	// create dummy PublishChunkPacket
	publishChunkPacket := NewPublishChunkPacket([20]byte{1, 2, 3, 4, 5}, []uint16{1, 2, 3, 4, 5})

	var deserializePublishChunk PublishChunkPacket
	testSerializeStruct(&publishChunkPacket, &deserializePublishChunk, t)
	checkEquals(publishChunkPacket, deserializePublishChunk, t)

	// create dummy AnswerNodesPacket
	answerNodesPacket := NewAnswerNodesPacket("filename.txt", [20]byte{1, 2, 3, 4, 5}, 1, [][4]byte{{1, 2, 3, 4}}, []uint16{1, 2, 3, 4, 5}, [][]uint16{{1, 2, 3, 4, 5}})

	var deserializeAnswerNodes AnswerNodesPacket
	testSerializeStruct(&answerNodesPacket, &deserializeAnswerNodes, t)
	checkEquals(answerNodesPacket, deserializeAnswerNodes, t)
}

func checkEquals(a interface{}, b interface{}, t *testing.T) {
	fmt.Printf("a: %v\n", a)
	fmt.Printf("b: %v\n", b)
	if !reflect.DeepEqual(a, b) {
		t.Fatalf("a != b")
	}
}

func TestSerializeStructInsideStruct(t *testing.T) {
	type TestStruct struct {
		InnerStruct struct {
			Number uint8
		}
		Number uint8
	}

	var struc TestStruct

	struc.InnerStruct.Number = 1
	struc.Number = 2

	var deserialize TestStruct

	testSerializeStruct(&struc, &deserialize, t)
	checkEquals(struc, deserialize, t)
}
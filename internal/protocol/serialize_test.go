package protocol

import (
	"bufio"
	"bytes"
	"testing"
)

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
	var packet PublishFilePacket
	packet.Create("test.txt", [20]byte{1, 2, 3, 4, 5}, [][20]byte{{6, 7, 8}, {9, 10, 11}})

	//create dummy writer
	buffer := bytes.Buffer{}

	writer := bufio.NewWriter(&buffer)

	// serialize packet
	err := Serialize(writer, &packet)
	if err != nil {
		t.Fatalf("error serializing packet: %v", err)
	}

	// flush buffer
	err = writer.Flush()
	if err != nil {
		t.Fatalf("error flushing buffer: %v", err)
	}

	// create dummy reader
	reader := bufio.NewReader(&buffer)

	// deserialize packet
	packet2, err := Deserialize(reader)
	if err != nil {
		t.Fatalf("error deserializing packet: %v", err)
	}
	var deserialized = packet2.(*PublishFilePacket)

	if packet.NameSize != deserialized.NameSize {
		t.Fatalf("packet.NameSize != deserialized.NameSize")
	}

	if packet.NumberOfChunks != deserialized.NumberOfChunks {
		t.Fatalf("packet.NumberOfChunks != deserialized.NumberOfChunks")
	}

	if packet.FileHash != deserialized.FileHash {
		t.Fatalf("packet.FileHash != deserialized.FileHash")
	}

	if packet.FileName != deserialized.FileName {
		t.Fatalf("packet.FileName != deserialized.FileName")
	}

	if len(packet.ChunkHashes) != len(deserialized.ChunkHashes) {
		t.Fatalf("len(packet.ChunkHashes) != len(deserialized.ChunkHashes)")
	}

	for i := 0; i < len(packet.ChunkHashes); i++ {
		if packet.ChunkHashes[i] != deserialized.ChunkHashes[i] {
			t.Fatalf("packet.ChunkHashes[i] != deserialized.ChunkHashes[i]")
		}
	}
}

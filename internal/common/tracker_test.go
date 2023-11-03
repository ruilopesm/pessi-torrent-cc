package common

import (
	"PessiTorrent/internal/serialization"
	"crypto/sha1"
	"testing"
)

func TestPublishFileSerializationAndDeserialization(t *testing.T) {
	// Create a sample PublishFilePacket instance

	chunkHashes := [][20]byte{
		sha1.Sum([]byte("chico")),
		sha1.Sum([]byte("daniel")),
		sha1.Sum([]byte("rui")),
	}
  
	var original PublishFilePacket
	original.Create("filename.txt", sha1.Sum([]byte("hello world")), chunkHashes)

	// Serialize the PublishFilePacket
  
	serializedData, err := serialization.Serialize(original)
	if err != nil {
		t.Errorf("Serialization error: %v", err)
		return
	}

	// Deserialize the data back into a PublishFilePacket
	var deserialized PublishFilePacket
  
	err = serialization.Deserialize(serializedData, &deserialized)
	if err != nil {
		t.Errorf("Deserialization error: %v", err)
		return
	}

	// Compare the original and deserialized structs
	if original.NameSize != deserialized.NameSize ||
		original.NumberOfChunks != deserialized.NumberOfChunks ||
		original.FileName != deserialized.FileName {
		t.Errorf("Serialization and deserialization do not match")
	}

	for i := 0; i < int(original.NumberOfChunks); i++ {
		if original.ChunkHashes[i] != deserialized.ChunkHashes[i] {
			t.Errorf("ChunkHashes do not match at index %d", i)
		}
	}
}

func TestPublishChunkSerializationAndDeserialization(t *testing.T) {
	// Create a sample PublishChunkPacket instance
	fileHash := sha1.Sum([]byte("example"))
	bitfield := []uint8{0, 2, 7, 10}

	var original PublishChunkPacket
	original.Create(fileHash, bitfield)

	// Serialize the PublishChunkPacket
	serializedData, err := serialization.Serialize(original)
	if err != nil {
		t.Errorf("Serialization error: %v", err)
		return
	}

	// Deserialize the data back into a PublishChunkPacket
	var deserialized PublishChunkPacket
	err = serialization.Deserialize(serializedData, &deserialized)
	if err != nil {
		t.Errorf("Deserialization error: %v", err)
		return
	}

	if len(original.FileHash) != len(deserialized.FileHash) {
		t.Errorf("FileHash length mismatch")
	}

	for i := 0; i < len(original.FileHash); i++ {
		if original.FileHash[i] != deserialized.FileHash[i] {
			t.Errorf("FileHash values do not match at index %d", i)
		}
	}

	if len(original.Bitfield) != len(deserialized.Bitfield) {
		t.Errorf("Bitfield length mismatch")
	}

	for i := 0; i < len(original.Bitfield); i++ {
		if original.Bitfield[i] != deserialized.Bitfield[i] {
			t.Errorf("Bitfield values do not match at index %d", i)
		}
	}
}

func TestRequestFileSerializationAndDeserialization(t *testing.T) {
	// Create a sample RequestFilePacket instance
	var original RequestFilePacket
	original.Create("example.txt")

	// Serialize the RequestFilePacket
	serializedData, err := serialization.Serialize(original)
	if err != nil {
		t.Errorf("Serialization error: %v", err)
		return
	}

	// Deserialize the data back into a RequestFilePacket
	var deserialized RequestFilePacket
	err = serialization.Deserialize(serializedData, &deserialized)
  
	if err != nil {
		t.Errorf("Deserialization error: %v", err)
		return
	}

	// Compare the original and deserialized structs
	if original.NameSize != deserialized.NameSize ||
		original.FileName != deserialized.FileName {
		t.Errorf("Serialization and deserialization do not match")
	}
}

func TestAlreadyExistsSerializationAndDeserialization(t *testing.T) {
	// Create a sample RequestFilePacket instance
	var original AlreadyExistsPacket
	original.Create("example.txt")

	// Serialize the RequestFilePacket
	serializedData, err := serialization.Serialize(original)
	if err != nil {
		t.Errorf("Serialization error: %v", err)
		return
	}

	// Deserialize the data back into a RequestFilePacket
	var deserialized AlreadyExistsPacket
	err = serialization.Deserialize(serializedData, &deserialized)
	if err != nil {
		t.Errorf("Deserialization error: %v", err)
		return
	}

	// Compare the original and deserialized structs
	if original.NameSize != deserialized.NameSize ||
		original.FileName != deserialized.FileName {
		t.Errorf("Serialization and deserialization do not match")
	}
}

func TestAnswerNodesSerializationAndDeserialization(t *testing.T) {
	// Create a sample AnswerNodesPacket instance
	bitfield := []uint8{0, 2, 7, 10}
	nodeIdentifier := [4]byte{128, 1, 1, 1}

	var original AnswerNodesPacket
	original.Create(42, nodeIdentifier, uint16(8081), bitfield)

	// Serialize the AnswerNodesPacket
	serializedData, err := serialization.Serialize(original)
	if err != nil {
		t.Errorf("Serialization error: %v", err)
		return
	}

	// Deserialize the data back into an AnswerNodesPacket
	var deserialized AnswerNodesPacket
	err = serialization.Deserialize(serializedData, &deserialized)
	if err != nil {
		t.Errorf("Deserialization error: %v", err)
		return
	}

	// Compare the original and deserialized structs
	if original.SequenceNumber != deserialized.SequenceNumber ||
		original.BitfieldSize != deserialized.BitfieldSize ||
		original.NodePort != deserialized.NodePort {
		t.Errorf("Serialization and deserialization do not match")
	}

	if len(original.NodeIdentifier) != len(deserialized.NodeIdentifier) {
		t.Errorf("NodeIdentifier length mismatch")
	}

	for i := 0; i < len(original.NodeIdentifier); i++ {
		if original.NodeIdentifier[i] != deserialized.NodeIdentifier[i] {
			t.Errorf("NodeIdentifier values do not match at index %d", i)
		}
	}

	if len(original.Bitfield) != len(deserialized.Bitfield) {
		t.Errorf("Bitfield length mismatch")
	}

	for i := 0; i < len(original.Bitfield); i++ {
		if original.Bitfield[i] != deserialized.Bitfield[i] {
			t.Errorf("Bitfield values do not match at index %d", i)
		}
	}
}

func TestSetBitsAndGetBit(t *testing.T) {
	bitfield := make([]uint8, 10)

	serialization.SetBit(bitfield, 0)
	if serialization.GetBit(bitfield, 0) == false {
		t.Error("Bit in position 0 not set to 1")
	}

	serialization.SetBit(bitfield, 5)
	if serialization.GetBit(bitfield, 5) == false {
		t.Error("Bit in position 5 not set to 1")
	}

	serialization.SetBit(bitfield, 10)
	if serialization.GetBit(bitfield, 10) == false {
		t.Error("Bit in position 10 not set to 1")
	}
}

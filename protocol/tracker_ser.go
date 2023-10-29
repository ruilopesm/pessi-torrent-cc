package protocol

import (
	"bytes"
	"encoding/binary"
)

type Serializer interface {
	Serialize() ([]byte, error)
	Deserialize([]byte) error
}

// PublishFile Packet

func (pf *PublishFile) Serialize() ([]byte, error) {
	buf := new(bytes.Buffer)

	// Serialize Packet Type
	binary.Write(buf, binary.BigEndian, uint8(1))
	// Serialize NameSize
	binary.Write(buf, binary.BigEndian, &pf.NameSize)
	// Serialize NumberOfChunks
	binary.Write(buf, binary.BigEndian, &pf.NumberOfChunks)
	// Serialize FileHash
	binary.Write(buf, binary.BigEndian, &pf.FileHash)
	// Serialize ChunkHashes
	for i := 0; i < int(pf.NumberOfChunks); i++ {
		binary.Write(buf, binary.BigEndian, &pf.ChunkHashes[i])
	}
	// Serialize FileName
	stringBytes := make([]byte, pf.NameSize)
	copy(stringBytes, []byte(pf.FileName))
	binary.Write(buf, binary.BigEndian, stringBytes)

	return buf.Bytes(), nil
}

func (pf *PublishFile) Deserialize(data []byte) error {
	buf := bytes.NewReader(data)
	// Deserialize Packet Type
	var packetType uint8
	binary.Read(buf, binary.BigEndian, &packetType)
	// Deserialize NameSize
	binary.Read(buf, binary.BigEndian, &pf.NameSize)
	// Deserialize NumberOfChunks
	binary.Read(buf, binary.BigEndian, &pf.NumberOfChunks)
	// Deserialize FileHash
	binary.Read(buf, binary.BigEndian, &pf.FileHash)
	// Deserialize ChunkHashes
	for i := 0; i < int(pf.NumberOfChunks); i++ {
		// Create a new array for each chunk
		var newChunkHash [20]byte
		binary.Read(buf, binary.BigEndian, &newChunkHash)
		// Append the newChunkHash to pf.ChunkHashes
		pf.ChunkHashes = append(pf.ChunkHashes, newChunkHash)
	}
	// Deserialize FileName
	fileNameBytes := make([]byte, pf.NameSize)
	binary.Read(buf, binary.BigEndian, &fileNameBytes)
	pf.FileName = string(fileNameBytes)

	return nil
}

// PublishChunk Packet

func (pc *PublishChunk) Serialize() ([]byte, error) {
	buf := new(bytes.Buffer)

	// Serialize Packet Type
	binary.Write(buf, binary.BigEndian, uint8(3))
	// Serialize Reserved
	binary.Write(buf, binary.BigEndian, uint8(0))
	// Serialize BitfieldSize
	binary.Write(buf, binary.BigEndian, pc.BitfieldSize)
	// Serialize FileHash
	binary.Write(buf, binary.BigEndian, pc.FileHash)
	// Serialize Bitfield
	binary.Write(buf, binary.BigEndian, pc.Bitfield)

	return buf.Bytes(), nil
}

func (pc *PublishChunk) Deserialize(data []byte) error {
	buf := bytes.NewReader(data)

	// Deserialize Packet Type
	var packetType uint8
	binary.Read(buf, binary.BigEndian, &packetType)
	// Deserialize Reserved
	var reserved uint8
	binary.Read(buf, binary.BigEndian, &reserved)
	// Deserialize BitfieldSize
	binary.Read(buf, binary.BigEndian, &pc.BitfieldSize)
	// Deserialize FileHash
	binary.Read(buf, binary.BigEndian, &pc.FileHash)
	// Deserialize Bitfield
	pc.Bitfield = make([]byte, pc.BitfieldSize)
	binary.Read(buf, binary.BigEndian, &pc.Bitfield)

	return nil
}

// RequestFile Packet

func (rf *RequestFile) Serialize() ([]byte, error) {
	buf := new(bytes.Buffer)

	// Serialize Packet Type
	binary.Write(buf, binary.BigEndian, uint8(4))
	// Serialize NameSize
	binary.Write(buf, binary.BigEndian, rf.NameSize)
	// Serialize FileName
	stringBytes := make([]byte, rf.NameSize)
	copy(stringBytes, []byte(rf.FileName))
	binary.Write(buf, binary.BigEndian, stringBytes)

	return buf.Bytes(), nil
}

func (rf *RequestFile) Deserialize(data []byte) error {
	buf := bytes.NewReader(data)

	// Deserialize Packet Type
	var packetType uint8
	binary.Read(buf, binary.BigEndian, &packetType)
	// Deserialize NameSize
	binary.Read(buf, binary.BigEndian, &rf.NameSize)
	// Deserialize FileName
	fileNameBytes := make([]byte, rf.NameSize)
	binary.Read(buf, binary.BigEndian, &fileNameBytes)
	rf.FileName = string(fileNameBytes)

	return nil
}

// AlreadyExists Packet

func (ae *AlreadyExists) Serialize() ([]byte, error) {
	buf := new(bytes.Buffer)

	// Serialize Packet Type
	binary.Write(buf, binary.BigEndian, uint8(2))
	// Serialize NameSize
	binary.Write(buf, binary.BigEndian, ae.NameSize)
	// Serialize Reserved bits
	binary.Write(buf, binary.BigEndian, uint16(0))
	// Serialize FileName
	stringBytes := make([]byte, ae.NameSize)
	copy(stringBytes, []byte(ae.FileName))
	binary.Write(buf, binary.BigEndian, stringBytes)

	return buf.Bytes(), nil
}

func (ae *AlreadyExists) Deserialize(data []byte) error {
	buf := bytes.NewReader(data)

	// Deserialize Packet Type
	var packetType uint8
	binary.Read(buf, binary.BigEndian, &packetType)
	// Deserialize NameSize
	binary.Read(buf, binary.BigEndian, &ae.NameSize)
	// Deserialize Reserved bits
	var reserved uint16
	binary.Read(buf, binary.BigEndian, &reserved)
	// Deserialize FileName
	fileNameBytes := make([]byte, ae.NameSize)
	binary.Read(buf, binary.BigEndian, &fileNameBytes)
	ae.FileName = string(fileNameBytes)

	return nil
}

// AnswerNodes Packet

func (an *AnswerNodes) Serialize() ([]byte, error) {
	buf := new(bytes.Buffer)

	// Serialize Packet Type
	binary.Write(buf, binary.BigEndian, uint8(5))
	// Serialize SequenceNumber
	binary.Write(buf, binary.BigEndian, an.SequenceNumber)
	// Serialize BitfieldSize
	binary.Write(buf, binary.BigEndian, an.BitfieldSize)
	// Serialize reserved
	binary.Write(buf, binary.BigEndian, uint16(0))
	// Serialize NodeIdentifier (IP address)
	binary.Write(buf, binary.BigEndian, an.NodeIdentifier)
	// Serialize NodePort
	binary.Write(buf, binary.BigEndian, an.NodePort)
	// Serialize Bitfield
	binary.Write(buf, binary.BigEndian, an.Bitfield)

	return buf.Bytes(), nil
}

func (an *AnswerNodes) Deserialize(data []byte) error {
	buf := bytes.NewReader(data)

	// Deserialize Packet Type
	var packetType uint8
	binary.Read(buf, binary.BigEndian, &packetType)
	// Deserialize SequenceNumber
	binary.Read(buf, binary.BigEndian, &an.SequenceNumber)
	// Deserialize BitfieldSize
	binary.Read(buf, binary.BigEndian, &an.BitfieldSize)
	// Deserialize reserved
	var reserved uint16
	binary.Read(buf, binary.BigEndian, &reserved)
	// Deserialize NodeIdentifier (IP address)
	binary.Read(buf, binary.BigEndian, &an.NodeIdentifier)
	// Deserialize NodePort
	binary.Read(buf, binary.BigEndian, &an.NodePort)
	// Deserialize Bitfield
	an.Bitfield = make([]byte, an.BitfieldSize)
	binary.Read(buf, binary.BigEndian, &an.Bitfield)

	return nil
}

func encodeBitField(bitfield []uint8) []byte {
	lastElement := bitfield[len(bitfield)-1]
	var size int
	size = (int(lastElement) / 8) + 1

	binaryBitfield := make([]byte, size)

	for value := range bitfield {
		setBit(binaryBitfield, int(bitfield[value]))
	}

	return binaryBitfield
}

func decodeBitField(binaryBitfield []byte) []uint8 {
	var bitfield []uint8
	size := len(binaryBitfield)

	for i := 0; i < size*8; i++ {
		if getBit(binaryBitfield, i) {
			bitfield = append(bitfield, uint8(i))
		}
	}

	return bitfield
}

// setBit sets the bit at the specified position to 1 (starting at 0).
func setBit(bitfield []byte, position int) {
	offset := int(position / 8)
	value := bitfield[offset]
	index := position - (8 * offset)

	mask := 1 << (7 - index)

	bitfield[offset] = value | uint8(mask)
}

// GetBit returns true if the bit value at a given position in the bitfield is set to 1 (starting at 0)
func getBit(bitfield []byte, position int) bool {
	offset := int(position / 8)
	value := bitfield[offset]
	index := position - (8 * offset)

	mask := uint8(1 << (7 - index))

	return (value & mask) == mask
}

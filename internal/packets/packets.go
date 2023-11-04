package packets

import (
	"PessiTorrent/internal/serialization"
	"bytes"
)

// NODE -> TRACKER

type PublishFilePacket struct {
	Type           uint8
	NameSize       uint8
	NumberOfChunks uint16
	FileHash       [20]byte
	FileName       string
	ChunkHashes    [][20]byte
}

func (pf *PublishFilePacket) Create(name string, fileHash [20]byte, chunkHashes [][20]byte) {
	pf.Type = uint8(PUBLISH_FILE_TYPE)
	pf.NameSize = uint8(len(name))
	pf.NumberOfChunks = uint16(len(chunkHashes))
	pf.FileHash = fileHash
	pf.FileName = name
	pf.ChunkHashes = chunkHashes
}

func (pf *PublishFilePacket) ReadString(reader *bytes.Reader) error {
	return serialization.ReadStringCallback(reader, &pf.FileName, int(pf.NameSize))
}

func (pf *PublishFilePacket) ReadSliceOfSliceByte20(reader *bytes.Reader) error {
	return serialization.ReadSliceOfSliceByte20Callback(reader, &pf.ChunkHashes, int(pf.NumberOfChunks))
}

type PublishChunkPacket struct {
	Type         uint8
	BitfieldSize uint16
	Reserved     uint8
	FileHash     [20]byte
	Bitfield     []byte
}

func (pc *PublishChunkPacket) Create(fileHash [20]byte, bitfield []uint8) {
	binaryBitField := serialization.EncodeBitField(bitfield)
	bitfieldSize := len(binaryBitField)

	pc.Type = uint8(PUBLISH_CHUNK_TYPE)
	pc.BitfieldSize = uint16(bitfieldSize)
	pc.Reserved = uint8(0)
	pc.FileHash = fileHash
	pc.Bitfield = binaryBitField
}

func (pc *PublishChunkPacket) ReadSliceByte(reader *bytes.Reader) error {
	return serialization.ReadSliceByteCallback(reader, &pc.Bitfield, int(pc.BitfieldSize))
}

type RequestFilePacket struct {
	Type     uint8
	NameSize uint8
	Reserved uint16
	FileName string
}

func (rf *RequestFilePacket) Create(fileName string) {
	rf.Type = uint8(REQUEST_FILE_TYPE)
	rf.NameSize = uint8(len(fileName))
	rf.Reserved = uint16(0)
	rf.FileName = fileName
}

func (rf *RequestFilePacket) ReadString(reader *bytes.Reader) error {
	return serialization.ReadStringCallback(reader, &rf.FileName, int(rf.NameSize))
}

// TRACKER -> NODE

type AlreadyExistsPacket struct {
	Type     uint8
	NameSize uint8
	Reserved uint16
	FileName string
}

func (ae *AlreadyExistsPacket) Create(fileName string) {
	ae.Type = uint8(ALREADY_EXISTS_TYPE)
	ae.NameSize = uint8(len(fileName))
	ae.Reserved = uint16(0)
	ae.FileName = fileName
}

func (ae *AlreadyExistsPacket) ReadString(reader *bytes.Reader) error {
	return serialization.ReadStringCallback(reader, &ae.FileName, int(ae.NameSize))
}

type AnswerNodesPacket struct {
	Type           uint8
	SequenceNumber uint8
	BitfieldSize   uint16
	Reserved       uint16
	NodeIdentifier [4]byte
	NodePort       uint16
	Bitfield       []byte
}

func (an *AnswerNodesPacket) Create(sequenceNumber uint8, nodeIdentifier [4]byte, nodePort uint16, bitfield []uint8) {
	binaryBitField := serialization.EncodeBitField(bitfield)
	bitfieldSize := len(binaryBitField)

	an.Type = uint8(ANSWER_NODES_TYPE)
	an.SequenceNumber = sequenceNumber
	an.BitfieldSize = uint16(bitfieldSize)
	an.Reserved = uint16(0)
	an.NodeIdentifier = nodeIdentifier
	an.NodePort = nodePort
	an.Bitfield = binaryBitField
}

func (an *AnswerNodesPacket) ReadSliceByte(reader *bytes.Reader) error {
	return serialization.ReadSliceByteCallback(reader, &an.Bitfield, int(an.BitfieldSize))
}

package common

import (
	"PessiTorrent/internal/serialization"
	"bytes"
)

const publishFileType = 1
const alreadyExistsType = 2
const publishChunkType = 3
const requestFileType = 4
const answerNodesType = 5

// NODE -> TRACKER

type PublishFile struct {
	Type           uint8
	NameSize       uint8
	NumberOfChunks uint16
	FileHash       [20]byte
	FileName       string
	ChunkHashes    [][20]byte
}

func (pf *PublishFile) Create(name string, fileHash [20]byte, chunkHashes [][20]byte) {
	pf.Type = uint8(publishFileType)
	pf.NameSize = uint8(len(name))
	pf.NumberOfChunks = uint16(len(chunkHashes))
	pf.FileHash = fileHash
	pf.FileName = name
	pf.ChunkHashes = chunkHashes
}

func (pf *PublishFile) ReadString(reader *bytes.Reader) error {
	return serialization.ReadStringCallback(reader, &pf.FileName, int(pf.NameSize))
}

func (pf *PublishFile) ReadSliceOfSliceByte20(reader *bytes.Reader) error {
	return serialization.ReadSliceOfSliceByte20Callback(reader, &pf.ChunkHashes, int(pf.NumberOfChunks))
}

type PublishChunk struct {
	Type         uint8
	BitfieldSize uint16
	Reserved     uint8
	FileHash     [20]byte
	Bitfield     []byte
}

func (pc *PublishChunk) Create(fileHash [20]byte, bitfield []uint8) {
	binaryBitField := serialization.EncodeBitField(bitfield)
	bitfieldSize := len(binaryBitField)

	pc.Type = uint8(publishChunkType)
	pc.BitfieldSize = uint16(bitfieldSize)
	pc.Reserved = uint8(0)
	pc.FileHash = fileHash
	pc.Bitfield = binaryBitField
}

func (pc *PublishChunk) ReadSliceByte(reader *bytes.Reader) error {
	return serialization.ReadSliceByteCallback(reader, &pc.Bitfield, int(pc.BitfieldSize))
}

type RequestFile struct {
	Type     uint8
	NameSize uint8
	Reserved uint16
	FileName string
}

func (rf *RequestFile) Create(fileName string) {
	rf.Type = uint8(requestFileType)
	rf.NameSize = uint8(len(fileName))
	rf.Reserved = uint16(0)
	rf.FileName = fileName
}

func (rf *RequestFile) ReadString(reader *bytes.Reader) error {
	return serialization.ReadStringCallback(reader, &rf.FileName, int(rf.NameSize))
}

// // TRACKER -> NODE

type AlreadyExists struct {
	Type     uint8
	NameSize uint8
	Reserved uint16
	FileName string
}

func (ae *AlreadyExists) Create(fileName string) {
	ae.Type = uint8(alreadyExistsType)
	ae.NameSize = uint8(len(fileName))
	ae.Reserved = uint16(0)
	ae.FileName = fileName
}

func (ae *AlreadyExists) ReadString(reader *bytes.Reader) error {
	return serialization.ReadStringCallback(reader, &ae.FileName, int(ae.NameSize))
}

type AnswerNodes struct {
	Type           uint8
	SequenceNumber uint8
	BitfieldSize   uint16
	Reserved       uint16
	NodeIdentifier [4]byte
	NodePort       uint16
	Bitfield       []byte
}

func (an *AnswerNodes) Create(sequenceNumber uint8, nodeIdentifier [4]byte, nodePort uint16, bitfield []uint8) {
	binaryBitField := serialization.EncodeBitField(bitfield)
	bitfieldSize := len(binaryBitField)

	an.Type = uint8(answerNodesType)
	an.SequenceNumber = sequenceNumber
	an.BitfieldSize = uint16(bitfieldSize)
	an.Reserved = uint16(0)
	an.NodeIdentifier = nodeIdentifier
	an.NodePort = nodePort
	an.Bitfield = binaryBitField
}

func (an *AnswerNodes) ReadSliceByte(reader *bytes.Reader) error {
	return serialization.ReadSliceByteCallback(reader, &an.Bitfield, int(an.BitfieldSize))
}

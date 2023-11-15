package packets

import (
	"PessiTorrent/internal/serialization"
	"bytes"
)

// NODE -> TRACKER

type InitPacket struct {
	Type     uint8
	Reserved uint8
	UDPPort  uint16
	IPAddr   [4]byte
}

func (ip *InitPacket) Create(ipAddr [4]byte, udpPort uint16) {
	ip.Type = uint8(InitType)
	ip.Reserved = uint8(0)
	ip.UDPPort = udpPort
	ip.IPAddr = ipAddr
}

type PublishFilePacket struct {
	Type           uint8
	NameSize       uint8
	NumberOfChunks uint16
	FileHash       [20]byte
	FileName       string
	ChunkHashes    [][20]byte
}

func (pf *PublishFilePacket) Create(name string, fileHash [20]byte, chunkHashes [][20]byte) {
	pf.Type = uint8(PublishFileType)
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

func (pc *PublishChunkPacket) Create(fileHash [20]byte, bitfield []uint16) {
	binaryBitField := serialization.EncodeBitField(bitfield)
	bitfieldSize := len(binaryBitField)

	pc.Type = uint8(PublishChunkType)
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
	rf.Type = uint8(RequestFileType)
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
	ae.Type = uint8(AlreadyExistsType)
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
	UDPPort        uint16
	BitfieldSize   uint16
	Reserved       uint16
	NodeIPAddr     [4]byte
	Bitfield       []byte
}

func (an *AnswerNodesPacket) Create(sequenceNumber uint8, ipAddr [4]byte, udpPort uint16, bitfield []uint16) {
	binaryBitField := serialization.EncodeBitField(bitfield)
	bitfieldSize := len(binaryBitField)

	an.Type = uint8(AnswerNodesType)
	an.SequenceNumber = sequenceNumber
	an.UDPPort = udpPort
	an.BitfieldSize = uint16(bitfieldSize)
	an.Reserved = uint16(0)
	an.NodeIPAddr = ipAddr
	an.Bitfield = binaryBitField
}

type RemoveFilePacket struct {
	Type     uint8
	NameSize uint8
	Reserved uint16
	FileName string
}

func (rf *RemoveFilePacket) Create(fileName string) {
	rf.Type = uint8(RemoveFileType)
	rf.NameSize = uint8(len(fileName))
	rf.Reserved = uint16(0)
	rf.FileName = fileName
}

func (rf *RemoveFilePacket) ReadString(reader *bytes.Reader) error {
	return serialization.ReadStringCallback(reader, &rf.FileName, int(rf.NameSize))
}

func (an *AnswerNodesPacket) ReadSliceByte(reader *bytes.Reader) error {
	return serialization.ReadSliceByteCallback(reader, &an.Bitfield, int(an.BitfieldSize))
}

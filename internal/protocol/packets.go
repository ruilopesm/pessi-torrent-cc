package protocol

// NODE -> TRACKER

type InitPacket struct {
	UDPPort uint16
	IPAddr  [4]byte
}

func NewInitPacket(ipAddr [4]byte, udpPort uint16) InitPacket {
	return InitPacket{
		UDPPort: udpPort,
		IPAddr:  ipAddr,
	}
}

func (ip *InitPacket) GetPacketType() uint8 {
	return InitType
}

type PublishFilePacket struct {
	NameSize       uint8
	NumberOfChunks uint16
	FileHash       [20]byte
	FileName       string
	ChunkHashes    [][20]byte
}

func NewPublishFilePacket(fileName string, fileHash [20]byte, chunkHashes [][20]byte) PublishFilePacket {
	return PublishFilePacket{
		NameSize:       uint8(len(fileName)),
		NumberOfChunks: uint16(len(chunkHashes)),
		FileHash:       fileHash,
		FileName:       fileName,
		ChunkHashes:    chunkHashes,
	}
}

func (pf *PublishFilePacket) GetPacketType() uint8 {
	return PublishFileType
}

type PublishChunkPacket struct {
	BitfieldSize uint16
	FileHash     [20]byte
	Bitfield     []uint8
}

func NewPublishChunkPacket(fileHash [20]byte, bitfield []uint16) PublishChunkPacket {
	binaryBitfield := EncodeBitField(bitfield)
	bitfieldSize := len(binaryBitfield)

	return PublishChunkPacket{
		BitfieldSize: uint16(bitfieldSize),
		FileHash:     fileHash,
		Bitfield:     binaryBitfield,
	}
}

func (pc *PublishChunkPacket) GetPacketType() uint8 {
	return PublishChunkType
}

type RequestFilePacket struct {
	NameSize uint8
	FileName string
}

func NewRequestFilePacket(fileName string) RequestFilePacket {
	return RequestFilePacket{
		NameSize: uint8(len(fileName)),
		FileName: fileName,
	}
}

func (rf *RequestFilePacket) GetPacketType() uint8 {
	return RequestFileType
}

// TRACKER -> NODE

type AlreadyExistsPacket struct {
	NameSize uint8
	Filename string
}

func NewAlreadyExistsPacket(filename string) AlreadyExistsPacket {
	return AlreadyExistsPacket{
		NameSize: uint8(len(filename)),
		Filename: filename,
	}
}

func (ae *AlreadyExistsPacket) GetPacketType() uint8 {
	return AlreadyExistsType
}

type AnswerNodesPacket struct {
	NumberOfNodes uint16
	Nodes         []NodeFileInfo
}

type NodeFileInfo struct {
	BitfieldSize uint16
	Port         uint16
	IPAddr       [4]byte
	Bitfield     []uint8
}

func NewAnswerNodesPacket(nNodes uint16, ipAddrs [][4]byte, ports []uint16, bitfields [][]uint16) AnswerNodesPacket {
	an := AnswerNodesPacket{
		NumberOfNodes: nNodes,
	}

	for i := 0; i < int(nNodes); i++ {
		bitfield := EncodeBitField(bitfields[i])

		node := NodeFileInfo{
			BitfieldSize: uint16(len(bitfield)),
			IPAddr:       ipAddrs[i],
			Port:         ports[i],
			Bitfield:     bitfield,
		}
		an.Nodes = append(an.Nodes, node)
	}

	return an
}

func (an *AnswerNodesPacket) GetPacketType() uint8 {
	return AnswerNodesType
}

type RemoveFilePacket struct {
	NameSize uint8
	FileName string
}

func NewRemoveFilePacket(fileName string) RemoveFilePacket {
	return RemoveFilePacket{
		NameSize: uint8(len(fileName)),
		FileName: fileName,
	}
}

func (rf *RemoveFilePacket) GetPacketType() uint8 {
	return RemoveFileType
}

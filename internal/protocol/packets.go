package protocol

// NODE -> TRACKER

type InitPacket struct {
	UDPPort uint16
	IPAddr  [4]byte
}

func (ip *InitPacket) GetPacketType() uint8 {
	return InitType
}

func NewInitPacket(ipAddr [4]byte, udpPort uint16) InitPacket {
	return InitPacket{
		UDPPort: udpPort,
		IPAddr:  ipAddr,
	}
}

type PublishFilePacket struct {
	NameSize       uint8
	NumberOfChunks uint16
	FileHash       [20]byte
	FileName       string
	ChunkHashes    [][20]byte
}

func (pf *PublishFilePacket) GetPacketType() uint8 {
	return PublishFileType
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

type PublishChunkPacket struct {
	BitfieldSize uint16
	FileHash     [20]byte
	Bitfield     []byte
}

func (pc *PublishChunkPacket) GetPacketType() uint8 {
	return PublishChunkType
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

type RequestFilePacket struct {
	NameSize uint8
	FileName string
}

func (rf *RequestFilePacket) GetPacketType() uint8 {
	return RequestFileType
}

func NewRequestFIlePacket(fileName string) RequestFilePacket {
	return RequestFilePacket{
		NameSize: uint8(len(fileName)),
		FileName: fileName,
	}
}

// TRACKER -> NODE

type AlreadyExistsPacket struct {
	NameSize uint8
	FileName string
}

func (ae *AlreadyExistsPacket) GetPacketType() uint8 {
	return AlreadyExistsType
}

func NewAlreadyExistsPacket(fileName string) AlreadyExistsPacket {
	return AlreadyExistsPacket{
		NameSize: uint8(len(fileName)),
		FileName: fileName,
	}
}

type AnswerNodesPacket struct {
	NumberOfNodes uint16
	Nodes         []NodeFileInfo
}

type NodeFileInfo struct {
	BitfieldSize uint16
	Port         uint16
	IPAddr       [4]byte
	Bitfield     []byte
}

func (an *AnswerNodesPacket) GetPacketType() uint8 {
	return AnswerNodesType
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

type RemoveFilePacket struct {
	NameSize uint8
	FileName string
}

func (rf *RemoveFilePacket) GetPacketType() uint8 {
	return RemoveFileType
}

func NewRemoveFilePacket(fileName string) RemoveFilePacket {
	return RemoveFilePacket{
		NameSize: uint8(len(fileName)),
		FileName: fileName,
	}
}

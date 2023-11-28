package protocol

// NODE -> TRACKER

type InitPacket struct {
	Name    string
	UDPPort uint16
}

func NewInitPacket(name string, udpPort uint16) InitPacket {
	return InitPacket{
		Name:    name,
		UDPPort: udpPort,
	}
}

func (ip *InitPacket) GetPacketType() uint8 {
	return InitType
}

type PublishFilePacket struct {
	FileName    string
	FileSize    uint64
	FileHash    [20]byte
	ChunkHashes [][20]byte
}

func NewPublishFilePacket(fileName string, fileSize uint64, fileHash [20]byte, chunkHashes [][20]byte) PublishFilePacket {
	return PublishFilePacket{
		FileName:    fileName,
		FileSize:    fileSize,
		FileHash:    fileHash,
		ChunkHashes: chunkHashes,
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

type PublishFileSuccessPacket struct {
	FileName string
}

func NewPublishFileSuccessPacket(fileName string) PublishFileSuccessPacket {
	return PublishFileSuccessPacket{
		FileName: fileName,
	}
}

func (pfs *PublishFileSuccessPacket) GetPacketType() uint8 {
	return PublishFileSuccessType
}

type RequestFilePacket struct {
	FileName string
}

func NewRequestFilePacket(fileName string) RequestFilePacket {
	return RequestFilePacket{
		FileName: fileName,
	}
}

func (rf *RequestFilePacket) GetPacketType() uint8 {
	return RequestFileType
}

// TRACKER -> NODE

type AlreadyExistsPacket struct {
	Filename string
}

func NewAlreadyExistsPacket(filename string) AlreadyExistsPacket {
	return AlreadyExistsPacket{
		Filename: filename,
	}
}

func (ae *AlreadyExistsPacket) GetPacketType() uint8 {
	return AlreadyExistsType
}

type NotFoundPacket struct {
	Filename string
}

func NewNotFoundPacket(filename string) NotFoundPacket {
	return NotFoundPacket{
		Filename: filename,
	}
}

func (nf *NotFoundPacket) GetPacketType() uint8 {
	return NotFoundType
}

type AnswerNodesPacket struct {
	FileName      string
	FileSize      uint64
	FileHash      [20]byte
	ChunkHashes   [][20]byte
	NumberOfNodes uint16
	Nodes         []NodeFileInfo
}

type NodeFileInfo struct {
	Name         string
	Port         uint16
	BitfieldSize uint16
	Bitfield     []uint8
}

func NewAnswerNodesPacket(fileName string, fileSize uint64, fileHash [20]byte, chunkHashes [][20]byte, nNodes uint16, names []string, ports []uint16, bitfields [][]uint16) AnswerNodesPacket {
	an := AnswerNodesPacket{
		FileName:      fileName,
		FileSize:      fileSize,
		FileHash:      fileHash,
		ChunkHashes:   chunkHashes,
		NumberOfNodes: nNodes,
	}

	for i := 0; i < int(nNodes); i++ {
		bitfield := EncodeBitField(bitfields[i])

		node := NodeFileInfo{
			BitfieldSize: uint16(len(bitfield)),
			Name:         names[i],
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
	FileName string
}

func NewRemoveFilePacket(fileName string) RemoveFilePacket {
	return RemoveFilePacket{
		FileName: fileName,
	}
}

func (rf *RemoveFilePacket) GetPacketType() uint8 {
	return RemoveFileType
}

// NODE -> NODE

type RequestChunksPacket struct {
	FileName string
	Chunks   []uint16
}

func NewRequestChunksPacket(fileName string, chunks []uint16) RequestChunksPacket {
	return RequestChunksPacket{
		FileName: fileName,
		Chunks:   chunks,
	}
}

func (rc *RequestChunksPacket) GetPacketType() uint8 {
	return RequestChunksType
}

type ChunkPacket struct {
	FileName     string
	Chunk        uint16
	ChunkContent []uint8
}

func NewChunkPacket(fileName string, chunk uint16, chunkContent []uint8) ChunkPacket {
	return ChunkPacket{
		FileName:     fileName,
		Chunk:        chunk,
		ChunkContent: chunkContent,
	}
}

func (c *ChunkPacket) GetPacketType() uint8 {
	return ChunkType
}

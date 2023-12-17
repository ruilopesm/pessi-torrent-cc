package protocol
// TODO: Change the packets that send IP addresses to send domain names instead

// NODE -> TRACKER

// InitPacket is sent by the node to the tracker when it starts
type InitPacket struct {
  Name  string
	UDPPort uint16
}

func NewInitPacket(name string, udpPort uint16) InitPacket {
	return InitPacket{
		Name:  name,
		UDPPort: udpPort,
	}
}

func (ip *InitPacket) GetPacketType() uint8 {
	return InitType
}

// PublishFilePacket is sent by the node to the tracker when it wants to publish a file
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

// PublishChunkPacket is sent by the node to the tracker when it wants to update the tracker about the chunks it has from a file
type PublishChunkPacket struct {
	FileHash     [20]byte
	BitfieldSize uint16
	Bitfield     []uint8
}

func NewPublishChunkPacket(fileHash [20]byte, bitfield []uint16) PublishChunkPacket {
	binaryBitfield := EncodeBitField(bitfield)
	bitfieldSize := len(binaryBitfield)

	return PublishChunkPacket{
		FileHash:     fileHash,
		BitfieldSize: uint16(bitfieldSize),
		Bitfield:     binaryBitfield,
	}
}

func (pc *PublishChunkPacket) GetPacketType() uint8 {
	return PublishChunkType
}

// RequestFilePacket is sent by the node to the tracker when it wants to download a file to get information about the file
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

// FileSuccessPacket is sent by the tracker to the node when it
// has successfully published(Type = PublishFileType)/removed(Type = RemoveFileType) a file
type FileSuccessPacket struct {
	FileName string
	Type     uint8
}

func NewPublishFileSuccessPacket(fileName string) FileSuccessPacket {
	return FileSuccessPacket{
		FileName: fileName,
		Type:     PublishFileType,
	}
}

func NewRemoveFileSuccessPacket(fileName string) FileSuccessPacket {
	return FileSuccessPacket{
		FileName: fileName,
		Type:     RemoveFileType,
	}
}

func (fs *FileSuccessPacket) GetPacketType() uint8 {
	return FileSuccessType
}

// AlreadyExistsPacket is sent by the tracker to the node when it wants to publish a file that already exists in the network
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

// NotFoundPacket is sent by the tracker to the node when it wants to download or remove a file that does not exist
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

// AnswerNodesPacket is sent by the tracker to the node when it wants to download a file to give information about the file
type AnswerNodesPacket struct {
	FileName    string
	FileSize    uint64
	FileHash    [20]byte
	ChunkHashes [][20]byte
	Nodes       []NodeFileInfo
}

type NodeFileInfo struct {
	Name     string
	Port     uint16
	Bitfield []uint8
}

func NewAnswerNodesPacket(fileName string, fileSize uint64, fileHash [20]byte, chunkHashes [][20]byte, names []string, ports []uint16, bitfields [][]uint16) AnswerNodesPacket {
	an := AnswerNodesPacket{
		FileName:    fileName,
		FileSize:    fileSize,
		FileHash:    fileHash,
		ChunkHashes: chunkHashes,
	}

	for i := 0; i < len(bitfields); i++ {
		bitfield := EncodeBitField(bitfields[i])

		node := NodeFileInfo{
			Name:   names[i],
			Port:     ports[i],
			Bitfield: bitfield,
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

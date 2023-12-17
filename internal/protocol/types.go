package protocol

const (
	InitType          = 0
	PublishFileType   = 1
	FileSuccessType   = 2
	AlreadyExistsType = 3
	NotFoundType      = 4
	PublishChunkType  = 5
	RequestFileType   = 6
	AnswerNodesType   = 7
	RemoveFileType    = 8
	RequestChunksType = 9
	ChunkType         = 10
)

type Packet interface {
	GetPacketType() uint8
}

func PacketStructFromType(packetType uint8) Packet {
	switch packetType {
	case InitType:
		return &InitPacket{}
	case PublishFileType:
		return &PublishFilePacket{}
	case FileSuccessType:
		return &FileSuccessPacket{}
	case AlreadyExistsType:
		return &AlreadyExistsPacket{}
	case NotFoundType:
		return &NotFoundPacket{}
	case PublishChunkType:
		return &PublishChunkPacket{}
	case RequestFileType:
		return &RequestFilePacket{}
	case AnswerNodesType:
		return &AnswerNodesPacket{}
	case RemoveFileType:
		return &RemoveFilePacket{}
	case RequestChunksType:
		return &RequestChunksPacket{}
	case ChunkType:
		return &ChunkPacket{}
	default:
		return nil
	}
}

package protocol

const (
	InitType                = 0
	PublishFileType         = 1
	FileSuccessType         = 2
	AlreadyExistsType       = 3
	NotFoundType            = 4
	UpdateChunksType        = 5
	RequestFileType         = 6
	UpdateFileType          = 7
	AnswerFileWithNodesType = 8
	AnswerNodesType         = 9
	RemoveFileType          = 10
	RequestChunksType       = 11
	ChunkType               = 12
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
	case UpdateChunksType:
		return &UpdateChunksPacket{}
	case RequestFileType:
		return &RequestFilePacket{}
	case UpdateFileType:
		return &UpdateFilePacket{}
	case AnswerFileWithNodesType:
		return &AnswerFileWithNodesPacket{}
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

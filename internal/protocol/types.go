package protocol

const (
	InitType = iota
	PublishFileType
	AlreadyExistsType
	PublishChunkType
	RequestFileType
	AnswerNodesType
	RemoveFileType
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
	case AlreadyExistsType:
		return &AlreadyExistsPacket{}
	case PublishChunkType:
		return &PublishChunkPacket{}
	case RequestFileType:
		return &RequestFilePacket{}
	case AnswerNodesType:
		return &AnswerNodesPacket{}
	case RemoveFileType:
		return &RemoveFilePacket{}
	default:
		return nil
	}
}

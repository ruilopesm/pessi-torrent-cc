package packets

const (
	InitType = iota
	PublishFileType
	AlreadyExistsType
	PublishChunkType
	RequestFileType
	AnswerNodesType
)

func PacketStructFromType(packetType uint8) interface{} {
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
	default:
		return nil
	}
}

package packets

const (
	INIT_TYPE = iota
	PUBLISH_FILE_TYPE
	ALREADY_EXISTS_TYPE
	PUBLISH_CHUNK_TYPE
	REQUEST_FILE_TYPE
	ANSWER_NODES_TYPE
  REMOVE_FILE_TYPE
)

func PacketStructFromType(packetType uint8) interface{} {
	switch packetType {
	case INIT_TYPE:
		return &InitPacket{}
	case PUBLISH_FILE_TYPE:
		return &PublishFilePacket{}
	case ALREADY_EXISTS_TYPE:
		return &AlreadyExistsPacket{}
	case PUBLISH_CHUNK_TYPE:
		return &PublishChunkPacket{}
	case REQUEST_FILE_TYPE:
		return &RequestFilePacket{}
	case ANSWER_NODES_TYPE:
		return &AnswerNodesPacket{}
  case REMOVE_FILE_TYPE:
    return &RemoveFilePacket{}
	default:
		return nil
	}
}

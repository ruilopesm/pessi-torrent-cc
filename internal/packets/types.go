package packets

const PUBLISH_FILE_TYPE = 1
const ALREADY_EXISTS_TYPE = 2
const PUBLISH_CHUNK_TYPE = 3
const REQUEST_FILE_TYPE = 4
const ANSWER_NODES_TYPE = 5

func PacketStructFromType(packetType uint8) interface{} {
	switch packetType {
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
	default:
		return nil
	}
}

package protocol


// NODE -> TRACKER

type InitPacket struct {
	Reserved uint8
	UDPPort  uint16
	IPAddr   [4]byte
}

func (ip *InitPacket) GetPacketType() uint8 {
	return InitType
}

func (ip *InitPacket) Create(ipAddr [4]byte, udpPort uint16) {
	ip.Reserved = uint8(0)
	ip.UDPPort = udpPort
	ip.IPAddr = ipAddr
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

func (pf *PublishFilePacket) Create(name string, fileHash [20]byte, chunkHashes [][20]byte) {
	pf.NameSize = uint8(len(name))
	pf.NumberOfChunks = uint16(len(chunkHashes))
	pf.FileHash = fileHash
	pf.FileName = name
	pf.ChunkHashes = chunkHashes
}

type PublishChunkPacket struct {
	BitfieldSize uint16
	Reserved     uint8
	FileHash     [20]byte
	Bitfield     []byte
}

func (pc *PublishChunkPacket) GetPacketType() uint8 {
	return PublishChunkType
}

func (pc *PublishChunkPacket) Create(fileHash [20]byte, bitfield []uint16) {
	binaryBitField := EncodeBitField(bitfield)
	bitfieldSize := len(binaryBitField)

	pc.BitfieldSize = uint16(bitfieldSize)
	pc.Reserved = uint8(0)
	pc.FileHash = fileHash
	pc.Bitfield = binaryBitField
}

type RequestFilePacket struct {
	NameSize uint8
	Reserved uint16
	FileName string
}

func (rf *RequestFilePacket) GetPacketType() uint8 {
	return RequestFileType
}

func (rf *RequestFilePacket) Create(fileName string) {
	rf.NameSize = uint8(len(fileName))
	rf.Reserved = uint16(0)
	rf.FileName = fileName
}

// TRACKER -> NODE

type AlreadyExistsPacket struct {
	NameSize uint8
	Reserved uint16
	FileName string
}

func (ae *AlreadyExistsPacket) GetPacketType() uint8 {
	return AlreadyExistsType
}

func (ae *AlreadyExistsPacket) Create(fileName string) {
	ae.NameSize = uint8(len(fileName))
	ae.Reserved = uint16(0)
	ae.FileName = fileName
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

// func (an *AnswerNodesPacket) Create(sequenceNumber uint8, ipAddr [4]byte, udpPort uint16, bitfield []uint16) {
func (an *AnswerNodesPacket) Create(nNodes uint16, ipAddrs [][4]byte, ports []uint16, bitfields [][]uint16) {
	an.NumberOfNodes = nNodes

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
}

type RemoveFilePacket struct {
	NameSize uint8
	Reserved uint16
	FileName string
}

func (rf *RemoveFilePacket) GetPacketType() uint8 {
	return RemoveFileType
}

func (rf *RemoveFilePacket) Create(fileName string) {
	rf.NameSize = uint8(len(fileName))
	rf.Reserved = uint16(0)
	rf.FileName = fileName
}

package protocol

// NODE -> TRACKER

type PublishFile struct {
	NameSize       uint8
	NumberOfChunks uint16
	FileHash       [20]byte
	FileName       string
	ChunkHashes    [][20]byte
}

func (pf *PublishFile) Create(name string, fileHash [20]byte, chunkHashes [][20]byte) {
	pf.NameSize = uint8(len(name))
	pf.NumberOfChunks = uint16(len(chunkHashes))
	pf.FileHash = fileHash
	pf.FileName = name
	pf.ChunkHashes = chunkHashes
}

type PublishChunk struct {
	BitfieldSize uint16
	FileHash     [20]byte
	Bitfield     []byte
}

func (pc *PublishChunk) Create(fileHash [20]byte, bitfield []uint8) {
	binaryBitField := encodeBitField(bitfield)
	bitfieldSize := len(binaryBitField)

	pc.BitfieldSize = uint16(bitfieldSize)
	pc.FileHash = fileHash
	pc.Bitfield = binaryBitField
}

type RequestFile struct {
	NameSize uint8
	FileName string
}

func (rf *RequestFile) Create(fileName string) {
	rf.NameSize = uint8(len(fileName))
	rf.FileName = fileName
}

// TRACKER -> NODE

type AlreadyExists struct {
	NameSize uint8
	FileName string
}

func (ae *AlreadyExists) Create(fileName string) {
	ae.NameSize = uint8(len(fileName))
	ae.FileName = fileName
}

type AnswerNodes struct {
	SequenceNumber uint8
	BitfieldSize   uint16
	NodeIdentifier [4]byte
	NodePort       uint16
	Bitfield       []byte
}

func (an *AnswerNodes) Create(sequenceNumber uint8, nodeIdentifier [4]byte, nodePort uint16, bitfield []uint8) {
	binaryBitField := encodeBitField(bitfield)
	bitfieldSize := len(binaryBitField)

	an.SequenceNumber = sequenceNumber
	an.BitfieldSize = uint16(bitfieldSize)
	an.NodeIdentifier = nodeIdentifier
	an.NodePort = nodePort
	an.Bitfield = binaryBitField
}

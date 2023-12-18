package main

import (
	"PessiTorrent/internal/filewriter"
	"PessiTorrent/internal/protocol"
	"PessiTorrent/internal/structures"
	"net"
	"time"
)

const (
	ChunkTimeout = 500 * time.Millisecond
)

type File struct {
	FileName string
	Path     string
}

func NewFile(fileName string, path string) File {
	return File{
		FileName: fileName,
		Path:     path,
	}
}

type ForDownloadFile struct {
	UpdatedByTracker bool // Whether the tracker has already sent the file info or not

	FileName   string
	FileHash   [20]byte
	FileSize   uint64
	FileWriter *filewriter.FileWriter

	// Last time the node sent a UpdateChunksPacket to the tracker
	LastServerChunksUpdate time.Time

	NumberOfChunks uint16
	Chunks         structures.SynchronizedList[ChunkInfo]

	Nodes structures.SynchronizedMap[string, *NodeInfo]
}

type ChunkInfo struct {
	Index      uint16
	Downloaded bool
	Hash       [20]byte
}

type NodeInfo struct {
	Address string
	// Chunk index -> Last time chunk was requested
	Chunks   structures.SynchronizedMap[uint16, *RequestInfo]
	Timeouts uint
}

type RequestInfo struct {
	TimeLastRequested time.Time
	NumberOfTries     uint
}

func NewForDownloadFile(fileName string) *ForDownloadFile {
	return &ForDownloadFile{
		UpdatedByTracker:       false,
		FileName:               fileName,
		LastServerChunksUpdate: time.Now(),
	}
}

func (f *ForDownloadFile) SetData(fileHash [20]byte, chunkHashes [][20]byte, fileSize uint64, numberOfChunks uint16) error {
	f.FileHash = fileHash
	f.FileSize = fileSize
	fileWriter, err := filewriter.NewFileWriter(f.FileName, fileSize, f.MarkChunkAsDownloaded)
	if err != nil {
		return err
	}
	f.FileWriter = fileWriter
	go f.FileWriter.Start()

	f.NumberOfChunks = numberOfChunks
	f.Chunks = structures.NewSynchronizedListWithInitialSize[ChunkInfo](uint(numberOfChunks))
	for i := 0; i < int(numberOfChunks); i++ {
		_ = f.Chunks.Set(uint(i), ChunkInfo{
			Index:      uint16(i),
			Downloaded: false,
			Hash:       chunkHashes[i],
		})
	}

	f.Nodes = structures.NewSynchronizedMap[string, *NodeInfo]()

	return nil
}

func (f *ForDownloadFile) IsFileDownloaded() bool {
	return f.LengthOfMissingChunks() == 0
}

func (f *ForDownloadFile) AddNode(nodeAddr *net.UDPAddr, bitfield []uint8) {
	nodeInfo := NodeInfo{
		Address: nodeAddr.String(),
		Chunks:  structures.NewSynchronizedMap[uint16, *RequestInfo](),
	}

	decoded := protocol.DecodeBitField(bitfield)
	for index, hasChunk := range decoded {
		if hasChunk {
			nodeInfo.Chunks.Put(uint16(index), &RequestInfo{TimeLastRequested: time.Time{}})
		}
	}

	f.Nodes.Put(nodeAddr.String(), &nodeInfo)
}

func (f *ForDownloadFile) MarkChunkAsRequested(chunkIndex uint16, nodeInfo *NodeInfo) {
	nodeInfo.Chunks.Put(chunkIndex, &RequestInfo{TimeLastRequested: time.Now()})
}

func (f *ForDownloadFile) MarkChunkAsDownloaded(chunkIndex uint16) {
	chunk, _ := f.Chunks.Get(uint(chunkIndex))
	chunk.Downloaded = true
	_ = f.Chunks.Set(uint(chunkIndex), chunk)
}

func (f *ForDownloadFile) ChunkAlreadyDownloaded(chunkIndex uint16) bool {
	chunk, _ := f.Chunks.Get(uint(chunkIndex))
	return chunk.Downloaded
}

func (f *ForDownloadFile) GetChunkHash(chunkIndex uint16) [20]byte {
	chunk, _ := f.Chunks.Get(uint(chunkIndex))
	return chunk.Hash
}

func (f *ForDownloadFile) GetMissingChunks() []uint {
	missingChunks := f.Chunks.IndexesWhere(func(chunk ChunkInfo) bool {
		return !chunk.Downloaded
	})

	return missingChunks
}

func (f *ForDownloadFile) GetNumberOfNodesWhichHaveChunk(chunkIndex uint16) uint {
	var numberOfNodes uint = 0
	f.Nodes.ForEach(func(nodeAddrString string, nodeInfo *NodeInfo) {
		_, ok := nodeInfo.Chunks.Get(chunkIndex)
		if ok {
			numberOfNodes++
		}
	})

	return numberOfNodes
}

func (f *ForDownloadFile) LengthOfMissingChunks() int {
	return len(f.GetMissingChunks())
}

func (n *NodeInfo) ShouldRequestChunk(chunkIndex uint16) bool {
	chunk, ok := n.GetLastTimeChunkWasRequested(chunkIndex)
	if !ok {
		return false
	}

	// Chunk was not requested yet or it was requested more than the chunk timeout ago
	// The expected time could be improved by calculating the average time between requests
	return chunk == time.Time{} || time.Since(chunk) > ChunkTimeout
}

func (n *NodeInfo) GetLastTimeChunkWasRequested(chunkIndex uint16) (time.Time, bool) {
	requestInfo, ok := n.Chunks.Get(chunkIndex)
	return requestInfo.TimeLastRequested, ok
}

func (f *ForDownloadFile) WriteChunkToDisk(chunkIndex uint16, chunkContent []uint8) {
	f.FileWriter.EnqueueChunkToWrite(chunkIndex, chunkContent)
}

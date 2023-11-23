package main

import (
	"PessiTorrent/internal/protocol"
	"PessiTorrent/internal/structures"
	"PessiTorrent/internal/transport"
)

type TrackedFile struct {
	FileName    string
	FileHash    [20]byte
	ChunkHashes [][20]byte
}

func NewTrackedFile(fileName string, fileHash [20]byte, chunkHashes [][20]byte) TrackedFile {
	return TrackedFile{
		FileName:    fileName,
		FileHash:    fileHash,
		ChunkHashes: chunkHashes,
	}
}

type Nodes struct {
	*structures.SynchronizedList[*NodeInfo]
}

type NodeInfo struct {
	conn    transport.TCPConnection
	udpPort uint16

	files structures.SynchronizedMap[*SharedFile]
}

type SharedFile struct {
	FileName    string
	FileHash    [20]byte
	ChunkHashes [][20]byte
	Bitfield    []uint16
}

func NewNodeInfo(conn transport.TCPConnection, udpPort uint16) NodeInfo {
	return NodeInfo{
		conn:    conn,
		udpPort: udpPort,
		files:   structures.NewSynchronizedMap[*SharedFile](),
	}
}

func NewSharedFile(fileName string, fileHash [20]byte, chunkHashes [][20]byte) SharedFile {
	size := len(chunkHashes)

	return SharedFile{
		FileName:    fileName,
		FileHash:    fileHash,
		ChunkHashes: chunkHashes,
		Bitfield:    protocol.NewCheckedBitfield(size),
	}
}

func (ni *NodeInfo) AddFile(file SharedFile) {
	ni.files.Put(file.FileName, &file)
}

func (ni *NodeInfo) RemoveFile(file SharedFile) {
	ni.files.Delete(file.FileName)
}

func (ni *NodeInfo) HasFile(fileName string) bool {
	_, ok := ni.files.Get(fileName)
	return ok
}

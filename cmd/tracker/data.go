package main

import (
	"PessiTorrent/internal/protocol"
	"PessiTorrent/internal/structures"
	"PessiTorrent/internal/transport"
)

type TrackedFile struct {
	FileName    string
	FileSize    uint64
	FileHash    [20]byte
	ChunkHashes [][20]byte
}

func NewTrackedFile(fileName string, fileSize uint64, fileHash [20]byte, chunkHashes [][20]byte) TrackedFile {
	return TrackedFile{
		FileName:    fileName,
		FileSize:    fileSize,
		FileHash:    fileHash,
		ChunkHashes: chunkHashes,
	}
}

type NodeInfo struct {
	name    string
	conn    transport.TCPConnection
	udpPort uint16

	files structures.SynchronizedMap[string, protocol.Bitfield]
}

func NewNodeInfo(conn transport.TCPConnection, udpPort uint16, name string) NodeInfo {
	return NodeInfo{
		name:    name,
		conn:    conn,
		udpPort: udpPort,
		files:   structures.NewSynchronizedMap[string, protocol.Bitfield](),
	}
}

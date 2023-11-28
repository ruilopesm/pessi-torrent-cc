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
	domain  string
	conn    transport.TCPConnection
	udpPort uint16

	files structures.SynchronizedMap[*SharedFile]
}

type SharedFile struct {
	FileName    string
	FileSize    uint64
	FileHash    [20]byte
	ChunkHashes [][20]byte
	Bitfield    []uint16
}

func NewNodeInfo(conn transport.TCPConnection, udpPort uint16, domain string) NodeInfo {
	return NodeInfo{
		domain:  domain,
		conn:    conn,
		udpPort: udpPort,
		files:   structures.NewSynchronizedMap[*SharedFile](),
	}
}

func NewSharedFile(fileName string, fileSize uint64, fileHash [20]byte, chunkHashes [][20]byte) SharedFile {
	return SharedFile{
		FileName:    fileName,
		FileSize:    fileSize,
		FileHash:    fileHash,
		ChunkHashes: chunkHashes,
		Bitfield:    protocol.NewCheckedBitfield(len(chunkHashes)),
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

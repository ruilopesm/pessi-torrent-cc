package main

import (
	"PessiTorrent/internal/protocol"
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
	FileName string

	FileHash    [20]byte
	ChunkHashes [][20]byte

	// Represents the chunks that have already been downloaded
	Downloaded []uint16
}

func NewForDownloadFile(fileName string, fileHash [20]byte, chunkHashes [][20]byte) ForDownloadFile {
	return ForDownloadFile{
		FileName:    fileName,
		FileHash:    fileHash,
		ChunkHashes: chunkHashes,
		Downloaded:  protocol.NewUncheckedBitfield(len(chunkHashes)),
	}
}

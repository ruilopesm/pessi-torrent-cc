package main

import (
	"PessiTorrent/internal/structures"
)

const (
	DownloadsFolder = "downloads"
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
	FileSize    uint64

	// Represents the chunks that have been downloaded
	DownloadedChunks structures.SynchronizedList[bool]
}

func NewForDownloadFile(fileName string) ForDownloadFile {
	return ForDownloadFile{
		FileName: fileName,
	}
}

func (f *ForDownloadFile) SetData(fileHash [20]byte, chunkHashes [][20]byte, fileSize uint64, numberOfChunks uint16) {
	f.FileHash = fileHash
	f.ChunkHashes = chunkHashes
	f.FileSize = fileSize

	f.DownloadedChunks = structures.NewSynchronizedListWithInitialSize[bool](uint(numberOfChunks))
}

func (f *ForDownloadFile) SetDownloadedChunk(chunk uint16) {
	_ = f.DownloadedChunks.Set(uint(chunk), true)
}

// Returns a list of the chunks that were not yet downloaded
func (f *ForDownloadFile) GetMissingChunks() []uint {
	chunks := f.DownloadedChunks.IndexesWhere(func(val bool) bool {
		return !val
	})

	return chunks
}

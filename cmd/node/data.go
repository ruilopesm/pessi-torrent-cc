package main

import "PessiTorrent/internal/structures"

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
	Downloaded structures.SynchronizedList[bool]
}

func NewForDownloadFile(fileName string) ForDownloadFile {
	return ForDownloadFile{
		FileName: fileName,
	}
}

func (f *ForDownloadFile) SetData(fileHash [20]byte, chunkHashes [][20]byte) {
	f.FileHash = fileHash
	f.ChunkHashes = chunkHashes
	f.Downloaded = structures.NewSynchronizedList[bool](uint(len(chunkHashes)))
}

func (f *ForDownloadFile) SetChunkDownloaded(chunk uint16) {
	f.Downloaded.Set(uint(chunk), true)
}

func (f *ForDownloadFile) AllChunksDownloaded(chunks []uint16) bool {
	for _, chunk := range chunks {
		v, _ := f.Downloaded.Get(uint(chunk))
		if !v {
			return false
		}
	}

	return true
}

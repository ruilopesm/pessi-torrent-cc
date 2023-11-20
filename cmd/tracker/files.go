package main

type File struct {
	filename    string
	fileHash    [20]byte
	chunkHashes [][20]byte
}

func NewFile(filename string, fileHash [20]byte, chunkHashes [][20]byte) File {
	return File{
		filename:    filename,
		fileHash:    fileHash,
		chunkHashes: chunkHashes,
	}
}

func (t *Tracker) AddFile(file File) {
	t.files.Put(file.filename, &file)
}

func (t *Tracker) RemoveFile(filename string) {
	t.files.Delete(filename)
}

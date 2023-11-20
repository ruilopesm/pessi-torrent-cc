package main

type File struct {
	filename    string
	filepath    string // Optional
	fileHash    [20]byte
	chunkHashes [][20]byte
	bitfield    []uint8 // Optional
}

func NewFile(filename string, fileHash [20]byte, chunkHashes [][20]byte) File {
	return File{
		filename:    filename,
		fileHash:    fileHash,
		chunkHashes: chunkHashes,
	}
}

func (f File) WithFilePath(filePath string) File {
	f.filepath = filePath
	return f
}

func (f File) WithBitfield(bitfield []byte) File {
	f.bitfield = bitfield
	return f
}

func (n *Node) AddFile(file File) {
	n.files.Put(file.filename, &file)
}

func (n *Node) RemoveFile(filename string) {
	n.files.Delete(filename)
}

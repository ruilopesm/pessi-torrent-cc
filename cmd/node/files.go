package main

import (
  "fmt"
  "os"
  "path/filepath"
	"PessiTorrent/internal/utils"
)

type File struct {
  filename string
  filepath string 
  chunkSize uint64 // bytes
	fileHash [20]byte
	chunkHashes   [][20]byte
}

func (n *Node) CreateFile(filePath string) (*File, error) {
  if _, err := os.Stat(filePath); os.IsNotExist(err) {
    return nil, fmt.Errorf("file does not exist: %v", filePath)
  }
  // open the file
  file, err := os.Open(filePath)
  if err != nil {
    return nil, err
  }
  defer file.Close()

  filename := filepath.Base(filePath)
  fileHash, err := utils.HashFile(file)
  if err != nil {
    return nil, err
  }

  chunkHashes := make([][20]byte, 0)
  chunkSize, err := utils.HashFileChunks(file, &chunkHashes)

  f := &File{
    filename: filename,
    filepath: filePath,
    chunkSize: chunkSize,
    fileHash: fileHash,
    chunkHashes: chunkHashes,
  }

  n.files.Lock()
  n.files.m[filename] = f
  n.files.Unlock()

  return f, nil
}

func (n *Node) RemoveFile(filename string) error {
  n.files.Lock()
  defer n.files.Unlock()

  delete(n.files.m, filename)

  return nil
}


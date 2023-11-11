package main

import (
  "fmt"
  "os"
  "path/filepath"
	"PessiTorrent/internal/utils"
	"PessiTorrent/internal/serialization"
)

// TODO: add bitfield of chunks the file has
type File struct {
  filename string
  filepath string 
	fileHash [20]byte
	chunkHashes   [][20]byte
  bitfield []byte
}

func (n *Node) AddSharedFile(filePath string) (*File, error) {
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
  _, err = utils.HashFileChunks(file, &chunkHashes)

  var bitfield []uint8
  // iterate over the file chunks
  for i := 0; i < len(chunkHashes); i++ {
    bitfield = append(bitfield, uint8(i))
  }

  f := &File{
    filename: filename,
    filepath: filePath,
    fileHash: fileHash,
    chunkHashes: chunkHashes,
    bitfield: serialization.EncodeBitField(bitfield),
  }

  n.files.Lock()
  n.files.m[filename] = f
  n.files.Unlock()

  return f, nil
}

// TODO: melhorar isto aqui
func (n *Node) AddDownloadFile(filename string, fileHash [20]byte, chunkHashes [][20]byte) (*File, error) {
  f := &File{
    filename: filename,
    filepath: "",
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


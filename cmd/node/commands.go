package main

import (
	"PessiTorrent/internal/protocol"
	"PessiTorrent/internal/utils"
	"fmt"
	"os"
	"path/filepath"
)

// request <file name>
func (n *Node) requestFile(args []string) error {
	filename := args[0]

	packet := protocol.NewRequestFilePacket(filename)
	n.conn.EnqueuePacket(&packet)

	return nil
}

// publish <path>
func (n *Node) publish(args []string) error {
	path := args[0]

	// Check if the path is a file or a directory
	switch info, err := os.Stat(path); {
	case err != nil:
		return err
	case info.IsDir():
		err = n.publishDirectory(path)
		if err != nil {
			return err
		}
	default:
		err = n.publishFile(path)
		if err != nil {
			return err
		}
	}

	return nil
}

func (n *Node) publishFile(path string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	filename := filepath.Base(path)

	fileHash, err := utils.HashFile(file)
	if err != nil {
		return err
	}

	chunkHashes := make([][20]byte, 0)
	err = utils.HashFileChunks(file, &chunkHashes)
	if err != nil {
		return err
	}

  bitfield := protocol.NewCheckedBitfield(len(chunkHashes))
	internal_file := NewFile(filename, fileHash, chunkHashes).WithFilePath(path).WithBitfield(protocol.EncodeBitField(bitfield))
	n.AddFile(internal_file)
	fmt.Printf("Added file %s to internal state\n", internal_file.filename)

	packet := protocol.NewPublishFilePacket(internal_file.filename, internal_file.fileHash, internal_file.chunkHashes)
	n.conn.EnqueuePacket(&packet)
	fmt.Println("Sent publish file packet to tracker")

	return nil
}

func (n *Node) publishDirectory(path string) error {
	err := filepath.WalkDir(path, func(current_path string, d os.DirEntry, err error) error {
		if path != current_path {
			switch d.IsDir() {
			case true:
				return n.publishDirectory(current_path)
			case false:
				return n.publishFile(current_path)
			}

			return nil
		}

		return nil
	})

	return err
}

// status
func (n *Node) status(_ []string) error {
	fmt.Printf("Currently connected to tracker with address %s\n", n.serverAddr)

	fmt.Println("Published files:")
	n.files.ForEach(func(filename string, file *File) {
		fmt.Printf("File: %v\n", file.filename)
		fmt.Printf("Filepath: %v\n", file.filepath)
		fmt.Printf("File hash: %v\n", file.fileHash)
		fmt.Printf("Chunk hashes: %v\n", file.chunkHashes)
		fmt.Printf("Bitfield: %b\n", file.bitfield)
	})

	return nil
}

// remove <file name>
func (n *Node) removeFile(args []string) error {
	filename := args[0]

	n.RemoveFile(filename)

	packet := protocol.NewRemoveFilePacket(filename)
	n.conn.EnqueuePacket(&packet)

	return nil
}

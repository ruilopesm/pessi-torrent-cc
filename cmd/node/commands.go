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

	forDownloadFile := NewForDownloadFile(filename)
	n.forDownload.Put(filename, &forDownloadFile)

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

	fileName := filepath.Base(path)

	fileHash, err := utils.HashFile(file)
	if err != nil {
		return err
	}

	chunkHashes := make([][20]byte, 0)
	err = utils.HashFileChunks(file, &chunkHashes)
	if err != nil {
		return err
	}

	newFile := NewFile(fileName, path)
	n.pending.Put(fileName, &newFile)
	fmt.Printf("Added file %s to pending files\n", fileName)

	packet := protocol.NewPublishFilePacket(fileName, fileHash, chunkHashes)
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

	fmt.Printf("\n")

	if n.pending.Len() != 0 {
		fmt.Println("Pending files:")
		n.pending.ForEach(func(filename string, file *File) {
			fmt.Printf("%s at %s\n", file.FileName, file.Path)
		})

		fmt.Printf("\n")
	}

	if n.published.Len() != 0 {
		fmt.Println("Published files:")
		n.published.ForEach(func(filename string, file *File) {
			fmt.Printf("%s at %s\n", file.FileName, file.Path)
		})

		fmt.Printf("\n")
	}

	if n.forDownload.Len() != 0 {
		fmt.Println("Files for download:")
		n.forDownload.ForEach(func(filename string, file *ForDownloadFile) {
			fmt.Printf("%s\n", file.FileName)
			fmt.Printf("Hash: %x\n", file.FileHash)
			fmt.Printf("Chunks: %d\n", len(file.ChunkHashes))
			fmt.Printf("Downloaded: %d\n", file.Downloaded.Len())
		})
	}

	return nil
}

// remove <file name>
func (n *Node) removeFile(args []string) error {
	filename := args[0]

	n.published.Delete(filename)

	packet := protocol.NewRemoveFilePacket(filename)
	n.conn.EnqueuePacket(&packet)

	fmt.Printf("Successfully removed file %s from the network", filename)

	return nil
}

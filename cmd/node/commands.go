package main

import (
	"PessiTorrent/internal/logger"
	"PessiTorrent/internal/protocol"
	"PessiTorrent/internal/utils"
	"os"
	"path/filepath"
)

// request <file name>
func (n *Node) requestFile(args []string) error {
	filename := args[0]

	packet := protocol.NewRequestFilePacket(filename)
	n.conn.EnqueuePacket(&packet)

	// Data of the file will be updated later, when the tracker responds back
	n.forDownload.Put(filename, &ForDownloadFile{})

	return nil
}

// publish <file name>
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
	fileSize, err := utils.HashFileChunks(file, &chunkHashes)
	if err != nil {
		return err
	}

	newFile := NewFile(fileName, path)
	n.pending.Put(fileName, &newFile)
	logger.Info("Added file %s to pending files", fileName)

	packet := protocol.NewPublishFilePacket(fileName, fileSize, fileHash, chunkHashes)
	n.conn.EnqueuePacket(&packet)
	logger.Info("Sent publish file packet to tracker")

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
	if n.connected {
		logger.Info("Connected to tracker on %s", n.trackerAddr)
	} else {
		logger.Info("Not connected to tracker. Run 'connect' in order to do so")
	}

	if n.pending.Len() != 0 {
		logger.Info("Pending files:")
		n.pending.ForEach(func(filename string, file *File) {
			logger.Info("%s at %s", file.FileName, file.Path)
		})
	}

	if n.published.Len() != 0 {
		logger.Info("Published files:")
		n.published.ForEach(func(filename string, file *File) {
			logger.Info("%s at %s", file.FileName, file.Path)
		})
	}

	if n.forDownload.Len() != 0 {
		logger.Info("Files for download:")
		n.forDownload.ForEach(func(fileName string, file *ForDownloadFile) {
			logger.Info("%s with size %d", fileName, file.FileSize)
			logger.Info("Chunks %d/%d", file.Chunks.Len(), file.NumberOfChunks)
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

	logger.Info("Successfully removed file %s from the network", filename)

	return nil
}

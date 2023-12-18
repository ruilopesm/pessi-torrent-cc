package main

import (
	"PessiTorrent/internal/logger"
	"PessiTorrent/internal/protocol"
	"PessiTorrent/internal/utils"
	"os"
	"path/filepath"
)

func (n *Node) connect(args []string) error {
	if n.connected {
		logger.Info("Already connected to tracker on %s", n.trackerAddr)
		return nil
	}

	go n.startTCP()

	return nil
}

// request <file name>
func (n *Node) requestFile(args []string) error {
	filename := args[0]

	packet := protocol.NewRequestFilePacket(filename)
	n.conn.EnqueuePacket(&packet)

	// Data of the file will be updated later, when the tracker responds back
	n.forDownload.Put(filename, NewForDownloadFile(filename))

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
			len := uint16(file.LengthOfMissingChunks())
			logger.Info("Chunks progress %d/%d (%.2f%%)", file.NumberOfChunks-len, file.NumberOfChunks, float64(file.NumberOfChunks-len)/float64(file.NumberOfChunks)*100)
		})
	}

	return nil
}

// remove <file name>
func (n *Node) removeFile(args []string) error {
	filename := args[0]

	packet := protocol.NewRemoveFilePacket(filename)
	n.conn.EnqueuePacket(&packet)

	return nil
}

// statistics
func (n *Node) statistics(_ []string) error {
	statistics := n.nodeStatistics
	logger.Info("Total uploaded: %d bytes", statistics.TotalUploaded)
	logger.Info("Total downloaded: %d bytes", statistics.TotalDownloaded)

	for addr := range statistics.nodeMap {
		logger.Info("Average download speed from %s: %.2f bytes/s", addr, statistics.getAverageDownloadSpeed(addr))
	}

	return nil
}

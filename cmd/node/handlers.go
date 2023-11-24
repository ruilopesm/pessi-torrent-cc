package main

import (
	"PessiTorrent/internal/protocol"
	"PessiTorrent/internal/transport"
	"PessiTorrent/internal/utils"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"time"
)

func (n *Node) handlePublishFileSuccessPacket(packet *protocol.PublishFileSuccessPacket, conn *transport.TCPConnection) {
	fmt.Printf("File %s published in the network successfully\n", packet.FileName)

	// Remove file from pending and add it to published, since tracker has accepted it
	file, _ := n.pending.Get(packet.FileName)
	n.published.Put(packet.FileName, file)
	n.pending.Delete(packet.FileName)
}

func (n *Node) handleAnswerNodesPacket(packet *protocol.AnswerNodesPacket, conn *transport.TCPConnection) {
	fmt.Printf("Answer nodes packet received from %s\n", conn.RemoteAddr())
	fmt.Printf("Number of nodes who got requested file: %d\n", packet.NumberOfNodes)

	// Update file in forDownload
	forDownloadFile, _ := n.forDownload.Get(packet.FileName)
	forDownloadFile.SetData(packet.FileHash, packet.ChunkHashes)

	// Request chunks
	for _, node := range packet.Nodes {
		chunks := protocol.DecodeBitField(node.Bitfield)
		udpAddr := &net.UDPAddr{
			IP:   node.IPAddr[:],
			Port: int(node.Port),
		}
		packet := protocol.NewRequestChunksPacket(packet.FileName, chunks)

		go n.requestChunks(&packet, udpAddr)
	}
}

func (n *Node) requestChunks(packet *protocol.RequestChunksPacket, addr *net.UDPAddr) {
	// Send request chunks packet
	n.srv.SendPacket(packet, addr)

	// Check, every 0.1s, if the chunk has been downloaded until 1s has passed
	// FIXME: Should be adjusted to the chunk size and number of requested chunks
	timeout := time.After(10 * time.Second)
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		file, _ := n.forDownload.Get(packet.FileName)

		select {
		case <-ticker.C:
			// Check if all requested chunks have been downloaded
			if file.AllChunksDownloaded(packet.Chunks) {
				fmt.Printf("All chunks of file %s have been downloaded\n", packet.FileName)
				return
			}

		case <-timeout:
			fmt.Printf("Chunks were not received within timeout\n")
			return
		}
	}
}

func (n *Node) handleAlreadyExistsPacket(packet *protocol.AlreadyExistsPacket, conn *transport.TCPConnection) {
	fmt.Printf("File %s already exists in the network\n", packet.Filename)

	// Remove file from pending, since tracker has rejected it
	n.pending.Delete(packet.Filename)
}

func (n *Node) handleNotFoundPacket(packet *protocol.NotFoundPacket, conn *transport.TCPConnection) {
	fmt.Printf("File %s was not found in the network\n", packet.Filename)

	// Remove file from downloading, since it does not exist
	n.forDownload.Delete(packet.Filename)
}

func (n *Node) handleRequestChunksPacket(packet *protocol.RequestChunksPacket, addr *net.UDPAddr) {
	fmt.Printf("Request chunks packet received from %s\n", addr)

	// Get file from published files
	file, ok := n.published.Get(packet.FileName)
	if !ok {
		fmt.Printf("File %s not found in published files\n", packet.FileName)
		return
	}

	// Send chunks to the node
	for _, chunk := range packet.Chunks {
		fmt.Printf("Sending chunk %d of file %s\n", chunk, packet.FileName)

		// Open file by the given path
		file, err := os.Open(file.Path)
		if err != nil {
			fmt.Printf("Error opening file: %v\n", err)
			return
		}

		stats, _ := file.Stat()
		chunkSize := utils.ChunkSize(uint64(stats.Size()))

		// Seek to the beginning of the chunk
		_, err = file.Seek(int64(uint64(chunk)*chunkSize), 0)
		if err != nil {
			fmt.Printf("Error seeking file: %v\n", err)
			return
		}

		// Read chunk
		chunkContent := make([]byte, chunkSize)
		read, err := file.Read(chunkContent)
		if err != nil && !errors.Is(err, io.EOF) {
			fmt.Printf("Error reading file: %v\n", err)
			return
		}

		// Send chunk
		packet := protocol.NewChunkPacket(packet.FileName, chunk, chunkContent[:read])
		n.srv.SendPacket(&packet, addr)
	}
}

func (n *Node) handleChunkPacket(packet *protocol.ChunkPacket, addr *net.UDPAddr) {
	fmt.Printf("Chunk %d of file %s downloaded\n", packet.Chunk, packet.FileName)

	// Set chunk as downloaded
	forDownloadFile, _ := n.forDownload.Get(packet.FileName)
	forDownloadFile.SetChunkDownloaded(packet.Chunk)

	// Write to file
	// FIXME: Folder should not be hardcoded and this is not thread-safe
	file, err := os.OpenFile(fmt.Sprintf("downloads/%s", packet.FileName), os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Printf("Error opening file: %v\n", err)
		return
	}

	_, err = file.Write(packet.ChunkContent)
	if err != nil {
		fmt.Printf("Error writing to file: %v\n", err)
		return
	}

	fmt.Printf("Chunk %d written to file %s\n", packet.Chunk, packet.FileName)
}

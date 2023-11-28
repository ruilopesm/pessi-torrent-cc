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

// Handler for when a node requests, to the tracker, the list of nodes who have a specific file
func (n *Node) handleAnswerNodesPacket(packet *protocol.AnswerNodesPacket, conn *transport.TCPConnection) {
	fmt.Printf("Answer nodes packet received from %s\n", conn.RemoteAddr())
	fmt.Printf("Number of nodes who got requested file: %d\n", packet.NumberOfNodes)

	// Update file in forDownload data structure
	forDownloadFile, ok := n.forDownload.Get(packet.FileName)
	if !ok {
		fmt.Printf("File %s not found in forDownload files\n", packet.FileName)
		return
	}

	forDownloadFile.SetData(packet.FileHash, packet.ChunkHashes, packet.FileSize, uint16(len(packet.ChunkHashes)))

	// Request file (in chunks) to the network
	go n.requestFileByChunks(packet, forDownloadFile)
}

func (n *Node) requestFileByChunks(packet *protocol.AnswerNodesPacket, forDownloadFile *ForDownloadFile) {
	responsesChannel := make(chan *protocol.ChunkPacket)
	defer close(responsesChannel)

	// Check for missing chunks
	missingChunks := forDownloadFile.GetMissingChunks()
	if len(missingChunks) == 0 {
		fmt.Printf("File %s downloaded successfully\n", packet.FileName)
		n.forDownload.Delete(packet.FileName)
		return
	}

	// Loop over nodes who have the file and request missing chunks
	for _, node := range packet.Nodes {
		missingChunksInNode := make([]uint16, 0)
		nodeAddrStr, err := n.dns.ResolveIP(node.Name)
		nodeAddr := net.ParseIP(nodeAddrStr)
		fmt.Println("Node name: ", node.Name)
		fmt.Println("Node addr: ", nodeAddr)

		if err != nil {
			fmt.Printf("Error resolving node IP: %v\n", err)
			continue
		}

		for _, chunk := range missingChunks {
			if protocol.GetBit(node.Bitfield, int(chunk)) {
				missingChunksInNode = append(missingChunksInNode, uint16(chunk))
			}
		}

		if len(missingChunksInNode) > 0 {
			udpAddr := &net.UDPAddr{
				IP:   nodeAddr,
				Port: int(node.Port),
			}

			fmt.Printf("Sending request chunks packet to %s\n", udpAddr)

			go n.sendRequestChunksPacket(packet.FileName, missingChunksInNode, udpAddr, responsesChannel)
		}
	}

	// FIXME: Should be calculated based on the total expected transfer time, number of chunks and number of nodes
	timeout := time.After(1 * time.Second)

	// Wait for all chunks to be downloaded or timeout
	for {
		select {
		case chunkPacket := <-responsesChannel:
			// Set received chunk as downloaded
			forDownloadFile.SetDownloadedChunk(chunkPacket.Chunk)
			fmt.Printf("Chunk %d of file %s downloaded\n", chunkPacket.Chunk, forDownloadFile.FileName)

		case <-timeout:
			// Retry to download missing chunks, if any
			n.requestFileByChunks(packet, forDownloadFile)
		}
	}
}

func (n *Node) sendRequestChunksPacket(fileName string, chunks []uint16, addr *net.UDPAddr, responsesChannel chan *protocol.ChunkPacket) {
	// Create new UDP socket to handle this request and its responses
	conn, err := net.ListenUDP("udp", nil)
	if err != nil {
		fmt.Printf("Error creating UDP socket: %v\n", err)
		return
	}
	defer conn.Close()

	sock := transport.NewUDPSocket(*conn, *addr)

	// Send request chunks packet
	packet := protocol.NewRequestChunksPacket(fileName, chunks)
	sock.SendPacket(&packet)

	// Read responses and send them through the responses channel
	for {
		responsePacket, err := sock.ReadPacket()
		if err != nil {
			fmt.Printf("Error reading packet: %v\n", err)
			return
		}

		switch packet := responsePacket.(type) {
		case *protocol.ChunkPacket:
			responsesChannel <- packet
		default:
			fmt.Printf("Unknown packet type: %v\n", packet)
		}
	}
}

// Handler for when a node requests chunks of a file to this node
func (n *Node) handleRequestChunksPacket(packet *protocol.RequestChunksPacket, addr *net.UDPAddr) {
	fmt.Printf("Request chunks packet received from %s\n", addr)

	// Get file from published files
	publishedFile, ok := n.published.Get(packet.FileName)
	if !ok {
		fmt.Printf("File %s not found in published files\n", packet.FileName)
		return
	}

	// Open file by the given path
	file, err := os.Open(publishedFile.Path)
	if err != nil {
		fmt.Printf("Error opening file: %v\n", err)
		return
	}

	stats, _ := file.Stat()
	chunkSize := utils.ChunkSize(uint64(stats.Size()))

	// Send requested chunks
	for _, chunk := range packet.Chunks {
		fmt.Printf("Sending chunk %d of file %s to %s\n", chunk, packet.FileName, addr)

		// Seek to the beginning of the chunk
		_, err = file.Seek(int64(uint64(chunk)*chunkSize), 0)
		if err != nil {
			fmt.Printf("Error seeking file: %v\n", err)
			return
		}

		// Read chunk bytes
		chunkContent := make([]byte, chunkSize)
		read, err := file.Read(chunkContent)
		if err != nil && !errors.Is(err, io.EOF) {
			fmt.Printf("Error reading file: %v\n", err)
			return
		}

		// Send chunk bytes
		packet := protocol.NewChunkPacket(packet.FileName, chunk, chunkContent[:read])
		n.srv.SendPacket(&packet, addr)
	}
}

// Handler for when a node publishes a file in the network
func (n *Node) handlePublishFileSuccessPacket(packet *protocol.PublishFileSuccessPacket, conn *transport.TCPConnection) {
	fmt.Printf("File %s published in the network successfully\n", packet.FileName)

	// Remove file from pending and add it to published, since tracker has accepted it
	file, _ := n.pending.Get(packet.FileName)
	n.published.Put(packet.FileName, file)
	n.pending.Delete(packet.FileName)
}

// Handler for when the file, the node is trying to publish, already exists in the network
func (n *Node) handleAlreadyExistsPacket(packet *protocol.AlreadyExistsPacket, conn *transport.TCPConnection) {
	fmt.Printf("File %s already exists in the network\n", packet.Filename)

	// Remove file from pending, since tracker has rejected it
	n.pending.Delete(packet.Filename)
}

// Handler for when the file, the node is trying to download, does not exist in the network
func (n *Node) handleNotFoundPacket(packet *protocol.NotFoundPacket, conn *transport.TCPConnection) {
	fmt.Printf("File %s was not found in the network\n", packet.Filename)

	// Remove file from downloading, since it does not exist
	n.forDownload.Delete(packet.Filename)
}

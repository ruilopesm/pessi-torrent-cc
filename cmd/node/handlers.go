package main

import (
	"PessiTorrent/internal/logger"
	"PessiTorrent/internal/protocol"
	"PessiTorrent/internal/transport"
	"PessiTorrent/internal/utils"
	"errors"
	"io"
	"net"
	"os"
	"time"
)

func (n *Node) HandlePackets(packet protocol.Packet, conn *transport.TCPConnection) {
	switch packet := packet.(type) {
	case *protocol.AnswerNodesPacket:
		n.handleAnswerNodesPacket(packet, conn)
	case *protocol.PublishFileSuccessPacket:
		n.handlePublishFileSuccessPacket(packet, conn)
	case *protocol.AlreadyExistsPacket:
		n.handleAlreadyExistsPacket(packet, conn)
	case *protocol.NotFoundPacket:
		n.handleNotFoundPacket(packet, conn)
	default:
		logger.Warn("Unknown packet type: %v", packet)
	}
}

func (n *Node) HandleUDPPackets(packet protocol.Packet, addr *net.UDPAddr) {
	switch data := packet.(type) {
	case *protocol.RequestChunksPacket:
		n.handleRequestChunksPacket(packet.(*protocol.RequestChunksPacket), addr)
	default:
		logger.Warn("Unknown packet type: %v", data)
	}
}

// Handler for when a node requests, to the tracker, the list of nodes who have a specific file
func (n *Node) handleAnswerNodesPacket(packet *protocol.AnswerNodesPacket, conn *transport.TCPConnection) {
	logger.Info("Answer nodes packet received from %s", conn.RemoteAddr())
	logger.Info("Number of nodes who got requested file: %d", packet.NumberOfNodes)

	// Update file in forDownload data structure
	forDownloadFile, ok := n.forDownload.Get(packet.FileName)
	if !ok {
		logger.Warn("File %s not found in forDownload files", packet.FileName)
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
		logger.Info("File %s downloaded successfully", packet.FileName)
		n.forDownload.Delete(packet.FileName)
		return
	}

	// Loop over nodes who have the file and request missing chunks
	for _, node := range packet.Nodes {
		missingChunksInNode := make([]uint16, 0)

		for _, chunk := range missingChunks {
			if protocol.GetBit(node.Bitfield, int(chunk)) {
				missingChunksInNode = append(missingChunksInNode, uint16(chunk))
			}
		}

		if len(missingChunksInNode) > 0 {
			udpAddr := &net.UDPAddr{
				IP:   node.IPAddr[:],
				Port: int(node.Port),
			}

			logger.Info("Sending request chunks packet to %s", udpAddr)

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
			logger.Info("Chunk %d of file %s downloaded", chunkPacket.Chunk, forDownloadFile.FileName)

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
		logger.Warn("Error creating UDP socket: %v", err)
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
			logger.Warn("Error reading packet: %v", err)
			return
		}

		switch packet := responsePacket.(type) {
		case *protocol.ChunkPacket:
			responsesChannel <- packet
		default:
			logger.Warn("Unknown packet type: %v", packet)
		}
	}
}

// Handler for when a node requests chunks of a file to this node
func (n *Node) handleRequestChunksPacket(packet *protocol.RequestChunksPacket, addr *net.UDPAddr) {
	logger.Info("Request chunks packet received from %s", addr)

	// Get file from published files
	publishedFile, ok := n.published.Get(packet.FileName)
	if !ok {
		logger.Warn("File %s not found in published files", packet.FileName)
		return
	}

	// Open file by the given path
	file, err := os.Open(publishedFile.Path)
	if err != nil {
		logger.Warn("Error opening file: %v", err)
		return
	}

	stats, _ := file.Stat()
	chunkSize := utils.ChunkSize(uint64(stats.Size()))

	// Send requested chunks
	for _, chunk := range packet.Chunks {
		logger.Info("Sending chunk %d of file %s to %s", chunk, packet.FileName, addr)

		// Seek to the beginning of the chunk
		_, err = file.Seek(int64(uint64(chunk)*chunkSize), 0)
		if err != nil {
			logger.Warn("Error seeking file: %v", err)
			return
		}

		// Read chunk bytes
		chunkContent := make([]byte, chunkSize)
		read, err := file.Read(chunkContent)
		if err != nil && !errors.Is(err, io.EOF) {
			logger.Warn("Error reading file: %v", err)
			return
		}

		// Send chunk bytes
		packet := protocol.NewChunkPacket(packet.FileName, chunk, chunkContent[:read])
		n.srv.SendPacket(&packet, addr)
	}
}

// Handler for when a node publishes a file in the network
func (n *Node) handlePublishFileSuccessPacket(packet *protocol.PublishFileSuccessPacket, conn *transport.TCPConnection) {
	logger.Info("File %s published in the network successfully", packet.FileName)

	// Remove file from pending and add it to published, since tracker has accepted it
	file, _ := n.pending.Get(packet.FileName)
	n.published.Put(packet.FileName, file)
	n.pending.Delete(packet.FileName)
}

// Handler for when the file, the node is trying to publish, already exists in the network
func (n *Node) handleAlreadyExistsPacket(packet *protocol.AlreadyExistsPacket, conn *transport.TCPConnection) {
	logger.Info("File %s already exists in the network", packet.Filename)

	// Remove file from pending, since tracker has rejected it
	n.pending.Delete(packet.Filename)
}

// Handler for when the file, the node is trying to download, does not exist in the network
func (n *Node) handleNotFoundPacket(packet *protocol.NotFoundPacket, conn *transport.TCPConnection) {
	logger.Info("File %s was not found in the network", packet.Filename)

	// Remove file from downloading, since it does not exist
	n.forDownload.Delete(packet.Filename)
}

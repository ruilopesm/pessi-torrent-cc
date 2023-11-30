package main

import (
	"PessiTorrent/internal/logger"
	"PessiTorrent/internal/protocol"
	"PessiTorrent/internal/transport"
	"PessiTorrent/internal/utils"
	"net"
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
		logger.Warn("Unknown packet type: %v.", packet)
	}
}

func (n *Node) HandleUDPPackets(packet protocol.Packet, addr *net.UDPAddr) {
	switch data := packet.(type) {
	case *protocol.ChunkPacket:
		n.handleChunkPacket(data, addr)
	default:
		logger.Warn("Unknown packet type: %v.", data)
	}
}

// Handler for when a node requests, to the tracker, the list of nodes who have a specific file
func (n *Node) handleAnswerNodesPacket(packet *protocol.AnswerNodesPacket, conn *transport.TCPConnection) {
	logger.Info("Answer nodes packet received from %s", conn.RemoteAddr())

	// Update file in forDownload data structure
	forDownloadFile, ok := n.forDownload.Get(packet.FileName)
	if !ok {
		logger.Warn("File %s not found in forDownload files", packet.FileName)
		return
	}

	forDownloadFile.SetData(packet.FileHash, packet.ChunkHashes, packet.FileSize, uint16(len(packet.ChunkHashes)))

	for _, node := range packet.Nodes {
		udpAddr := net.UDPAddr{
			IP:   node.IPAddr[:],
			Port: int(node.Port),
		}
		forDownloadFile.AddNode(&udpAddr, node.Bitfield)
	}

	logger.Info("File %s information internally updated. Run 'status' in order to check for downloading files", packet.FileName)
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

func (n *Node) handleChunkPacket(packet *protocol.ChunkPacket, addr *net.UDPAddr) {
	// Check if hash of chunk is correct
	forDownloadFile, ok := n.forDownload.Get(packet.FileName)
	if !ok {
		logger.Warn("File %s not found in forDownload files", packet.FileName)
		return
	}

	// Discard packet if chunk is already downloaded
	if forDownloadFile.ChunkAlreadyDownloaded(packet.Chunk) {
		return
	}

	// Discard packet if hash of chunk is not correct
	if forDownloadFile.GetChunkHash(packet.Chunk) != utils.HashChunk(packet.ChunkContent) {
		return
	}

	// Update chunk info
	forDownloadFile.MarkChunkAsDownloaded(packet.Chunk)

	logger.Info("Chunk %d of file %s received from %s", packet.Chunk, packet.FileName, addr)
}

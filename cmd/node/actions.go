package main

import (
	"PessiTorrent/internal/connection"
	"PessiTorrent/internal/protocol"
	"fmt"
)

func (n *Node) handlePublishFilePacket(packet *protocol.PublishFilePacket, conn *connection.Connection) {
	fmt.Printf("Publish file with name %s packet received from %s\n", packet.FileName, conn.RemoteAddr())

	file := NewFile(packet.FileName, packet.FileHash, packet.ChunkHashes)
	n.forDownload.Put(packet.FileName, &file)
}

func (n *Node) handlePublishFileSuccessPacket(packet *protocol.PublishFileSuccessPacket, conn *connection.Connection) {
	fmt.Printf("File %s published in the network successfully\n", packet.FileName)

	// Remove file from pending and add it to published, since tracker has accepted it
	file, _ := n.pending.Get(packet.FileName)
	n.published.Put(packet.FileName, file)
	n.pending.Delete(packet.FileName)
}

func (n *Node) handleAnswerNodesPacket(packet *protocol.AnswerNodesPacket, conn *connection.Connection) {
	fmt.Printf("Answer nodes packet received from %s\n", conn.RemoteAddr())
	fmt.Printf("Number of nodes who got requested file: %d\n", packet.NumberOfNodes)

	for _, node := range packet.Nodes {
		fmt.Printf("Node %v:%d has the file chunks %b\n", node.IPAddr, node.Port, node.Bitfield)
	}
}

func (n *Node) handleAlreadyExistsPacket(packet *protocol.AlreadyExistsPacket, conn *connection.Connection) {
	fmt.Printf("File %s already exists in the network\n", packet.Filename)

	// Remove file from pending, since tracker has rejected it
	n.pending.Delete(packet.Filename)
}

func (n *Node) handleNotFoundPacket(packet *protocol.NotFoundPacket, conn *connection.Connection) {
	fmt.Printf("File %s was not found in the network\n", packet.Filename)
}

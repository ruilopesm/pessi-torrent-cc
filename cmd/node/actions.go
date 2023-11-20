package main

import (
	"PessiTorrent/internal/connection"
	"PessiTorrent/internal/protocol"
	"fmt"
)

func (n *Node) handlePublishFilePacket(packet *protocol.PublishFilePacket, conn *connection.Connection) {
	fmt.Printf("Publish file packet received from %s\n", conn.RemoteAddr())

	file := NewFile(packet.FileName, packet.FileHash, packet.ChunkHashes)
	n.files.Put(packet.FileName, &file)
}

func (n *Node) handleAnswerNodesPacket(packet *protocol.AnswerNodesPacket, conn *connection.Connection) {
	fmt.Printf("Answer nodes packet received from %s\n", conn.RemoteAddr())
	fmt.Printf("Number of nodes who got requested file: %d\n", packet.NumberOfNodes)

	for _, node := range packet.Nodes {
		fmt.Printf("Node %v:%d has the file chunks %b\n", node.IPAddr, node.Port, node.Bitfield)
	}
}

func (n *Node) handleAlreadyExistsPacket(packet *protocol.AlreadyExistsPacket, conn *connection.Connection) {
	fmt.Printf("File %s already exists\n", packet.Filename)
}

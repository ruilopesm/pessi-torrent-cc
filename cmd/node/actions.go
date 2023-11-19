package main

import (
	"PessiTorrent/internal/connection"
	"PessiTorrent/internal/protocol"
	"fmt"
)

func (n *Node) handlePublishFilePacket(packet *protocol.PublishFilePacket, conn *connection.Connection) {
	fmt.Printf("publish file packet received from %s\n", conn.RemoteAddr())

	f := File{
		filename:    packet.FileName,
		fileHash:    packet.FileHash,
		chunkHashes: packet.ChunkHashes,
	}
	n.files.Put(packet.FileName, &f)
}

func (n *Node) handleAnswerNodesPacket(packet *protocol.AnswerNodesPacket, conn *connection.Connection) {
	fmt.Printf("answer nodes packet received from %s\n", conn.RemoteAddr())
	fmt.Printf("number of nodes who got requested file: %d\n", packet.NumberOfNodes)

	for _, node := range packet.Nodes {
		fmt.Printf("node %v:%d has the file chunks %b\n", node.IPAddr, node.Port, node.Bitfield)
	}
}

func (n *Node) handleAlreadyExistsPacket(packet *protocol.AlreadyExistsPacket, conn *connection.Connection) {
	fmt.Printf("file %s already exists\n", packet.FileName)
}

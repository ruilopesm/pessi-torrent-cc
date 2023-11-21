package main

import (
	"PessiTorrent/internal/protocol"
	"PessiTorrent/internal/transport"
	"fmt"
	"net"
)

func (n *Node) handlePublishFilePacket(packet *protocol.PublishFilePacket, conn *transport.TCPConnection) {
	fmt.Printf("Publish file with name %s packet received from %s\n", packet.FileName, conn.RemoteAddr())

	file := NewFile(packet.FileName, packet.FileHash, packet.ChunkHashes)
	n.forDownload.Put(packet.FileName, &file)
}

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

	// Dial, using udp, to each node and request the chunks
	for _, node := range packet.Nodes {
		fmt.Printf("Node %v:%d has the file chunks %b\n", node.IPAddr, node.Port, node.Bitfield)

		packet := protocol.NewRequestChunksPacket("filename.txt")
		udpAddr := net.UDPAddr{
			IP:   node.IPAddr[:],
			Port: int(node.Port),
		}
		n.srv.SendPacket(&packet, &udpAddr)
	}
}

func (n *Node) handleAlreadyExistsPacket(packet *protocol.AlreadyExistsPacket, conn *transport.TCPConnection) {
	fmt.Printf("File %s already exists in the network\n", packet.Filename)

	// Remove file from pending, since tracker has rejected it
	n.pending.Delete(packet.Filename)
}

func (n *Node) handleNotFoundPacket(packet *protocol.NotFoundPacket, conn *transport.TCPConnection) {
	fmt.Printf("File %s was not found in the network\n", packet.Filename)
}

func (n *Node) handleRequestChunksPacket(packet *protocol.RequestChunksPacket, addr *net.UDPAddr) {
	fmt.Printf("Request chunks packet received from %s\n", addr)
	fmt.Printf("File name: %s\n", packet.FileName)
}

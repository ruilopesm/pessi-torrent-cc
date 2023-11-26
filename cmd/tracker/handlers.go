package main

import (
	"PessiTorrent/internal/protocol"
	"PessiTorrent/internal/transport"
	"PessiTorrent/internal/utils"
	"fmt"
)

func (t *Tracker) handleInitPacket(packet *protocol.InitPacket, conn *transport.TCPConnection) {
	fmt.Printf("Init packet received from %s\n", conn.RemoteAddr())

	newNode := NewNodeInfo(*conn, packet.UDPPort)
	t.nodes.Add(&newNode)

	fmt.Printf("Registered node with data: %v, %v\n", packet.IPAddr, packet.UDPPort)
}

func (t *Tracker) handlePublishFilePacket(packet *protocol.PublishFilePacket, conn *transport.TCPConnection) {
	fmt.Printf("Publish file packet received from %s\n", conn.RemoteAddr())

	// If file already exists
	if t.files.Contains(packet.FileName) {
		fmt.Printf("File %s published from %s already exists\n", packet.FileName, conn.RemoteAddr())

		aePacket := protocol.NewAlreadyExistsPacket(packet.FileName)
		conn.EnqueuePacket(&aePacket)
		return
	}

	// Add file to the tracker
	file := NewTrackedFile(packet.FileName, packet.FileSize, packet.FileHash, packet.ChunkHashes)
	t.files.Put(packet.FileName, &file)

	// Add file to the node's list of files
	t.nodes.ForEach(func(node *NodeInfo) {
		if node.conn.RemoteAddr() == conn.RemoteAddr() {
			newSharedFile := NewSharedFile(packet.FileName, packet.FileSize, packet.FileHash, packet.ChunkHashes)
			node.AddFile(newSharedFile)
		}
	})

	// Send response back to the node
	pfsPacket := protocol.NewPublishFileSuccessPacket(packet.FileName)
	conn.EnqueuePacket(&pfsPacket)
}

func (t *Tracker) handleRequestFilePacket(packet *protocol.RequestFilePacket, conn *transport.TCPConnection) {
	fmt.Printf("Request file packet received from %s\n", conn.RemoteAddr())

	if t.files.Contains(packet.FileName) {
		var nNodes uint16
		var ipAddrs [][4]byte
		var ports []uint16
		var bitfields [][]uint16

		t.nodes.ForEach(func(node *NodeInfo) {
			if file, exists := node.files.Get(packet.FileName); exists {
				nNodes++
				ipAddrs = append(ipAddrs, utils.TCPAddrToBytes(node.conn.RemoteAddr()))
				ports = append(ports, uint16(node.udpPort))
				bitfields = append(bitfields, file.Bitfield)
			}
		})

		file, _ := t.files.Get(packet.FileName)

		// Send file name, hash and chunks hashes
		anPacket := protocol.NewAnswerNodesPacket(file.FileName, file.FileSize, file.FileHash, file.ChunkHashes, nNodes, ipAddrs, ports, bitfields)
		conn.EnqueuePacket(&anPacket)
	} else {
		fmt.Printf("File %s requested from %s does not exist\n", packet.FileName, conn.RemoteAddr())

		nfPacket := protocol.NewNotFoundPacket(packet.FileName)
		conn.EnqueuePacket(&nfPacket)
	}
}

func (t *Tracker) handleRemoveFilePacket(packet *protocol.RemoveFilePacket, conn *transport.TCPConnection) {
	fmt.Printf("Remove file packet received from %s\n", conn.RemoteAddr())

	t.nodes.ForEach(func(node *NodeInfo) {
		if node.conn.RemoteAddr() == conn.RemoteAddr() {
			node.files.Delete(packet.FileName)
		}
	})
}

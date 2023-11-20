package main

import (
	"PessiTorrent/internal/connection"
	"PessiTorrent/internal/protocol"
	"PessiTorrent/internal/structures"
	"PessiTorrent/internal/utils"
	"fmt"
)

func (t *Tracker) handleInitPacket(packet *protocol.InitPacket, conn *connection.Connection) {
	fmt.Printf("Init packet received from %s\n", conn.RemoteAddr())

	info := NodeInfo{
		conn:    *conn,
		udpPort: packet.UDPPort,
		files:   structures.NewSynchronizedMap[NodeFile](),
	}
	t.nodes.Add(&info)

	fmt.Printf("Registered node with data: %v, %v\n", packet.IPAddr, packet.UDPPort)
}

func (t *Tracker) handlePublishFilePacket(packet *protocol.PublishFilePacket, conn *connection.Connection) {
	fmt.Printf("Publish file packet received from %s\n", conn.RemoteAddr())

	// Add file to the tracker
	file := NewFile(packet.FileName, packet.FileHash, packet.ChunkHashes)
	t.AddFile(file)

	// Add file to the node's list of files
	t.nodes.ForEach(func(node *NodeInfo) {
		if node.conn.RemoteAddr() == conn.RemoteAddr() {
			f := NodeFile{
				file:            &file,
				chunksAvailable: protocol.NewCheckedBitfield(len(file.chunkHashes)),
			}
			node.files.Put(file.filename, f)
		}
	})
}

func (t *Tracker) handleRequestFilePacket(packet *protocol.RequestFilePacket, conn *connection.Connection) {
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
				bitfields = append(bitfields, file.chunksAvailable)
			}
		})

		// Send file/chunks hashes
		file, _ := t.files.Get(packet.FileName)
		pfPacket := protocol.NewPublishFilePacket(file.filename, file.fileHash, file.chunkHashes)
		conn.EnqueuePacket(&pfPacket)

		// Send nodes info
		anPacket := protocol.NewAnswerNodesPacket(nNodes, ipAddrs, ports, bitfields)
		conn.EnqueuePacket(&anPacket)
	} else {
		// TODO: send file not found packet
		fmt.Printf("File %s not found\n", packet.FileName)
	}
}

func (t *Tracker) handleRemoveFilePacket(packet *protocol.RemoveFilePacket, conn *connection.Connection) {
	fmt.Printf("Remove file packet received from %s\n", conn.RemoteAddr())

	t.nodes.ForEach(func(node *NodeInfo) {
		if node.conn.RemoteAddr() != conn.RemoteAddr() {
			node.files.Delete(packet.FileName)
		}
	})
}

package main

import (
	"PessiTorrent/internal/connection"
	"PessiTorrent/internal/protocol"
	"PessiTorrent/internal/structures"
	"PessiTorrent/internal/utils"
	"fmt"
)

func (t *Tracker) handleInitPacket(packet *protocol.InitPacket, conn *connection.Connection) {
	fmt.Printf("init packet received from %s\n", conn.RemoteAddr())

	info := NodeInfo{
		conn:    *conn,
		udpPort: packet.UDPPort,
		files:   structures.NewSynchronizedMap[NodeFile](),
	}
	t.nodes.Add(&info)

	fmt.Printf("registered node with data: %v, %v\n", packet.IPAddr, packet.UDPPort)
}

func (t *Tracker) handlePublishFilePacket(packet *protocol.PublishFilePacket, conn *connection.Connection) {
	fmt.Printf("publish file packet received from %s\n", conn.RemoteAddr())

	// Add file to the tracker
	file := File{
		name:     packet.FileName,
		fileHash: packet.FileHash,
		hashes:   packet.ChunkHashes,
	}
	t.files.Put(file.name, &file)

	// Add file to the node's list of files
	t.nodes.ForEach(func(node *NodeInfo) {
		if node.conn.RemoteAddr() == conn.RemoteAddr() {
			chunksAvailable := make([]uint16, len(file.hashes))
			for i := 0; i < len(file.hashes); i++ {
				chunksAvailable[i] = uint16(i)
			}

			f := NodeFile{
				file:            &file,
				chunksAvailable: chunksAvailable,
			}
			node.files.Put(file.name, f)
		}
	})
}

func (t *Tracker) handleRequestFilePacket(packet *protocol.RequestFilePacket, conn *connection.Connection) {
	fmt.Printf("request file packet received from %s\n", conn.RemoteAddr())

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

		// Send file hashes
		file, _ := t.files.Get(packet.FileName)
		var pfPacket protocol.PublishFilePacket
		pfPacket.Create(file.name, file.fileHash, file.hashes)
		conn.EnqueuePacket(&pfPacket)

		// Send nodes info
		var anPacket protocol.AnswerNodesPacket
		anPacket.Create(nNodes, ipAddrs, ports, bitfields)
		conn.EnqueuePacket(&anPacket)
	} else {
		// TODO: send file not found packet
		fmt.Printf("file %s not found\n", packet.FileName)
	}
}

func (t *Tracker) handleRemoveFilePacket(packet *protocol.RemoveFilePacket, conn *connection.Connection) {
	fmt.Printf("remove file packet received from %s\n", conn.RemoteAddr())

	t.nodes.ForEach(func(node *NodeInfo) {
		if node.conn.RemoteAddr() != conn.RemoteAddr() {
			node.files.Delete(packet.FileName)
		}
	})
}

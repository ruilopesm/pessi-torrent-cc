package main

import (
	"PessiTorrent/internal/connection"
	"PessiTorrent/internal/protocol"
	"PessiTorrent/internal/structures"
	"PessiTorrent/internal/utils"
	"fmt"
)

// TODO: remove this
// func (t *Tracker) HandleprotocolDispatcher(packet interface{}, packetType uint8, conn *connection.Connection) {
// 	switch packetType {
// 	case protocol.InitType:
// 		t.handleInitPacket(packet.(*protocol.InitPacket), conn)
// 	case protocol.PublishFileType:
// 		t.handlePublishFilePacket(packet.(*protocol.PublishFilePacket), conn)
// 	case protocol.RequestFileType:
// 		t.handleRequestFilePacket(packet.(*protocol.RequestFilePacket), conn)
// 	case protocol.RemoveFileType:
// 		t.handleRemoveFilePacket(packet.(*protocol.RemoveFilePacket), conn)
// 	default:
// 		fmt.Println("unknown packet type")
// 	}
// }

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
			f := NodeFile{
				file: &file,
				chunksAvailable: make([]uint16, len(file.hashes)),
			}
			node.files.Put(file.name, f)
		}
	})
}

func (t *Tracker) handleRequestFilePacket(packet *protocol.RequestFilePacket, conn *connection.Connection) {
	fmt.Printf("request file packet received from %s\n", conn.RemoteAddr())

	if t.files.Contains(packet.FileName) {
		var sequenceNumber uint8 = 0
		var packetsToSend []protocol.AnswerNodesPacket

    var nodes []NodeInfo
		t.nodes.ForEach(func(node *NodeInfo) {
			if file, exists := node.files.Get(packet.FileName); exists {
				var packet protocol.AnswerNodesPacket
				packet.Create(
					sequenceNumber,
					utils.TCPAddrToBytes(node.conn.RemoteAddr()),
					node.udpPort,
					file.chunksAvailable,
				)
				packetsToSend = append(packetsToSend, packet)
				sequenceNumber++
			}
		})

		file, _ := t.files.Get(packet.FileName)
		var packet protocol.PublishFilePacket
		packet.Create(file.name, file.fileHash, file.hashes)
		conn.EnqueuePacket(&packet)
		// Send protocol in reverse order
		for i := len(packetsToSend) - 1; i >= 0; i-- {
			conn.EnqueuePacket(&packetsToSend[i])
		}
	} else {
		// TODO: send file not found packet
		fmt.Printf("file %s not found\n", packet.FileName)
		return
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

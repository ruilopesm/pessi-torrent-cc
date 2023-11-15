package main

import (
	"PessiTorrent/internal/connection"
	"PessiTorrent/internal/packets"
	"PessiTorrent/internal/structures"
	"PessiTorrent/internal/utils"
	"fmt"
)

func (t *Tracker) HandlePacketsDispatcher(packet interface{}, packetType uint8, conn *connection.Connection) {
	switch packetType {
	case packets.InitType:
		t.handleInitPacket(packet.(*packets.InitPacket), conn)
	case packets.PublishFileType:
		t.handlePublishFilePacket(packet.(*packets.PublishFilePacket), conn)
	case packets.RequestFileType:
		t.handleRequestFilePacket(packet.(*packets.RequestFilePacket), conn)
	case packets.REMOVE_FILE_TYPE:
		t.handleRemoveFilePacket(packet.(*packets.RemoveFilePacket), conn)
	default:
		fmt.Println("unknown packet type")
	}
}

func (t *Tracker) handleInitPacket(packet *packets.InitPacket, conn *connection.Connection) {
	fmt.Printf("init packet received from %s\n", conn.RemoteAddr())

	info := NodeInfo{
		conn:    *conn,
		udpPort: packet.UDPPort,
		files:   structures.NewSynchronizedMap[NodeFile](),
	}
	t.nodes.Add(&info)

	fmt.Printf("registered node with data: %v, %v\n", packet.IPAddr, packet.UDPPort)
}

func (t *Tracker) handlePublishFilePacket(packet *packets.PublishFilePacket, conn *connection.Connection) {
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
		if node.conn.RemoteAddr() != conn.RemoteAddr() {
			f := NodeFile{
				file:            &file,
				chunksAvailable: make([]uint8, len(file.hashes)),
			}
			node.files.Put(file.name, f)
		}
	})
}

func (t *Tracker) handleRequestFilePacket(packet *packets.RequestFilePacket, conn *connection.Connection) {
	fmt.Printf("request file packet received from %s\n", conn.RemoteAddr())

	if t.files.Contains(packet.FileName) {
		var sequenceNumber uint8 = 0
		var packetsToSend []packets.AnswerNodesPacket

		t.nodes.ForEach(func(node *NodeInfo) {
			if file, exists := node.files.Get(packet.FileName); exists {
				var packet packets.AnswerNodesPacket
				packet.Create(
					sequenceNumber,
					utils.TCPAddrToBytes(node.conn.RemoteAddr()),
					node.udpPort,
					file.chunksAvailable,
				)
				packetsToSend = append(packetsToSend, packet)

				fmt.Printf("sent answer nodes packet to %s\n", conn.RemoteAddr())
				sequenceNumber++
			}
		})

		file, _ := t.files.Get(packet.FileName)
		var packet packets.PublishFilePacket
		packet.Create(file.name, file.fileHash, file.hashes)
		err := conn.WritePacket(packet)
		if err != nil {
			fmt.Printf("error sending publish file packet to %s\n", conn.RemoteAddr())
		}

		// Send packets in reverse order
		for i := len(packetsToSend) - 1; i >= 0; i-- {
			err := conn.WritePacket(packetsToSend[i])
			if err != nil {
				fmt.Printf("error sending answer nodes packet to %s\n", conn.RemoteAddr())
			}
		}
	} else {
		// TODO: send file not found packet
		fmt.Printf("file %s not found\n", packet.FileName)
		return
	}
}

func (t *Tracker) handleRemoveFilePacket(packet *packets.RemoveFilePacket, conn *connection.Connection) {
	fmt.Printf("remove file packet received from %s\n", conn.RemoteAddr())

	t.nodes.ForEach(func(node *NodeInfo) {
		if node.conn.RemoteAddr() != conn.RemoteAddr() {
			node.files.Delete(packet.FileName)
		}
	})
}

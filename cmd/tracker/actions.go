package main

import (
	"PessiTorrent/internal/connection"
	"PessiTorrent/internal/packets"
	"PessiTorrent/internal/utils"
	"fmt"
)

func (t *Tracker) HandlePacketsDispatcher(packet interface{}, packetType uint8, conn *connection.Connection) {
	switch packetType {
	case packets.INIT_TYPE:
		t.handleInitPacket(packet.(*packets.InitPacket), conn)
	case packets.PUBLISH_FILE_TYPE:
		t.handlePublishFilePacket(packet.(*packets.PublishFilePacket), conn)
	case packets.REQUEST_FILE_TYPE:
		t.handleRequestFilePacket(packet.(*packets.RequestFilePacket), conn)
	default:
		fmt.Println("unknown packet type")
	}
}

func (t *Tracker) handleInitPacket(packet *packets.InitPacket, conn *connection.Connection) {
	fmt.Printf("init packet received from %s\n", conn.RemoteAddr())

	t.nodes.Lock()
	defer t.nodes.Unlock()

	info := &NodeInfo{
		conn:    *conn,
		udpPort: packet.UDPPort,
		files:   SynchronizedMap[NodeFile]{m: make(map[string]NodeFile)},
	}

	t.nodes.l = append(t.nodes.l, info)
	fmt.Printf("registered node with data: %v, %v\n", packet.IPAddr, packet.UDPPort)
}

func (t *Tracker) handlePublishFilePacket(packet *packets.PublishFilePacket, conn *connection.Connection) {
	fmt.Printf("publish file packet received from %s\n", conn.RemoteAddr())
	t.files.Lock()
	defer t.files.Unlock()

	file := &File{
		name:     packet.FileName,
		fileHash: packet.FileHash,
		hashes:   packet.ChunkHashes,
	}

	t.files.m[file.name] = file

	t.nodes.Lock()
	defer t.nodes.Unlock()

	// Add file to the node's list of files
	for _, node := range t.nodes.l {
		if node.conn.RemoteAddr() == conn.RemoteAddr() {
			node.files.Lock()
			defer node.files.Unlock()
			node.files.m[file.name] = NodeFile{
				file: file,
				// FIXME: should be a bitfield all zeros
				chunksAvailable: []uint8{0, 2, 7, 10},
			}
		}
	}
}

func (t *Tracker) handleRequestFilePacket(packet *packets.RequestFilePacket, conn *connection.Connection) {
	fmt.Printf("request file packet received from %s\n", conn.RemoteAddr())
	t.files.RLock()
	defer t.files.RUnlock()

	if _, ok := t.files.m[packet.FileName]; ok {
		t.nodes.RLock()
		defer t.nodes.RUnlock()

		var sequenceNumber uint8 = 0
		var packetsToSend []packets.AnswerNodesPacket

		for _, node := range t.nodes.l {
			node.files.RLock()
			defer node.files.RUnlock()

			if file, ok := node.files.m[packet.FileName]; ok {
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
		}

    fileName := packet.FileName
    file := t.files.m[fileName]
    var packet packets.PublishFilePacket
    packet.Create(file.name, file.fileHash, file.hashes)
    conn.WritePacket(packet)


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

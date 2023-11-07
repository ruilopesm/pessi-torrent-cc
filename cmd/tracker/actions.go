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
		udpPort: packet.UDPListenPort,
		files:   SynchronizedMap[NodeFile]{m: make(map[string]NodeFile)},
	}

	t.nodes.l = append(t.nodes.l, info)
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
	t.files.Lock()
	defer t.files.Unlock()

	if _, ok := t.files.m[packet.FileName]; ok {
		t.nodes.Lock()
		defer t.nodes.Unlock()

		for _, node := range t.nodes.l {
			node.files.Lock()
			defer node.files.Unlock()

			if _, ok := node.files.m[packet.FileName]; ok {
        nodeIp, err := utils.IPv4ToByteArray(conn.RemoteAddr())

        // TODO: store this toSend packets inside an array and send them afterwards
				var toSend packets.AnswerNodesPacket
				toSend.Create(
					// FIXME: should be a sequence number
					0,
					nodeIp,
					node.udpPort,
					node.files.m[packet.FileName].chunksAvailable,
				)
				err = conn.WritePacket(toSend)
				if err != nil {
					fmt.Println(err)
				}
			}
		}
	} else {
		// TODO: send file not found packet
		fmt.Printf("file %s not found\n", packet.FileName)
		return
	}
}

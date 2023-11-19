package main

import (
  "PessiTorrent/internal/protocol"
  "PessiTorrent/internal/connection"
  "fmt"
)

func (n *Node) handleFileHashesPacket(packet *protocol.PublishFilePacket, conn *connection.Connection) {
  fmt.Printf("publish file packet received from %s\n", conn.RemoteAddr())

	f := File{
		filename:    packet.FileName,
		fileHash:    packet.FileHash,
		chunkHashes: packet.ChunkHashes,
	}
  n.files.Put(packet.FileName, &f)
}

func (n *Node) handleFile

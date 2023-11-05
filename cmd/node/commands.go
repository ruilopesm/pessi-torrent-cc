package main

import (
	"PessiTorrent/internal/packets"
	"crypto/sha1"
	"fmt"
)

// request <file>
func (n *Node) requestFile(args []string) error {
	filename := args[0]

	var packet packets.RequestFilePacket
	packet.Create(filename)
	err := n.conn.WritePacket(packet)
	if err != nil {
		return err
	}

	// Read response packets
	for {
		receivedPacket, _, err := n.conn.ReadPacket()
		if err != nil {
			return err
		}

		fmt.Printf("node %v has file %s\n", receivedPacket.(*packets.AnswerNodesPacket).NodeIdentifier, filename)

		if receivedPacket.(*packets.AnswerNodesPacket).SequenceNumber == 0 {
			fmt.Println("no more packets")
			break
		}
	}

	return nil
}

// publish <file>
func (n *Node) publishFile(args []string) error {
	filename := args[0]

	var packet packets.PublishFilePacket
	// FIXME: create hash from file content and create hashes for all chunks
	packet.Create(filename, sha1.Sum([]byte(filename)), make([][20]byte, 0))
	err := n.conn.WritePacket(packet)
	if err != nil {
		return err
	}

	return nil
}

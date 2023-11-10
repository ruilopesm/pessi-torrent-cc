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
	// TODO: handle file not found packet
	responsePacket, _, err := n.conn.ReadPacket()
	fmt.Printf("node %v, %v has the file\n", responsePacket.(*packets.AnswerNodesPacket).NodeIPAddr, responsePacket.(*packets.AnswerNodesPacket).UDPPort)
	if err != nil {
		return err
	}

	// Read response packets
	for {
		if responsePacket.(*packets.AnswerNodesPacket).SequenceNumber == 0 {
			break
		}

		responsePacket, _, err = n.conn.ReadPacket()
		fmt.Printf("node %v, %v has the file\n", responsePacket.(*packets.AnswerNodesPacket).NodeIPAddr, responsePacket.(*packets.AnswerNodesPacket).UDPPort)
		if err != nil {
			return err
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

package main

import (
	"PessiTorrent/internal/connection"
	"PessiTorrent/internal/packets"
	"log"
	"net"
)

type Node struct {
	serverAddr string
}

func NewNode(serverAddr string) *Node {
	return &Node{
		serverAddr: serverAddr,
	}
}

func (n *Node) Start() error {
	conn, err := net.Dial("tcp", n.serverAddr)
	if err != nil {
		return err
	}
	defer conn.Close()

	c := connection.NewConnection(conn)

	files := []string{"file.txt", "file2.txt", "file3.txt", "file4.txt", "file5.txt"}

	for _, file := range files {
		var packet packets.RequestFilePacket
		packet.Create(file)
		err = c.WritePacket(packet)
		if err != nil {
			return err
		}
	}

	return nil
}

func main() {
	node := NewNode("localhost:8080")
	err := node.Start()
	if err != nil {
		log.Fatal(err)
	}
}

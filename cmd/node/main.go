package main

import (
	"PessiTorrent/internal/cli"
	"PessiTorrent/internal/connection"
	"log"
	"net"
)

type Node struct {
	serverAddr string
	conn       connection.Connection
	quitch     chan struct{}
}

func NewNode(serverAddr string) Node {
	return Node{
		serverAddr: serverAddr,
		quitch:     make(chan struct{}),
	}
}

func (n *Node) Start() error {
	conn, err := net.Dial("tcp", n.serverAddr)
	if err != nil {
		return err
	}
	defer conn.Close()

	n.conn = connection.NewConnection(conn)
	cli := cli.NewCLI(n.stop)
	cli.AddCommand("requestFile", "<file>", 1, n.requestFile)

	go cli.Start()

	<-n.quitch

	return nil
}

func (n *Node) stop() {
	n.quitch <- struct{}{}
}

func main() {
	node := NewNode("localhost:42069")
	err := node.Start()
	if err != nil {
		log.Fatal(err)
	}
}

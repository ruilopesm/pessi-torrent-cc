package main

import (
	"PessiTorrent/internal/cli"
	"PessiTorrent/internal/connection"
	"log"
	"net"
	"os"
)

type Node struct {
	serverAddr string
	conn       *connection.Connection
	quitch     chan struct{}
	commands   map[string]cli.Command
}

func NewNode(serverAddr string) *Node {
	return &Node{
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
	n.SetCommands()

	go cli.StartCLI(n.commands, n.quitch)

	<-n.quitch

	return nil
}

func main() {
	ip := os.Args[1]
	port := os.Args[2]

	node := NewNode(ip + ":" + port)
	err := node.Start()
	if err != nil {
		log.Fatal(err)
	}
}

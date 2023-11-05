package main

import (
	"PessiTorrent/internal/cli"
	"PessiTorrent/internal/connection"
	"PessiTorrent/internal/packets"
	"fmt"
	"log"
	"net"
	"os"
)

type Node struct {
	serverAddr string
	UDPAddr    net.UDPAddr
	conn       connection.Connection
	quitch     chan struct{}
}

func NewNode(serverAddr string, udpListenPort string) Node {
	s, err := net.ResolveUDPAddr("udp4", ":"+udpListenPort)
	if err != nil {
		fmt.Println("error resolving udp addr: ", err)
		os.Exit(1)
	}

	return Node{
		serverAddr: serverAddr,
		UDPAddr:    *s,
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

	var packet packets.InitPacket
	packet.Create(n.UDPAddr.String(), uint16(n.UDPAddr.Port))
	err = n.conn.WritePacket(packet)
	if err != nil {
		return err
	}

	cli := cli.NewCLI(n.stop)
	cli.AddCommand("request", "<file>", 1, n.requestFile)
	cli.AddCommand("publish", "<file>", 1, n.publishFile)
	go cli.Start()

	<-n.quitch

	return nil
}

func (n *Node) stop() {
	n.quitch <- struct{}{}
}

func main() {
	if len(os.Args) != 2 {
		log.Fatal("Usage: node <udp port>")
	}

	node := NewNode("localhost:42069", os.Args[1])
	err := node.Start()
	if err != nil {
		log.Fatal(err)
	}
}

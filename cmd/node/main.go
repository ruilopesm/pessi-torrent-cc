package main

import (
	"PessiTorrent/internal/cli"
	"PessiTorrent/internal/connection"
	"PessiTorrent/internal/protocol"
	"PessiTorrent/internal/structures"
	"PessiTorrent/internal/utils"
	"fmt"
	"net"
	"os"
)

type Node struct {
	ipAddr     [4]byte
	serverAddr string
	udpPort    uint16
	conn       connection.Connection
	files      structures.SynchronizedMap[*File]
	quitch     chan struct{}
}

func NewNode(serverAddr string, listenUDPPort string) Node {
	udpPort, err := utils.StrToUDPPort(listenUDPPort)
	if err != nil {
		fmt.Println("Error parsing udp port:", err)
		os.Exit(1)
	}

	return Node{
		serverAddr: serverAddr,
		udpPort:    uint16(udpPort),
		files:      structures.NewSynchronizedMap[*File](),
		quitch:     make(chan struct{}),
	}
}

func (n *Node) handlePacket(packet interface{}, conn *connection.Connection) {
	switch data := packet.(type) {
	case *protocol.PublishFilePacket:
		n.handlePublishFilePacket(packet.(*protocol.PublishFilePacket), conn)
	case *protocol.AnswerNodesPacket:
		n.handleAnswerNodesPacket(packet.(*protocol.AnswerNodesPacket), conn)
	case *protocol.AlreadyExistsPacket:
		n.handleAlreadyExistsPacket(packet.(*protocol.AlreadyExistsPacket), conn)
	default:
		fmt.Println("Unknown packet type received:", data)
	}
}

func (n *Node) Start() error {
	conn, err := net.Dial("tcp4", n.serverAddr)
	if err != nil {
		return err
	}
	defer conn.Close()

	n.conn = connection.NewConnection(conn, n.handlePacket)
	n.ipAddr = utils.TCPAddrToBytes(conn.LocalAddr())
	go n.conn.Start()

	// TODO: Listen on udp

	// Notify tracker of the node's existence
	packet := protocol.NewInitPacket(n.ipAddr, n.udpPort)
	n.conn.EnqueuePacket(&packet)

	go n.StartCLI()

	<-n.quitch

	return nil
}

func (n *Node) StartCLI() {
	c := cli.NewCLI(n.Stop)
	c.AddCommand("publish", "<file name>", "", 1, n.publish)
	c.AddCommand("request", "<file name>", "", 1, n.requestFile)
	c.AddCommand("status", "", "Show the status of the node", 0, n.status)
	c.AddCommand("remove", "<file name>", "", 1, n.removeFile)
	c.Start()
}

func (n *Node) Stop() {
	n.quitch <- struct{}{}
}

func main() {
	if len(os.Args) != 3 {
		fmt.Println("Usage: node <server ip:port> <udp port>")
		return
	}

	// TODO: check if port is inside udp range

	node := NewNode(os.Args[1], os.Args[2])
	err := node.Start()
	if err != nil {
		fmt.Println("Error starting node:", err)
	}
}

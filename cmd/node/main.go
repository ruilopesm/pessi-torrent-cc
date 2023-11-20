package main

import (
	"PessiTorrent/internal/cli"
	"PessiTorrent/internal/connection"
	"PessiTorrent/internal/protocol"
	"PessiTorrent/internal/structures"
	"PessiTorrent/internal/utils"
	"fmt"
	"log"
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
	go n.conn.Start()
	n.ipAddr = utils.TCPAddrToBytes(conn.LocalAddr())

	// TODO: Listen on udp

	packet := protocol.NewInitPacket(n.ipAddr, n.udpPort)
	n.conn.EnqueuePacket(&packet)

	cli := cli.NewCLI(n.stop)
	cli.AddCommand("request", "<file name>", 1, n.requestFile)
	cli.AddCommand("publish", "<path>", 1, n.publish)
	cli.AddCommand("status", "", 0, n.status)
	cli.AddCommand("remove", "<file name>", 1, n.removeFile)
	go cli.Start()

	<-n.quitch

	return nil
}

func (n *Node) stop() {
	n.quitch <- struct{}{}
}

func main() {
	if len(os.Args) != 3 {
		log.Fatal("Usage: node <server address> <udp port>")
		return
	}

	// TODO: check if port is inside udp range

	node := NewNode(os.Args[1], os.Args[2])
	err := node.Start()
	if err != nil {
		log.Fatal(err)
	}
}

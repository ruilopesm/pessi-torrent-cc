package main

import (
	"PessiTorrent/internal/cli"
	"PessiTorrent/internal/protocol"
	"PessiTorrent/internal/structures"
	"PessiTorrent/internal/transport"
	"PessiTorrent/internal/utils"
	"fmt"
	"net"
	"os"
)

type Node struct {
	serverAddr string
	udpPort    uint16

	conn transport.TCPConnection
	srv  transport.UDPServer

	published   structures.SynchronizedMap[*File]
	pending     structures.SynchronizedMap[*File]
	forDownload structures.SynchronizedMap[*ForDownloadFile]

	quitch chan struct{}
}

func NewNode(serverAddr string, listenUDPPort string) Node {
	udpPort, err := utils.StrToUDPPort(listenUDPPort)
	if err != nil {
		fmt.Println("Error parsing UDP port:", err)
		os.Exit(1)
	}

	return Node{
		serverAddr: serverAddr,
		udpPort:    uint16(udpPort),

		pending:     structures.NewSynchronizedMap[*File](),
		published:   structures.NewSynchronizedMap[*File](),
		forDownload: structures.NewSynchronizedMap[*ForDownloadFile](),

		quitch: make(chan struct{}),
	}
}

func (n *Node) handleTCPPackets(packet interface{}, conn *transport.TCPConnection) {
	switch data := packet.(type) {
	case *protocol.PublishFileSuccessPacket:
		n.handlePublishFileSuccessPacket(packet.(*protocol.PublishFileSuccessPacket), conn)
	case *protocol.AnswerNodesPacket:
		n.handleAnswerNodesPacket(packet.(*protocol.AnswerNodesPacket), conn)
	case *protocol.AlreadyExistsPacket:
		n.handleAlreadyExistsPacket(packet.(*protocol.AlreadyExistsPacket), conn)
	case *protocol.NotFoundPacket:
		n.handleNotFoundPacket(packet.(*protocol.NotFoundPacket), conn)
	default:
		fmt.Println("Unknown packet type received:", data)
	}
}

func (n *Node) handleUDPPackets(packet interface{}, addr *net.UDPAddr) {
	switch data := packet.(type) {
	case *protocol.RequestChunksPacket:
		n.handleRequestChunksPacket(packet.(*protocol.RequestChunksPacket), addr)
	default:
		fmt.Println("Unknown packet type received:", data)
	}
}

func (n *Node) Start() error {
	// Dial tracker using TCP
	conn, err := net.Dial("tcp4", n.serverAddr)
	if err != nil {
		return err
	}
	defer conn.Close()

	n.conn = transport.NewTCPConnection(conn, n.handleTCPPackets)
	go n.conn.Start()

	// Listen on UDP
	udpAddr := net.UDPAddr{
		IP:   net.IPv4zero,
		Port: int(n.udpPort),
	}

	udpConn, err := net.ListenUDP("udp4", &udpAddr)
	if err != nil {
		return err
	}
	defer udpConn.Close()

	n.srv = transport.NewUDPServer(*udpConn, n.handleUDPPackets)
	go n.srv.Start()

	fmt.Println("Node listening UDP on", udpConn.LocalAddr())

	// Notify tracker of the node's existence
	ipAddr := utils.TCPAddrToBytes(n.conn.LocalAddr())
	packet := protocol.NewInitPacket(ipAddr, n.udpPort)
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
		fmt.Println("Usage: node <server ip:port> <UDP port>")
		return
	}

	// TODO: check if port is inside UDP range

	node := NewNode(os.Args[1], os.Args[2])
	err := node.Start()
	if err != nil {
		fmt.Println("Error starting node:", err)
	}
}

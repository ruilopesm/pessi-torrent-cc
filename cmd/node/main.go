package main

import (
	"PessiTorrent/internal/cli"
	"PessiTorrent/internal/config"
	"PessiTorrent/internal/logger"
	"PessiTorrent/internal/protocol"
	"PessiTorrent/internal/structures"
	"PessiTorrent/internal/transport"
	"PessiTorrent/internal/utils"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
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

	n.conn = transport.NewTCPConnection(conn, n.handleTCPPackets, n.Stop)
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

	n.srv = transport.NewUDPServer(*udpConn, n.handleUDPPackets, func() {})
	go n.srv.Start()

	// Notify tracker of the node's existence
	ipAddr := utils.TCPAddrToBytes(n.conn.LocalAddr())
	packet := protocol.NewInitPacket(ipAddr, n.udpPort)
	n.conn.EnqueuePacket(&packet)

	console := cli.NewConsole()
	logger.SetLogger(&console)
	defer console.Close()

	go n.StartCLI(console)

	logger.Info("Node listening UDP on %v", udpConn.LocalAddr())

	<-n.quitch

	return nil
}

func (n *Node) StartCLI(console cli.Console) {
	c := cli.NewCLI(n.Stop, console)
	c.AddCommand("publish", "<file name>", "", 1, n.publish)
	c.AddCommand("request", "<file name>", "", 1, n.requestFile)
	c.AddCommand("status", "", "Show the status of the node", 0, n.status)
	c.AddCommand("remove", "<file name>", "", 1, n.removeFile)
	c.Start()
}

func (n *Node) Stop() {
	n.srv.Stop()
	n.quitch <- struct{}{}
}

func main() {
	// TODO: check if port is inside udp range

	conf, err := config.NewConfig(config.ConfigPath)
	if err != nil {
		log.Panic("Error reading config:", err)
		return
	}
	trackerAddr := conf.Tracker.Host + ":" + strconv.Itoa(conf.Tracker.Port)
	udpPort := strconv.Itoa(conf.Node.Port)

	flag.StringVar(&trackerAddr, "t", trackerAddr, "The address of the tracker")
	flag.StringVar(&udpPort, "p", udpPort, "The port to listen on")
	flag.Parse()

	node := NewNode(trackerAddr, udpPort)

	err = node.Start()
	if err != nil {
		fmt.Println("Error starting node: ", err)
	}
}

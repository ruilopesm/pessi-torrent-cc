package main

import (
	"PessiTorrent/internal/cli"
	"PessiTorrent/internal/logger"
	"PessiTorrent/internal/protocol"
	"PessiTorrent/internal/structures"
	"PessiTorrent/internal/ticker"
	"PessiTorrent/internal/transport"
	"PessiTorrent/internal/utils"
	"net"
)

type Node struct {
	trackerAddr string
	udpPort     uint16
	connected   bool // Whether the node is connected to the tracker or not

	conn transport.TCPConnection
	srv  transport.UDPServer
	tck  ticker.Ticker

	published   structures.SynchronizedMap[string, *File]
	pending     structures.SynchronizedMap[string, *File]
	forDownload structures.SynchronizedMap[string, *ForDownloadFile]

	quitChannel chan struct{}
}

func NewNode(trackerAddr string, udpPort uint16) Node {
	return Node{
		trackerAddr: trackerAddr,
		udpPort:     udpPort,

		pending:     structures.NewSynchronizedMap[string, *File](),
		published:   structures.NewSynchronizedMap[string, *File](),
		forDownload: structures.NewSynchronizedMap[string, *ForDownloadFile](),

		quitChannel: make(chan struct{}),
	}
}

func (n *Node) Start() {
	go n.startTCP()
	go n.startUDP()
	go n.startCLI()
	go n.startTicker()

	<-n.quitChannel
}

func (n *Node) startTCP() {
	conn, err := net.Dial("tcp4", n.trackerAddr)
	if err != nil {
		logger.Error("Tracker is not yet ready: %s", err)
		return
	}

	n.connected = true
	n.conn = transport.NewTCPConnection(conn, n.HandlePackets, n.Stop)
	go n.conn.Start()

	logger.Info("Connected to tracker on %s", n.trackerAddr)

	// Notify tracker of node's existence
	ipAddr := utils.TCPAddrToBytes(n.conn.LocalAddr())
	packet := protocol.NewInitPacket(ipAddr, n.udpPort)
	n.conn.EnqueuePacket(&packet)
}

func (n *Node) startUDP() {
	udpAddr := net.UDPAddr{
		IP:   net.IPv4zero,
		Port: int(n.udpPort),
	}

	conn, err := net.ListenUDP("udp4", &udpAddr)
	if err != nil {
		logger.Error("Failed to start UDP server: %s", err)
		return
	}

	n.srv = transport.NewUDPServer(*conn, n.HandleUDPPackets, func() {})
	go n.srv.Start()

	logger.Info("UDP server started on %s", udpAddr.String())
}

func (n *Node) startCLI() {
	console := cli.NewConsole()
	defer console.Close()
	logger.SetLogger(&console)

	c := cli.NewCLI(n.Stop, console)
	c.AddCommand("publish", "<file name>", "", 1, n.publish)
	c.AddCommand("request", "<file name>", "", 1, n.requestFile)
	c.AddCommand("status", "", "Show the status of the node", 0, n.status)
	c.AddCommand("remove", "<file name>", "", 1, n.removeFile)
	c.Start()
}

func (n *Node) startTicker() {
	tck := ticker.NewTicker(n.tick)
	tck.Start()
	n.tck = tck
}

func (n *Node) tick() {
	n.forDownload.ForEach(func(fileName string, file *ForDownloadFile) {
		missingChunks := file.GetMissingChunks()

		file.Nodes.ForEach(func(nodeAddr *net.UDPAddr, node *NodeInfo) {
			missingChunksPerNode := make([]uint16, 0)
			for _, chunk := range missingChunks {
				if node.ShouldRequestChunk(uint16(chunk)) {
					missingChunksPerNode = append(missingChunksPerNode, uint16(chunk))
				}
			}

			if len(missingChunksPerNode) > 0 {
				logger.Info("Requesting %d chunks to %s", len(missingChunksPerNode), nodeAddr)
				packet := protocol.NewRequestChunksPacket(fileName, missingChunksPerNode)
				n.srv.EnqueueRequest(&packet, nodeAddr)
			}
		})
	})
}

func (n *Node) Stop() {
	n.srv.Stop()
	n.tck.Stop()
	n.quitChannel <- struct{}{}
	close(n.quitChannel)
}

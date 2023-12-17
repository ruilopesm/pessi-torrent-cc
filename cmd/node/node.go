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
	"sort"
	"time"
)

const (
	UpdateServerChunksInterval = 5 * time.Second
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

	nodeStatistics *NodeStatistics

	// Last time the node sent a UpdateChunksPacket to the tracker
	lastServerChunksUpdate time.Time

	quitChannel chan struct{}
}

func NewNode(trackerAddr string, udpPort uint16) Node {
	return Node{
		trackerAddr: trackerAddr,
		udpPort:     udpPort,

		pending:     structures.NewSynchronizedMap[string, *File](),
		published:   structures.NewSynchronizedMap[string, *File](),
		forDownload: structures.NewSynchronizedMap[string, *ForDownloadFile](),

		nodeStatistics: NewNodeStatistics(),

		lastServerChunksUpdate: time.Now(),

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
		logger.Error("No tracker to connect found on %s. Try again later with the 'connect' command", n.trackerAddr)
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
	c.AddCommand("connect", "<tracker address>", "Connect to the tracker", 1, n.connect)
	c.AddCommand("publish", "<file name>", "", 1, n.publish)
	c.AddCommand("request", "<file name>", "", 1, n.requestFile)
	c.AddCommand("status", "", "Show the status of the node", 0, n.status)
	c.AddCommand("statistics", "", "Show the statistics of the node", 0, n.statistics)
	c.AddCommand("remove", "<file name>", "", 1, n.removeFile)
	c.Start()
}

func (n *Node) startTicker() {
	tck := ticker.NewTicker(n.tick)
	tck.Start()
	n.tck = tck
}

func (n *Node) updateServerChunks(file *ForDownloadFile) {
	bitfield := make([]bool, 0)
	file.Chunks.ForEach(func(chunkInfo ChunkInfo) {
		bitfield = append(bitfield, chunkInfo.Downloaded)
	})

	encondedBitfield := protocol.EncodeBitField(bitfield)

	packet := protocol.NewUpdateChunksPacket(file.FileName, encondedBitfield)
	n.conn.EnqueuePacket(&packet)
}

func (n *Node) tick() {
	n.forDownload.Lock()
	defer n.forDownload.Unlock()

	for fileName, file := range n.forDownload.M {
		if time.Now().Sub(n.lastServerChunksUpdate) > UpdateServerChunksInterval || file.IsFileDownloaded() {
			n.lastServerChunksUpdate = time.Now()
			n.updateServerChunks(file)
			logger.Info("Sent update chunks packet to tracker for file %s", fileName)
		}

		if file.IsFileDownloaded() {
			logger.Info("File %s was successfully downloaded", fileName)
			file.FileWriter.Stop()
			delete(n.forDownload.M, fileName)
			continue
		}

		missingChunks := file.GetMissingChunks()

		// Sort missing chunks by rarity
		sort.Slice(missingChunks, func(i, j int) bool {
			missingChunkI := uint16(missingChunks[i])
			missingChunkJ := uint16(missingChunks[j])

			return file.GetNumberOfNodesWhichHaveChunk(missingChunkI) < file.GetNumberOfNodesWhichHaveChunk(missingChunkJ)
		})

		file.Nodes.ForEach(func(nodeAddrString string, nodeInfo *NodeInfo) {
			missingChunksPerNode := make([]uint16, 0)
			for _, chunk := range missingChunks {
				if nodeInfo.ShouldRequestChunk(uint16(chunk)) {
					missingChunksPerNode = append(missingChunksPerNode, uint16(chunk))
				}
			}

			nodeAddr, _ := net.ResolveUDPAddr("udp4", nodeAddrString)

			if len(missingChunksPerNode) > 0 {
				logger.Info("Requesting %d chunks to %s", len(missingChunksPerNode), nodeAddr)
				packet := protocol.NewRequestChunksPacket(fileName, missingChunksPerNode)
				n.srv.EnqueueRequest(&packet, nodeAddr)

				// Mark chunks as requested
				for _, chunkIndex := range missingChunksPerNode {
					file.MarkChunkAsRequested(chunkIndex, nodeInfo)
				}
			}
		})
	}
}

func (n *Node) Stop() {
	n.srv.Stop()
	n.tck.Stop()
	n.quitChannel <- struct{}{}
	close(n.quitChannel)
}

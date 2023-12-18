package main

import (
	"PessiTorrent/internal/cli"
	"PessiTorrent/internal/dns"
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
	UpdateServerChunksInterval  = 5 * time.Second
	MaxChunksPerRequest         = 100
	ChunkRequestTimeoutDuration = 500 * time.Millisecond
	MaxTriesPerChunk            = 3
	MaxNodeTimeouts             = 3
	TickInterval                = 100 * time.Millisecond
)

type Node struct {
	dns *dns.DNS

	trackerAddr string
	udpPort     uint16
	connected   bool // Whether the node is connected to the tracker or not

	conn transport.TCPConnection
	srv  transport.UDPServer
	tck  ticker.Ticker

	published      structures.SynchronizedMap[string, *File]
	pending        structures.SynchronizedMap[string, *File]
	forDownload    structures.SynchronizedMap[string, *ForDownloadFile]
	downloadedFile structures.SynchronizedMap[string, *File]
	downloadPath   string

	nodeStatistics *NodeStatistics

	quitChannel chan struct{}
}

func NewNode(trackerAddr string, udpPort uint16, dnsAddr string) Node {
	return Node{
		dns: dns.NewDNS(dnsAddr),

		trackerAddr: trackerAddr,
		udpPort:     udpPort,

		pending:      structures.NewSynchronizedMap[string, *File](),
		published:    structures.NewSynchronizedMap[string, *File](),
		forDownload:  structures.NewSynchronizedMap[string, *ForDownloadFile](),
		downloadPath: "./downloads",

		nodeStatistics: NewNodeStatistics(),

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
	domain, err := n.dns.ResolveDomain(net.IP(ipAddr[:]).String())
	if err != nil {
		logger.Error("Error resolving domain: %v", err)
		return
	}

	packet := protocol.NewInitPacket(domain, n.udpPort)
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
	c.AddCommand("publish", "<file name | directory>", "", 1, n.publish)
	c.AddCommand("request", "<file name>", "", 1, n.requestFile)
	c.AddCommand("status", "", "Show the status of the node", 0, n.status)
	c.AddCommand("statistics", "", "Show the statistics of the node", 0, n.statistics)
	c.AddCommand("path", "<download folder path>", "Set download path", 1, n.setDownloadPath)
	c.AddCommand("remove", "<file name>", "", 1, n.removeFile)
	c.Start()
}

func (n *Node) startTicker() {
	tck := ticker.NewTicker(TickInterval, n.tick)
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
		if !file.UpdatedByTracker {
			continue
		}

		if time.Now().Sub(file.LastServerChunksUpdate) > UpdateServerChunksInterval || file.IsFileDownloaded() {
			file.LastServerChunksUpdate = time.Now()
			n.updateServerChunks(file)
			logger.Info("Sent update chunks packet to tracker for file %s", fileName)

			// Also request to update our nodes info about the file
			packet := protocol.NewUpdateFilePacket(fileName)
			n.conn.EnqueuePacket(&packet)
		}

		if file.IsFileDownloaded() {
			timeToDownload := time.Since(file.DownloadStarted)
			logger.Info("File %s was successfully downloaded in %s", fileName, timeToDownload.String())
			file.FileWriter.Stop()

			newFile := NewFile(file.FileName, file.FilePath)
			n.published.Put(file.FileName, &newFile)

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

		nodes := file.Nodes.Values()
		sort.Slice(nodes, func(i, j int) bool {
			return n.nodeStatistics.getAverageDownloadSpeed(nodes[i].Address) < n.nodeStatistics.getAverageDownloadSpeed(nodes[j].Address)
		})

		chunksToRequest := make(map[*NodeInfo][]uint16)

		for _, nodeInfo := range nodes {
			chunksToRequest[nodeInfo] = make([]uint16, 0)

			for len(missingChunks) > 0 && len(chunksToRequest[nodeInfo]) < MaxChunksPerRequest {
				chunk := missingChunks[0]
				missingChunks = missingChunks[1:] // Pop first element

				requestInfo, ok := nodeInfo.Chunks.Get(uint16(chunk))

				if time.Now().Sub(requestInfo.TimeLastRequested) >= ChunkRequestTimeoutDuration || !ok {
					requestInfo.NumberOfTries++
					if requestInfo.NumberOfTries >= MaxTriesPerChunk {
						logger.Warn("Node %s is not responding.", nodeInfo.Address)
						nodeInfo.Timeouts++
						if nodeInfo.Timeouts >= MaxNodeTimeouts {
							logger.Warn("Node %s has timed out 3 times. Removing it from file %s", nodeInfo.Address, file.FileName)
							file.Nodes.Delete(nodeInfo.Address)
						}
					} else {
						chunksToRequest[nodeInfo] = append(chunksToRequest[nodeInfo], uint16(chunk)) // Queue chunk
					}
				}
			}
		}

		for nodeInfo, chunksToRequest := range chunksToRequest {
			nodeAddr, _ := net.ResolveUDPAddr("udp4", nodeInfo.Address)
			n.RequestChunks(chunksToRequest, nodeAddr, file, nodeInfo)
		}
	}
}

func (n *Node) RequestChunks(chunkIndexes []uint16, nodeAddr *net.UDPAddr, file *ForDownloadFile, nodeInfo *NodeInfo) {
	if len(chunkIndexes) <= 0 {
		return
	}

	packet := protocol.NewRequestChunksPacket(file.FileName, chunkIndexes)
	n.srv.EnqueueRequest(&packet, nodeAddr)

	// Mark chunks as requested
	for _, chunkIndex := range chunkIndexes {
		file.MarkChunkAsRequested(chunkIndex, nodeInfo)
	}
}

func (n *Node) Stop() {
	n.srv.Stop()
	n.tck.Stop()
	n.quitChannel <- struct{}{}
	close(n.quitChannel)
}

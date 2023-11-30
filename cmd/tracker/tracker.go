package main

import (
	"PessiTorrent/internal/logger"
	"PessiTorrent/internal/structures"
	"PessiTorrent/internal/transport"
	"net"
)

type Tracker struct {
	tcpPort  uint16
	listener net.Listener

	files structures.SynchronizedMap[string, *TrackedFile]
	nodes structures.SynchronizedList[*NodeInfo]

	quitChannel chan struct{}
}

func NewTracker(port uint16) Tracker {
	return Tracker{
		tcpPort: port,
		files:   structures.NewSynchronizedMap[string, *TrackedFile](),
		nodes:   structures.NewSynchronizedList[*NodeInfo](),

		quitChannel: make(chan struct{}),
	}
}

func (t *Tracker) Start() {
	go t.startTCP()

	<-t.quitChannel
}

func (t *Tracker) Stop() {
	t.quitChannel <- struct{}{}
	close(t.quitChannel)
}

func (t *Tracker) startTCP() {
	tcpAddr := net.TCPAddr{
		IP:   net.IPv4zero,
		Port: int(t.tcpPort),
	}

	listener, err := net.Listen("tcp4", tcpAddr.String())
	if err != nil {
		logger.Error("Failed to start TCP server: %s", err)
		t.Stop()
		return
	}

	logger.Info("TCP server started on %s", tcpAddr.String())
	t.listener = listener

	t.acceptConnections()
}

func (t *Tracker) acceptConnections() {
	for {
		cn, err := t.listener.Accept()
		if err != nil {
			logger.Error("Failed to accept connection: %s", err)
			continue
		}

		conn := transport.NewTCPConnection(cn, t.HandlePackets, func() {})
		logger.Info("Node %s connected", conn.RemoteAddr())

		go conn.Start()
	}
}

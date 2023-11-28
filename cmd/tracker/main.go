package main

import (
	"PessiTorrent/internal/config"
	"PessiTorrent/internal/logger"
	"PessiTorrent/internal/protocol"
	"PessiTorrent/internal/structures"
	"PessiTorrent/internal/transport"
	"flag"
	"fmt"
	"log"
	"net"
	"strconv"
)

type Tracker struct {
	addr     string
	listener net.Listener

	files structures.SynchronizedMap[*TrackedFile]
	nodes structures.SynchronizedList[*NodeInfo]

	quitch chan struct{}
}

func NewTracker(listenPort string) Tracker {
	return Tracker{
		addr: net.IPv4zero.String() + ":" + listenPort,

		files: structures.NewSynchronizedMap[*TrackedFile](),
		nodes: structures.NewSynchronizedList[*NodeInfo](),

		quitch: make(chan struct{}),
	}
}

func (t *Tracker) Start() error {
	ln, err := net.Listen("tcp4", t.addr)
	if err != nil {
		return err
	}
	defer ln.Close()
	logger.Info("Tracker listening TCP on %s", ln.Addr())

	t.listener = ln

	go t.acceptLoop()

	<-t.quitch

	return nil
}

func (t *Tracker) acceptLoop() {
	for {
		c, err := t.listener.Accept()
		if err != nil {
			fmt.Println("Accept error:", err)
			continue
		}
		conn := transport.NewTCPConnection(c, t.handlePacket, func() {})
		logger.Info("Node %s connected", conn.RemoteAddr())

		go conn.Start()
	}
}

func (t *Tracker) handlePacket(packet interface{}, conn *transport.TCPConnection) {
	switch data := packet.(type) {
	case *protocol.InitPacket:
		t.handleInitPacket(packet.(*protocol.InitPacket), conn)
	case *protocol.PublishFilePacket:
		t.handlePublishFilePacket(packet.(*protocol.PublishFilePacket), conn)
	case *protocol.RequestFilePacket:
		t.handleRequestFilePacket(packet.(*protocol.RequestFilePacket), conn)
	case *protocol.RemoveFilePacket:
		t.handleRemoveFilePacket(packet.(*protocol.RemoveFilePacket), conn)
	default:
		fmt.Println("Unknown packet type received: ", data)
	}
}

func main() {
	// TODO: check if port is inside tcp range

	conf, err := config.NewConfig(config.ConfigPath)
	if err != nil {
		fmt.Println("Error reading config:", err)
		return
	}
	udpPort := strconv.Itoa(conf.Tracker.Port)

	flag.StringVar(&udpPort, "p", udpPort, "The port to listen on")
	flag.Parse()

	tracker := NewTracker(udpPort)
	err = tracker.Start()
	if err != nil {
		log.Fatal(err)
	}
}

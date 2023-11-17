package main

import (
	"PessiTorrent/internal/connection"
	"PessiTorrent/internal/protocol"
	"PessiTorrent/internal/structures"
	"fmt"
	"log"
	"net"
	"os"
)

type Tracker struct {
	listenAddr string
	ln         net.Listener
	files      structures.SynchronizedMap[*File]
	nodes      structures.SynchronizedList[*NodeInfo]
	quitch     chan struct{}
}

type File struct {
	name     string
	fileHash [20]byte
	hashes   [][20]byte
}

type NodeInfo struct {
	conn    connection.Connection
	udpPort uint16
	files   structures.SynchronizedMap[NodeFile]
}

type NodeFile struct {
	file            *File
	chunksAvailable []uint16
}

func NewTracker(listenPort string) Tracker {
	return Tracker{
		listenAddr: "0.0.0.0:" + listenPort,
		files:      structures.NewSynchronizedMap[*File](),
		nodes:      structures.NewSynchronizedList[*NodeInfo](16),
		quitch:     make(chan struct{}),
	}
}

func (t *Tracker) Start() error {
	ln, err := net.Listen("tcp4", t.listenAddr)
	if err != nil {
		return err
	}
	defer ln.Close()
	fmt.Println("tracker listening tcp on", ln.Addr())

	t.ln = ln

	go t.acceptLoop()

	<-t.quitch

	return nil
}

func (t *Tracker) acceptLoop() {
	for {
		c, err := t.ln.Accept()
		if err != nil {
			fmt.Println("accept error:", err)
			continue
		}
		conn := connection.NewConnection(c, t.handlePacket)
		fmt.Printf("node %s connected\n", conn.RemoteAddr())

		go conn.Start()
	}
}

func (t *Tracker) handlePacket(packet interface{}, conn *connection.Connection) {
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
		fmt.Println("unknown packet type received: ", data)
	}
}

func main() {
	if len(os.Args) != 2 {
		fmt.Println("usage: tracker <listen port>")
		return
	}

	// TODO: check if port is inside tcp range

	tracker := NewTracker(os.Args[1])
	err := tracker.Start()
	if err != nil {
		log.Fatal(err)
	}
}

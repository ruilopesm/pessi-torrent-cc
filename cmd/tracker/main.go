package main

import (
	"PessiTorrent/internal/connection"
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
		conn := connection.NewConnection(c)
		fmt.Printf("node %s connected\n", conn.RemoteAddr())

		go t.handleConnection(&conn)
	}
}

func (t *Tracker) handleConnection(conn *connection.Connection) {
	defer conn.Close()

	for {
		packet, packetType, err := conn.ReadPacket()
		if err != nil {
			if err.Error() != "EOF" {
				fmt.Println("read error:", err)
			} else {
				fmt.Printf("node %s disconnected\n", conn.RemoteAddr())
				t.removeNode(conn)
			}

			return
		}

		t.HandlePacketsDispatcher(packet, packetType, conn)
	}
}

func (t *Tracker) removeNode(conn *connection.Connection) {
	t.nodes.ForEach(func(node *NodeInfo) {
		if node.conn.RemoteAddr() == conn.RemoteAddr() {
			fmt.Println("removed node", node.conn.RemoteAddr())
			t.nodes.Remove(node)

			return
		}
	})
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

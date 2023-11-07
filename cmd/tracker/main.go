package main

import (
	"PessiTorrent/internal/connection"
	"fmt"
	"log"
	"net"
	"os"
	"sync"
)

type Tracker struct {
	listenAddr string
	ln         net.Listener
	files      SynchronizedMap[*File]
	nodes      SynchronizedList[*NodeInfo]
	quitch     chan struct{}
}

type SynchronizedMap[T any] struct {
	m map[string]T
	sync.RWMutex
}

type SynchronizedList[T any] struct {
	l []T
	sync.RWMutex
}

type File struct {
	name     string
	fileHash [20]byte
	hashes   [][20]byte
}

type NodeInfo struct {
	conn    connection.Connection
	udpPort uint16
	files   SynchronizedMap[NodeFile]
}

type NodeFile struct {
	file            *File
	chunksAvailable []uint8
}

func NewTracker(listenAddr string) Tracker {
	s, err := net.ResolveTCPAddr("tcp4", ":"+listenAddr)
	if err != nil {
		fmt.Println("error resolving tcp addr: ", err)
		os.Exit(1)
	}

	return Tracker{
		listenAddr: s.String(),
		files:      SynchronizedMap[*File]{m: make(map[string]*File)},
		nodes:      SynchronizedList[*NodeInfo]{l: make([]*NodeInfo, 0)},
		quitch:     make(chan struct{}),
	}
}

func (t *Tracker) Start() error {
	ln, err := net.Listen("tcp4", t.listenAddr)
	if err != nil {
		return err
	}
	defer ln.Close()
	fmt.Println("tracker listening tcp on ", ln.Addr())

	t.ln = ln

	go t.acceptLoop()

	<-t.quitch

	return nil
}

func (t *Tracker) acceptLoop() {
	for {
		c, err := t.ln.Accept()
		if err != nil {
			fmt.Println("accept error: ", err)
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
				fmt.Println("read error: ", err)
			} else {
				fmt.Printf("node %s disconnected\n", conn.RemoteAddr())
			}

			return
		}

		t.HandlePacketsDispatcher(packet, packetType, conn)
	}
}

func main() {
	tracker := NewTracker("42069")
	err := tracker.Start()
	if err != nil {
		log.Fatal(err)
	}
}

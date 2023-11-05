package main

import (
	"PessiTorrent/internal/connection"
	"fmt"
	"log"
	"net"
	"sync"
)

type Tracker struct {
	listenAddr string
	ln         net.Listener
	nodesMap   SynchronizedMap
	quitch     chan struct{}
}

type SynchronizedMap struct {
	m map[net.Addr]*connection.Connection
	sync.RWMutex
}

func NewTracker(listenAddr string) *Tracker {
	return &Tracker{
		listenAddr: listenAddr,
		nodesMap:   SynchronizedMap{m: make(map[net.Addr]*connection.Connection)},
		quitch:     make(chan struct{}),
	}
}

func (t *Tracker) Start() error {
	ln, err := net.Listen("tcp", t.listenAddr)
	if err != nil {
		return err
	}
	defer ln.Close()

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
		go t.handleConnection(conn)
	}
}

func (t *Tracker) handleConnection(conn *connection.Connection) {
	t.nodesMap.Lock()
	t.nodesMap.m[conn.RemoteAddr()] = conn
	t.nodesMap.Unlock()

	t.readLoop(conn)
}

func (t *Tracker) readLoop(conn *connection.Connection) {
	defer conn.Close()

	for {
		packet, err := conn.ReadPacket()
		if err != nil {
			if err.Error() != "EOF" {
				fmt.Println("read error: ", err)
			} else {
				fmt.Printf("node %s disconnected\n", conn.RemoteAddr())
			}

			return
		}

		fmt.Println("packet received: ", packet)
	}
}

func main() {
  tracker := NewTracker(":42069")
	err := tracker.Start()
	if err != nil {
		log.Fatal(err)
	}
}

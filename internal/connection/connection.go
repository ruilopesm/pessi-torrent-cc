package connection

import (
	"PessiTorrent/internal/protocol"
	"fmt"
	"net"
)

type Connection struct {
	net.Conn
	writeQueue chan interface{}
}

func NewConnection(conn net.Conn) Connection {
	return Connection{
		conn,
		make(chan interface{}),
	}
}

func (conn *Connection) Start() {
	go conn.writeLoop()
	go conn.readLoop()
}

func (conn *Connection) writeLoop() {
	for {
		packet := <-conn.writeQueue
		err := protocol.Serialize(conn, packet)
		if err != nil {
			_ = fmt.Errorf("Error serializing packet: %v", err)
		}
	}
}

func (conn *Connection) EnqueuePacket(packet interface{}) {
	conn.writeQueue <- packet
}

func (conn *Connection) readLoop() {
	for {
		packet, err := protocol.Deserialize(conn)
		if err != nil {
			if err.Error() == "EOF" {
				fmt.Println("Connection closed")
				return
			}
			_ = fmt.Errorf("Error reading packet: %v", err)
		}

		conn.handlePacket(packet)
	}
}

func (conn *Connection) handlePacket(packet interface{}) {
	switch data := packet.(type) {
	case *protocol.InitPacket:
		fmt.Println("init packet received")
	case *protocol.PublishFilePacket:
		fmt.Println("publish file packet received")
	default:
		fmt.Println("unknown packet type received: ", data)
	}
}

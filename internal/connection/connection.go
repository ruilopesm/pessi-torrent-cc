package connection

import (
	"PessiTorrent/internal/protocol"
	"fmt"
	"net"
)

type PacketHandler func(packet interface{}, conn *Connection)

type Connection struct {
	net.Conn
	writeQueue   chan interface{}
	handlePacket PacketHandler
}

func NewConnection(conn net.Conn, handlePacket PacketHandler) Connection {
	return Connection{
		conn,
		make(chan interface{}),
		handlePacket,
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
			fmt.Println("error serializing packet: ", err)
			return
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
			fmt.Println("error reading packet: ", err)
			return
		}

		conn.handlePacket(packet, conn)
	}
}

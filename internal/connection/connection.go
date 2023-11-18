package connection

import (
	"PessiTorrent/internal/protocol"
	"errors"
	"fmt"
	"io"
	"net"
)

type PacketHandler func(packet interface{}, conn *Connection)

type Connection struct {
	net.Conn
	writeQueue   chan protocol.Packet
	handlePacket PacketHandler
}

func NewConnection(conn net.Conn, handlePacket PacketHandler) Connection {
	return Connection{
		conn,
		make(chan protocol.Packet),
		handlePacket,
	}
}

func (conn *Connection) Start() {
	go conn.writeLoop()
	go conn.readLoop()
}

func (conn *Connection) Stop() {
	err := conn.Close()
	if err != nil {
		fmt.Println("error closing connection: ", err)
	}
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

func (conn *Connection) EnqueuePacket(packet protocol.Packet) {
	conn.writeQueue <- packet
}

func (conn *Connection) readLoop() {
	for {
		packet, err := protocol.Deserialize(conn)
		if err != nil {
			if errors.Is(err, io.EOF) || errors.Is(err, net.ErrClosed) {
				fmt.Println("Connection closed")
				return
			}
			fmt.Println("error reading packet: ", err)
			return
		}

		conn.handlePacket(packet, conn)
	}
}

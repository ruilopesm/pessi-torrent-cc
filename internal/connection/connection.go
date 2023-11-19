package connection

import (
	"PessiTorrent/internal/protocol"
	"bufio"
	"errors"
	"fmt"
	"io"
	"net"
)

type PacketHandler func(packet interface{}, conn *Connection)

type Connection struct {
	connection   net.Conn
	readWrite    bufio.ReadWriter
	writeQueue   chan protocol.Packet
	handlePacket PacketHandler
}

func NewConnection(conn net.Conn, handlePacket PacketHandler) Connection {
	return Connection{
		conn,
		*bufio.NewReadWriter(bufio.NewReader(conn), bufio.NewWriter(conn)),
		make(chan protocol.Packet),
		handlePacket,
	}
}

func (conn *Connection) Start() {
	go conn.writeLoop()
	go conn.readLoop()
}

func (conn *Connection) Stop() {
	err := conn.connection.Close()
	if err != nil {
		fmt.Println("error closing connection: ", err)
	}
}

func (conn *Connection) writeLoop() {
	for {
		packet := <-conn.writeQueue
		err := protocol.SerializePacket(conn.readWrite, packet)
		if err != nil {
			fmt.Println("error serializing packet: ", err)
			return
		}

		err = conn.readWrite.Flush()
		if err != nil {
			fmt.Println("error flushing buffer: ", err)
			return
		}
	}
}

func (conn *Connection) EnqueuePacket(packet protocol.Packet) {
	conn.writeQueue <- packet
}

func (conn *Connection) readLoop() {
	for {
		packet, err := protocol.DeserializePacket(conn.readWrite)
		if err != nil {
			if errors.Is(err, io.EOF) || errors.Is(err, net.ErrClosed) {
				fmt.Println("connection closed")
				return
			}
			fmt.Println("error reading packet: ", err)
			return
		}

		conn.handlePacket(packet, conn)
	}
}

func (conn *Connection) RemoteAddr() net.Addr {
	return conn.connection.RemoteAddr()
}

package transport

import (
	"PessiTorrent/internal/protocol"
	"bufio"
	"errors"
	"fmt"
	"io"
	"net"
)

type TCPPacketHandler func(packet interface{}, conn *TCPConnection)

type TCPConnection struct {
	connection   net.Conn
	readWrite    bufio.ReadWriter
	writeQueue   chan protocol.Packet
	handlePacket TCPPacketHandler
}

func NewTCPConnection(conn net.Conn, handlePacket TCPPacketHandler) TCPConnection {
	return TCPConnection{
		conn,
		*bufio.NewReadWriter(bufio.NewReader(conn), bufio.NewWriter(conn)),
		make(chan protocol.Packet),
		handlePacket,
	}
}

func (conn *TCPConnection) Start() {
	go conn.writeLoop()
	go conn.readLoop()
}

func (conn *TCPConnection) Stop() {
	err := conn.connection.Close()
	if err != nil {
		fmt.Println("Error closing TCP connection:", err)
	}
}

func (conn *TCPConnection) writeLoop() {
	for {
		packet := <-conn.writeQueue
		err := protocol.SerializePacket(conn.readWrite, packet)
		if err != nil {
			fmt.Println("Error serializing packet:", err)
			return
		}

		err = conn.readWrite.Flush()
		if err != nil {
			fmt.Println("Error flushing buffer:", err)
			return
		}
	}
}

func (conn *TCPConnection) EnqueuePacket(packet protocol.Packet) {
	conn.writeQueue <- packet
}

func (conn *TCPConnection) readLoop() {
	for {
		packet, err := protocol.DeserializePacket(conn.readWrite)
		if err != nil {
			if errors.Is(err, io.EOF) || errors.Is(err, net.ErrClosed) {
				fmt.Println("Connection from", conn.RemoteAddr(), "closed")
				return
			}
			fmt.Println("Error reading packet:", err)
			return
		}

		conn.handlePacket(packet, conn)
	}
}

func (conn *TCPConnection) LocalAddr() net.Addr {
	return conn.connection.LocalAddr()
}

func (conn *TCPConnection) RemoteAddr() net.Addr {
	return conn.connection.RemoteAddr()
}

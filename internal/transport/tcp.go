package transport

import (
	"PessiTorrent/internal/logger"
	"PessiTorrent/internal/protocol"
	"bufio"
	"errors"
	"fmt"
	"io"
	"net"
	"syscall"
)

type TCPPacketHandler func(packet interface{}, conn *TCPConnection)

type TCPConnection struct {
	connection   net.Conn
	readWrite    bufio.ReadWriter
	writeQueue   chan protocol.Packet
	handlePacket TCPPacketHandler
	onQuit       func()
}

func NewTCPConnection(conn net.Conn, handlePacket TCPPacketHandler, onQuit func()) TCPConnection {
	return TCPConnection{
		conn,
		*bufio.NewReadWriter(bufio.NewReader(conn), bufio.NewWriter(conn)),
		make(chan protocol.Packet),
		handlePacket,
		onQuit,
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

	close(conn.writeQueue)
	conn.onQuit()
}

func (conn *TCPConnection) writeLoop() {
	for {
		packet := <-conn.writeQueue
		if packet == nil {
			return
		}

		err := protocol.SerializePacket(conn.readWrite, packet)
		if err != nil {
			fmt.Println("Error serializing packet:", err)
			continue
		}

		err = conn.readWrite.Flush()
		if err != nil {
			fmt.Println("Error flushing buffer:", err)
			continue
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
			if errors.Is(err, io.EOF) || errors.Is(err, net.ErrClosed) || errors.Is(err, syscall.WSAECONNRESET) {
				logger.Info("Connection from %s closed", conn.RemoteAddr())
				conn.Stop()
				return
			}

			fmt.Println("Error deserializing packet:", err)
			continue
		}

		go conn.handlePacket(packet, conn)
	}
}

func (conn *TCPConnection) LocalAddr() net.Addr {
	return conn.connection.LocalAddr()
}

func (conn *TCPConnection) RemoteAddr() net.Addr {
	return conn.connection.RemoteAddr()
}

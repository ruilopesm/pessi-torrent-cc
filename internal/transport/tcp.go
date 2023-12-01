package transport

import (
	"PessiTorrent/internal/logger"
	"PessiTorrent/internal/protocol"
	"bufio"
	"errors"
	"io"
	"net"
	"strings"
)

type TCPPacketHandler func(packet protocol.Packet, conn *TCPConnection)

type TCPConnection struct {
	connection   net.Conn
	readWrite    bufio.ReadWriter
	writeQueue   chan protocol.Packet
	handlePacket TCPPacketHandler
	onClose      func()
}

func NewTCPConnection(conn net.Conn, handlePacket TCPPacketHandler, onClose func()) TCPConnection {
	return TCPConnection{
		conn,
		*bufio.NewReadWriter(bufio.NewReader(conn), bufio.NewWriter(conn)),
		make(chan protocol.Packet),
		handlePacket,
		onClose,
	}
}

func (conn *TCPConnection) Start() {
	go conn.writeLoop()
	go conn.readLoop()
}

func (conn *TCPConnection) Stop() {
	err := conn.connection.Close()
	if err != nil {
		logger.Error("Error closing TCP connection:", err)
	}

	close(conn.writeQueue)
	conn.onClose()
}

func (conn *TCPConnection) writeLoop() {
	for {
		packet := <-conn.writeQueue
		if packet == nil {
			return
		}

		err := protocol.SerializePacket(conn.readWrite, packet)
		if err != nil {
			logger.Error("Error serializing packet:", err)
			continue
		}

		err = conn.readWrite.Flush()
		if err != nil {
			logger.Error("Error flushing buffer:", err)
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
			if errors.Is(err, io.EOF) || errors.Is(err, net.ErrClosed) || strings.Contains(err.Error(), "read tcp4") {
				logger.Info("Connection from %s closed", conn.RemoteAddr())
				conn.Stop()
				return
			}

			logger.Error("Error deserializing packet:", err)
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

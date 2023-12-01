package transport

import (
	"PessiTorrent/internal/logger"
	"PessiTorrent/internal/protocol"
	"bytes"
	"errors"
	"net"
)

const (
	UDPMaxPacketSize = 65515 // 65535 - 20 (UDP header)
)

type UDPPacketHandler func(packet protocol.Packet, addr *net.UDPAddr)

type UDPServer struct {
	connection    net.UDPConn
	readBuffer    []byte
	requestsQueue chan RequestChunk
	handlePacket  UDPPacketHandler
	onClose       func()
}

type RequestChunk struct {
	packet protocol.Packet
	addr   *net.UDPAddr
}

func NewUDPServer(conn net.UDPConn, handlePacket UDPPacketHandler, onClose func()) UDPServer {
	return UDPServer{
		conn,
		make([]byte, UDPMaxPacketSize),
		make(chan RequestChunk),
		handlePacket,
		onClose,
	}
}

func (srv *UDPServer) Start() {
	go srv.writeLoop()
	go srv.readLoop()
}

func (srv *UDPServer) Stop() {
	srv.connection.Close()
}

func (srv *UDPServer) writeLoop() {
	for {
		request := <-srv.requestsQueue
		if request.packet == nil {
			return
		}

		buffer := new(bytes.Buffer)
		err := protocol.SerializePacket(buffer, request.packet)
		if err != nil {
			logger.Error("Error serializing packet:", err)
			continue
		}

		_, err = srv.connection.WriteToUDP(buffer.Bytes(), request.addr)
		if err != nil {
			logger.Error("Error sending packet:", err)
		}
	}
}

func (srv *UDPServer) readLoop() {
	for {
		n, addr, err := srv.connection.ReadFromUDP(srv.readBuffer)
		if err != nil {
			if errors.Is(err, net.ErrClosed) {
				srv.onClose()
				break
			}
			logger.Error("Error reading from UDP connection:", err)
			continue
		}

		packet, err := protocol.DeserializePacket(bytes.NewReader(srv.readBuffer[:n]))
		if err != nil {
			logger.Error("Error deserializing packet:", err)
			continue
		}

		go srv.handlePacket(packet, addr)
	}
}

func (srv *UDPServer) SendPacket(packet protocol.Packet, addr *net.UDPAddr) {
	buffer := new(bytes.Buffer)
	err := protocol.SerializePacket(buffer, packet)
	if err != nil {
		logger.Error("Error serializing packet:", err)
		return
	}

	_, err = srv.connection.WriteToUDP(buffer.Bytes(), addr)
	if err != nil {
		logger.Error("Error sending packet:", err)
	}
}

func (srv *UDPServer) EnqueueRequest(packet protocol.Packet, addr *net.UDPAddr) {
	srv.requestsQueue <- RequestChunk{packet, addr}
}

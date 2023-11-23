package transport

import (
	"PessiTorrent/internal/protocol"
	"bytes"
	"fmt"
	"net"
)

const (
	BufferSize = 1_000_000 // 1 MB
)

type UDPPacketHandler func(packet interface{}, addr *net.UDPAddr)

type UDPServer struct {
	connection   net.UDPConn
	readBuffer   []byte
	writeBuffer  bytes.Buffer
	handlePacket UDPPacketHandler
}

func NewUDPServer(conn net.UDPConn, handlePacket UDPPacketHandler) UDPServer {
	return UDPServer{
		conn,
		make([]byte, BufferSize),
		bytes.Buffer{},
		handlePacket,
	}
}

func (srv *UDPServer) Start() {
	go srv.readLoop()
}

func (srv *UDPServer) Stop() {
	err := srv.connection.Close()
	if err != nil {
		fmt.Println("Error closing UDP connection:", err)
	}
}

func (srv *UDPServer) readLoop() {
	for {
		n, addr, err := srv.connection.ReadFromUDP(srv.readBuffer)
		if err != nil {
			fmt.Println("Error reading from UDP connection:", err)
			continue
		}

		packet, err := protocol.DeserializePacket(bytes.NewReader(srv.readBuffer[:n]))
		if err != nil {
			fmt.Println("Error deserializing packet:", err)
			continue
		}

		srv.handlePacket(packet, addr)
	}
}

func (srv *UDPServer) SendPacket(packet protocol.Packet, addr *net.UDPAddr) {
	srv.writeBuffer.Reset()
	err := protocol.SerializePacket(&srv.writeBuffer, packet)
	if err != nil {
		fmt.Println("Error serializing packet:", err)
		return
	}

	_, err = srv.connection.WriteToUDP(srv.writeBuffer.Bytes(), addr)
	if err != nil {
		fmt.Println("Error writing to UDP connection:", err)
		return
	}
}

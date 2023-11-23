package transport

import (
	"PessiTorrent/internal/protocol"
	"bytes"
	"fmt"
	"net"
)

const (
	UDPMaxPacketSize = 65515 // 65535 - 20 (UDP header)
)

type UDPPacketHandler func(packet interface{}, addr *net.UDPAddr)

type UDPServer struct {
	connection   net.UDPConn
	readBuffer   []byte
	handlePacket UDPPacketHandler
}

func NewUDPServer(conn net.UDPConn, handlePacket UDPPacketHandler) UDPServer {
	return UDPServer{
		conn,
		make([]byte, UDPMaxPacketSize),
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

		go srv.handlePacket(packet, addr)
	}
}

func (srv *UDPServer) SendPacket(packet protocol.Packet, addr *net.UDPAddr) {
	buffer := new(bytes.Buffer)
	err := protocol.SerializePacket(buffer, packet)
	if err != nil {
		fmt.Println("Error serializing packet:", err)
		return
	}

	_, err = srv.connection.WriteToUDP(buffer.Bytes(), addr)
	if err != nil {
		fmt.Println("Error sending packet:", err)
	}
}

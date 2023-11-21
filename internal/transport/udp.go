package transport

import (
	"PessiTorrent/internal/protocol"
	"bytes"
	"errors"
	"fmt"
	"io"
	"net"
)

type UDPPacketHandler func(packet interface{}, addr *net.UDPAddr)

type UDPServer struct {
	connection   net.UDPConn
	writeBuffer  bytes.Buffer
	readBuffer   bytes.Buffer
	handlePacket UDPPacketHandler
}

func NewUDPServer(conn net.UDPConn, handlePacket UDPPacketHandler) UDPServer {
	return UDPServer{
		conn,
		*bytes.NewBuffer(make([]byte, 0)),
		*bytes.NewBuffer(make([]byte, 0)),
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
		n, addr, err := srv.connection.ReadFromUDP(srv.readBuffer.Bytes())
		if err != nil {
			fmt.Println("Error reading from UDP connection:", err)
			return
		}

		packet, err := protocol.DeserializePacket(&srv.readBuffer)
		if err != nil {
			if !errors.Is(err, io.EOF) && !errors.Is(err, net.ErrClosed) {
				fmt.Println("Error deserializing packet:", err)
				return
			}
		}

		srv.handlePacket(packet, addr)

		srv.readBuffer.Truncate(n)
	}
}

func (srv *UDPServer) SendPacket(packet protocol.Packet, addr *net.UDPAddr) {
	err := protocol.SerializePacket(&srv.writeBuffer, packet)
	if err != nil {
		fmt.Println("Error serializing packet:", err)
		return
	}

	_, err = srv.connection.WriteToUDP(srv.writeBuffer.Bytes(), addr)
	if err != nil {
		fmt.Println("Error sending packet:", err)
		return
	}

	srv.writeBuffer.Reset()
}

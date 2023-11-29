package transport

import (
	"PessiTorrent/internal/protocol"
	"bytes"
	"errors"
	"fmt"
	"net"
)

const (
	UDPMaxPacketSize = 65515 // 65535 - 20 (UDP header)
)

type UDPPacketHandler func(packet protocol.Packet, addr *net.UDPAddr)

type UDPServer struct {
	connection   net.UDPConn
	readBuffer   []byte
	handlePacket UDPPacketHandler
	onClose      func()
}

func NewUDPServer(conn net.UDPConn, handlePacket UDPPacketHandler, onClose func()) UDPServer {
	return UDPServer{
		conn,
		make([]byte, UDPMaxPacketSize),
		handlePacket,
		onClose,
	}
}

func (srv *UDPServer) Start() {
	go srv.readLoop()
}

func (srv *UDPServer) Stop() {
	srv.connection.Close()
}

func (srv *UDPServer) readLoop() {
	for {
		n, addr, err := srv.connection.ReadFromUDP(srv.readBuffer)
		if err != nil {
			if errors.Is(err, net.ErrClosed) {
				srv.onClose()
				break
			}
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

type UDPSocket struct {
	connection net.UDPConn
	toSend     net.UDPAddr
}

func NewUDPSocket(conn net.UDPConn, toSend net.UDPAddr) UDPSocket {
	return UDPSocket{
		conn,
		toSend,
	}
}

func (sock *UDPSocket) SendPacket(packet protocol.Packet) {
	buffer := new(bytes.Buffer)
	err := protocol.SerializePacket(buffer, packet)
	if err != nil {
		fmt.Println("Error serializing packet:", err)
		return
	}

	_, err = sock.connection.WriteToUDP(buffer.Bytes(), &sock.toSend)
	if err != nil {
		fmt.Println("Error sending packet:", err)
	}
}

func (sock *UDPSocket) ReadPacket() (protocol.Packet, error) {
	buffer := make([]byte, UDPMaxPacketSize)
	n, _, err := sock.connection.ReadFromUDP(buffer)
	if err != nil {
		return nil, err
	}

	packet, err := protocol.DeserializePacket(bytes.NewReader(buffer[:n]))
	if err != nil {
		return nil, err
	}

	return packet, nil
}

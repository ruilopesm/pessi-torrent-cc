package utils

import (
	"net"
	"testing"
)

func TestTCPAddrToBytes(t *testing.T) {
	tcpAddr := &net.TCPAddr{
		IP:   net.IPv4(192, 168, 1, 1),
		Port: 8080,
	}

	result := TCPAddrToBytes(tcpAddr)
	expected := [4]byte{192, 168, 1, 1}
	if result != expected {
		t.Errorf("TCPAddrToBytes: expected %v, got %v", expected, result)
	}
}

func TestUDPAddrToBytes(t *testing.T) {
	udpAddr := &net.UDPAddr{
		IP:   net.IPv4(192, 168, 1, 2),
		Port: 8080,
	}

	result := UDPAddrToBytes(udpAddr)
	expected := [4]byte{192, 168, 1, 2}
	if result != expected {
		t.Errorf("UDPAddrToBytes: expected %v, got %v", expected, result)
	}
}

package utils

import (
	"net"
	"testing"
)

func TestTCPAddrToBytes(t *testing.T) {
	// Create a sample TCP address
	tcpAddr := &net.TCPAddr{
		IP:   net.IPv4(192, 168, 1, 1),
		Port: 8080,
	}

	// Call the function
	result := TCPAddrToBytes(tcpAddr)

	// Expected result
	expected := [4]byte{192, 168, 1, 1}

	// Check if the result matches the expectation
	if result != expected {
		t.Errorf("TCPAddrToBytes: expected %v, got %v", expected, result)
	}
}

func TestUDPAddrToBytes(t *testing.T) {
	// Create a sample UDP address
	udpAddr := &net.UDPAddr{
		IP:   net.IPv4(192, 168, 1, 2),
		Port: 8080,
	}

	// Call the function
	result := UDPAddrToBytes(udpAddr)

	// Expected result
	expected := [4]byte{192, 168, 1, 2}

	// Check if the result matches the expectation
	if result != expected {
		t.Errorf("UDPAddrToBytes: expected %v, got %v", expected, result)
	}
}

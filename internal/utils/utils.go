package utils

import (
	"net"
)

func TCPAddrToBytes(addr net.Addr) [4]byte {
	ip := addr.(*net.TCPAddr).IP.To4()
	var result [4]byte
	copy(result[:], ip)

	return result
}

func UDPAddrToBytes(addr net.Addr) [4]byte {
	ip := addr.(*net.UDPAddr).IP.To4()
	var result [4]byte
	copy(result[:], ip)

	return result
}

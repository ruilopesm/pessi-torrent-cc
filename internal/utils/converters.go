package utils

import (
	"net"
	"strconv"
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

func StrToUDPPort(port string) (uint16, error) {
	udpPort, err := strconv.ParseUint(port, 10, 16)
	if err != nil {
		var zero uint16
		return zero, err
	}

	return uint16(udpPort), nil
}

func StrToUDPAddr(addr string) ([4]byte, error) {
	ip, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		var zero [4]byte
		return zero, err
	}

	return UDPAddrToBytes(ip), nil
}

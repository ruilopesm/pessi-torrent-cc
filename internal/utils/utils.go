package utils

import (
  "fmt"
  "net"
)


func IPv4ToByteArray(addr net.Addr) ([4]byte, error) {
	var ipv4Bytes [4]byte

	// Check if the provided address is a network address
	ipNet, ok := addr.(*net.IPAddr)
	if !ok {
		return ipv4Bytes, fmt.Errorf("not an IP address")
	}

	// Ensure the IP address is an IPv4 address
	ip := ipNet.IP.To4()
	if ip == nil {
		return ipv4Bytes, fmt.Errorf("not an IPv4 address")
	}

	copy(ipv4Bytes[:], ip)
	return ipv4Bytes, nil
}

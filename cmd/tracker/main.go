package main

import (
	"PessiTorrent/internal/common"
	"PessiTorrent/internal/serialization"
	"fmt"
	"net"
)

func handleClient(conn net.Conn) {
	defer conn.Close()

	buf := make([]byte, 1024)

	for {
		_, err := conn.Read(buf)
		if err != nil {
			fmt.Println("client disconnected")
			return
		}

		packetType := uint(buf[0])
		packet := common.PacketStructFromPacketType(packetType)
		err = serialization.Deserialize(buf, packet)
		if err != nil {
			fmt.Printf("couldn't deserialize struct from %s\n", buf)
			continue
		}

		fmt.Println("deserialized packet:", packet)
	}
}

func main() {
	listen, err := net.Listen("tcp", "localhost:8081")
	if err != nil {
		fmt.Println("error listening:", err)
		return
	}

	defer listen.Close()

	fmt.Println("server listening on port 8081")

	for {
		conn, err := listen.Accept()
		if err != nil {
			fmt.Println("error accepting connection:", err)
			continue
		}

		go handleClient(conn)
	}
}

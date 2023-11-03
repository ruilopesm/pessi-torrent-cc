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
		n, err := conn.Read(buf)
		if err != nil {
			fmt.Println("Connection closed.")
			return
		}

		data := buf[:n]

		message := string(data)
		fmt.Printf("Received: %s\n", message)

		packetType := uint(data[0])
		packet := common.PacketStructFromPacketType(packetType)
		if err = serialization.Deserialize(data, packet); err != nil {
			fmt.Printf("couldn't deserialize struct from %s\n", message)
			continue
		}

		// Process the received message (custom protocol logic)
		response := "Server received: " + message

		_, err = conn.Write([]byte(response))
		if err != nil {
			fmt.Println("Error writing to client:", err)
			return
		}
	}
}

func main() {
	port := "8081"
	listen, err := net.Listen("tcp", ":"+port)
	if err != nil {
		fmt.Println("Error listening:", err)
		return
	}
	defer listen.Close()

	fmt.Println("Server listening on port " + port)

	for {
		conn, err := listen.Accept()
		if err != nil {
			fmt.Println("Error accepting connection:", err)
			continue
		}

		go handleClient(conn)
	}
}

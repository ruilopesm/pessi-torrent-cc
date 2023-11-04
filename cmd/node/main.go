package main

import (
	"PessiTorrent/internal/common"
	"PessiTorrent/internal/serialization"
	"fmt"
	"net"
	"os"
)

func sendMessage(conn net.Conn, message []byte) error {
	_, err := conn.Write(message)
	if err != nil {
		return err
	}

	fmt.Println("sent bytes: ", message)

	return nil
}

func main() {
	conn, err := net.Dial("tcp", "localhost:8081")

	if err != nil {
		fmt.Println("error connecting to the server:", err)
		os.Exit(1)
	}

	defer conn.Close()

	var message common.RequestFilePacket
	message.Create("example.txt")

	serializedMessage, err := serialization.Serialize(message)
	if err != nil {
		fmt.Println("error serializing the message:", err)
		os.Exit(1)
	}

	err = sendMessage(conn, serializedMessage)
	if err != nil {
		fmt.Println("error sending the message:", err)
		os.Exit(1)
	}
}

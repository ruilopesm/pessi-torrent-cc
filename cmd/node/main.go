package main

import (
	"fmt"
	"net"
	"os"
)

func sendMessage(conn net.Conn, message string) (string, error) {
	_, err := conn.Write([]byte(message))
	if err != nil {
		return "", err
	}

	fmt.Println("Sent:", message)

	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	if err != nil {
		return "", err
	}

	response := string(buf[:n])
	return response, nil
}

func main() {
	// serverAddr := "localhost:8081"
	address := os.Args[1]
	port := os.Args[2]
	serverAddr := address + ":" + port

	conn, err := net.Dial("tcp", serverAddr)

	if err != nil {
		fmt.Println("Error connecting to the server:", err)
		os.Exit(1)
	}
	defer conn.Close()

	message := "HELLO WORLD"

	response, err := sendMessage(conn, message)
	if err != nil {
		fmt.Println("Error sending or receiving data:", err)
		os.Exit(1)
	}

	fmt.Println("Received:", response)
}

package main

import (
	"PessiTorrent/internal/packets"
	"bufio"
	"fmt"
	"os"
	"strings"
)

func (n *Node) Cli() error {
	reader := bufio.NewReader(os.Stdin)

	for {
		select {
		case <-n.quitch:
			fmt.Println("Exiting CLI")
			return nil
		default:
			fmt.Print("> ")
			input, _ := reader.ReadString('\n')
			input = strings.TrimSpace(input)

			// Split the input into command and arguments
			parts := strings.Fields(input)
			if len(parts) == 0 {
				continue
			}

			command := parts[0]
			args := parts[1:]

			switch command {

			case "GET":
				if len(args) == 1 {
					n.get(args[0])
				} else {
					fmt.Println("Usage: GET <file>")
				}

			case "EXIT":
				fmt.Println("Exiting CLI")
				close(n.quitch)
				return nil

			default:
				fmt.Println("Invalid command. Supported commands: GET, EXIT")
			}
		}
	}
}

func (n *Node) get(fileName string) error {
	var packet packets.RequestFilePacket
	packet.Create(fileName)
	err := n.conn.WritePacket(packet)
	if err != nil {
		return err
	}

	return nil
}

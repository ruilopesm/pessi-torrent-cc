package main

import (
	"PessiTorrent/internal/cli"
	"PessiTorrent/internal/packets"
	"fmt"
)

func (n *Node) SetCommands() {
	n.commands = map[string]cli.Command{
		"GET": {
			Usage:   "get <filename>",
			Execute: n.get,
		},
		"EXIT": {
			Usage:   "exit",
			Execute: n.exit,
		},
		"HELP": {
			Usage:   "help",
			Execute: n.help,
		},
	}
}

// commands

func (n *Node) get(args []string) error {
	var packet packets.RequestFilePacket
	packet.Create(args[0])
	err := n.conn.WritePacket(packet)
	if err != nil {
		return err
	}

	return nil
}

func (n *Node) exit(args []string) error {
	close(n.quitch)

	return nil
}

func (n *Node) help(args []string) error {
	fmt.Println("Available commands:")
	for command, cmd := range n.commands {
		fmt.Printf("%s - %s\n", command, cmd.Usage)
	}
	return nil
}

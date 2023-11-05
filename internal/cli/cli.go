package cli

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

type CLI struct {
	commands     map[string]Command
	shutdownHook func()
}

type Command struct {
	Name         string
	Usage        string
	NumberOfArgs int
	Execute      func(args []string) error
}

func NewCLI(shutdownHook func()) *CLI {
	return &CLI{
		commands:     make(map[string]Command),
		shutdownHook: shutdownHook,
	}
}

func (c *CLI) AddCommand(name string, usage string, numberOfArgs int, execute func(args []string) error) {
	c.commands[name] = Command{
		Name:         name,
		Usage:        usage,
		NumberOfArgs: numberOfArgs,
		Execute:      execute,
	}
}

func (c *CLI) Start() error {
	c.help()

	for {
		fmt.Print("> ")

		reader := bufio.NewReader(os.Stdin)
		input, err := reader.ReadString('\n')
		if err != nil {
			return err
		}
		input = strings.TrimSuffix(input, "\n")
		parts := strings.Split(input, " ")

		// Check if the command was previously registered
		if cmd, ok := c.commands[parts[0]]; ok {
			args := parts[1:]
			if len(args) != cmd.NumberOfArgs {
				fmt.Printf("Wrong number of arguments for %s\n", cmd.Name)
				fmt.Printf("Usage: %s %s\n", cmd.Name, cmd.Usage)
				continue
			}

			err := cmd.Execute(args)
			if err != nil {
				fmt.Printf("Error executing command %s: %s\n", cmd.Name, err)
			}
		} else if parts[0] == "exit" {
			c.shutdownHook()
			break
		} else {
			fmt.Printf("Unknown command %s\n", parts[0])
			c.help()
			continue
		}
	}

	return nil
}

func (c *CLI) help() {
	fmt.Println("Available commands:")

	for _, cmd := range c.commands {
		fmt.Printf("\t%s\t%s\n", cmd.Name, cmd.Usage)
	}

	fmt.Println("\texit\tExit the program")
}

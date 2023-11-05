package cli

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

type Command struct {
	Usage   string
	Execute func(args []string) error
}

type CommandsSetter interface {
	SetCommands()
}

func StartCLI(commands map[string]Command, quitch chan struct{}) error {
	reader := bufio.NewReader(os.Stdin)

	for {
		select {
		case <-quitch:
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

			command := strings.ToUpper(parts[0])
			args := parts[1:]

			cmd, exists := commands[command]
			if exists {
				err := cmd.Execute(args)
				if err != nil {
					continue
				}
			} else {
				cmd, _ = commands["HELP"]
				cmd.Execute(nil)
			}
		}
	}
}

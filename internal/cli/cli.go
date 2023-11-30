package cli

import (
	"PessiTorrent/internal/logger"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"golang.org/x/term"
)

type CLI struct {
	commands     map[string]Command
	shutdownHook func()
	console      Console
}

func NewCLI(shutdownHook func(), console Console) CLI {
	return CLI{
		commands:     make(map[string]Command),
		shutdownHook: shutdownHook,
		console:      console,
	}
}

type Console struct {
	Term     *term.Terminal
	OldState *term.State
}

func NewConsole() Console {
	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		panic(err)
	}

	screen := struct {
		io.Reader
		io.Writer
	}{os.Stdin, os.Stdout}

	terminal := term.NewTerminal(screen, "> ")
	return Console{
		Term:     terminal,
		OldState: oldState,
	}
}

func (c *Console) ReadInput() (string, error) {
	return c.Term.ReadLine()
}

func (c *Console) Close() {
	err := term.Restore(int(os.Stdin.Fd()), c.OldState)
	if err != nil {
		panic(err)
	}
}

func (c *Console) Info(message string, args ...any) {
	message = fmt.Sprintf(message, args...)
	_, _ = c.Term.Write([]byte(message + "\n"))
}

func (c *Console) Warn(message string, args ...any) {
	message = fmt.Sprintf(message, args...)
	_, _ = c.Term.Write([]byte(message + "\n"))
}

func (c *Console) Error(message string, args ...any) {
	message = fmt.Sprintf(message, args...)
	_, _ = c.Term.Write([]byte(message + "\n"))
}

type Command struct {
	Name         string
	Usage        string
	Description  string
	NumberOfArgs int
	Execute      func(args []string) error
}

func (c *CLI) AddCommand(name string, usage string, description string, numberOfArgs int, execute func(args []string) error) {
	c.commands[name] = Command{
		Name:         name,
		Usage:        usage,
		Description:  description,
		NumberOfArgs: numberOfArgs,
		Execute:      execute,
	}
}

func (c *CLI) Start() {
	logger.Info("Type 'help' for a list of available commands")

	for {
		input, err := c.console.ReadInput()
		if err != nil {
			if errors.Is(err, io.EOF) {
				c.shutdownHook()
				break
			}
			log.Panicf("Error reading input: %s\n", err)
		}

		input = strings.TrimSuffix(input, "\n")
		input = strings.TrimSuffix(input, "\r") // Windows

		parts := strings.Split(input, " ")

		// Check if the command was previously registered
		if cmd, ok := c.commands[parts[0]]; ok {
			args := parts[1:]
			if len(args) != cmd.NumberOfArgs {
				logger.Warn("Wrong number of arguments for %s.", cmd.Name)
				logger.Warn("Usage: %s %s.", cmd.Name, cmd.Usage)
				continue
			}

			err := cmd.Execute(args)
			if err != nil {
				logger.Warn("Error executing command %s: %s.", cmd.Name, err)
			}
		} else if parts[0] == "exit" {
			c.shutdownHook()
			break
		} else if parts[0] == "help" {
			c.help()
		} else {
			logger.Info("Unknown command %s.", parts[0])
			c.help()
			continue
		}
	}
}

func (c *CLI) help() {
	logger.Info("Available commands:")

	for _, cmd := range c.commands {
		if cmd.Usage != "" {
			logger.Info("\t%s\t%s\t%s", cmd.Name, cmd.Usage, cmd.Description)
		} else {
			logger.Info("\t%s\t%s", cmd.Name, cmd.Description)
		}
	}

	logger.Info("\thelp\tShow this help")
	logger.Info("\texit\tExit the program")
}

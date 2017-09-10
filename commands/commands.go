package commands

import (
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
)

// commMap maps the name of a command to the function which executes the command
var commMap = map[string]command{
	"ping": pingCommand,
}

// HelpCommand is the name of the command which lists the commands available
// and can give information about a specific command.
const HelpCommand = "help"

// command is a function which executes the given command and arguments on
// the provided discord session.
type command func(*discordgo.Session, *discordgo.MessageCreate, []string) error

// pingCommand is a casic command which will responsd "Pong!" to any ping.
func pingCommand(s *discordgo.Session, m *discordgo.MessageCreate, _ []string) error {
	s.ChannelMessageSend(m.ChannelID, "Pong!")
	return nil
}

// Parse parses a msg and executes the command, if it exists.
func Parse(s *discordgo.Session, m *discordgo.MessageCreate, msg string) error {
	args := strings.Fields(msg)
	if c, ok := commMap[args[0]]; ok {
		if err := c(s, m, args); err != nil {
			return fmt.Errorf("failed to execute command: %s", err)
		}
		return nil
	}
	return fmt.Errorf("Failed to recognise the command %q", args[0])
}

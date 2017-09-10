package commands

import (
	"context"
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
)

// commMap maps the name of a command to the function which executes the command
var commMap = map[string]command{
	"ping": pingCommand,
}

// HelpCommand is the name of the command which explains what commands are
// available from the bot as well as giving help on individual commands.
const HelpCommand = "help"

// command holds the information that makes up a command
type command struct {
	// help is a string which explains the purpose and usage of the command
	help string
	// exec is a function which executes the command message on the given
	// discord session.
	exec func(context.Context, *discordgo.Session, *discordgo.MessageCreate, []string) error
}

// pingCommand is a basic command which will responsd "Pong!" to any ping.
func pingCommand(ctx context.Context, s *discordgo.Session, m *discordgo.MessageCreate, _ []string) error {
	s.ChannelMessageSend(m.ChannelID, "Pong!")
	return nil
}

// Execute parses a msg and executes the command, if it exists.
func Execute(s *discordgo.Session, m *discordgo.MessageCreate, msg string) error {
	args := strings.Fields(msg)
	if c, ok := commMap[args[0]]; ok {
		if err := c(s, m, args); err != nil {
			return fmt.Errorf("failed to execute command %q: %s", args[0], err)
		}
		return nil
	}
	return fmt.Errorf("Failed to recognise the command %q", args[0])
}

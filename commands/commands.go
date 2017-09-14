package commands

import (
	"context"
	"fmt"
	"strings"

	"../logging"
	"github.com/bwmarrin/discordgo"
)

// commMap maps the name of a command to the function which executes the command
var commMap = map[string]*command{
	"ping": &command{
		help: "Responds 'Pong!' to and 'ping'",
		exec: pingCommand,
	},
}

// HelpCommand is the name of the command which lists the commands available
// and can give information about a specific command.
const HelpCommand = "help"

// command is a function which executes the given command and arguments on
// the provided discord session.
type command struct {
	help string
	exec func(context.Context, *discordgo.Session, *discordgo.MessageCreate, []string) error
}

// pingCommand is a basic command which will responds "Pong!" to any ping.
func pingCommand(ctx context.Context, s *discordgo.Session, m *discordgo.MessageCreate, _ []string) error {
	l, ok := logging.FromContext(ctx)
	s.ChannelMessageSend(m.ChannelID, "Pong!")
	if ok {
		l.Infof("Responding 'Pong!' to ping command")
	}
	return nil
}

// showHelpCommand lists all of the commands available and explains what they do.
func showHelpCommand(ctx context.Context, s *discordgo.Session, m *discordgo.MessageCreate, _ []string) error {
	if l, ok := logging.FromContext(ctx); ok {
		l.Infof("Responding to help command")
	}
	var out string
	for name, c := range commMap {
		out += fmt.Sprintf("%s: %s\n", name, c.help)
	}
	s.ChannelMessageSend(m.ChannelID, out)
	return nil
}

// isHelpCommand tells you whether the given message arguments are calling the help command.
func isHelpCommand(msgArgs []string) bool {
	if len(msgArgs) < 1 {
		return false
	}
	return msgArgs[0] == HelpCommand
}

// Execute parses a msg and executes the command, if it exists.
func Execute(ctx context.Context, s *discordgo.Session, m *discordgo.MessageCreate, msg string) error {
	args := strings.Fields(msg)
	// the help command is a special case because the help command must loop though
	// the map of all other commands.
	if isHelpCommand(args) {
		if err := showHelpCommand(ctx, s, m, args); err != nil {
			return fmt.Errorf("failed to execute help command: %s", err)
		}
		return nil
	}
	if c, ok := commMap[args[0]]; ok {
		if err := c.exec(ctx, s, m, args); err != nil {
			return fmt.Errorf("failed to execute command: %s", err)
		}
		return nil
	}
	return fmt.Errorf("Failed to recognise the command %q", args[0])
}

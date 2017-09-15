package commands

import (
	"context"
	"fmt"
	"reflect"
	"runtime"
	"strings"

	"github.com/UCCNetworkingSociety/Netsoc-Discord-Bot/logging"
	"github.com/bwmarrin/discordgo"
)

var (
	// commMap maps the name of a command to the function which executes the command
	commMap map[string]*command

	// savedAliases stores all shortcut alias commands
	savedAliases = map[string]string{}
)

// HelpCommand is the name of the command which lists the commands available
// and can give information about a specific command.
const HelpCommand = "help"

// command is a function which executes the given command and arguments on
// the provided discord session.
type command struct {
	help string
	exec func(context.Context, *discordgo.Session, *discordgo.MessageCreate, []string) error
}

func init() {
	commMap = map[string]*command{}

	err := LoadFromStorage("./storage/aliases.json", &savedAliases)
	if err != nil {
		panic(err)
	}

	for key, value := range savedAliases {
		commMap[key] = &command{
			help: value,
			exec: printShortcut,
		}
	}

	// Put registered commands after alias registration
	// to ensure an alias doesn't overwrite their
	// functionality
	commMap["ping"] = &command{
		help: "Responds 'Pong!' to and 'ping'.",
		exec: pingCommand,
	}

	commMap["alias"] = &command{
		help: "Sets a shortcut command. Usage: !alias keyword url_link_to_resource",
		exec: aliasCommand,
	}

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

// aliasCommand sets string => string shortcut that can be called later to print a value
func aliasCommand(ctx context.Context, s *discordgo.Session, m *discordgo.MessageCreate, c []string) error {
	l, loggerOk := logging.FromContext(ctx)

	if len(c) < 3 {
		s.ChannelMessageSend(m.ChannelID, "Too few arguments supplied. Refer to !help for usage.")
		return fmt.Errorf("Too few arguments supplied for set command")
	} else if len(c) > 3 {
		s.ChannelMessageSend(m.ChannelID, "Too many arguments supplied. Refer to !help for usage.")
		return fmt.Errorf("Too many arguments supplied for set command")
	}

	// Ensure user has permission to use this command
	isAllowed := IsAllowed(ctx, s, m.Author.ID, "alias")

	if isAllowed {
		if _, ok := commMap[c[1]]; (ok && GetFunctionName(printShortcut) != GetFunctionName(commMap[c[1]].exec)) || c[1] == HelpCommand {
			// If key already exists and who's function is not printShortcut OR is not the "help" command
			s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("%s is a registered command and cannot be set as an alias.", c[1]))
			if loggerOk {
				l.Infof("%s attempted to overwrite command %s with %s", m.Author, c[1], c[2])
			}
		} else {
			// If key does not exist (or who's function is printShortcut)
			commMap[c[1]] = &command{
				help: c[2],
				exec: printShortcut,
			}

			savedAliases[c[1]] = c[2]
			WriteToStorage("./storage/aliases.json", savedAliases)

			if loggerOk {
				l.Infof("%s has set an alias for %s => %s", m.Author, c[1], c[2])
			}

			s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("%s has set an alias for %s => %s", m.Author, c[1], c[2]))
		}
	} else {
		s.ChannelMessageSend(m.ChannelID, "You do not have permissions to use this command.")
		if loggerOk {
			l.Infof("%q is not allowed to execute the alias command", m.Author)
		}
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

// printShortcut uses the help text of the command to print the shortcut's value
func printShortcut(ctx context.Context, s *discordgo.Session, m *discordgo.MessageCreate, args []string) error {
	s.ChannelMessageSend(m.ChannelID, commMap[args[0]].help)
	return nil
}

// GetFunctionName returns the name of a given function
func GetFunctionName(i interface{}) string {
	return runtime.FuncForPC(reflect.ValueOf(i).Pointer()).Name()
}

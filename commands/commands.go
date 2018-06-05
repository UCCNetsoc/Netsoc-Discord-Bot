package commands

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/bwmarrin/discordgo"
)

var (
	// commMap maps the name of a command to the function which executes the command
	commMap map[string]Command
)

// HelpCommand is the name of the command which lists the commands available
// and can give information about a specific command.
const HelpCommand = "help"

// messageContextValue is used to wrap the type of any values stored in the
// context propogated to a command function.
type messageContextValue string

// Command defines functions which every kind of command needs
type Command interface {
	Help() string
	Exec(context.Context, *discordgo.Session, *discordgo.MessageCreate, []string) error
}

// textCommand is a command which returns a textual response
type textCommand struct {
	command  func(ctx context.Context, args []string) (string, error)
	helpText string
}

// commandFunc returns the internal user-defined command function. This is used for
// testing mainly.
func (c *textCommand) commandFunc() func(ctx context.Context, args []string) (string, error) {
	return c.command
}

// Help implements Command.Help
func (c *textCommand) Help() string {
	return c.helpText
}

// Exec implements Command.Exec
func (c *textCommand) Exec(ctx context.Context, s *discordgo.Session, m *discordgo.MessageCreate, args []string) error {
	textResponse, err := c.command(ctx, args)
	if err != nil {
		errToSend := fmt.Errorf("Failed to execute text-command with args %q: %s", args, err)
		if _, sendErr := s.ChannelMessageSend(m.ChannelID, errToSend.Error()); sendErr != nil {
			return fmt.Errorf("Failed to send error message to the channal %q: %s", m.ChannelID, sendErr)
		}
		return errToSend
	}

	if _, err := s.ChannelMessageSend(m.ChannelID, textResponse); err != nil {
		return fmt.Errorf("Failed to send message to the channal %q: %s", m.ChannelID, err)
	}
	return nil
}

// embedCommand is a command which returns an embed message
type embedCommand struct {
	command  func(ctx context.Context, args []string) (*discordgo.MessageEmbed, error)
	helpText string
}

// Help implements Command.Help
func (c *embedCommand) Help() string {
	return c.helpText
}

// Exec implements Command.Exec
func (c *embedCommand) Exec(ctx context.Context, s *discordgo.Session, m *discordgo.MessageCreate, args []string) error {
	embed, err := c.command(ctx, args)
	if err != nil {
		errToSend := fmt.Errorf("Failed to execute embed-command with args %q: %s", args, err)
		if _, sendErr := s.ChannelMessageSend(m.ChannelID, errToSend.Error()); sendErr != nil {
			return fmt.Errorf("Failed to send error message to the channal %q: %s", m.ChannelID, sendErr)
		}
		return errToSend
	}

	if _, err := s.ChannelMessageSendEmbed(m.ChannelID, embed); err != nil {
		return fmt.Errorf("Failed to send embed to the channal %q: %s", m.ChannelID, err)
	}
	return nil
}

func init() {
	commMap = make(map[string]Command)
	commMap, err := withAliasCommands(commMap)
	if err != nil {
		log.Fatalf("Failed to initilise alias commands: %s", err)
	}

	// Put registered commands after alias registration
	// to ensure an alias doesn't overwrite their
	// functionality
	commMap["ping"] = &textCommand{
		helpText: "Responds 'Pong!' to and 'ping'.",
		command:  pingCommand,
	}

	commMap["alias"] = &textCommand{
		helpText: "Sets a shortcut command. Usage: !alias keyword url_link_to_resource",
		command:  aliasCommand,
	}

	commMap[HelpCommand] = &embedCommand{
		helpText: "If followed by a command name, it shows the details of the command",
		command:  showHelpCommand,
	}

	commMap["info"] = &embedCommand{
		helpText: "Displays some info about NetsocBot",
		command:  infoCommand,
	}

	commMap["config"] = &embedCommand{
		helpText: "Displays the config for NetsocBot",
		command:  configCommand,
	}

	commMap["inspire"] = &textCommand{
		helpText: "Gives an inspirational quote",
		command:  inspireCommand,
	}

	commMap["minecraft"] = &textCommand{
		helpText: "Check to see who is currently online on the NetsocCraft server",
		command:  minecraftCommand,
	}

	commMap["unalias"] = &textCommand{
		helpText: "Remove an alias from the stored aliases",
		command:  unAliasCommand,
	}
}

// Execute parses a msg and executes the command, if it exists.
func Execute(ctx context.Context, s *discordgo.Session, m *discordgo.MessageCreate, msg string) error {
	args := strings.Fields(msg)
	if c, ok := commMap[args[0]]; ok {
		// Ensure user has permission to use this command
		if !IsAllowed(ctx, s, m.Author.ID, args[0]) {
			if _, err := s.ChannelMessageSend(m.ChannelID, "You do not have permissions to use this command."); err != nil {
				return fmt.Errorf("Failed to send permssion denial message to the channal %q: %s", m.ChannelID, err)
			}
			return fmt.Errorf("%q is not allowed to execute the command %q", m.Author, args[0])
		}
		ctx = context.WithValue(ctx, messageContextValue("ChannelID"), m.ChannelID)
		ctx = context.WithValue(ctx, messageContextValue("Session"), s)
		if err := c.Exec(ctx, s, m, args); err != nil {
			return fmt.Errorf("failed to execute command: %s", err)
		}
		return nil
	}
	return fmt.Errorf("Failed to recognise the command %q", args[0])
}

package commands

import (
	"context"
	"fmt"
	"os/exec"
	"reflect"
	"runtime"
	"strconv"
	"strings"

	"github.com/UCCNetworkingSociety/Netsoc-Discord-Bot/config"
	"github.com/UCCNetworkingSociety/Netsoc-Discord-Bot/logging"
	"github.com/bwmarrin/discordgo"
	"github.com/ulule/deepcopier"
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
	exec func(context.Context, *discordgo.Session, *discordgo.MessageCreate, []string) (*discordgo.Message, error)
}

func init() {
	commMap = map[string]*command{}

	LoadFromStorage("storage/aliases.json", &savedAliases)

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

	commMap[HelpCommand] = &command{
		help: "If followed by a command name, it shows the details of the command",
		exec: showHelpCommand,
	}

	commMap["top"] = &command{
		help: "Prints the output of `top -b -n 1`",
		exec: topCommand,
	}

	commMap["sensors"] = &command{
		help: "Displays temperature of the server",
		exec: sensorsCommand,
	}

	commMap["info"] = &command{
		help: "Displays some info about NetsocBot",
		exec: infoCommand,
	}

	commMap["config"] = &command{
		help: "Displays the config for NetsocBot",
		exec: configCommand,
	}
}

func configCommand(ctx context.Context, s *discordgo.Session, m *discordgo.MessageCreate, _ []string) (*discordgo.Message, error) {
	l, loggerOk := logging.FromContext(ctx)
	if loggerOk {
		l.Infof("Responding to config command", nil)
	}

	// Ensure user has permission to use this command
	if !IsAllowed(ctx, s, m.Author.ID, "config") {
		returnMsg, _ := s.ChannelMessageSend(m.ChannelID, "You do not have permissions to use this command.")
		if loggerOk {
			l.Infof("%q is not allowed to execute the config command", m.Author)
		}
		return returnMsg, nil
	}

	tmpconf := &config.Config{}
	deepcopier.Copy(config.GetConfig()).To(tmpconf)

	returnMsg, _ := s.ChannelMessageSendEmbed(m.ChannelID, &discordgo.MessageEmbed{
		Color: 0,

		Fields: []*discordgo.MessageEmbedField{
			{Name: "Config:", Value: tmpconf.String(), Inline: true},
		},
	})

	return returnMsg, nil
}

func infoCommand(ctx context.Context, s *discordgo.Session, m *discordgo.MessageCreate, _ []string) (*discordgo.Message, error) {
	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)

	returnMsg, _ := s.ChannelMessageSendEmbed(m.ChannelID, &discordgo.MessageEmbed{
		Color: 0,

		Fields: []*discordgo.MessageEmbedField{
			{Name: "Memory Usage:", Value: "```" + strconv.Itoa(int(mem.Alloc/1024/1024)) + "MB" + "```", Inline: true},
			{Name: "Goroutines:", Value: "```" + strconv.Itoa(runtime.NumGoroutine()) + "```", Inline: true},
			{Name: "Go Version:", Value: "```" + runtime.Version() + "```", Inline: true},
			{Name: "Usable Cores:", Value: "```" + strconv.Itoa(runtime.NumCPU()) + "```", Inline: true},
		},
	})

	return returnMsg, nil
}

func sensorsCommand(ctx context.Context, s *discordgo.Session, m *discordgo.MessageCreate, _ []string) (*discordgo.Message, error) {
	l, ok := logging.FromContext(ctx)
	if ok {
		l.Infof("Responding to top command")
	}

	cmd := exec.Command("sensors")
	stdout, err := cmd.Output()
	if err != nil {
		l.Errorf("sensors command error %s", err)
		return nil, err
	}

	returnMsg, _ := s.ChannelMessageSend(m.ChannelID, "```"+string(stdout)+"```")

	return returnMsg, nil
}

func topCommand(ctx context.Context, s *discordgo.Session, m *discordgo.MessageCreate, _ []string) (*discordgo.Message, error) {
	l, ok := logging.FromContext(ctx)
	if ok {
		l.Infof("Responding to top command")
	}

	cmd := exec.Command("top", "-b", "-n", "1")
	stdout, err := cmd.Output()
	if err != nil {
		l.Errorf("top command error %s", err)
		return nil, err
	}

	returnMsg, _ := s.ChannelMessageSend(m.ChannelID, "```"+string(stdout[:1994])+"```")

	return returnMsg, nil
}

// pingCommand is a basic command which will responds "Pong!" to any ping.
func pingCommand(ctx context.Context, s *discordgo.Session, m *discordgo.MessageCreate, _ []string) (*discordgo.Message, error) {
	l, ok := logging.FromContext(ctx)
	returnMsg, _ := s.ChannelMessageSend(m.ChannelID, "Pong!")
	if ok {
		l.Infof("Responding 'Pong!' to ping command")
	}
	return returnMsg, nil
}

// aliasCommand sets string => string shortcut that can be called later to print a value
func aliasCommand(ctx context.Context, s *discordgo.Session, m *discordgo.MessageCreate, c []string) (*discordgo.Message, error) {
	l, loggerOk := logging.FromContext(ctx)

	if len(c) < 3 {
		returnMsg, _ := s.ChannelMessageSend(m.ChannelID, "Too few arguments supplied. Refer to !help for usage.")
		return returnMsg, fmt.Errorf("Too few arguments supplied for set command")
	} else if len(c) > 3 {
		returnMsg, _ := s.ChannelMessageSend(m.ChannelID, "Too many arguments supplied. Refer to !help for usage.")
		return returnMsg, fmt.Errorf("Too many arguments supplied for set command")
	}

	// Ensure user has permission to use this command
	if !IsAllowed(ctx, s, m.Author.ID, "alias") {
		returnMsg, _ := s.ChannelMessageSend(m.ChannelID, "You do not have permissions to use this command.")
		if loggerOk {
			l.Infof("%q is not allowed to execute the alias command", m.Author)
		}
		return returnMsg, nil
	}

	if _, ok := commMap[c[1]]; ok && GetFunctionName(printShortcut) != GetFunctionName(commMap[c[1]].exec) {
		// If key already exists and who's function is not printShortcut OR is not the "help" command
		returnMsg, _ := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("%s is a registered command and cannot be set as an alias.", c[1]))
		if loggerOk {
			l.Infof("%s attempted to overwrite command %s with %s", m.Author, c[1], c[2])
		}
		return returnMsg, nil
	}

	// If key does not exist (or who's function is printShortcut)
	commMap[c[1]] = &command{
		help: c[2],
		exec: printShortcut,
	}

	savedAliases[c[1]] = c[2]
	if err := WriteToStorage("./storage/aliases.json", savedAliases); err != nil {
		l.Errorf("Error writing alias to file")
		returnMsg, _ := s.ChannelMessageSend(m.ChannelID, "Error writing alias to file.")
		return returnMsg, err
	}

	if loggerOk {
		l.Infof("%s has set an alias for %s => %s", m.Author, c[1], c[2])
	}

	returnMsg, _ := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("%s has set an alias for %s => %s", m.Author, c[1], c[2]))

	return returnMsg, nil
}

// showHelpCommand lists all of the commands available and explains what they do.
func showHelpCommand(ctx context.Context, s *discordgo.Session, m *discordgo.MessageCreate, msg []string) (*discordgo.Message, error) {
	if l, ok := logging.FromContext(ctx); ok {
		l.Infof("Responding to help command")
	}

	if len(msg) == 2 {
		if c, ok := commMap[msg[1]]; ok {
			returnMsg, _ := s.ChannelMessageSendEmbed(m.ChannelID, &discordgo.MessageEmbed{
				Color: 0,

				Fields: []*discordgo.MessageEmbedField{
					{Name: msg[1], Value: c.help},
				},
			})
			return returnMsg, nil
		}

		returnMsg, _ := s.ChannelMessageSend(m.ChannelID, "Command not found.")
		return returnMsg, nil
	}

	returnMsg, _ := s.ChannelMessageSendEmbed(m.ChannelID, &discordgo.MessageEmbed{
		Color: 0,

		Fields: func() []*discordgo.MessageEmbedField {
			var out []*discordgo.MessageEmbedField

			for name, c := range commMap {
				out = append(out, &discordgo.MessageEmbedField{
					Name:  name,
					Value: c.help,
				})
			}

			return out
		}(),
	})
	return returnMsg, nil
}

// Execute parses a msg and executes the command, if it exists.
func Execute(ctx context.Context, s *discordgo.Session, m *discordgo.MessageCreate, msg string) error {
	args := strings.Fields(msg)
	fmt.Println(commMap)
	// the help command is a special case because the help command must loop though
	// the map of all other commands.
	if c, ok := commMap[args[0]]; ok {
		if err, _ := c.exec(ctx, s, m, args); err != nil {
			return fmt.Errorf("failed to execute command: %#v", err)
		}
		return nil
	}
	return fmt.Errorf("Failed to recognise the command %q", args[0])
}

// printShortcut uses the help text of the command to print the shortcut's value
func printShortcut(ctx context.Context, s *discordgo.Session, m *discordgo.MessageCreate, args []string) (*discordgo.Message, error) {
	returnMsg, _ := s.ChannelMessageSend(m.ChannelID, commMap[args[0]].help)
	return returnMsg, nil
}

// GetFunctionName returns the name of a given function
func GetFunctionName(i interface{}) string {
	return runtime.FuncForPC(reflect.ValueOf(i).Pointer()).Name()
}

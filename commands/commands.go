package commands

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/url"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/Necroforger/dgwidgets"
	"github.com/UCCNetworkingSociety/Netsoc-Discord-Bot/config"
	"github.com/UCCNetworkingSociety/Netsoc-Discord-Bot/logging"
	"github.com/bwmarrin/discordgo"
)

var (
	// commMap maps the name of a command to the function which executes the command
	commMap map[string]*command

	aliasMap map[string]*command
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
	aliasMap = map[string]*command{}

	LoadFromStorage("storage/aliases.json", &savedAliases)

	for key, value := range savedAliases {
		aliasMap[key] = &command{
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

	commMap["inspire"] = &command{
		help: "Gives an inspirational quote",
		exec: inspireCommand,
	}

	commMap["minecraft"] = &command{
		help: "Check to see who is currently online on the NetsocCraft server",
		exec: minecraftCommand,
	}
}

// minecraftCommand checks the stats of minecraft.netsoc.co for that moment
// data comes from http://minecraft.netsoc.co/standalone/dynmap_NetsocCraft.json
func minecraftCommand(ctx context.Context, s *discordgo.Session, m *discordgo.MessageCreate, _ []string) (*discordgo.Message, error) {
	l, loggerOk := logging.FromContext(ctx)
	if loggerOk {
		l.Infof("Responding to minecraft command")
	}

	resp, err := http.Get("http://minecraft.netsoc.co/standalone/dynmap_NetsocCraft.json")

	if err != nil {
		return nil, fmt.Errorf("Failed to retrieve data from the Minecraft Server: %v", err)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("Failed to read the response body: %v", err)
	}

	// Unmarshal the data and get all the player lists I guess
	q := &struct {
		CurrentCount int  `json:"currentcount"`
		HasStorm     bool `json:"hasStorm"`
		IsThundering bool `json:"isThundering"`
		ConfigHash   int  `json:"confighash"`
		ServerTime   int  `json:"servertime"`
		TimeStamp    int  `json:"timestamp"`
		Players      []struct {
			// Players nested JSON structure
			World   string  `json:"world"`
			Armour  int     `json:"armor"`
			Name    string  `json:"name"`
			X       float64 `json:"x"`
			Y       float64 `json:"y"`
			Z       float64 `json:"z"`
			Sort    int     `json:"sort"`
			Type    string  `json:"type"`
			Account string  `json:"account"`
		} `json:"players"`
		Updates []interface{} `json:"updates"`
	}{}

	if err := json.Unmarshal(body, q); err != nil {
		return nil, fmt.Errorf("Failed to parse response json %q: %v", string(body), err)
	}

	// Create a discord message containing a list of all the players currently online
	var msg string
	if len(q.Players) == 0 {
		msg = "Nobody home :("
	} else {
		msg = "```markdown\n"
		for i, player := range q.Players {
			msg += fmt.Sprintf("%d. %s\n", i+1, player.Name)
		}
		msg += "```"
	}

	// Attempt to send the message to the discord
	retMsg, err := s.ChannelMessageSend(m.ChannelID, msg)
	if err != nil {
		return nil, fmt.Errorf("Failed to send message to the channal %q: %v", m.ChannelID, err)
	}
	if loggerOk {
		l.Infof("Sending minecraft information: %d users active", len(q.Players))
	}
	return retMsg, nil
}

// inspireCommand gets an inspirational quote from forismatic.com
func inspireCommand(ctx context.Context, s *discordgo.Session, m *discordgo.MessageCreate, _ []string) (*discordgo.Message, error) {
	l, loggerOk := logging.FromContext(ctx)
	if loggerOk {
		l.Infof("Responding to inspire command", nil)
	}

	resp, err := http.PostForm("http://api.forismatic.com/api/1.0/",
		url.Values{
			"method": {"getQuote"},
			"format": {"json"},
			"key":    {strconv.Itoa(rand.Intn(1000000))},
			"lang":   {"en"},
		})
	if err != nil {
		return nil, fmt.Errorf("Failed to get the quote from API: %v", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("Failed to read the response body: %v", err)
	}

	q := &struct {
		QuoteText   string `json:"quoteTxt"`
		QuoteAuthor string `json:"quoteAuthor"`
	}{}
	if err := json.Unmarshal(body, q); err != nil {
		return nil, fmt.Errorf("Failed to parse response json %q: %v", string(body), err)
	}

	retMsg, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("%q - %s", q.QuoteText, q.QuoteAuthor))
	if err != nil {
		return nil, fmt.Errorf("Failed to send message to the channal %q: %v", m.ChannelID, err)
	}
	if loggerOk {
		l.Infof("Sending quote %q", q.QuoteText)
	}
	return retMsg, nil
}

func configCommand(ctx context.Context, s *discordgo.Session, m *discordgo.MessageCreate, _ []string) (*discordgo.Message, error) {
	l, loggerOk := logging.FromContext(ctx)
	if loggerOk {
		l.Infof("Responding to config command", nil)
	}

	// Ensure user has permission to use this command
	if !IsAllowed(ctx, s, m.Author.ID, "config") {
		retMsg, err := s.ChannelMessageSend(m.ChannelID, "You do not have permissions to use this command.")
		if err != nil {
			return nil, fmt.Errorf("Failed to send message to the channal %q: %v", m.ChannelID, err)
		}
		if loggerOk {
			l.Infof("%q is not allowed to execute the config command", m.Author)
		}
		return retMsg, nil
	}

	retMsg, err := s.ChannelMessageSendEmbed(m.ChannelID, &discordgo.MessageEmbed{
		Color: 0,

		Fields: []*discordgo.MessageEmbedField{
			{Name: "Config:", Value: config.GetConfig().String(), Inline: true},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("Failed to send message to the channal %q: %v", m.ChannelID, err)
	}

	return retMsg, nil
}

func infoCommand(ctx context.Context, s *discordgo.Session, m *discordgo.MessageCreate, _ []string) (*discordgo.Message, error) {
	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)

	retMsg, err := s.ChannelMessageSendEmbed(m.ChannelID, &discordgo.MessageEmbed{
		Color: 0,

		Fields: []*discordgo.MessageEmbedField{
			{Name: "Memory Usage:", Value: "```" + strconv.Itoa(int(mem.Alloc/1024/1024)) + "MB" + "```", Inline: true},
			{Name: "Goroutines:", Value: "```" + strconv.Itoa(runtime.NumGoroutine()) + "```", Inline: true},
			{Name: "Go Version:", Value: "```" + runtime.Version() + "```", Inline: true},
			{Name: "Usable Cores:", Value: "```" + strconv.Itoa(runtime.NumCPU()) + "```", Inline: true},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("Failed to send message to the channal %q: %v", m.ChannelID, err)
	}

	return retMsg, nil
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

	retMsg, err := s.ChannelMessageSend(m.ChannelID, "```"+string(stdout)+"```")
	if err != nil {
		return nil, fmt.Errorf("Failed to send message to the channal %q: %v", m.ChannelID, err)
	}

	return retMsg, nil
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

	retMsg, err := s.ChannelMessageSend(m.ChannelID, "```"+string(stdout[:1994])+"```")
	if err != nil {
		return nil, fmt.Errorf("Failed to send message to the channal %q: %v", m.ChannelID, err)
	}
	return retMsg, nil
}

// pingCommand is a basic command which will responds "Pong!" to any ping.
func pingCommand(ctx context.Context, s *discordgo.Session, m *discordgo.MessageCreate, _ []string) (*discordgo.Message, error) {
	l, ok := logging.FromContext(ctx)
	retMsg, err := s.ChannelMessageSend(m.ChannelID, "Pong!")
	if err != nil {
		return nil, fmt.Errorf("Failed to send message to the channal %q: %v", m.ChannelID, err)
	}
	if ok {
		l.Infof("Responding 'Pong!' to ping command")
	}
	return retMsg, nil
}

// aliasCommand sets string => string shortcut that can be called later to print a value
func aliasCommand(ctx context.Context, s *discordgo.Session, m *discordgo.MessageCreate, c []string) (*discordgo.Message, error) {
	l, loggerOk := logging.FromContext(ctx)

	switch {
	case len(c) == 1:
		p := dgwidgets.NewPaginator(s, m.ChannelID)
		p.Add(&discordgo.MessageEmbed{
			Color: 0,

			Fields: []*discordgo.MessageEmbedField{
				{
					Name: "Aliases",
					Value: func() string {
						var out []string
						for alias := range aliasMap {
							out = append(out, "**"+alias+"**")
						}
						return strings.Join(out, "\n")
					}(),
				},
			},
		})

		for aliasName, alias := range aliasMap {
			embed := &discordgo.MessageEmbed{
				Title:       aliasName,
				Description: alias.help,
			}

			resp, err := http.Head(alias.help)
			if err != nil {
				p.Add(embed)
				continue
			} else {
				content := strings.TrimPrefix(resp.Header.Get("Content-Type"), "image/")
				if content == "gif" || content == "jpeg" || content == "png" {
					embed.Image = &discordgo.MessageEmbedImage{
						URL: alias.help,
					}
				}
			}
			defer resp.Body.Close()

			p.Add(embed)
		}

		p.SetPageFooters()
		p.Loop = true
		p.ColourWhenDone = 0xFF0000
		p.DeleteReactionsWhenDone = true
		p.Spawn()
		p.Widget.Timeout = time.Minute * 5

		return nil, nil
	case len(c) == 2:
		retMsg, err := s.ChannelMessageSend(m.ChannelID, "Too few arguments supplied. Refer to !help for usage.")
		if err != nil {
			return nil, fmt.Errorf("Failed to send message to the channal %q: %v", m.ChannelID, err)
		}
		return retMsg, errors.New("Too few arguments supplied for set command")
	}

	// Ensure user has permission to use this command
	if !IsAllowed(ctx, s, m.Author.ID, "alias") {
		retMsg, err := s.ChannelMessageSend(m.ChannelID, "You do not have permissions to use this command.")
		if err != nil {
			return nil, fmt.Errorf("Failed to send message to the channal %q: %v", m.ChannelID, err)
		}
		if loggerOk {
			l.Infof("%q is not allowed to execute the alias command", m.Author)
		}
		return retMsg, nil
	}

	_, ok1 := aliasMap[c[1]]
	if _, ok := commMap[c[1]]; ok1 || ok {
		// If key already exists
		retMsg, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("%s is a registered command/alias and cannot be set as an alias.", c[1]))
		if err != nil {
			return nil, fmt.Errorf("Failed to send message to the channal %q: %v", m.ChannelID, err)
		}
		if loggerOk {
			l.Infof("%s attempted to overwrite command %s with %s", m.Author, c[1], strings.Join(c[2:], " "))
		}
		return retMsg, nil
	}

	// If key does not exist (or who's function is printShortcut)
	aliasMap[c[1]] = &command{
		help: strings.Join(c[2:], " "),
		exec: printShortcut,
	}

	savedAliases[c[1]] = strings.Join(c[2:], " ")
	if err := WriteToStorage("./storage/aliases.json", savedAliases); err != nil {
		l.Errorf("Error writing alias to file")
		retMsg, err := s.ChannelMessageSend(m.ChannelID, "Error writing alias to file.")
		if err != nil {
			return nil, fmt.Errorf("Failed to send message to the channal %q: %v", m.ChannelID, err)
		}
		return retMsg, nil
	}

	if loggerOk {
		l.Infof("%s has set an alias for %s => %s", m.Author, c[1], strings.Join(c[2:], " "))
	}

	retMsg, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("%s has set an alias for %s => %s", m.Author, c[1], strings.Join(c[2:], " ")))
	if err != nil {
		return nil, fmt.Errorf("Failed to send message to the channal %q: %v", m.ChannelID, err)
	}

	return retMsg, nil
}

// showHelpCommand lists all of the commands available and explains what they do.
func showHelpCommand(ctx context.Context, s *discordgo.Session, m *discordgo.MessageCreate, msg []string) (*discordgo.Message, error) {
	if l, ok := logging.FromContext(ctx); ok {
		l.Infof("Responding to help command")
	}

	if len(msg) == 2 {
		if c, ok := commMap[msg[1]]; ok {
			retMsg, err := s.ChannelMessageSendEmbed(m.ChannelID, &discordgo.MessageEmbed{
				Color: 0,

				Fields: []*discordgo.MessageEmbedField{
					{Name: msg[1], Value: c.help},
				},
			})
			if err != nil {
				return nil, fmt.Errorf("Failed to send message to the channal %q: %v", m.ChannelID, err)
			}
			return retMsg, nil
		}

		retMsg, err := s.ChannelMessageSend(m.ChannelID, "Command not found.")
		if err != nil {
			return nil, fmt.Errorf("Failed to send message to the channal %q: %v", m.ChannelID, err)
		}
		return retMsg, nil
	}

	retMsg, err := s.ChannelMessageSendEmbed(m.ChannelID, &discordgo.MessageEmbed{
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
	if err != nil {
		return nil, fmt.Errorf("Failed to send message to the channal %q: %v", m.ChannelID, err)
	}
	return retMsg, nil
}

// Execute parses a msg and executes the command, if it exists.
func Execute(ctx context.Context, s *discordgo.Session, m *discordgo.MessageCreate, msg string) error {
	args := strings.Fields(msg)

	if c, ok := commMap[args[0]]; ok {
		if _, err := c.exec(ctx, s, m, args); err != nil {
			return fmt.Errorf("failed to execute command: %#v", err)
		}
		return nil
	} else if val, ok := aliasMap[args[0]]; ok {
		if _, err := s.ChannelMessageSend(m.ChannelID, val.help); err != nil {
			return fmt.Errorf("Failed to send message to the channal %q: %v", m.ChannelID, err)
		}
		return nil
	}
	return fmt.Errorf("Failed to recognise the command %q", args[0])
}

// printShortcut uses the help text of the command to print the shortcut's value
func printShortcut(ctx context.Context, s *discordgo.Session, m *discordgo.MessageCreate, args []string) (*discordgo.Message, error) {
	return s.ChannelMessageSend(m.ChannelID, commMap[args[0]].help)
}

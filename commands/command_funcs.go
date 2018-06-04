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
	"runtime"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/Necroforger/dgwidgets"
	"github.com/UCCNetworkingSociety/Netsoc-Discord-Bot/config"
	"github.com/UCCNetworkingSociety/Netsoc-Discord-Bot/logging"
	"github.com/bwmarrin/discordgo"
)

// pingCommand is a basic command which will responds "Pong!" to any ping.
func pingCommand(ctx context.Context, _ []string) (string, error) {
	if l, ok := logging.FromContext(ctx); ok {
		l.Infof("Responding 'Pong!' to ping command")
	}
	return "Pong!", nil
}

// minecraftCommand checks the stats of minecraft.netsoc.co for that moment
// data comes from http://minecraft.netsoc.co/standalone/dynmap_NetsocCraft.json
func minecraftCommand(ctx context.Context, _ []string) (string, error) {
	l, loggerOk := logging.FromContext(ctx)
	if loggerOk {
		l.Infof("Responding to minecraft command")
	}

	resp, err := http.Get("http://minecraft.netsoc.co/standalone/dynmap_NetsocCraft.json")
	if err != nil {
		return "", fmt.Errorf("Failed to retrieve data from the Minecraft Server: %s", err)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("Failed to read the response body: %s", err)
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
		return "", fmt.Errorf("Failed to parse response json %q: %s", string(body), err)
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

	if loggerOk {
		l.Infof("Sending minecraft information: %d users active", len(q.Players))
	}
	return "", nil
}

// inspireCommand gets an inspirational quote from forismatic.com
func inspireCommand(ctx context.Context, _ []string) (string, error) {
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
		return "", fmt.Errorf("Failed to get the quote from API: %s", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("Failed to read the response body: %s", err)
	}

	q := &struct {
		QuoteText   string `json:"quoteText"`
		QuoteAuthor string `json:"quoteAuthor"`
	}{}
	if err := json.Unmarshal(body, q); err != nil {
		return "", fmt.Errorf("Failed to parse response json %q: %s", string(body), err)
	}

	if loggerOk {
		l.Infof("Sending quote %q", q.QuoteText)
	}

	return fmt.Sprintf("%q - %s", q.QuoteText, q.QuoteAuthor), nil
}

// showHelpCommand lists all of the commands available and explains what they do.
func showHelpCommand(ctx context.Context, args []string) (*discordgo.MessageEmbed, error) {
	if l, ok := logging.FromContext(ctx); ok {
		l.Infof("Responding to help command")
	}

	if len(args) == 2 {
		if comm, ok := commMap[args[1]]; ok {
			return &discordgo.MessageEmbed{
				Color: 0,

				Fields: []*discordgo.MessageEmbedField{
					{Name: args[1], Value: comm.Help()},
				},
			}, nil
		}

		return nil, fmt.Errorf("Failed to find command %q", args[1])
	}

	return &discordgo.MessageEmbed{
		Color: 0,

		Fields: func() []*discordgo.MessageEmbedField {
			var out []*discordgo.MessageEmbedField

			for name, c := range commMap {
				out = append(out, &discordgo.MessageEmbedField{
					Name:  name,
					Value: c.Help(),
				})
			}

			return out
		}(),
	}, nil
}

// configCommand returns a message with the current configuration of the bot
func configCommand(ctx context.Context, _ []string) (*discordgo.MessageEmbed, error) {
	l, loggerOk := logging.FromContext(ctx)
	if loggerOk {
		l.Infof("Responding to config command", nil)
	}

	return &discordgo.MessageEmbed{
		Color: 0,

		Fields: []*discordgo.MessageEmbedField{
			{Name: "Config:", Value: config.GetConfig().String(), Inline: true},
		},
	}, nil
}

// infoCommand returns a message with the current resource usage of the bot
func infoCommand(ctx context.Context, _ []string) (*discordgo.MessageEmbed, error) {
	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)

	return &discordgo.MessageEmbed{
		Color: 0,

		Fields: []*discordgo.MessageEmbedField{
			{Name: "Memory Usage:", Value: "```" + strconv.Itoa(int(mem.Alloc/1024/1024)) + "MB" + "```", Inline: true},
			{Name: "Goroutines:", Value: "```" + strconv.Itoa(runtime.NumGoroutine()) + "```", Inline: true},
			{Name: "Go Version:", Value: "```" + runtime.Version() + "```", Inline: true},
			{Name: "Usable Cores:", Value: "```" + strconv.Itoa(runtime.NumCPU()) + "```", Inline: true},
		},
	}, nil
}

// aliasCommand sets string => string shortcut that can be called later to print a value
func aliasCommand(ctx context.Context, args []string) (string, error) {
	l, loggerOk := logging.FromContext(ctx)

	switch {
	case len(args) == 1:
		// shop a paginator of all aliases
		p := dgwidgets.NewPaginator(ctx.Value("Session").(*discordgo.Session), ctx.Value("ChannelID").(string))
		p.Add(&discordgo.MessageEmbed{
			Color: 0,

			Fields: []*discordgo.MessageEmbedField{
				{
					Name: "Aliases",
					Value: func() string {
						var sortedAliases []string
						for a := range aliasMap {
							sortedAliases = append(sortedAliases, fmt.Sprintf("**%s**", a))
						}
						sort.Strings(sortedAliases)
						for i := range sortedAliases {
							sortedAliases[i] = fmt.Sprintf("%d) %s", i+2, sortedAliases[i])
						}
						return strings.Join(sortedAliases, "\n")
					}(),
				},
			},
		})

		for aliasName, aliasValue := range aliasMap {
			embed := &discordgo.MessageEmbed{
				Title:       aliasName,
				Description: aliasValue.Help(),
			}

			resp, err := http.Head(aliasValue.Help())
			if err != nil {
				p.Add(embed)
				continue
			}
			defer resp.Body.Close()
			content := strings.TrimPrefix(resp.Header.Get("Content-Type"), "image/")
			if content == "gif" || content == "jpeg" || content == "png" {
				embed.Image = &discordgo.MessageEmbedImage{
					URL: aliasValue.Help(),
				}
			}
			p.Add(embed)
		}

		p.SetPageFooters()
		p.Loop = true
		p.ColourWhenDone = 0xFF0000
		p.DeleteReactionsWhenDone = true
		p.Spawn()
		p.Widget.Timeout = time.Minute * 5

		return "", nil
	case len(args) == 2:
		// no version with 2 args
		return "", errors.New("Too few arguments supplied. Refer to !help for usage")
	default:
		// has at least 3 args, so setting a new alias
		var (
			_, aliasExists   = aliasMap[args[1]]
			_, commandExists = commMap[args[1]]
		)
		if aliasExists || commandExists {
			return "", fmt.Errorf("%q is a registered command/alias and cannot be set as an alias", args[1])
		}

		aliasMap[args[1]] = &textCommand{
			helpText: strings.Join(args[2:], " "),
			command: func(_ context.Context, args []string) (string, error) {
				return strings.Join(args[2:], " "), nil
			},
		}

		savedAliases[args[1]] = strings.Join(args[2:], " ")
		if err := WriteToStorage("./storage/aliases.json", savedAliases); err != nil {

			return "", errors.New("Error writing alias to file")
		}

		if loggerOk {
			l.Infof("Set an alias for %s => %s", args[1], strings.Join(args[2:], " "))
		}

		return fmt.Sprintf("Set an alias for %s => %s", args[1], strings.Join(args[2:], " ")), nil
	}
}

// unAliasCommand takes an existing alias and removes it from the alias map
func unAliasCommand(ctx context.Context, msg []string) (string, error) {
	l, loggerOk := logging.FromContext(ctx)
	if loggerOk {
		l.Infof("Responding to unalias command")
	}
	if len(msg) != 2 {
		return "", errors.New("Please indicate an alias to unset")
	}
	toRemove := msg[1]
	if _, ok := savedAliases[toRemove]; !ok {
		return "", fmt.Errorf("Alias %q doesn't exist", toRemove)
	}

	delete(savedAliases, toRemove)
	delete(aliasMap, toRemove)

	if err := WriteToStorage("./storage/aliases.json", savedAliases); err != nil {
		return "", fmt.Errorf("Failed to removing alias from file: %s", err)
	}

	return fmt.Sprintf("Removing alias %q", toRemove), nil
}

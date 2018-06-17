package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/url"
	"runtime"
	"strconv"

	"github.com/UCCNetworkingSociety/Netsoc-Discord-Bot/config"
	"github.com/bwmarrin/discordgo"
	"github.com/golang/glog"
)

var (
	minecraftAPIURL           = "http://minecraft.netsoc.co/standalone/dynmap_NetsocCraft.json"
	inspirationalQuotesAPIURL = "http://api.forismatic.com/api/1.0/"
)

// pingCommand is a basic command which will responds "Pong!" to any ping.
func pingCommand(ctx context.Context, _ []string) (string, error) {
	glog.Infof("Responding 'Pong!' to ping command")
	return "Pong!", nil
}

// minecraftCommand checks the stats of minecraft.netsoc.co for that moment
// data comes from http://minecraft.netsoc.co/standalone/dynmap_NetsocCraft.json
func minecraftCommand(ctx context.Context, _ []string) (string, error) {
	glog.Infof("Responding to minecraft command")

	resp, err := http.Get(minecraftAPIURL)
	if err != nil {
		return "", fmt.Errorf("Failed to retrieve data from the Minecraft Server: %s", err)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("Failed to read the response body: %s", err)
	}

	// Unmarshal the data and get all the player lists I guess
	q := &struct {
		Players []struct {
			Name string `json:"name"`
		} `json:"players"`
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

	glog.Infof("Sending minecraft information: %d users active", len(q.Players))
	return msg, nil
}

// inspireCommand gets an inspirational quote from forismatic.com
func inspireCommand(ctx context.Context, _ []string) (string, error) {
	glog.Infof("Responding to inspire command")

	resp, err := http.PostForm(inspirationalQuotesAPIURL,
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

	glog.Infof("Sending quote %q", q.QuoteText)
	return fmt.Sprintf("%q - %s", q.QuoteText, q.QuoteAuthor), nil
}

// showHelpCommand lists all of the commands available and explains what they do.
func showHelpCommand(ctx context.Context, args []string) (*discordgo.MessageEmbed, error) {
	glog.Infof("Responding to help command")

	if len(args) == 2 {
		// getting help about a specific command
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

	// getting help about all commands
	var allCommands []*discordgo.MessageEmbedField
	for name, c := range commMap {
		if _, isAlias := aliasMap[name]; isAlias {
			continue
		}
		allCommands = append(allCommands, &discordgo.MessageEmbedField{
			Name:  name,
			Value: c.Help(),
		})
	}
	return &discordgo.MessageEmbed{
		Color:  0,
		Fields: allCommands,
	}, nil
}

// configCommand returns a message with the current configuration of the bot
func configCommand(ctx context.Context, _ []string) (*discordgo.MessageEmbed, error) {
	glog.Infof("Responding to config command")

	return &discordgo.MessageEmbed{
		Color: 0,

		Fields: []*discordgo.MessageEmbedField{
			{Name: "Config:", Value: config.GetConfig().String(), Inline: true},
		},
	}, nil
}

// infoCommand returns a message with the current resource usage of the bot
func infoCommand(ctx context.Context, _ []string) (*discordgo.MessageEmbed, error) {
	glog.Infof("Responding to info command")

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

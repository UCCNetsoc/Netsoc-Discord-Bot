package commands

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/Necroforger/dgwidgets"
	"github.com/UCCNetworkingSociety/Netsoc-Discord-Bot/logging"
	"github.com/bwmarrin/discordgo"
)

var (
	// note that this is relative to the $(pwd) as opposed to the `commands` directory
	aliasStorageFilepath = "aliases.json"
	aliasMap             = make(map[string]*alias)
)

type aliasKind int

const (
	// if adding a new kind, add it *after* the last one in the list or you will
	// break the existing `aliases.json` file
	kindIMAGE aliasKind = iota
	kindOTHER
)

// alias represents the alias value and meta information used to display the alias
type alias struct {
	Value string    `json: "Value"`
	Kind  aliasKind `json: "Kind"`
}

// loadFromStorage opens the alias file at `aliasStorageFilepath` and loads
// the aliases into `aliasMap`.
func loadFromStorage() error {
	if _, err := os.Stat(aliasStorageFilepath); os.IsNotExist(err) {
		if err := ioutil.WriteFile(aliasStorageFilepath, []byte("{}"), 0744); err != nil {
			return fmt.Errorf("Failed to create alias file: %s", err)
		}
	}

	file, err := ioutil.ReadFile(aliasStorageFilepath)
	if err != nil {
		return fmt.Errorf("failed to read aliases file: %s", err)
	}

	if err := json.Unmarshal(file, &aliasMap); err != nil {
		return fmt.Errorf("failed to unmarshal alias JSON: %s", err)
	}
	return nil
}

// writeToStorage overwrites the aliases held within `aliasMap` to the storage file.
// Note that if an alias was removed from `aliasMap`, the removed alias will also
// be removed from the storage file.
func writeToStorage() error {
	aliasBytes, err := json.Marshal(aliasMap)
	if err != nil {
		return fmt.Errorf("Failed to marshal JSON: %s", err)
	}
	if err := ioutil.WriteFile(aliasStorageFilepath, aliasBytes, 0744); err != nil {
		return fmt.Errorf("Failed to write alias file: %s", err)
	}
	return nil
}

// withAliasCommands takes a map of command names to the command itself and
// adds in a command where the command name is the alias key and the command
// itself is a command which returns the alias value.
func withAliasCommands(commMap map[string]Command) (map[string]Command, error) {
	if err := loadFromStorage(); err != nil {
		return nil, fmt.Errorf("failed to load aliases from storage: %s", err)
	}

	for aliasName, info := range aliasMap {
		info := info
		commMap[aliasName] = &textCommand{
			helpText: info.Value,
			command: func(_ context.Context, _ []string) (string, error) {
				return info.Value, nil
			},
		}
	}
	return commMap, nil
}

// aliasCommand sets string => string shortcut that can be called later to print a value
func aliasCommand(ctx context.Context, args []string) (string, error) {
	l, loggerOk := logging.FromContext(ctx)
	if loggerOk {
		l.Infof("Responding to alias command")
	}

	if len(args) == 2 {
		return "", errors.New("Too few arguments supplied. Refer to !help for usage")
	}

	// one argument will respond with a paginator which details all the aliases
	// in existence and allows the discord user to scroll through them.
	if len(args) == 1 {
		// show a paginator of all aliases
		p := dgwidgets.NewPaginator(
			ctx.Value(messageContextValue("Session")).(*discordgo.Session),
			ctx.Value(messageContextValue("ChannelID")).(string))

		var sortedAliases []string
		for a := range aliasMap {
			sortedAliases = append(sortedAliases, a)
		}
		sort.Strings(sortedAliases)

		var aliasListing []string
		for i, a := range sortedAliases {
			aliasListing = append(aliasListing, fmt.Sprintf("%d) **%s**", i+2, a))
		}

		p.Add(&discordgo.MessageEmbed{
			Color: 0,

			Fields: []*discordgo.MessageEmbedField{
				{
					Name:  "Aliases",
					Value: strings.Join(aliasListing, "\n"),
				},
			},
		})

		for _, aliasName := range sortedAliases {
			info := aliasMap[aliasName]
			embed := &discordgo.MessageEmbed{
				Title:       aliasName,
				Description: info.Value,
			}

			// if the link is to an image, we add the image itself to the embed
			kind := info.Kind
			if kind == kindIMAGE {
				embed.Image = &discordgo.MessageEmbedImage{
					URL: info.Value,
				}
			}
			p.Add(embed)
		}

		p.SetPageFooters()
		p.Loop = true
		p.ColourWhenDone = 0xFF0000
		p.DeleteReactionsWhenDone = true
		p.Widget.Timeout = 2 * time.Minute
		if err := p.Spawn(); err != nil {
			return "", fmt.Errorf("Failed to spawn paginator: %s", err)
		}
		return "", nil
	}

	// if we have 3 or more arguments, it means we are setting a new alias. The
	// first argument is the 'alias' keyword, then the alias-key and then the alias-value.
	var (
		_, aliasExists   = aliasMap[args[1]]
		_, commandExists = commMap[args[1]]
	)
	if aliasExists || commandExists {
		return "", fmt.Errorf("%q is a registered command/alias and cannot be set as an alias", args[1])
	}

	newAliasValue := strings.Join(args[2:], " ")
	newAlias := &alias{Value: newAliasValue}

	resp, err := http.Head(newAliasValue)
	if err == nil {
		// this is a URL. We now check if it is an image
		defer resp.Body.Close()
		contentType := resp.Header.Get("Content-Type")
		switch {
		case strings.Contains(contentType, "image"):
			newAlias.Kind = kindIMAGE
		default:
			newAlias.Kind = kindOTHER
		}
	} else {
		newAlias.Kind = kindOTHER
	}

	// record the new alias in the command map and save it to storage
	aliasMap[args[1]] = newAlias
	if err := writeToStorage(); err != nil {
		return "", fmt.Errorf("Failed to write new alias to storage file: %s", err)
	}

	// reload the command map in commands.go to encorporate the new alias command
	commMap, err = withAliasCommands(commMap)
	if err != nil {
		return "", fmt.Errorf("Failed to reload the command map with the new alias: %s", err)
	}

	if loggerOk {
		l.Infof("Set an alias for %s => %s", args[1], newAliasValue)
	}

	return fmt.Sprintf("Set an alias for %s => %s", args[1], newAliasValue), nil

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
	if _, ok := aliasMap[toRemove]; !ok {
		return "", fmt.Errorf("Alias %q doesn't exist", toRemove)
	}
	delete(aliasMap, toRemove)
	delete(commMap, toRemove)
	if err := writeToStorage(); err != nil {
		return "", fmt.Errorf("Failed to write new alias storage file: %s", err)
	}

	if loggerOk {
		l.Infof("Removed alias %q", toRemove)
	}
	return fmt.Sprintf("Removing alias %q", toRemove), nil
}

package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
)

var (
	// Set defaults for the config to ensure it works
	// or warns even if a particular setting is missing
	conf = &Config{
		Prefix:        "!",
		Token:         "warn",
		HelpChannelID: "warn",
		BotHostName:   "0.0.0.0:4201",
		GuildID:       "291573897730588684",
		SysAdminTag:   "<@&318907623476822016>",
		Permissions: map[string][]string{
			"alias": []string{
				"Chairperson",
				"Equipments Officer",
				"Events Officer",
				"Finance Officer",
				"HLM",
				"PRO",
				"Secretary",
				"SysAdmin",
			},
		},
	}
)

// Config represetns the bot configuration loaded from the JSON
// file "./config.json".
type Config struct {
	// Prefix is the string that will prefix all commands
	// which this not will listen for.
	Prefix string `json:"prefix"`
	// Token is the Discord bot user token.
	Token string `json:"token"`
	// HelpChannelID is the channel ID to which help messages from
	// netsoc-admin will be sent.
	HelpChannelID string `json:"helpChannelID"`
	// BotHostName is the address which the bot can be reached at
	// over the internet. This is used by netsocadmin to reach the
	// '/help' endpoint.
	BotHostName string `json:"botHostName"`
	// SysAdminTag is the tag which, when included in a discord message,
	// will result in a notification being sent to all SysAdmins so they
	// can be notified of the help message.
	GuildID     string `json:"guildID"`
	SysAdminTag string `json:"sysAdminTag"`

	// Defines which roles can execute commands (if applicable)
	Permissions map[string][]string `json:"permissions"`
}

// LoadConfig loads the configuration information found in ./config.json
func LoadConfig() error {
	file, err := ioutil.ReadFile("config.json")
	if err != nil {
		return fmt.Errorf("failed to read configuration file: %v", err)
	}

	if len(file) < 1 {
		return errors.New("Configuration file 'config.json' was empty")
	}

	conf = &Config{}
	if err := json.Unmarshal(file, conf); err != nil {
		return fmt.Errorf("failed to unmarshal configuration JSON: %s", err)
	}

	return nil
}

// GetConfig gets the loaded configuration
func GetConfig() *Config {
	return conf
}

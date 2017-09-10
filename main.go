package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"./commands"
	"./logging"

	"github.com/bwmarrin/discordgo"
)

var (
	conf *config
	l    *logging.Logger
	dg   *discordgo.Session
)

// config represetns the bot configuration loaded from the JSON
// file "./config.json".
type config struct {
	// Prefix is the string that will prefix all commands
	// which this not will listen for.
	Prefix string `json:"prefix"`
	// Token is the Discord bot user token.
	Token string `json:"token"`
	// HelpChannelId is the channel ID to which help messages from
	// netsoc-admin will be sent.
	HelpChannelId string `json:"helpChannelId"`
}

// helpBody represents the help message which is sent from netsoc-admin.
type helpBody struct {
	User    string `json:"user"`
	Email   string `json:"email"`
	Subject string `json:"subject"`
	Message string `json:"message"`
}

func main() {
	if err := loadConfig(); err != nil {
		log.Fatalf("Failed to load configuration JSON: %s", err)
	}

	var err error
	l, err = logging.New()
	if err != nil {
		log.Fatalf("Failed to create bot's logger: %s", err)
	}
	defer l.Close()

	l.Infof("Starting bot..")
	dg, err = discordgo.New("Bot " + conf.Token)
	if err != nil {
		l.Errorf("Failed to create Discord session: %s", err)
		return
	}
	dg.AddHandler(messageCreate)

	if err := dg.Open(); err != nil {
		l.Errorf("Failed to open websocket connection: %s", err)
		return
	}
	defer dg.Close()

	if err := dg.UpdateStatus(0, conf.Prefix+commands.HelpCommand); err != nil {
		l.Errorf("Failed to set bot's status")
		return
	}
	l.Infof("Bot succesfully started")

	http.HandleFunc("/help", help)
	if err := http.ListenAndServe(":4201", nil); err != nil {
		l.Errorf("Failed to serve HTTP: %s", err)
	}
}

func help(w http.ResponseWriter, r *http.Request) {
	resp := &helpBody{}

	bytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		l.Errorf("Failed to read request body: %s", err)
		dg.ChannelMessageSend(conf.HelpChannelId, "help request error, check logs")
		return
	}
	r.Body.Close()
	err = json.Unmarshal(bytes, resp)
	if err != nil {
		l.Errorf("Failed to unmarshal request JSON %q: %s", bytes, err)
		dg.ChannelMessageSend(conf.HelpChannelId, "help request error, check logs")
		return
	}
	msg := fmt.Sprintf("```From: %s\nEmail: %s\n\nSubject: %s\n\n%s```", resp.User, resp.Email, resp.Subject, resp.Message)
	dg.ChannelMessageSend(conf.HelpChannelId, msg)
}

// messageCreate is an event handler which is called whenever a new message
// is created in the Discord server.
func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.Bot || !strings.HasPrefix(m.Content, conf.Prefix) ||
		strings.TrimPrefix(m.Content, conf.Prefix) == "" {
		return
	}
	c := strings.TrimPrefix(m.Content, conf.Prefix)
	l.Infof("Received command: %q", c)
	if err := commands.Execute(s, m, c); err != nil {
		l.Errorf("Failed to execute command %q: %s", c, err)
	}
}

// loadConfig loads teh configuration information found in ./config.json
func loadConfig() error {
	file, err := ioutil.ReadFile("config.json")
	if err != nil {
		return fmt.Errorf("failed to read configuration file: ", err)
	}

	if len(file) < 1 {
		return errors.New("Configuration file 'config.json' was empty")
	}

	conf = &config{}
	if err := json.Unmarshal(file, conf); err != nil {
		return fmt.Errorf("failed to unmarshal configuration JSON: %s", err)
	}

	return nil
}

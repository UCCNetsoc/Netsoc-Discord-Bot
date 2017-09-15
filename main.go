package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	
	"github.com/UCCNetworkingSociety/Netsoc-Discord-Bot/commands"
	"github.com/UCCNetworkingSociety/Netsoc-Discord-Bot/config"
	"github.com/UCCNetworkingSociety/Netsoc-Discord-Bot/logging"
	

/* 	"./commands"
	"./config"
	"./logging"
 */
	"github.com/bwmarrin/discordgo"
)

var (
	conf *config.Config
	l    *logging.Logger
	dg   *discordgo.Session
)

// helpBody represents the help message which is sent from netsoc-admin.
type helpBody struct {
	User    string `json:"user"`
	Email   string `json:"email"`
	Subject string `json:"subject"`
	Message string `json:"message"`
}

func main() {
	if err := config.LoadConfig(); err != nil {
		log.Fatalf("Failed to load configuration JSON: %s", err)
	}
	conf = config.GetConfig()

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
	l.Infof("Bot successfully started")

	http.HandleFunc("/help", help)
	if err := http.ListenAndServe(conf.BotHostName, nil); err != nil {
		l.Errorf("Failed to serve HTTP: %s", err)
	}
}

func help(w http.ResponseWriter, r *http.Request) {
	resp := &helpBody{}

	bytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		l.Errorf("Failed to read request body: %s", err)
		dg.ChannelMessageSend(conf.HelpChannelID, "help request error, check logs")
		return
	}
	defer r.Body.Close()

	err = json.Unmarshal(bytes, resp)
	if err != nil {
		l.Errorf("Failed to unmarshal request JSON %q: %s", bytes, err)
		dg.ChannelMessageSend(conf.HelpChannelID, "help request error, check logs")
		return
	}
	
	msg := fmt.Sprintf("%s Help pls\n\n```From: %s\nEmail: %s\n\nSubject: %s\n\n%s```", conf.SysAdminTag, resp.User, resp.Email, resp.Subject, resp.Message)
	dg.ChannelMessageSend(conf.HelpChannelID, msg)
}

// messageCreate is an event handler which is called whenever a new message
// is created in the Discord server.
func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.Bot || !strings.HasPrefix(m.Content, conf.Prefix) || strings.TrimPrefix(m.Content, conf.Prefix) == "" {
		return
	}

	c := strings.TrimPrefix(m.Content, conf.Prefix)

	l.Infof("Received command %q from %q", c, m.Author)

	ctx := logging.NewContext(context.Background(), l)

	if err := commands.Execute(ctx, s, m, c); err != nil {
		l.Errorf("Failed to execute command %q: %s", c, err)
	}
}

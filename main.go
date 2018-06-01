package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/UCCNetworkingSociety/Netsoc-Discord-Bot/commands"
	"github.com/UCCNetworkingSociety/Netsoc-Discord-Bot/config"
	"github.com/UCCNetworkingSociety/Netsoc-Discord-Bot/logging"

	"github.com/bwmarrin/discordgo"
	"github.com/fsnotify/fsnotify"
)

var (
	conf *config.Config
	l    *logging.Logger
	dg   *discordgo.Session
)

// handlerWithError wraps a regular http handler function. It cleans
// up the logging and error response writing.
type handlerWithError func(w http.ResponseWriter, r *http.Request) error

// ServeHTTP implements the http.Handler interface
func (h handlerWithError) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	uri := r.URL.RequestURI()
	l.Infof("Received request for %q", uri)
	if err := h(w, r); err != nil {
		http.Error(w, err.Error(), 500)
		l.Errorf("Failed to handle request for %q: %v", uri, err)
	}
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
		l.Errorf("Failed to set bot's status: %v", err)
		return
	}
	l.Infof("Bot successfully started")

	l.Infof("Watching config.json")
	watcherDone, err := watchConfig()
	if err != nil {
		l.Errorf("Failed to create configuration file watcher: %v", err)
		return
	}

	l.Infof("Serving http server on %s", conf.BotHostName)
	http.Handle("/help", handlerWithError(help))
	http.Handle("/alert", handlerWithError(alertHandler))
	if err := http.ListenAndServe(conf.BotHostName, nil); err != nil {
		watcherDone <- struct{}{}
		l.Errorf("Failed to serve HTTP: %s", err)
	}
}

// help sends a help message to the help discord channel
func help(w http.ResponseWriter, r *http.Request) error {
	defer r.Body.Close()

	var resp struct {
		User    string `json:"user"`
		Email   string `json:"email"`
		Subject string `json:"subject"`
		Message string `json:"message"`
	}
	if err := json.NewDecoder(r.Body).Decode(&resp); err != nil {
		err = fmt.Errorf("Failed to unmarshal request JSON: %s", err)
		if _, dgErr := dg.ChannelMessageSend(conf.HelpChannelID, fmt.Sprintf("help request error: %v", err)); dgErr != nil {
			return fmt.Errorf("failed to send failure notice to discord (%v): %v", err, dgErr)
		}
		return err
	}

	msg := fmt.Sprintf("%s Help pls\n\n```From: %s\nEmail: %s\n\nSubject: %s\n\n%s```", conf.SysAdminTag, resp.User, resp.Email, resp.Subject, resp.Message)
	if _, err := dg.ChannelMessageSend(conf.HelpChannelID, msg); err != nil {
		return fmt.Errorf("Failed to send help request to discord: %v", err)
	}
	return nil
}

// alertHandler relays a prometheus alert to the alerting channel
func alertHandler(w http.ResponseWriter, r *http.Request) error {
	defer r.Body.Close()

	var resp struct {
		Status string `json:"status"`
		Alerts []struct {
			Annotations map[string]string `json:"annotations"`
		} `json:"alerts"`
	}

	if err := json.NewDecoder(r.Body).Decode(&resp); err != nil {
		err = fmt.Errorf("Failed to unmarshal request JSON: %s", err)
		if _, dgErr := dg.ChannelMessageSend(conf.AlertsChannelID, fmt.Sprintf("alerts request error: %v", err)); dgErr != nil {
			return fmt.Errorf("failed to send failure notice to discord (%v): %v", err, dgErr)
		}
		return err
	}

	msg := fmt.Sprintf("%s Alerts are %s:", conf.SysAdminTag, resp.Status)
	for _, a := range resp.Alerts {
		msg += fmt.Sprintf("\n - %s", a.Annotations["summary"])
	}
	if _, err := dg.ChannelMessageSend(conf.AlertsChannelID, msg); err != nil {
		return fmt.Errorf("failed to send alert to discord: %v", err)
	}
	return nil
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

// watchConfig monitors for changes in the config JSON file and reloads the
// config values if there is
func watchConfig() (chan struct{}, error) {
	// creates a new file watcher
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		l.Errorf("Failed to create filesystem watcher for config file: %v", err)
	}

	done := make(chan struct{})

	go func() {
		for {
			select {
			// watch for events
			case event := <-watcher.Events:
				l.Infof("Watcher event recieved: %s", event)
				config.LoadConfig()

			// watch for errors
			case err := <-watcher.Errors:
				l.Errorf("Config file watcher error: %v", err)

			case <-done:
				l.Infof("Watcher shutting down")
				return
			}
		}
	}()

	// out of the box fsnotify can watch a single file, or a single directory
	if err := watcher.Add("./config.json"); err != nil {
		l.Errorf("Failed to add config file to watcher: %v", err)
	}

	return done, nil
}

package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"strings"

	"github.com/UCCNetworkingSociety/Netsoc-Discord-Bot/commands"
	"github.com/UCCNetworkingSociety/Netsoc-Discord-Bot/config"
	"github.com/golang/glog"

	"github.com/bwmarrin/discordgo"
	"github.com/fsnotify/fsnotify"
)

var (
	conf *config.Config
	dg   *discordgo.Session
)

// handlerWithError wraps a regular http handler function. It cleans
// up the logging and error response writing.
type handlerWithError func(w http.ResponseWriter, r *http.Request) error

// ServeHTTP implements the http.Handler interface
func (h handlerWithError) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	uri := r.URL.RequestURI()
	glog.Infof("Received request for %q", uri)
	if err := h(w, r); err != nil {
		http.Error(w, err.Error(), 500)
		glog.Errorf("Failed to handle request for %q: %s", uri, err)
	}
}

func main() {
	// this must be done for the glog library
	flag.Parse()
	defer glog.Flush()

	var err error
	if err = config.LoadConfig(); err != nil {
		glog.Fatalf("Failed to load configuration JSON: %s", err)
	}
	conf = config.GetConfig()

	glog.Infof("Starting bot..")
	dg, err = discordgo.New("Bot " + conf.Token)
	if err != nil {
		glog.Fatalf("Failed to create Discord session: %s", err)
	}

	dg.AddHandler(messageCreate)

	if err := dg.Open(); err != nil {
		glog.Fatalf("Failed to open websocket connection: %s", err)
	}
	defer dg.Close()

	if err := dg.UpdateStatus(0, conf.Prefix+commands.HelpCommand); err != nil {
		glog.Fatalf("Failed to set bot's status: %s", err)
	}
	glog.Infof("Bot successfully started")

	glog.Infof("Watching config.json")
	watcherDone, err := watchConfig()
	if err != nil {
		glog.Fatalf("Failed to create configuration file watcher: %s", err)
	}
	defer func() {
		watcherDone <- struct{}{}
	}()

	glog.Infof("Serving http server on %s", conf.BotHostName)
	close := make(chan struct{})
	<-close
}

// messageCreate is an event handler which is called whenever a new message
// is created in the Discord server.
func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.Bot || !strings.HasPrefix(m.Content, conf.Prefix) || strings.TrimPrefix(m.Content, conf.Prefix) == "" {
		return
	}
	c := strings.TrimPrefix(m.Content, conf.Prefix)
	glog.Infof("Received command %q from %q", c, m.Author)
	if err := commands.Execute(context.Background(), s, m, c); err != nil {
		glog.Errorf("Failed to execute command %q: %s", c, err)
	}
}

// watchConfig monitors for changes in the config JSON file and reloads the
// config values if there is
func watchConfig() (chan struct{}, error) {
	// creates a new file watcher
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("Failed to create filesystem watcher for config file: %s", err)
	}

	done := make(chan struct{})

	go func() {
		for {
			select {
			// watch for events
			case event := <-watcher.Events:
				glog.Infof("Watcher event recieved: %s", event)
				config.LoadConfig()

			// watch for errors
			case err := <-watcher.Errors:
				glog.Errorf("Config file watcher error: %s", err)

			case <-done:
				glog.Infof("Watcher shutting down")
				return
			}
		}
	}()

	// out of the box fsnotify can watch a single file, or a single directory
	if err := watcher.Add("./config.json"); err != nil {
		glog.Errorf("Failed to add config file to watcher: %s", err)
	}

	return done, nil
}

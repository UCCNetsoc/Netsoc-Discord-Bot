package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"

	"./commands"

	"github.com/bwmarrin/discordgo"
)

var (
	conf     *config
	logF     *os.File
	infoLog  *log.Logger
	errorLog *log.Logger
	dg       = &discordgo.Session{}
)

// config represetns the bot configuration loaded from the JSON
// file "./config.json".
type config struct {
	// game is the Game to which this bit pertains.
	game string `json:"game"`
	// prefix is the string that will prefix all commands
	// which this not will listen for.
	prefix string `json:"prefix"`
	// token is the Discord bot user token.
	token string `json:"token"`
	// inDev is true if we are in a developemnt environment.
	inDev bool `json:"indev"`
}

// helpBody represents the help message which is sent from netsoc-admin.
type helpBody struct {
	user    string `json:"user"`
	email   string `json:"email"`
	subject string `json:"subject"`
	message string `json:"message"`
}

func main() {
	if err := loadConfig(); err != nil {
		log.Fatalf("Failed to load configuration JSON: %s", err)
	}

	loadLog()
	defer logF.Close()

	log.SetOutput(logF)

	infoLog = log.New(logF, "INFO:  ", log.Ldate|log.Ltime)
	errorLog = log.New(logF, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)
	if conf.inDev {
		errorLog = log.New(os.Stdout, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)
	}

	var err error
	// Create a new Discord session using the provided bot token.
	dg, err = discordgo.New("Bot " + conf.token)
	if err != nil {
		errorLog.Println("Error creating Discord session,", err)
		return
	}

	// Open a websocket connection to Discord and begin listening.
	err = dg.Open()
	if err != nil {
		errorLog.Println("Error opening connection,", err)
		return
	}
	defer dg.Close()

	dg.AddHandler(messageCreate)

	setInitialGame(dg)

	fmt.Fprintln(logF, "")
	infoLog.Println(`/*********BOT RESTARTED*********\`)

	// Wait here until CTRL-C or other term signal is received.
	fmt.Println("Bot is now running. Press CTRL-C to exit.")

	http.HandleFunc("/help", help)

	errorLog.Fatalln(http.ListenAndServe(":4201", nil))
}

func help(w http.ResponseWriter, r *http.Request) {
	resp := helpBody{}

	bytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		dg.ChannelMessageSend("354748497683283979", err.Error())
		return
	}

	err = json.Unmarshal(bytes, &resp)
	if err != nil {
		dg.ChannelMessageSend("354748497683283979", err.Error())
		return
	}

	dg.ChannelMessageSend("354748497683283979", fmt.Sprintf("```From: %s\nEmail: %s\n\nSubject: %s\n\n%s```", resp.user, resp.email, resp.subject, resp.message))
}

// messageCreate is an event handler which is called whenever a new message
// is created in the Discord server.
func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.Bot || !strings.HasPrefix(m.Content, conf.prefix) || strings.TrimPrefix(m.Content, conf.prefix) == "" {
		return
	}
	c := strings.TrimPrefix(m.Content, conf.prefix)
	if err := commands.Parse(s, m, c); err != nil {
		errorLog.Printf("Failed to execute command %q: %s", c, err)
	}
}

func loadLog() *os.File {
	var err error
	logF, err = os.OpenFile("log.log", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		fmt.Println(err)
		os.Exit(2)
	}
	return logF
}

func setInitialGame(s *discordgo.Session) {
	err := s.UpdateStatus(0, conf.game)
	if err != nil {
		errorLog.Println("Update status err:", err)
		return
	}
	infoLog.Println("set initial game to ", conf.game)
	return
}

// loadConfig loads teh configuration information found in ./config.json
func loadConfig() error {
	file, err := ioutil.ReadFile("config.json")
	if err != nil {
		errorLog.Println("Config open err", err)
		return fmt.Errorf("failed to read configuration file: ", err)
	}

	if len(file) < 1 {
		infoLog.Println("config.json is empty")
		return errors.New("Configuration file 'config.json' was empty")
	}

	if err := json.Unmarshal(file, conf); err != nil {
		errorLog.Println("Config unmarshal err", err)
		return fmt.Errorf("failed to unmarshal configuration JSON: %s", err)
	}

	return nil
}

func saveConfig() {
	out, err := json.MarshalIndent(conf, "", "  ")
	if err != nil {
		errorLog.Println("Config marshall err:", err)
		return
	}

	err = ioutil.WriteFile("config.json", out, 0600)
	if err != nil {
		errorLog.Println("Save config err:", err)
	}
	return
}

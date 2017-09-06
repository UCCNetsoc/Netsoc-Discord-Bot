package main

import (
	"strings"
	"os"
	"github.com/bwmarrin/discordgo"
	"log"
	"fmt"
	"io/ioutil"
	"errors"
	"encoding/json"
	"net/http"
)

var (
	conf = &config{}
	logF *os.File
	infoLog *log.Logger
	errorLog *log.Logger
	errEmptyFile = errors.New("file is empty")	
	dg = &discordgo.Session{}
)

func main() {
	loadLog()
	defer logF.Close()

	log.SetOutput(logF)

	infoLog = log.New(logF, "INFO:  ", log.Ldate|log.Ltime)	
	errorLog = log.New(logF, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)
	if conf.InDev {
		errorLog = log.New(os.Stdout, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)		
	}

	loadConfig()

	var err error
	// Create a new Discord session using the provided bot token.
	dg, err = discordgo.New("Bot " + conf.Token)
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

	//Register handlers
	dg.AddHandler(messageCreate)
	//dg.AddHandler(membJoin)

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

	dg.ChannelMessageSend("354748497683283979", fmt.Sprintf("```From: %s\nEmail: %s\n\nSubject: %s\n\n%s```", resp.User, resp.Email, resp.Subject, resp.Message))
}

func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.Bot || !strings.HasPrefix(m.Content, conf.Prefix) || strings.TrimPrefix(m.Content, conf.Prefix) == ""{
		return
	}

	parseCommand(s, m, strings.TrimPrefix(m.Content, conf.Prefix))
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
	err := s.UpdateStatus(0, conf.Game)
	if err != nil {
		errorLog.Println("Update status err:", err)
		return
	}
	infoLog.Println("set initial game to ", conf.Game)
	return
}

func loadConfig() error {
	file, err := ioutil.ReadFile("config.json")
	if err != nil {
		errorLog.Println("Config open err", err)
		return err
	}

	if len(file) < 1 {
		infoLog.Println("config.json is empty")
		return errEmptyFile
	}

	err = json.Unmarshal(file, conf)
	if err != nil {
		errorLog.Println("Config unmarshal err", err)
		return err
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

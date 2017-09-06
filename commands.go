package main 

import (
	"github.com/bwmarrin/discordgo"
	"strings"
	_ "github.com/bwmarrin/discordgo"
)

var (
	commMap = make(map[string]command)
	ping = command{"ping", "", false, pingCommand}.add()
)

func parseCommand(s *discordgo.Session, m *discordgo.MessageCreate, msg string) {
	msglist := strings.Fields(msg)

	if command, ok := commMap[msglist[0]]; ok {
		command.Exec(s, m, msglist)
	}
}

func (c command) add() command {
	commMap[strings.ToLower(c.Name)] = c
	return c
}

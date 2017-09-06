package main

import (
	"github.com/bwmarrin/discordgo"
)

func pingCommand(s *discordgo.Session, m *discordgo.MessageCreate, msglist []string) {
	s.ChannelMessageSend(m.ChannelID, "Pong!")
}
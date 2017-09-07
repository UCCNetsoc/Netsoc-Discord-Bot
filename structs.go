package main

import (
	"github.com/bwmarrin/discordgo"
)

type command struct {
	Name string
	Help string

	AdminOnly bool

	Exec func(*discordgo.Session, *discordgo.MessageCreate, []string)
}

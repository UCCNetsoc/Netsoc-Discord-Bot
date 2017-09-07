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

type helpBody struct {
	User    string `json:"user"`
	Email   string `json:"email"`
	Subject string `json:"subject"`
	Message string `json:"message"`
}

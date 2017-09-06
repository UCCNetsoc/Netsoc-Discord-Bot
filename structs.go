package main

import (
	"github.com/bwmarrin/discordgo"
)

type config struct {
	Game   string `json:"game"`
	Prefix string `json:"prefix"`
	Token  string `json:"token"`

	InDev bool `json:"indev"`
}

type command struct {
	Name string
	Help string

	AdminOnly bool

	Exec func(*discordgo.Session, *discordgo.MessageCreate, []string)
}

type helpBody struct {
	User string `json:"user"`
	Email string `json:"email"`
	Subject string `json:"subject"`
	Message string `json:"message"`
}
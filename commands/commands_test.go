package commands

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/bwmarrin/discordgo"
	"github.com/stretchr/testify/assert"
)

func TestPingCommand(t *testing.T) {
	httpClient, mux, server := testServer()
	defer server.Close()

	mux.HandleFunc("/api/v6/channels/354748497683283979/messages", func(w http.ResponseWriter, r *http.Request) {
		assertMethod(t, "POST", r)
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"id":"360750208365363201","channel_id":"354748497683283979","content":"Pong!","timestamp":"2017-09-22T11:32:32.089000+00:00","edited_timestamp":"","mention_roles":[],"tts":false,"mention_everyone":false,"author":{"id":"357194797175537674","email":"","username":"Netsoc Bot","avatar":"d6ed896aeeed69d4a93db58c5249b646","discriminator":"0255","token":"","verified":false,"mfa_enabled":false,"bot":true},"attachments":[],"embeds":[],"mentions":[],"reactions":null,"type":0}`)
	})

	expected := &discordgo.Message{
		ID:              "360750208365363201",
		ChannelID:       "354748497683283979",
		Content:         "Pong!",
		Timestamp:       "2017-09-22T11:32:32.089000+00:00",
		EditedTimestamp: "",
		MentionRoles:    []string{},
		Tts:             false,
		MentionEveryone: false,
		Author: &discordgo.User{
			ID:            "357194797175537674",
			Email:         "",
			Username:      "Netsoc Bot",
			Avatar:        "d6ed896aeeed69d4a93db58c5249b646",
			Discriminator: "0255",
			Token:         "",
			Verified:      false,
			MFAEnabled:    false,
			Bot:           true,
		},
		Attachments: []*discordgo.MessageAttachment{},
		Embeds:      []*discordgo.MessageEmbed{},
		Mentions:    []*discordgo.User{},
		Reactions:   []*discordgo.MessageReactions(nil),
		Type:        0,
	}

	ctx, s, m := getTestCommandParams(httpClient)
	msgContent := []string{}

	err, returnMsg := pingCommand(ctx, s, m, msgContent)
	assert.Nil(t, err)
	assert.Equal(t, expected, returnMsg)
}

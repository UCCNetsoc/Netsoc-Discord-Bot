package commands

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/bwmarrin/discordgo"
	"github.com/stretchr/testify/assert"

	"github.com/UCCNetworkingSociety/Netsoc-Discord-Bot/config"
	"github.com/UCCNetworkingSociety/Netsoc-Discord-Bot/logging"
)

// testServer returns an http Client, ServeMux, and Server. The client proxies
// requests to the server and handlers can be registered on the mux to handle
// requests. The caller must close the test server.
func testServer() (*http.Client, *http.ServeMux, *httptest.Server) {
	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	transport := &RewriteTransport{&http.Transport{
		Proxy: func(req *http.Request) (*url.URL, error) {
			return url.Parse(server.URL)
		},
	}}
	client := &http.Client{Transport: transport}
	return client, mux, server
}

// RewriteTransport rewrites https requests to http to avoid TLS cert issues
// during testing.
type RewriteTransport struct {
	Transport http.RoundTripper
}

// RoundTrip rewrites the request scheme to http and calls through to the
// composed RoundTripper or if it is nil, to the http.DefaultTransport.
func (t *RewriteTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.URL.Scheme = "http"
	if t.Transport == nil {
		return http.DefaultTransport.RoundTrip(req)
	}
	return t.Transport.RoundTrip(req)
}

// assertMethod tests that the incoming Request method matches the expected method
func assertMethod(t *testing.T, expectedMethod string, req *http.Request) {
	assert.Equal(t, expectedMethod, req.Method)
}

// assertQuery tests that the Request has the expected url query key/val pairs
func assertQuery(t *testing.T, expected map[string]string, req *http.Request) {
	queryValues := req.URL.Query()
	expectedValues := url.Values{}
	for key, value := range expected {
		expectedValues.Add(key, value)
	}
	assert.Equal(t, expectedValues, queryValues)
}

// assertPostForm tests that the Request has the expected key values pairs url
// encoded in its Body
func assertPostForm(t *testing.T, expected map[string]string, req *http.Request) {
	req.ParseForm() // parses request Body to put url.Values in r.Form/r.PostForm
	expectedValues := url.Values{}
	for key, value := range expected {
		expectedValues.Add(key, value)
	}
	assert.Equal(t, expectedValues, req.Form)
}

// createTestMessage populates an example message sent to the discord API
func createTestMessage(content string) *discordgo.MessageCreate {
	return &discordgo.MessageCreate{
		&discordgo.Message{
			ID:              "1",
			ChannelID:       "354748497683283979",
			Content:         content,
			Timestamp:       discordgo.Timestamp("0"),
			EditedTimestamp: discordgo.Timestamp("0"),
			MentionRoles:    []string{},
			Tts:             false,
			MentionEveryone: false,
			Author: &discordgo.User{
				ID:            "1",
				Email:         "test@test.com",
				Username:      "test",
				Avatar:        "",
				Discriminator: "1",
				Token:         "1",
				Verified:      true,
				MFAEnabled:    false,
				Bot:           false,
			},
			Attachments: []*discordgo.MessageAttachment{},
			Embeds:      []*discordgo.MessageEmbed{},
			Mentions:    []*discordgo.User{},
			Reactions:   []*discordgo.MessageReactions{},
			Type:        1,
		},
	}
}

// getTestCommandParams generates the needed parameters to run a command
func getTestCommandParams(client *http.Client) (context.Context, *discordgo.Session, *discordgo.MessageCreate) {
	config.LoadConfig()
	conf := config.GetConfig()
	l, _ := logging.New()
	defer l.Close()
	dg, _ := discordgo.New("Bot " + conf.Token)
	dg.Client = client
	ctx := logging.NewContext(context.Background(), l)
	msg := createTestMessage("!ping")
	return ctx, dg, msg
}

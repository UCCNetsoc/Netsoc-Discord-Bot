package commands

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPingCommand(t *testing.T) {
	response, err := pingCommand(context.Background(), []string{"!ping"})
	assert.Nil(t, err)
	assert.Equal(t, "Pong!", response)
}

func TestMinecraftCommand(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/has-players", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, `{"players": [{"name": "gilly"}]}`)
	}))
	mux.HandleFunc("/no-players", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, `{"players": []}`)
	}))
	ts := httptest.NewServer(mux)
	defer ts.Close()

	minecraftAPIURL = ts.URL + "/has-players"
	got, err := minecraftCommand(context.Background(), []string{"minecraft"})
	if err != nil {
		t.Fatalf("minecraftCommand error: %s", err)
	}

	if !strings.Contains(got, "gilly") {
		t.Errorf("Response did not contain an online user: %q", got)
	}

	minecraftAPIURL = ts.URL + "/no-players"
	got, err = minecraftCommand(context.Background(), []string{"minecraft"})
	if err != nil {
		t.Fatalf("minecraftCommand error: %s", err)
	}

	if !strings.Contains(strings.ToLower(got), "nobody") {
		t.Errorf("Response didn't say that 'nobody' was online: %q", got)
	}
}

func TestInspireCommand(t *testing.T) {
	var (
		quoteText   = "If you think your users are idiots, only idiots will use it."
		quoteAuthor = "Linus Torvalds"
	)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, `{"quoteText": %q, "quoteAuthor":%q}`, quoteText, quoteAuthor)
	}))
	defer ts.Close()

	inspirationalQuotesAPIURL = ts.URL
	got, err := inspireCommand(context.Background(), []string{"inspire"})
	if err != nil {
		t.Fatalf("inspireCommand error: %s", err)
	}

	if !strings.Contains(got, quoteText) {
		t.Errorf("Response did not contain the quote: %q", got)
	}
	if !strings.Contains(got, quoteAuthor) {
		t.Errorf("Response did not contain the quote author: %q", got)
	}
}

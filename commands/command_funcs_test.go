package commands

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPingCommand(t *testing.T) {
	response, err := pingCommand(context.Background(), []string{"!ping"})
	assert.Nil(t, err)
	assert.Equal(t, "Pong!", response)
}

package commands

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStringInSlice(t *testing.T) {
	list := []string{
		"Hello",
		"World",
	}

	assert.Equal(t, StringInSlice("Hello", list), true)
	assert.Equal(t, StringInSlice("Henlo", list), false)

}

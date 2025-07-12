package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseCustomMessagesWithData(t *testing.T) {
	contacts := parseCustomMessages("George:12345,not-valid,Thomas:6789023")

	assert.Len(t, contacts, 2)
	assert.Equal(t, "George", contacts[0].Name)
	assert.Equal(t, "12345", contacts[0].Phone)
	assert.Equal(t, "Thomas", contacts[1].Name)
	assert.Equal(t, "6789023", contacts[1].Phone)
}

func TestParseCustomMessagesEmpty(t *testing.T) {
	contacts := parseCustomMessages("")

	assert.Empty(t, contacts)
}

func TestParseCustomMessagesInvalid(t *testing.T) {
	commands := parseBotkaCommands("help:sos,volleyball:vqq123,no_message:taj333,shout:vsichni")

	assert.Equal(t, "sos", commands.Help)
	assert.Equal(t, "vqq123", commands.Volleyball)
	assert.Equal(t, "taj333", commands.NoMessage)
	assert.Equal(t, "vsichni", commands.Shout)
}

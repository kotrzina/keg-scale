package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseCustomMessagesWithData(t *testing.T) {
	contacts := parseCustomMessages("George:12345,not-valid,Thomas:6789023")

	assert.Equal(t, 2, len(contacts))
	assert.Equal(t, "George", contacts[0].Name)
	assert.Equal(t, "12345", contacts[0].Phone)
	assert.Equal(t, "Thomas", contacts[1].Name)
	assert.Equal(t, "6789023", contacts[1].Phone)
}

func TestParseCustomMessagesEmpty(t *testing.T) {
	contacts := parseCustomMessages("")

	assert.Equal(t, 0, len(contacts))
}

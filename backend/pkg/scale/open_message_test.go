package scale

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestScale_shouldSendOpen(t *testing.T) {
	cases := []struct {
		name        string
		openBefore  time.Duration
		closeBefore time.Duration
		shouldSend  bool
	}{
		{"should send open", 24 * time.Hour, 19 * time.Hour, true},
		{"once per 12 hours", 9 * time.Hour, 19 * time.Hour, false},
		{"at least 3 hours closed", 24 * time.Hour, 2 * time.Hour, false},
	}

	for _, tt := range cases {
		s := &Scale{
			pub: pub{
				openedAt: time.Now().Add(-tt.openBefore),
				closedAt: time.Now().Add(-tt.closeBefore),
			},
		}

		assert.Equal(t, tt.shouldSend, s.shouldSendOpen(), tt.name)
	}
}

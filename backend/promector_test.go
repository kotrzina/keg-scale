package main

import (
	"testing"
	"time"
)

func TestFormatStep(t *testing.T) {
	cases := []struct {
		in   time.Duration
		want string
	}{
		{10 * time.Second, "10s"},
		{time.Minute, "1m"},
		{10 * time.Minute, "10m"},
		{20 * time.Minute, "20m"},
		{time.Hour, "1h"},
		{12 * time.Hour, "12h"},
		{4*time.Hour + 30*time.Minute, "4h30m"},
	}

	for _, c := range cases {
		got := formatStep(c.in)
		if got != c.want {
			t.Errorf("formatStep(%q) == %q, want %q", c.in, got, c.want)
		}
	}
}

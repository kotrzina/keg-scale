package main

import (
	"strings"
	"time"
)

func strip(s string) string {
	var result strings.Builder
	for i := 0; i < len(s); i++ {
		b := s[i]
		if ('0' <= b && b <= '9') ||
			b == '.' {
			result.WriteByte(b)
		}
	}
	return result.String()
}

func formatDate(t time.Time) string {
	if t.Unix() <= 0 {
		return ""
	}

	return t.In(getTz()).Format("2006-01-02 15:04:05")
}

func formatTime(t time.Time) string {
	if t.Unix() <= 0 {
		return ""
	}

	return t.In(getTz()).Format("15:04")
}

func getTz() *time.Location {
	tz, _ := time.LoadLocation("Europe/Prague")
	return tz
}

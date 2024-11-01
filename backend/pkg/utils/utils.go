package utils

import (
	"strings"
	"time"
)

func Strip(s string) string {
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

func FormatDate(t time.Time) string {
	if t.Unix() <= 0 {
		return ""
	}

	return t.In(getTz()).Format("2006-01-02 15:04:05")
}

func FormatTime(t time.Time) string {
	if t.Unix() <= 0 {
		return ""
	}

	return t.In(getTz()).Format("15:04")
}

func getTz() *time.Location {
	tz, _ := time.LoadLocation("Europe/Prague")
	return tz
}

func GetOkJson() []byte {
	return []byte(`{"is_ok":true}`)
}

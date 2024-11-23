package utils

import (
	"os"
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

func FormatDateShort(t time.Time) string {
	if t.Unix() <= 0 {
		return ""
	}

	return t.In(getTz()).Format("02. 01.")
}

func FormatTime(t time.Time) string {
	if t.Unix() <= 0 {
		return ""
	}

	return t.In(getTz()).Format("15:04")
}

func getTz() *time.Location {
	tz, err := time.LoadLocation("Europe/Prague")
	if err != nil {
		os.Stderr.WriteString("Failed to load timezone: " + err.Error())
		os.Exit(1)
	}
	return tz
}

func GetOkJSON() []byte {
	return []byte(`{"is_ok":true}`)
}

func FormatBeer(n int) string {
	if n == 1 {
		return "pivo"
	}

	if n < 5 {
		return "piva"
	}

	return "piv"
}

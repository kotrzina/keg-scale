package utils

import (
	"fmt"
	"os"
	"strings"
	"time"

	"mvdan.cc/xurls/v2"
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

	return t.In(GetTz()).Format("2006-01-02 15:04:05")
}

func FormatWeekday(t time.Time) string {
	return t.In(GetTz()).Weekday().String()
}

func FormatDateShort(t time.Time) string {
	if t.Unix() <= 0 {
		return ""
	}

	return t.In(GetTz()).Format("02. 01.")
}

func FormatTime(t time.Time) string {
	if t.Unix() <= 0 {
		return ""
	}

	return t.In(GetTz()).Format("15:04")
}

func GetTz() *time.Location {
	tz, err := time.LoadLocation("Europe/Prague")
	if err != nil {
		_, _ = os.Stderr.WriteString("Failed to load timezone: " + err.Error())
		os.Exit(1)
	}
	return tz
}

type Ok struct {
	IsOk bool `json:"is_ok"`
}

func GetOk() Ok {
	return Ok{IsOk: true}
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

// UnwrapHTML replaces all URLs in the string with HTML links
// and replaces newlines with <br/>
// handles use cases like the link is the last word in the sentence with a dot
func UnwrapHTML(s string) string {
	rxStrict := xurls.Strict()
	links := rxStrict.FindAllString(s, -1)
	for _, link := range links {
		s = strings.ReplaceAll(s, link, fmt.Sprintf("<a target=\"_blank\" href=\"%s\">%s</a>", link, link))
	}

	s = strings.ReplaceAll(s, "\n", "<br/>")
	return s
}

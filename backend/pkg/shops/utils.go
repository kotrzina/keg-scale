package shops

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

var reParsePriceString = regexp.MustCompile(`^(\d+).\d+`)

func parsePriceString(s string) (int, error) {
	s = strings.ReplaceAll(s, " ", "")
	matches := reParsePriceString.FindStringSubmatch(s)
	if len(matches) != 2 {
		return 0, fmt.Errorf("could not parse %s as price", s)
	}

	f, err := strconv.Atoi(matches[1])
	if err != nil {
		return 0, fmt.Errorf("could not convert %s to int", matches[1])
	}

	return f, nil
}

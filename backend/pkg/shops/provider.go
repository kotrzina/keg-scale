package shops

import "strings"

type StockType string

const (
	StockTypeAvailable StockType = "available"
	StockTypeUnknown   StockType = "unknown"
)

type Provider interface {
	GetItems() ([]ProviderItem, error)
}

type ProviderItem struct {
	Name  string    `json:"name"`
	Link  string    `json:"url"`
	Price string    `json:"price"`
	stock StockType // currently not exported for language model
}

// title sanitization
func sanitizeTitle(s string) string {
	titlePrefixes := []string{"sud", "pivo"}
	uselessWords := []string{"AKCE"}

	for _, prefix := range titlePrefixes {
		s = strings.TrimPrefix(s, prefix)
	}

	for _, word := range uselessWords {
		s = strings.ReplaceAll(s, word, "")
	}

	// remove whitespaces
	s = strings.ReplaceAll(s, "  ", " ")

	return strings.TrimSpace(s)
}

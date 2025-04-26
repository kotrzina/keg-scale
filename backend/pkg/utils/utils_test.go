package utils

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestStrip(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"Hello, World!", ""},
		{"Hello, 123!", "123"},
		{"Hello, 123.74! hello", "123.74"},
		{"Hello, 0.12", "0.12"},
	}

	for _, test := range tests {
		if got := Strip(test.input); got != test.expected {
			t.Errorf("Strip(%q) = %q; want %q", test.input, got, test.expected)
		}
	}
}

func TestGetOkJson(t *testing.T) {
	got, err := json.Marshal(GetOk())
	require.NoError(t, err)
	require.Contains(t, string(got), "ok")
}

func TestFormatBeer(t *testing.T) {
	tests := []struct {
		input    int
		expected string
	}{
		{1, "pivo"},
		{2, "piva"},
		{3, "piva"},
		{4, "piva"},
		{5, "piv"},
		{6, "piv"},
		{7, "piv"},
		{8, "piv"},
		{9, "piv"},
		{10, "piv"},
	}

	for _, test := range tests {
		if got := FormatBeer(test.input); got != test.expected {
			t.Errorf("FormatBeer(%d) = %q; want %q", test.input, got, test.expected)
		}
	}
}

func TestUnwrapHTML(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"", ""},
		{"🍺🍺🍺", "🍺🍺🍺"},
		{"Hello, World!", "Hello, World!"},
		{"The Link is https://example.com right?", "The Link is <a target=\"_blank\" href=\"https://example.com\">https://example.com</a> right?"},
		{"http://notsecured.cz page", "<a target=\"_blank\" href=\"http://notsecured.cz\">http://notsecured.cz</a> page"},
		{"Two links:\n- http://notsecured.cz\n- https://go.lang", "Two links:<br/>- <a target=\"_blank\" href=\"http://notsecured.cz\">http://notsecured.cz</a><br/>- <a target=\"_blank\" href=\"https://go.lang\">https://go.lang</a>"},
		{"Link is last word in the sentence https://dotpage.cz/hell.", "Link is last word in the sentence <a target=\"_blank\" href=\"https://dotpage.cz/hell\">https://dotpage.cz/hell</a>."},
		{"Fotky z tenisového turnaje 2023 najdete na této adrese: https://www.rajce.idnes.cz/dao/album/tenis-veselice-2023\n\nPokracovani", "Fotky z tenisového turnaje 2023 najdete na této adrese: <a target=\"_blank\" href=\"https://www.rajce.idnes.cz/dao/album/tenis-veselice-2023\">https://www.rajce.idnes.cz/dao/album/tenis-veselice-2023</a><br/><br/>Pokracovani"},
	}

	for _, test := range tests {
		if got := UnwrapHTML(test.input); got != test.expected {
			t.Errorf("UnwrapHTML(%q) = %q; want %q", test.input, got, test.expected)
		}
	}
}

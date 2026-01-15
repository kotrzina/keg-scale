package ai

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSanitizeBeerTitle(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"sud Pilsner Urquell 50l", "Pilsner Urquell 50l"},
		{"pivo Kozel 30l", "Kozel 30l"},
		{"Gambrinus AKCE 50l", "Gambrinus 50l"},
		{"sud AKCE Staropramen", "Staropramen"},
		{"  Multiple   spaces  ", "Multiple spaces"},
		{"Normal Beer Name", "Normal Beer Name"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := sanitizeBeerTitle(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestRemoveUnwantedHTML(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "remove svg",
			input:    `<div><svg width="100"><path d="M0 0"/></svg>text</div>`,
			expected: `<div>text</div>`,
		},
		{
			name:     "remove img",
			input:    `<div><img src="test.jpg">text</div>`,
			expected: `<div>text</div>`,
		},
		{
			name:     "remove style attribute",
			input:    `<div style="color: red;">text</div>`,
			expected: `<div>text</div>`,
		},
		{
			name:     "remove class attribute",
			input:    `<div class="my-class">text</div>`,
			expected: `<div>text</div>`,
		},
		{
			name:     "collapse multiple spaces",
			input:    `<div>text    with   spaces</div>`,
			expected: `<div>text with spaces</div>`,
		},
		{
			name:     "remove space before closing tag",
			input:    `<div >text</div >`,
			expected: `<div>text</div>`,
		},
		{
			name:     "remove space between tags",
			input:    `<div> </div> <span>text</span>`,
			expected: `<div></div><span>text</span>`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := removeUnwantedHTML(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestRemoveHrefs(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{`<a href="https://example.com">link</a>`, `<a >link</a>`},
		{`<a href="/path">link</a>`, `<a >link</a>`},
		{`no links here`, `no links here`},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := removeHrefs(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

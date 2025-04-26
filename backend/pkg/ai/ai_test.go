package ai

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRenderPrompt(t *testing.T) {
	rendered := renderPrompt()

	assert.Contains(t, rendered, "Botka")
	assert.NotContains(t, rendered, "${datetime}")
	assert.Greater(t, len(rendered), 200)
	assert.Less(t, len(rendered), 3000)
}

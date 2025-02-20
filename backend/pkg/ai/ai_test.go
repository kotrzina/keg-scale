package ai

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRenderPrompt(t *testing.T) {
	rendered := renderPrompt()

	assert.Contains(t, rendered, "Botka")
	assert.NotContains(t, rendered, "${datetime}")
	assert.True(t, len(rendered) > 200)
	assert.True(t, len(rendered) < 3000)
}

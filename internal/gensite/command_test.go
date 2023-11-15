package gensite

import (
	_ "embed"
	"github.com/stretchr/testify/assert"
	"testing"
)

//go:embed command_test.html
var commandTestHtml string

func TestCommand_Html(t *testing.T) {
	cmd := command{
		Section:     "section",
		Name:        "name",
		Description: "description",
	}

	result, err := cmd.html()
	assert.NoError(t, err)
	assert.Equal(t, commandTestHtml, string(result))
}

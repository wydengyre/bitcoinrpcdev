package gensite

import (
	_ "embed"
	"fmt"
)

//go:embed command.html
var commandHtml string

type command struct {
	Version     string
	Section     string
	Name        string
	DateTime    string
	Description string
}

var commandTmpl = mustBtcTemplate("command", commandHtml)

func (c *command) html() ([]byte, error) {
	rendered, err := commandTmpl.render(c)
	if err != nil {
		e := fmt.Errorf("failed to render command %s html: %w", c.Name, err)
		return nil, e
	}
	return rendered, nil
}

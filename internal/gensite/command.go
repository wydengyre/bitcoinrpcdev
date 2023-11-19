package gensite

import (
	_ "embed"
	"fmt"
	"html/template"
)

//go:embed command.html
var commandHtml string

//go:embed command.css
var commandCss string

type command struct {
	Version     string
	Section     string
	Name        string
	DateTime    string
	Description string
}

var commandTmpl = mustBtcTemplate("command", commandHtml)

func (c *command) html() ([]byte, error) {
	renderMap := structToMap(c)
	renderMap["css"] = template.CSS(commandCss)
	rendered, err := commandTmpl.render(renderMap)
	if err != nil {
		e := fmt.Errorf("failed to render command %s html: %w", c.Name, err)
		return nil, e
	}
	return rendered, nil
}

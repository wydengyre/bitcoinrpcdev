package gensite

import (
	_ "embed"
	"fmt"
	"html/template"
	"slices"
)

//go:embed section.html
var sectionHtml string

//go:embed section.css
var sectionCss string

var sectionTmpl = mustBtcTemplate("command", sectionHtml)

type section struct {
	Name     string
	Version  string
	Commands []string
}

func (s *section) html() ([]byte, error) {
	slices.Sort(s.Commands)
	renderMap := structToMap(s)
	renderMap["css"] = template.CSS(sectionCss)
	rendered, err := sectionTmpl.render(renderMap)
	if err != nil {
		e := fmt.Errorf("failed to render section html: %w", err)
		return nil, e
	}
	return rendered, nil
}

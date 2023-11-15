package gensite

import (
	_ "embed"
	"fmt"
	"slices"
)

//go:embed section.html
var sectionHtml string

var sectionTmpl = mustBtcTemplate("command", sectionHtml)

type section struct {
	Name     string
	Version  string
	Commands []string
}

func (s *section) html() ([]byte, error) {
	slices.Sort(s.Commands)
	rendered, err := sectionTmpl.render(s)
	if err != nil {
		e := fmt.Errorf("failed to render section html: %w", err)
		return nil, e
	}
	return rendered, nil
}

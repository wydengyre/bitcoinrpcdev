package gensite

import (
	_ "embed"
	"fmt"
)

//go:embed index.html
var indexHtml string

type index struct {
	Latest   string
	Versions []string
}

var indexTmpl = mustBtcTemplate("index", indexHtml)

func (i *index) html() ([]byte, error) {
	rendered, err := indexTmpl.render(i)
	if err != nil {
		e := fmt.Errorf("failed to render index html: %w", err)
		return nil, e
	}
	return rendered, nil
}

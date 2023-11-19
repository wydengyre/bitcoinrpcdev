package gensite

import (
	_ "embed"
	"fmt"
	"html/template"
)

//go:embed index.html
var indexHtml string

//go:embed index.css
var indexCss string

type index struct {
	Latest   string
	Versions []string
}

var indexTmpl = mustBtcTemplate("index", indexHtml)

func (i *index) html() ([]byte, error) {
	renderMap := structToMap(i)
	renderMap["css"] = template.CSS(indexCss)
	rendered, err := indexTmpl.render(renderMap)
	if err != nil {
		e := fmt.Errorf("failed to render index html: %w", err)
		return nil, e
	}
	return rendered, nil
}

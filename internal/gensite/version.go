package gensite

import (
	_ "embed"
	"fmt"
	"html/template"
	"slices"
)

//go:embed version.html
var versionHtml string

//go:embed version.css
var versionCss string

type version struct {
	Name     string
	Sections map[string][]string
}

var versionTmpl = mustBtcTemplate("version", versionHtml)

func (v *version) html() ([]byte, error) {
	toRender := make(map[string]interface{}, 3)
	toRender["Version"] = v
	toRender["SectionsAlpha"] = alphaKeys(v.Sections)
	toRender["css"] = template.CSS(versionCss)
	rendered, err := versionTmpl.render(toRender)
	if err != nil {
		e := fmt.Errorf("failed to render version html: %w", err)
		return nil, e
	}
	return rendered, nil
}

func alphaKeys(m map[string][]string) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	slices.Sort(keys)
	return keys
}

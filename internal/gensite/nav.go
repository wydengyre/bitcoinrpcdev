package gensite

import (
	_ "embed"
	"html/template"
)

//go:embed nav.html
var navHtml string

func mustAddNav(t *template.Template) *template.Template {
	template.Must(addNav(t))
	return t
}

func addNav(t *template.Template) (*template.Template, error) {
	return t.New("nav").Parse(navHtml)
}

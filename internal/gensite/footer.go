package gensite

import "html/template"
import _ "embed"

//go:embed footer.html
var footerHtml string

func mustAddFooter(t *template.Template, err error) *template.Template {
	template.Must(t, err)
	template.Must(addFooter(t))
	return t
}

func addFooter(t *template.Template) (*template.Template, error) {
	return t.New("footer").Parse(footerHtml)
}

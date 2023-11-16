package gensite

import (
	"github.com/go-git/go-git/v5"
	"html/template"
)
import _ "embed"

//go:embed footer.html
var footerHtml string

var gitRepo = mustGetGitRepo()
var gitHash = mustGetGitHash()
var gitModified = mustGetGitModified()

func mustAddFooter(t *template.Template, err error) *template.Template {
	template.Must(t, err)
	template.Must(addFooter(t))
	return t
}

func addFooter(t *template.Template) (*template.Template, error) {
	return t.New("footer").Parse(footerHtml)
}

func addFooterData(m map[string]interface{}) {
	m["gitHash"] = gitHash
	m["gitHashShort"] = gitHash[:7]
	m["gitModified"] = gitModified
	m["datetime"] = nowStr()
}

func mustGetGitRepo() *git.Repository {
	h, err := git.PlainOpen(".")
	if err != nil {
		panic(err)
	}
	return h
}

func mustGetGitHash() string {
	h, err := gitRepo.Head()
	if err != nil {
		panic(err)
	}
	return h.Hash().String()
}

func mustGetGitModified() bool {
	w, err := gitRepo.Worktree()
	if err != nil {
		panic(err)
	}
	s, err := w.Status()
	if err != nil {
		panic(err)
	}
	return !s.IsClean()
}

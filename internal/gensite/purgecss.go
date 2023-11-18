package gensite

import (
	"bytes"
	_ "embed"
	"errors"
	"fmt"
	"github.com/andybalholm/cascadia"
	"golang.org/x/net/html"
	"os/exec"
)

//go:embed purgecss.mjs
var purgeCSS string
var cssMatcher = cascadia.MustCompile("style")

func purgeStyleCss(h []byte) ([]byte, error) {
	doc, err := html.Parse(bytes.NewReader(h))
	if err != nil {
		e := fmt.Errorf("failed to parse html: %w", err)
		return nil, e
	}
	// find the style tag
	n := cascadia.Query(doc, cssMatcher)
	if n == nil {
		e := errors.New("failed to find style tag in html " + string(h))
		return nil, e
	}
	css := n.FirstChild.Data
	n.FirstChild.Data = ""

	buf := bytes.Buffer{}
	err = html.Render(&buf, doc)
	if err != nil {
		e := fmt.Errorf("failed to render html without css: %w", err)
		return nil, e
	}

	minCss, err := purgecss(buf.String(), css)
	if err != nil {
		e := fmt.Errorf(`failed to purge css %s
from html %s
error: %w`, css, string(h), err)
		return nil, e
	}
	n.FirstChild.Data = string(minCss)

	buf = bytes.Buffer{}
	err = html.Render(&buf, doc)
	if err != nil {
		e := fmt.Errorf("failed to render html with purged css: %w", err)
		return nil, e
	}
	return buf.Bytes(), nil
}

func purgecss(html, css string) ([]byte, error) {
	cmd := nodeCommand(purgeCSS, html, css)
	out, err := cmd.Output()
	if err != nil {
		var exitErr *exec.ExitError
		ok := errors.As(err, &exitErr)
		if ok {
			e := fmt.Errorf("purgecss.mjs stderr %s: %w", exitErr.Stderr, exitErr)
			return nil, e
		}
		e := fmt.Errorf("error running purgecss.mjs: %w", err)
		return nil, e
	}

	return out, nil
}

func nodeCommand(script string, arg ...string) *exec.Cmd {
	nodeArgs := append([]string{"--input-type=module", "--eval=" + script, "--"}, arg...)
	return exec.Command("node", nodeArgs...)
}

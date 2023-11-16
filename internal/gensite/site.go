package gensite

import (
	"bytes"
	"fmt"
	"github.com/tdewolff/minify/v2"
	"github.com/tdewolff/minify/v2/html"
	"log/slog"
	"net/url"
	"os"
	"path/filepath"
)

const canonicalHome = `https://bitcoinrpc.dev/`
const mimeHtml = "text/html"

// site is a map of page paths to page contents
type site map[string][]byte

type htmler interface {
	html() ([]byte, error)
}

var m = minify.New()

func init() {
	m.AddFunc(mimeHtml, html.Minify)
}

// newFs creates a new site
func newSite() site {
	return make(site)
}

// add adds an html file to the site
func (s site) add(path string, hr htmler) error {
	slog.Debug("adding", "path", path)
	h, err := hr.html()
	if err != nil {
		return fmt.Errorf("failed to render html: %w", err)
	}
	ch, err := addCanonicalUrl(h, path)
	if err != nil {
		return fmt.Errorf("failed to add canonical url: %w", err)
	}
	mh, err := minifyHtml(ch)
	if err != nil {
		return fmt.Errorf("failed to minify html: %w", err)
	}
	s[path] = mh
	return nil
}

func addCanonicalUrl(html []byte, path string) ([]byte, error) {
	canonicalUrl, err := url.JoinPath(canonicalHome, path)
	if err != nil {
		return nil, fmt.Errorf("failed to join canonical url: %w", err)
	}
	tag := fmt.Sprintf(`<link rel="canonical" href="%s">`, canonicalUrl)
	out := bytes.Replace(html, []byte(`</head>`), []byte(tag+`</head>`), 1)
	return out, nil
}

// write writes the site to the filesystem
func (s site) write(rootPath string) error {
	slog.Info("writing site", "rootPath", rootPath)
	for path, content := range s {
		outPath := filepath.Join(rootPath, path)
		outDir := filepath.Dir(outPath)

		err := os.MkdirAll(outDir, 0755)
		if err != nil {
			return fmt.Errorf("failed to create directory %s: %w", outDir, err)
		}

		err = os.WriteFile(outPath, content, 0644)
		if err != nil {
			return fmt.Errorf("failed to write file %s: %w", outPath, err)
		}
		slog.Info("wrote", "path", outPath)
	}
	return nil
}

func minifyHtml(html []byte) ([]byte, error) {
	return m.Bytes(mimeHtml, html)
}

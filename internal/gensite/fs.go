package gensite

import (
	"fmt"
	"os"
	"path/filepath"
)

// site is a map of page paths to page contents
type site map[string][]byte

type htmler interface {
	html() ([]byte, error)
}

// newFs creates a new site
func newSite() site {
	return make(site)
}

// add adds an html file to the site
func (s site) add(path string, hr htmler) error {
	h, err := hr.html()
	if err != nil {
		return fmt.Errorf("failed to render html: %w", err)
	}
	mh, err := minifyHtml(h)
	if err != nil {
		return fmt.Errorf("failed to minify html: %w", err)
	}
	s[path] = mh
	return nil
}

// write writes the site to the filesystem
func (s site) write(rootPath string) error {
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
	}
	return nil
}

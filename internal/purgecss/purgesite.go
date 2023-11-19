package purgecss

import (
	"bitcoinrpcschema/internal/bitcoind"
	"bitcoinrpcschema/internal/gensite"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sync"
)

// PurgeSite creates purged css files for the site
//
// One css file will be created for each type of page in the site.
func PurgeSite(cssPath, outPath string) error {
	css, err := os.ReadFile(cssPath)
	if err != nil {
		e := fmt.Errorf("error reading css file: %w", err)
		return e
	}
	siteCss, err := purgeSite(css)
	if err != nil {
		e := fmt.Errorf("error purging site: %w", err)
		return e
	}
	for page, content := range siteCss {
		outPath := filepath.Join(outPath, page)

		err = os.WriteFile(outPath, content, 0644)
		if err != nil {
			return fmt.Errorf("failed to write file %s: %w", outPath, err)
		}
		slog.Info("wrote", "path", outPath)
	}
	return nil
}

func purgeSite(css []byte) (map[string][]byte, error) {
	db := bitcoind.RpcDb{
		bitcoind.ReleaseVersion{Major: 1, Minor: 2, Patch: 3}: {
			"section1": {
				{Name: "cmd1", Help: "help1"},
			},
		},
	}
	dbBytes, err := db.Marshal()
	if err != nil {
		e := fmt.Errorf("error marshalling db: %w", err)
		return nil, e
	}
	site, err := gensite.GenSite(dbBytes)
	if err != nil {
		e := fmt.Errorf("error generating site: %w", err)
		return nil, e
	}
	pages := map[string][]byte{
		"index":   site["index.html"],
		"version": site["1.2.3/index.html"],
		"section": site["1.2.3/section1/index.html"],
		"command": site["1.2.3/section1/cmd1/index.html"],
	}

	cssContent := string(css)
	cssMap := make(map[string][]byte, len(pages))
	var cssMapMutex sync.Mutex
	wg := sync.WaitGroup{}
	errs := make([]error, len(pages))
	i := 0
	for page, content := range pages {
		wg.Add(1)
		iCopy := i // deal with closure issues
		go func(page string, content []byte) {
			defer wg.Done()
			purged, err := purgecss(string(content), cssContent)
			if err != nil {
				e := fmt.Errorf("error purging css for page %s: %w", page, err)
				errs[iCopy] = e
				return
			}
			cssPath := page + ".css"
			cssMapMutex.Lock()
			cssMap[cssPath] = purged
			cssMapMutex.Unlock()
		}(page, content)
		i++
	}
	wg.Wait()
	err = errors.Join(errs...)
	if err != nil {
		e := fmt.Errorf("error purging css: %w", err)
		return nil, e
	}
	return cssMap, nil
}

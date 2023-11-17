package gensite_test

import (
	"bitcoinrpcschema/internal/bitcoind"
	"bitcoinrpcschema/internal/gensite"
	"bytes"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/net/html"
	"log/slog"
	"net/url"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

// generatedSite is a map of the generated path to the contents of the file
var generatedSite map[string][]byte

func TestMain(m *testing.M) {
	panicOnErr := func(err error) {
		if err != nil {
			panic(err)
		}
	}

	logger := slog.Default()
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError})))
	defer slog.SetDefault(logger)

	// this could be generalized to an end-to-end integration test if we make mock bitcoind archives
	// and serve them locally
	webDir, err := os.MkdirTemp("", "bitcoinrpcdev-gensite-test")
	panicOnErr(err)
	cleanup := func() {
		err := os.RemoveAll(webDir)
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "error removing temp dir %s: %v\n", webDir, err)
		}
	}
	db := bitcoind.RpcDb{
		bitcoind.ReleaseVersion{Major: 1, Minor: 2, Patch: 3}: {
			"section1": {
				{Name: "cmd1", Help: "help1"},
				{Name: "cmd2", Help: "help2"},
			},
			"section2": {
				{Name: "cmd3", Help: "help3"},
				{Name: "cmd4", Help: "help4"},
			},
		},
		bitcoind.ReleaseVersion{Major: 2, Minor: 3, Patch: 4}: {
			"section1": {
				{Name: "cmd1", Help: "help1"},
				{Name: "cmd2", Help: "help2"},
			},
			"section2": {
				{Name: "cmd3", Help: "help3"},
				{Name: "cmd4", Help: "help4-old"},
			},
		},
	}

	var exitCode int
	func() {
		defer cleanup()
		dbBytes, err := db.Marshal()
		panicOnErr(err)
		err = gensite.Gen(dbBytes, webDir)
		panicOnErr(err)
		generatedSite, err = readSite(webDir)
		panicOnErr(err)
	}()
	exitCode = m.Run()
	os.Exit(exitCode)
}

// test the pages we expect are generated
func TestGeneratedPages(t *testing.T) {
	expected := []string{
		"1.2.3/index.html",
		"1.2.3/section1/cmd1/index.html",
		"1.2.3/section1/cmd2/index.html",
		"1.2.3/section1/index.html",
		"1.2.3/section2/cmd3/index.html",
		"1.2.3/section2/cmd4/index.html",
		"1.2.3/section2/index.html",
		"2.3.4/index.html",
		"2.3.4/section1/cmd1/index.html",
		"2.3.4/section1/cmd2/index.html",
		"2.3.4/section1/index.html",
		"2.3.4/section2/cmd3/index.html",
		"2.3.4/section2/cmd4/index.html",
		"2.3.4/section2/index.html",
		"index.html",
	}
	generated := make([]string, 0, len(generatedSite))
	for path := range generatedSite {
		generated = append(generated, path)
	}
	slices.Sort(generated)
	assert.Equal(t, expected, generated)
}

func TestCrawl(t *testing.T) {
	visited := make(map[string][]byte, len(generatedSite))
	err := crawl(generatedSite, func(path string, content []byte) error {
		visited[path] = content
		return nil
	})
	require.NoError(t, err)
	assert.Equal(t, generatedSite, visited)
}

func crawl(site map[string][]byte, f func(path string, content []byte) error) error {
	remainingPages := make(map[string]struct{}, len(site))
	for path := range site {
		remainingPages[path] = struct{}{}
	}

	var visit func(path string) error
	visit = func(path string) error {
		_, ok := remainingPages[path]
		if !ok {
			return nil
		}
		delete(remainingPages, path)
		page, ok := site[path]
		if !ok {
			return fmt.Errorf("missing page %s, this should be impossible", path)
		}

		err := f(path, page)
		if err != nil {
			return fmt.Errorf("error processing %s: %w", path, err)
		}

		links, err := relativeLinks(page)
		if err != nil {
			return fmt.Errorf("error parsing links for page %s: %w", path, err)
		}
		for _, link := range links {
			nextPath, err := joinRelativeLink(path, link)
			if err != nil {
				return fmt.Errorf("error joining link %s from %s: %w", link, path, err)
			}
			err = visit(nextPath)
			if err != nil {
				return fmt.Errorf("error visiting %s from %s: %w", nextPath, path, err)
			}
		}
		return nil
	}
	return visit("index.html")
}

func relativeLinks(h []byte) ([]string, error) {
	doc, err := html.Parse(bytes.NewReader(h))
	if err != nil {
		return nil, err
	}

	var links []string
	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "a" {
			for _, a := range n.Attr {
				if a.Key == "href" && !strings.HasPrefix(a.Val, "http://") && !strings.HasPrefix(a.Val, "https://") {
					links = append(links, a.Val)
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(doc)
	return links, nil
}

func joinRelativeLink(base, link string) (nextPath string, err error) {
	if link[0] == '/' {
		nextPath, err = url.JoinPath(link, "index.html")
	} else if strings.HasSuffix(link, ".html") {
		nextPath, err = url.JoinPath(filepath.Dir(base), link)
	} else {
		nextPath, err = url.JoinPath(filepath.Dir(base), link, "index.html")
	}
	if err != nil {
		return "", err
	}
	if nextPath[0] == '/' {
		nextPath = nextPath[1:]
	}
	return
}

// readSite reads the site from the given path and returns a map of the path to the contents of the file
func readSite(path string) (map[string][]byte, error) {
	site := make(map[string][]byte)

	curDir, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = os.Chdir(curDir)
	}()

	err = os.Chdir(path)
	if err != nil {
		return nil, err
	}
	err = filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		b, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		site[path] = b
		return nil
	})
	if err != nil {
		return nil, err
	}
	return site, nil
}

package downloader

import (
	"errors"
	"fmt"
	"github.com/go-git/go-billy/v5/memfs"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/storage/memory"
	"io"
	"log/slog"
	"os"
	"path"
)

func DownloadGitRpcs(repoUrl string, versions []releaseVersion) error {
	rpcs, err := getGitRpcs(repoUrl, versions)
	if err != nil {
		e := fmt.Errorf("failed to get rpcs from git repo %s: %w", repoUrl, err)
		return e
	}
	for v, files := range rpcs {
		dir := path.Join("bitcoin-core", "bitcoin-"+v.String())
		for p, content := range files {
			fullPath := path.Join(dir, p)
			err := os.WriteFile(fullPath, content, 0644)
			if err != nil {
				e := fmt.Errorf("failed to write bitcoin rpc file %s: %w", fullPath, err)
				return e
			}
		}
	}
	return nil
}

func getGitRpcs(repoUrl string, versions []releaseVersion) (map[releaseVersion]map[string][]byte, error) {
	co := git.CloneOptions{
		Progress: os.Stderr,
		URL:      repoUrl,
	}
	fs := memfs.New()
	slog.Info("cloning bitcoin repo: " + repoUrl)
	r, err := git.Clone(memory.NewStorage(), fs, &co)
	if err != nil {
		e := fmt.Errorf("failed to clone bitcoin repo: %w", err)
		return nil, e
	}

	rpcs := make(map[releaseVersion]map[string][]byte, len(versions))
	for _, v := range versions {
		rpcs[v], err = getVersionRpcCppFiles(r, v)
		if err != nil {
			e := fmt.Errorf("failed to get rpc cpp files for version %v: %w", v, err)
			return nil, e
		}
	}
	return rpcs, nil
}

func getVersionRpcCppFiles(r *git.Repository, v releaseVersion) (map[string][]byte, error) {
	tagName := "v" + v.String()
	t, err := r.Tag(tagName)
	if err != nil {
		e := fmt.Errorf("failed to get tag: %w", err)
		return nil, e
	}
	w, err := r.Worktree()
	if err != nil {
		e := fmt.Errorf("failed to get worktree: %w", err)
		return nil, e
	}
	err = w.Checkout(&git.CheckoutOptions{Branch: t.Name()})
	if err != nil {
		e := fmt.Errorf("failed to check out: %w", err)
		return nil, e
	}

	wantedPaths := []string{"src/rpc/register.h"}
	ps, err := w.Filesystem.ReadDir("src/rpc")
	if err != nil {
		e := fmt.Errorf("failed to read dir: %w", err)
		return nil, e
	}
	for _, p := range ps {
		if p.IsDir() {
			continue
		}
		if path.Ext(p.Name()) != ".cpp" {
			continue
		}
		wantedPaths = append(wantedPaths, "src/rpc/"+p.Name())
	}

	files := make(map[string][]byte, len(wantedPaths))
	for _, p := range wantedPaths {
		f, err := w.Filesystem.Open(p)
		if err != nil {
			e := fmt.Errorf("failed to open file: %w", err)
			return nil, e
		}
		content, err := io.ReadAll(f)
		closeErr := f.Close()
		if err != nil {
			e := fmt.Errorf("failed to read file: %w", err)
			return nil, errors.Join(e, closeErr)
		}
		if closeErr != nil {
			e := fmt.Errorf("failed to close file: %w", closeErr)
			return nil, e
		}

		fileName := path.Base(p)
		files[fileName] = content
	}
	return files, nil
}

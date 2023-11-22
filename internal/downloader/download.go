package downloader

import (
	"archive/tar"
	"compress/gzip"
	"errors"
	"fmt"
	"github.com/panjf2000/ants/v2"
	"golang.org/x/net/html"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"sync"
)

// how many major versions to keep
const keepVersions = 3

// seems reasonable
const maxDownloadStreams = 3

type releaseVersion struct {
	major uint
	minor uint
	patch uint
}

func (rv releaseVersion) String() string {
	if rv.patch == 0 {
		return fmt.Sprintf("%d.%d", rv.major, rv.minor)
	}
	return fmt.Sprintf("%d.%d.%d", rv.major, rv.minor, rv.patch)
}

func (rv releaseVersion) cmp(other releaseVersion) int {
	if rv.major > other.major {
		return 1
	}
	if rv.major < other.major {
		return -1
	}
	if rv.minor > other.minor {
		return 1
	}
	if rv.minor < other.minor {
		return -1
	}
	if rv.patch > other.patch {
		return 1
	}
	if rv.patch < other.patch {
		return -1
	}
	return 0
}

var releaseVersionRe = regexp.MustCompile(`^bitcoin-core-(\d+)\.(\d+)\.?(\d+)?/$`)

func parseReleaseVersion(s string) (releaseVersion, error) {
	var releaseVersion releaseVersion
	matches := releaseVersionRe.FindStringSubmatch(s)
	if len(matches) < 3 {
		return releaseVersion, fmt.Errorf("invalid release version: %s", s)
	}
	rvMaj, err := strconv.Atoi(matches[1])
	if err != nil {
		e := fmt.Errorf("invalid release version, error parsing major version: %s, %w", s, err)
		return releaseVersion, e
	}
	releaseVersion.major = uint(rvMaj)
	rvMin, err := strconv.Atoi(matches[2])
	if err != nil {
		e := fmt.Errorf("invalid release version, error parsing minor version: %s, %w", s, err)
		return releaseVersion, e
	}
	releaseVersion.minor = uint(rvMin)
	if len(matches) == 4 && matches[3] != "" {
		rvPatch, err := strconv.Atoi(matches[3])
		if err != nil {
			e := fmt.Errorf("invalid release version, error parsing patch version: %s, %w", s, err)
			return releaseVersion, e
		}
		releaseVersion.patch = uint(rvPatch)
	}
	return releaseVersion, nil
}

func Get(rootPath, binUrl, gitUrl string) error {
	r, err := http.Get(binUrl)
	if err != nil {
		return err
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(r.Body)

	var versions []releaseVersion
	var latestVersion releaseVersion
	doc := html.NewTokenizer(r.Body)
	for tokenType := doc.Next(); tokenType != html.ErrorToken; {
		token := doc.Token()
		if tokenType == html.StartTagToken && token.Data == "a" {
			for _, attr := range token.Attr {
				if attr.Key == "href" {
					v, err := parseReleaseVersion(attr.Val)
					if err == nil {
						versions = append(versions, v)
						if latestVersion.cmp(v) < 0 {
							latestVersion = v
						}
					} else {
						slog.Debug(fmt.Sprintf("error parsing release version: %s\n", err))
					}
				}
			}
		}
		tokenType = doc.Next()
	}

	minMajor := latestVersion.major - keepVersions + 1
	const approxReleasesPerMajorVersion = 4
	keptVersions := make([]releaseVersion, 0, keepVersions*approxReleasesPerMajorVersion)
	for _, version := range versions {
		if version.major >= minMajor {
			keptVersions = append(keptVersions, version)
		}
	}

	slog.Info(fmt.Sprintf("versions to download: %v\n", keptVersions))

	var wg sync.WaitGroup
	downloadedVersions := make([]releaseVersion, 0, len(keptVersions))
	errs := make([]error, 0, len(keptVersions))
	p, err := ants.NewPoolWithFunc(maxDownloadStreams, func(i interface{}) {
		defer wg.Done()
		err := downloadRelease(rootPath, binUrl, i.(releaseVersion))
		_, ok := err.(errorNotFound)
		if ok {
			slog.Info(fmt.Sprintf("release unavailable: %s\n", i.(releaseVersion)))
		} else if err != nil {
			e := fmt.Errorf("error downloading release: %w", err)
			errs = append(errs, e)
		} else {
			slog.Info(fmt.Sprintf("downloaded version %s\n", i))
			downloadedVersions = append(downloadedVersions, i.(releaseVersion))
		}
	})
	if err != nil {
		return fmt.Errorf("error creating download pool: %w", err)
	}
	defer p.Release()
	for _, version := range keptVersions {
		wg.Add(1)
		slog.Info(fmt.Sprintf("downloading version %s\n", version))
		err := p.Invoke(version)
		if err != nil {
			return fmt.Errorf("error invoking download pool: %w", err)
		}
	}
	wg.Wait()
	if len(errs) > 0 {
		joined := errors.Join(errs...)
		return fmt.Errorf("errors downloading releases: %w", joined)
	}

	return DownloadGitRpcs(gitUrl, downloadedVersions)
}

type errorNotFound struct {
	release string
}

func (e errorNotFound) Error() string {
	return "not found: " + e.release
}

var bitcoindPathRe = regexp.MustCompile(`^bitcoin-\d+\.\d+(?:\.\d+)?/bin/bitcoind$`)

func downloadRelease(rootPath, binUrl string, version releaseVersion) error {
	plat := "x86_64-linux-gnu"
	if runtime.GOARCH == "arm64" {
		plat = "aarch64-linux-gnu"
	}
	releaseUrl := fmt.Sprintf("%sbitcoin-core-%s/bitcoin-%s-%s.tar.gz", binUrl, version, version, plat)

	r, err := http.Get(releaseUrl)
	if err != nil {
		return fmt.Errorf("error downloading release: %w", err)
	}
	defer silentClose(r.Body)

	if r.StatusCode == http.StatusNotFound {
		return errorNotFound{release: releaseUrl}
	}
	if r.StatusCode != http.StatusOK {
		return fmt.Errorf("error downloading release from %s, HTTP request failed with status: %d", releaseUrl, r.StatusCode)
	}

	gzReader, err := gzip.NewReader(r.Body)
	if err != nil {
		return err
	}
	defer silentClose(gzReader)

	tarReader := tar.NewReader(gzReader)

	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("error reading tar file: %w", err)
		}

		if !bitcoindPathRe.MatchString(header.Name) {
			continue
		}

		slog.Info("uncompressing " + header.Name)

		filePath := fmt.Sprintf("%s/%s", rootPath, header.Name)
		dirPath := filePath
		if dirPath[len(dirPath)-1] != '/' {
			dirPath = filepath.Dir(filePath)
		}
		if err := os.MkdirAll(dirPath, 0755); err != nil {
			return fmt.Errorf("error creating directories for file %s: %w", filePath, err)
		}
		if dirPath == filePath {
			continue
		}

		file, err := os.Create(filePath)
		if err != nil {
			return fmt.Errorf("error creating file %s: %w", filePath, err)
		}
		err = func() error {
			defer silentClose(file)
			_, err = io.Copy(file, tarReader)
			if err != nil {
				return fmt.Errorf("error copying file: %w", err)
			}
			err = file.Chmod(os.FileMode(header.Mode))
			if err != nil {
				return fmt.Errorf("error setting file permissions: %w", err)
			}
			return nil
		}()
		if err != nil {
			return err
		}
	}

	return nil
}

func silentClose(c io.Closer) {
	_ = c.Close()
}

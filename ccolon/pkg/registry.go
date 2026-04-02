package pkg

import (
	"archive/tar"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// PackageInfo represents the metadata fetched from the registry's package.json.
type PackageInfo struct {
	Name        string            `json:"name"`
	Description string            `json:"description,omitempty"`
	Latest      string            `json:"latest"`
	Versions    map[string]string `json:"versions,omitempty"` // version -> tarball filename
}

// RegistryClient talks to a GitHub-based CColon package registry.
type RegistryClient struct {
	// BaseURL is the GitHub repo URL, e.g. "https://github.com/TRC-Loop/ccolon-registry"
	BaseURL string
}

// NewRegistryClient creates a client for the given registry URL.
func NewRegistryClient(registryURL string) *RegistryClient {
	return &RegistryClient{BaseURL: strings.TrimRight(registryURL, "/")}
}

// rawURL converts the GitHub repo URL to a raw.githubusercontent.com path.
func (rc *RegistryClient) rawURL(path string) string {
	// "https://github.com/TRC-Loop/ccolon-registry" ->
	// "https://raw.githubusercontent.com/TRC-Loop/ccolon-registry/main/<path>"
	base := strings.Replace(rc.BaseURL, "https://github.com/", "https://raw.githubusercontent.com/", 1)
	return base + "/main/" + path
}

// releaseURL returns the download URL for a release asset.
func (rc *RegistryClient) releaseURL(name, version string) string {
	tag := name + "-" + version
	return rc.BaseURL + "/releases/download/" + tag + "/" + tag + ".tar.gz"
}

// FetchPackageInfo retrieves package metadata from the registry.
func (rc *RegistryClient) FetchPackageInfo(name string) (*PackageInfo, error) {
	url := rc.rawURL("packages/" + name + "/package.json")
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch package info: %s", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("package %q not found in registry", name)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("registry returned status %d for package %q", resp.StatusCode, name)
	}

	var info PackageInfo
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return nil, fmt.Errorf("invalid package.json for %q: %s", name, err)
	}
	return &info, nil
}

// Download fetches a package tarball and extracts it into destDir.
// It prints a progress bar to stdout while downloading.
func (rc *RegistryClient) Download(name, version, destDir string) error {
	url := rc.releaseURL(name, version)
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("download failed: %s", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return fmt.Errorf("release %s@%s not found in registry", name, version)
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download returned status %d", resp.StatusCode)
	}

	totalSize := resp.ContentLength
	label := fmt.Sprintf("%s@%s", name, version)

	// Wrap the body in a progress reader
	pr := &progressReader{
		reader: resp.Body,
		total:  totalSize,
		label:  label,
	}

	if err := extractTarGz(pr, destDir); err != nil {
		return fmt.Errorf("extraction failed: %s", err)
	}

	// Clear the progress line
	fmt.Printf("\r%-60s\n", fmt.Sprintf("downloaded %s", label))
	return nil
}

// progressReader wraps an io.Reader and prints a progress bar on each read.
type progressReader struct {
	reader  io.Reader
	total   int64
	current int64
	label   string
}

func (pr *progressReader) Read(p []byte) (int, error) {
	n, err := pr.reader.Read(p)
	pr.current += int64(n)

	barWidth := 20
	if pr.total > 0 {
		pct := float64(pr.current) / float64(pr.total)
		filled := int(pct * float64(barWidth))
		if filled > barWidth {
			filled = barWidth
		}
		bar := strings.Repeat("#", filled) + strings.Repeat("-", barWidth-filled)
		fmt.Printf("\r[%s] %3.0f%% downloading %s", bar, pct*100, pr.label)
	} else {
		// Unknown total size, show bytes downloaded
		fmt.Printf("\r[--------] %d bytes downloading %s", pr.current, pr.label)
	}

	return n, err
}

// extractTarGz reads a tar.gz stream from r and extracts it into destDir.
func extractTarGz(r io.Reader, destDir string) error {
	gz, err := gzip.NewReader(r)
	if err != nil {
		return fmt.Errorf("gzip error: %s", err)
	}
	defer gz.Close()

	tr := tar.NewReader(gz)
	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		target := filepath.Join(destDir, header.Name)

		// Prevent path traversal
		if !strings.HasPrefix(filepath.Clean(target), filepath.Clean(destDir)) {
			return fmt.Errorf("tar entry %q escapes destination", header.Name)
		}

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, 0755); err != nil {
				return err
			}
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
				return err
			}
			f, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, os.FileMode(header.Mode))
			if err != nil {
				return err
			}
			if _, err := io.Copy(f, tr); err != nil {
				f.Close()
				return err
			}
			f.Close()
		}
	}
	return nil
}

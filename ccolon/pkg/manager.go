package pkg

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

func packagesDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("cannot determine home directory: %s", err)
	}
	dir := filepath.Join(home, ".ccolon", "packages")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", err
	}
	return dir, nil
}

// ParseInstallArg parses "https://github.com/user/repo@version" into repo URL and version.
// If no version is specified, returns "" as version (meaning latest).
func ParseInstallArg(arg string) (repoURL string, version string) {
	// Handle @version suffix
	if idx := strings.LastIndex(arg, "@"); idx > 0 {
		// Make sure @ is not part of the URL scheme
		candidate := arg[idx+1:]
		prefix := arg[:idx]
		if !strings.Contains(candidate, "/") && !strings.Contains(candidate, ":") {
			return prefix, candidate
		}
	}
	return arg, ""
}

// FetchManifest downloads and parses the ccolon.json from a GitHub repo.
// If version is empty, uses the default branch (main).
func FetchManifest(repoURL, version string) (*Manifest, error) {
	ref := "main"
	if version != "" {
		ref = version
	}
	rawURL := githubRawURL(repoURL, ref, "ccolon.json")
	resp, err := http.Get(rawURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch manifest: %s", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode == 404 {
		return nil, fmt.Errorf("no ccolon.json found at %s (ref: %s)", repoURL, ref)
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("failed to fetch manifest: HTTP %d", resp.StatusCode)
	}
	var m Manifest
	if err := json.NewDecoder(resp.Body).Decode(&m); err != nil {
		return nil, fmt.Errorf("invalid ccolon.json: %s", err)
	}
	return &m, nil
}

func localPackagesDir() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(cwd, "ccolon_packages")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", err
	}
	return dir, nil
}

// Install downloads a package from a GitHub repo and installs it locally.
func Install(repoURL, version string, local bool) error {
	manifest, err := FetchManifest(repoURL, version)
	if err != nil {
		return err
	}
	if manifest.Name == "" {
		return fmt.Errorf("package has no name in ccolon.json")
	}

	ver := version
	if ver == "" {
		ver = manifest.Version
	}
	if ver == "" {
		ver = "latest"
	}

	var pkgDir string
	if local {
		pkgDir, err = localPackagesDir()
	} else {
		pkgDir, err = packagesDir()
	}
	if err != nil {
		return err
	}
	installDir := filepath.Join(pkgDir, manifest.Name+"@"+ver)

	// Remove existing installation
	os.RemoveAll(installDir)
	if err := os.MkdirAll(installDir, 0755); err != nil {
		return err
	}

	// Download the tarball from GitHub
	ref := version
	if ref == "" {
		ref = "main"
	}
	tarURL := repoURL + "/archive/refs/tags/" + ref + ".tar.gz"
	if version == "" {
		tarURL = repoURL + "/archive/refs/heads/main.tar.gz"
	}

	fmt.Printf("downloading %s@%s ...\n", manifest.Name, ver)
	resp, err := http.Get(tarURL)
	if err != nil {
		return fmt.Errorf("download failed: %s", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("download failed: HTTP %d (is the version tag correct?)", resp.StatusCode)
	}

	pr := &progressReader{
		reader: resp.Body,
		total:  resp.ContentLength,
		label:  fmt.Sprintf("%s@%s", manifest.Name, ver),
	}

	if err := extractTarGz(pr, installDir); err != nil {
		return fmt.Errorf("extraction failed: %s", err)
	}
	fmt.Printf("\rinstalled %s@%s to %s\n", manifest.Name, ver, installDir)

	// Flatten: GitHub tarballs extract into a subdirectory (repo-name-ref/)
	// Move contents up one level if there's exactly one subdirectory
	entries, _ := os.ReadDir(installDir)
	if len(entries) == 1 && entries[0].IsDir() {
		subDir := filepath.Join(installDir, entries[0].Name())
		subEntries, _ := os.ReadDir(subDir)
		for _, e := range subEntries {
			src := filepath.Join(subDir, e.Name())
			dst := filepath.Join(installDir, e.Name())
			os.Rename(src, dst)
		}
		os.Remove(subDir)
	}

	// Save repo URL into the manifest for reference
	manifest.Repository = repoURL
	_ = SaveManifest(installDir, manifest)

	return nil
}

// Remove deletes an installed package.
func Remove(name string) error {
	pkgDir, err := packagesDir()
	if err != nil {
		return err
	}

	// Find all versions of this package
	entries, err := os.ReadDir(pkgDir)
	if err != nil {
		return err
	}
	removed := 0
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), name+"@") {
			path := filepath.Join(pkgDir, e.Name())
			if err := os.RemoveAll(path); err != nil {
				return fmt.Errorf("failed to remove %s: %s", e.Name(), err)
			}
			fmt.Printf("removed %s\n", e.Name())
			removed++
		}
	}
	if removed == 0 {
		return fmt.Errorf("package '%s' is not installed", name)
	}
	return nil
}

// InstalledPackage holds info about a locally installed package.
type InstalledPackage struct {
	Name       string
	Version    string
	Path       string
	Repository string
}

// List returns all installed packages.
func List() ([]InstalledPackage, error) {
	pkgDir, err := packagesDir()
	if err != nil {
		return nil, err
	}
	entries, err := os.ReadDir(pkgDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var packages []InstalledPackage
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		parts := strings.SplitN(e.Name(), "@", 2)
		name := parts[0]
		ver := ""
		if len(parts) == 2 {
			ver = parts[1]
		}
		repo := ""
		pkgPath := filepath.Join(pkgDir, e.Name())
		if m, err := LoadManifest(pkgPath); err == nil {
			repo = m.Repository
		}
		packages = append(packages, InstalledPackage{
			Name:       name,
			Version:    ver,
			Path:       pkgPath,
			Repository: repo,
		})
	}
	return packages, nil
}

// Init creates a ccolon.json in the current directory.
func Init() error {
	if _, err := os.Stat("ccolon.json"); err == nil {
		return fmt.Errorf("ccolon.json already exists")
	}
	dir, err := os.Getwd()
	if err != nil {
		return err
	}
	name := filepath.Base(dir)
	m := &Manifest{
		Name:         name,
		Version:      "0.1.0",
		Dependencies: make(map[string]string),
	}
	if err := SaveManifest(".", m); err != nil {
		return err
	}
	fmt.Printf("created ccolon.json for '%s'\n", name)
	return nil
}

// githubRawURL converts a GitHub repo URL to a raw content URL.
func githubRawURL(repoURL, ref, path string) string {
	// https://github.com/user/repo -> https://raw.githubusercontent.com/user/repo/ref/path
	base := strings.Replace(repoURL, "https://github.com/", "https://raw.githubusercontent.com/", 1)
	base = strings.TrimRight(base, "/")
	return base + "/" + ref + "/" + path
}

// progressReader wraps an io.Reader and prints a progress bar.
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
		fmt.Printf("\r[--------] %d bytes downloading %s", pr.current, pr.label)
	}
	return n, err
}

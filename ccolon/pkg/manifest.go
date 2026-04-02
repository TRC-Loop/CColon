package pkg

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

const DefaultRegistry = "https://github.com/TRC-Loop/ccolon-registry"

// Manifest represents a ccolon.json project file.
type Manifest struct {
	Name         string            `json:"name"`
	Version      string            `json:"version"`
	Description  string            `json:"description,omitempty"`
	Dependencies map[string]string `json:"dependencies,omitempty"`
	Registry     string            `json:"registry,omitempty"`
	Type         string            `json:"type,omitempty"`  // "ccl" (default) or "go"
	Entry        string            `json:"entry,omitempty"` // entry point file
}

// RegistryURL returns the effective registry URL, checking the manifest field
// first, then the CCOLON_REGISTRY env var, and falling back to the default.
func (m *Manifest) RegistryURL() string {
	if m.Registry != "" {
		return m.Registry
	}
	if env := os.Getenv("CCOLON_REGISTRY"); env != "" {
		return env
	}
	return DefaultRegistry
}

// LoadManifest reads and parses a ccolon.json from the given directory.
func LoadManifest(dir string) (*Manifest, error) {
	path := filepath.Join(dir, "ccolon.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("could not read ccolon.json: %s", err)
	}
	var m Manifest
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("invalid ccolon.json: %s", err)
	}
	return &m, nil
}

// SaveManifest writes a Manifest as ccolon.json in the given directory.
func SaveManifest(dir string, m *Manifest) error {
	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return fmt.Errorf("could not encode manifest: %s", err)
	}
	data = append(data, '\n')
	path := filepath.Join(dir, "ccolon.json")
	return os.WriteFile(path, data, 0644)
}

package pkg

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/TRC-Loop/ccolon/compiler"
	"github.com/TRC-Loop/ccolon/vm"
)

// LoadPackages scans ~/.ccolon/packages/ and loads all installed packages
// into the VM. CColon packages are registered as file import sources.
// Go native packages are compiled and loaded as plugins.
func LoadPackages(machine *vm.VM, compileSource func(string) (*compiler.FuncObject, error)) error {
	pkgDir, err := packagesDir()
	if err != nil {
		return nil // no packages dir, nothing to load
	}

	entries, err := os.ReadDir(pkgDir)
	if err != nil {
		return nil
	}

	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		dir := filepath.Join(pkgDir, e.Name())
		manifestPath := filepath.Join(dir, "ccolon.json")
		data, err := os.ReadFile(manifestPath)
		if err != nil {
			continue // skip packages without manifest
		}

		var m Manifest
		if err := json.Unmarshal(data, &m); err != nil {
			continue
		}

		switch m.Type {
		case "go":
			if err := LoadGoPlugin(dir, machine); err != nil {
				fmt.Fprintf(os.Stderr, "warning: failed to load Go package '%s': %s\n", m.Name, err)
			}
		default:
			// CCL package: register as an importable module by running the entry file
			entry := m.Entry
			if entry == "" {
				entry = "lib.ccl"
			}
			entryPath := filepath.Join(dir, entry)
			source, err := os.ReadFile(entryPath)
			if err != nil {
				continue // skip if entry file doesn't exist
			}
			fn, err := compileSource(string(source))
			if err != nil {
				fmt.Fprintf(os.Stderr, "warning: failed to compile package '%s': %s\n", m.Name, err)
				continue
			}
			// Run the package code to register its globals/functions
			if err := machine.Run(fn); err != nil {
				fmt.Fprintf(os.Stderr, "warning: failed to load package '%s': %s\n", m.Name, err)
			}
		}
	}
	return nil
}

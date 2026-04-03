package pkg

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/TRC-Loop/ccolon/compiler"
	"github.com/TRC-Loop/ccolon/vm"
)

// LoadPackages scans ~/.ccolon/packages/ and (optionally) ./ccolon_packages/
// and registers all installed packages as importable modules.
func LoadPackages(machine *vm.VM, compileSource func(string) (*compiler.FuncObject, error)) error {
	dirs := []string{}

	// global packages
	if pkgDir, err := packagesDir(); err == nil {
		dirs = append(dirs, pkgDir)
	}

	// local project packages
	if cwd, err := os.Getwd(); err == nil {
		localDir := filepath.Join(cwd, "ccolon_packages")
		if info, err := os.Stat(localDir); err == nil && info.IsDir() {
			dirs = append(dirs, localDir)
		}
	}

	for _, pkgDir := range dirs {
		entries, err := os.ReadDir(pkgDir)
		if err != nil {
			continue
		}

		for _, e := range entries {
			if !e.IsDir() {
				continue
			}
			dir := filepath.Join(pkgDir, e.Name())
			manifestPath := filepath.Join(dir, "ccolon.json")
			data, err := os.ReadFile(manifestPath)
			if err != nil {
				continue
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
				// CCL package: compile and run entry file, capture globals as module
				entry := m.Entry
				if entry == "" {
					entry = "lib.ccl"
				}
				entryPath := filepath.Join(dir, entry)
				source, err := os.ReadFile(entryPath)
				if err != nil {
					continue
				}
				fn, err := compileSource(string(source))
				if err != nil {
					fmt.Fprintf(os.Stderr, "warning: failed to compile package '%s': %s\n", m.Name, err)
					continue
				}

				// Capture globals before and after running the package
				globalsBefore := machine.GlobalNames()
				if err := machine.Run(fn); err != nil {
					fmt.Fprintf(os.Stderr, "warning: failed to load package '%s': %s\n", m.Name, err)
					continue
				}
				globalsAfter := machine.GlobalNames()

				// Find new globals added by the package
				mod := &vm.ModuleValue{
					Name:    m.Name,
					Methods: make(map[string]*vm.NativeFuncValue),
				}
				beforeSet := make(map[string]bool)
				for _, n := range globalsBefore {
					beforeSet[n] = true
				}
				for _, n := range globalsAfter {
					if !beforeSet[n] {
						val := machine.GetGlobal(n)
						if fn, ok := val.(*vm.FuncValue); ok {
							// Wrap CColon function as a callable module method
							mod.Methods[n] = &vm.NativeFuncValue{
								Name: m.Name + "." + n,
								Fn: func(captured *vm.FuncValue) func(args []vm.Value) (vm.Value, error) {
									return func(args []vm.Value) (vm.Value, error) {
										return machine.CallFunc(captured, args)
									}
								}(fn),
							}
						}
					}
				}

				machine.RegisterModule(m.Name, mod)
			}
		}
	}
	return nil
}

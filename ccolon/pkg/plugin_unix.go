//go:build linux || darwin

package pkg

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"plugin"

	"github.com/TRC-Loop/ccolon/vm"
)

// LoadGoPlugin compiles and loads a Go native package.
// The package directory must contain Go source files with a Register function.
func LoadGoPlugin(pkgDir string, machine *vm.VM) error {
	soPath := filepath.Join(pkgDir, "plugin.so")

	// Compile the plugin
	cmd := exec.Command("go", "build", "-buildmode=plugin", "-o", soPath, ".")
	cmd.Dir = pkgDir
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to compile Go plugin in %s: %s\n%s", pkgDir, err, string(output))
	}

	// Load the plugin
	p, err := plugin.Open(soPath)
	if err != nil {
		return fmt.Errorf("failed to load plugin %s: %s", soPath, err)
	}

	// Look up the Register function
	sym, err := p.Lookup("Register")
	if err != nil {
		return fmt.Errorf("plugin %s has no Register function: %s", pkgDir, err)
	}

	registerFn, ok := sym.(func(*vm.VM))
	if !ok {
		return fmt.Errorf("plugin %s Register function has wrong signature (expected func(*vm.VM))", pkgDir)
	}

	registerFn(machine)
	return nil
}

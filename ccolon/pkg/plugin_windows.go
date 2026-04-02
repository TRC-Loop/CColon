//go:build windows

package pkg

import (
	"fmt"

	"github.com/TRC-Loop/ccolon/vm"
)

// LoadGoPlugin is not supported on Windows.
func LoadGoPlugin(pkgDir string, machine *vm.VM) error {
	return fmt.Errorf("Go native plugins are not supported on Windows")
}

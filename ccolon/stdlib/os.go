package stdlib

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/TRC-Loop/ccolon/vm"
)

func NewOsModule() *vm.ModuleValue {
	return &vm.ModuleValue{
		Name: "os",
		Methods: map[string]*vm.NativeFuncValue{
			"exec": {
				Name: "os.exec",
				Fn: func(args []vm.Value) (vm.Value, error) {
					if len(args) < 1 {
						return nil, fmt.Errorf("os.exec() takes at least 1 argument")
					}
					cmdStr, ok := args[0].(*vm.StringValue)
					if !ok {
						return nil, fmt.Errorf("os.exec() requires a string command")
					}

					var cmd *exec.Cmd
					if runtime.GOOS == "windows" {
						cmd = exec.Command("cmd", "/c", cmdStr.Val)
					} else {
						cmd = exec.Command("sh", "-c", cmdStr.Val)
					}

					output, err := cmd.CombinedOutput()
					exitCode := int64(0)
					if err != nil {
						if exitErr, ok := err.(*exec.ExitError); ok {
							exitCode = int64(exitErr.ExitCode())
						} else {
							return nil, fmt.Errorf("os.exec() failed: %s", err)
						}
					}

					return &vm.DictValue{
						Entries: map[string]vm.Value{
							"output":   &vm.StringValue{Val: string(output)},
							"exitcode": &vm.IntValue{Val: exitCode},
						},
						Order: []string{"output", "exitcode"},
					}, nil
				},
			},
			"env": {
				Name: "os.env",
				Fn: func(args []vm.Value) (vm.Value, error) {
					if len(args) < 1 || len(args) > 2 {
						return nil, fmt.Errorf("os.env() takes 1-2 arguments, got %d", len(args))
					}
					key, ok := args[0].(*vm.StringValue)
					if !ok {
						return nil, fmt.Errorf("os.env() requires a string key")
					}
					val, exists := os.LookupEnv(key.Val)
					if !exists {
						if len(args) == 2 {
							return args[1], nil
						}
						return &vm.NilValue{}, nil
					}
					return &vm.StringValue{Val: val}, nil
				},
			},
			"setenv": {
				Name: "os.setenv",
				Fn: func(args []vm.Value) (vm.Value, error) {
					if len(args) != 2 {
						return nil, fmt.Errorf("os.setenv() takes 2 arguments, got %d", len(args))
					}
					key, ok := args[0].(*vm.StringValue)
					if !ok {
						return nil, fmt.Errorf("os.setenv() requires string arguments")
					}
					val, ok := args[1].(*vm.StringValue)
					if !ok {
						return nil, fmt.Errorf("os.setenv() requires string arguments")
					}
					os.Setenv(key.Val, val.Val)
					return &vm.NilValue{}, nil
				},
			},
			"cwd": {
				Name: "os.cwd",
				Fn: func(args []vm.Value) (vm.Value, error) {
					dir, err := os.Getwd()
					if err != nil {
						return nil, fmt.Errorf("os.cwd() failed: %s", err)
					}
					return &vm.StringValue{Val: dir}, nil
				},
			},
			"chdir": {
				Name: "os.chdir",
				Fn: func(args []vm.Value) (vm.Value, error) {
					if len(args) != 1 {
						return nil, fmt.Errorf("os.chdir() takes 1 argument, got %d", len(args))
					}
					dir, ok := args[0].(*vm.StringValue)
					if !ok {
						return nil, fmt.Errorf("os.chdir() requires a string")
					}
					if err := os.Chdir(dir.Val); err != nil {
						return nil, fmt.Errorf("os.chdir() failed: %s", err)
					}
					return &vm.NilValue{}, nil
				},
			},
			"args": {
				Name: "os.args",
				Fn: func(args []vm.Value) (vm.Value, error) {
					elements := make([]vm.Value, len(os.Args))
					for i, a := range os.Args {
						elements[i] = &vm.StringValue{Val: a}
					}
					return &vm.ListValue{Elements: elements}, nil
				},
			},
			"platform": {
				Name: "os.platform",
				Fn: func(args []vm.Value) (vm.Value, error) {
					return &vm.StringValue{Val: runtime.GOOS}, nil
				},
			},
			"arch": {
				Name: "os.arch",
				Fn: func(args []vm.Value) (vm.Value, error) {
					return &vm.StringValue{Val: runtime.GOARCH}, nil
				},
			},
			"exit": {
				Name: "os.exit",
				Fn: func(args []vm.Value) (vm.Value, error) {
					code := int64(0)
					if len(args) >= 1 {
						if v, ok := args[0].(*vm.IntValue); ok {
							code = v.Val
						}
					}
					os.Exit(int(code))
					return &vm.NilValue{}, nil
				},
			},
			"mkdir": {
				Name: "os.mkdir",
				Fn: func(args []vm.Value) (vm.Value, error) {
					if len(args) != 1 {
						return nil, fmt.Errorf("os.mkdir() takes 1 argument, got %d", len(args))
					}
					dir, ok := args[0].(*vm.StringValue)
					if !ok {
						return nil, fmt.Errorf("os.mkdir() requires a string")
					}
					if err := os.MkdirAll(dir.Val, 0755); err != nil {
						return nil, fmt.Errorf("os.mkdir() failed: %s", err)
					}
					return &vm.NilValue{}, nil
				},
			},
			"remove": {
				Name: "os.remove",
				Fn: func(args []vm.Value) (vm.Value, error) {
					if len(args) != 1 {
						return nil, fmt.Errorf("os.remove() takes 1 argument, got %d", len(args))
					}
					path, ok := args[0].(*vm.StringValue)
					if !ok {
						return nil, fmt.Errorf("os.remove() requires a string")
					}
					if err := os.RemoveAll(path.Val); err != nil {
						return nil, fmt.Errorf("os.remove() failed: %s", err)
					}
					return &vm.NilValue{}, nil
				},
			},
			"exists": {
				Name: "os.exists",
				Fn: func(args []vm.Value) (vm.Value, error) {
					if len(args) != 1 {
						return nil, fmt.Errorf("os.exists() takes 1 argument, got %d", len(args))
					}
					path, ok := args[0].(*vm.StringValue)
					if !ok {
						return nil, fmt.Errorf("os.exists() requires a string")
					}
					_, err := os.Stat(path.Val)
					return &vm.BoolValue{Val: err == nil}, nil
				},
			},
			"listdir": {
				Name: "os.listdir",
				Fn: func(args []vm.Value) (vm.Value, error) {
					if len(args) != 1 {
						return nil, fmt.Errorf("os.listdir() takes 1 argument, got %d", len(args))
					}
					dir, ok := args[0].(*vm.StringValue)
					if !ok {
						return nil, fmt.Errorf("os.listdir() requires a string")
					}
					entries, err := os.ReadDir(dir.Val)
					if err != nil {
						return nil, fmt.Errorf("os.listdir() failed: %s", err)
					}
					elements := make([]vm.Value, len(entries))
					for i, e := range entries {
						elements[i] = &vm.StringValue{Val: e.Name()}
					}
					return &vm.ListValue{Elements: elements}, nil
				},
			},
			"hostname": {
				Name: "os.hostname",
				Fn: func(args []vm.Value) (vm.Value, error) {
					name, err := os.Hostname()
					if err != nil {
						return nil, fmt.Errorf("os.hostname() failed: %s", err)
					}
					return &vm.StringValue{Val: name}, nil
				},
			},
			"envlist": {
				Name: "os.envlist",
				Fn: func(args []vm.Value) (vm.Value, error) {
					d := &vm.DictValue{
						Entries: make(map[string]vm.Value),
						Order:   []string{},
					}
					for _, e := range os.Environ() {
						parts := strings.SplitN(e, "=", 2)
						if len(parts) == 2 {
							d.Entries[parts[0]] = &vm.StringValue{Val: parts[1]}
							d.Order = append(d.Order, parts[0])
						}
					}
					return d, nil
				},
			},
		},
	}
}

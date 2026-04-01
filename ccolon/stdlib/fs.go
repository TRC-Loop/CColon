package stdlib

import (
	"fmt"
	"os"

	"github.com/TRC-Loop/ccolon/vm"
)

func NewFsModule() *vm.ModuleValue {
	return &vm.ModuleValue{
		Name: "fs",
		Methods: map[string]*vm.NativeFuncValue{
			"open": {
				Name: "fs.open",
				Fn: func(args []vm.Value) (vm.Value, error) {
					if len(args) != 2 {
						return nil, fmt.Errorf("fs.open() takes 2 arguments (path, mode), got %d", len(args))
					}
					path, ok := args[0].(*vm.StringValue)
					if !ok {
						return nil, fmt.Errorf("fs.open() path must be a string")
					}
					mode, ok := args[1].(*vm.StringValue)
					if !ok {
						return nil, fmt.Errorf("fs.open() mode must be a string")
					}

					var file *os.File
					var err error
					switch mode.Val {
					case "r":
						file, err = os.Open(path.Val)
					case "w":
						file, err = os.Create(path.Val)
					case "a":
						file, err = os.OpenFile(path.Val, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
					default:
						return nil, fmt.Errorf("fs.open() mode must be 'r', 'w', or 'a', got '%s'", mode.Val)
					}
					if err != nil {
						return nil, fmt.Errorf("fs.open() failed: %s", err.Error())
					}

					return &vm.FileValue{
						Path: path.Val,
						Mode: mode.Val,
						File: file,
					}, nil
				},
			},
			"exists": {
				Name: "fs.exists",
				Fn: func(args []vm.Value) (vm.Value, error) {
					if len(args) != 1 {
						return nil, fmt.Errorf("fs.exists() takes 1 argument, got %d", len(args))
					}
					path, ok := args[0].(*vm.StringValue)
					if !ok {
						return nil, fmt.Errorf("fs.exists() path must be a string")
					}
					_, err := os.Stat(path.Val)
					return &vm.BoolValue{Val: err == nil}, nil
				},
			},
			"remove": {
				Name: "fs.remove",
				Fn: func(args []vm.Value) (vm.Value, error) {
					if len(args) != 1 {
						return nil, fmt.Errorf("fs.remove() takes 1 argument, got %d", len(args))
					}
					path, ok := args[0].(*vm.StringValue)
					if !ok {
						return nil, fmt.Errorf("fs.remove() path must be a string")
					}
					if err := os.Remove(path.Val); err != nil {
						return nil, fmt.Errorf("fs.remove() failed: %s", err.Error())
					}
					return &vm.NilValue{}, nil
				},
			},
			"mkdir": {
				Name: "fs.mkdir",
				Fn: func(args []vm.Value) (vm.Value, error) {
					if len(args) < 1 || len(args) > 2 {
						return nil, fmt.Errorf("fs.mkdir() takes 1-2 arguments (path, exist_ok=false), got %d", len(args))
					}
					path, ok := args[0].(*vm.StringValue)
					if !ok {
						return nil, fmt.Errorf("fs.mkdir() path must be a string")
					}

					existOk := false
					if len(args) == 2 {
						b, ok := args[1].(*vm.BoolValue)
						if !ok {
							return nil, fmt.Errorf("fs.mkdir() exist_ok must be a bool")
						}
						existOk = b.Val
					}

					if existOk {
						if err := os.MkdirAll(path.Val, 0755); err != nil {
							return nil, fmt.Errorf("fs.mkdir() failed: %s", err.Error())
						}
					} else {
						if err := os.Mkdir(path.Val, 0755); err != nil {
							return nil, fmt.Errorf("fs.mkdir() failed: %s", err.Error())
						}
					}
					return &vm.NilValue{}, nil
				},
			},
			"read": {
				Name: "fs.read",
				Fn: func(args []vm.Value) (vm.Value, error) {
					if len(args) != 1 {
						return nil, fmt.Errorf("fs.read() takes 1 argument (path), got %d", len(args))
					}
					path, ok := args[0].(*vm.StringValue)
					if !ok {
						return nil, fmt.Errorf("fs.read() path must be a string")
					}
					data, err := os.ReadFile(path.Val)
					if err != nil {
						return nil, fmt.Errorf("fs.read() failed: %s", err.Error())
					}
					return &vm.StringValue{Val: string(data)}, nil
				},
			},
			"write": {
				Name: "fs.write",
				Fn: func(args []vm.Value) (vm.Value, error) {
					if len(args) != 2 {
						return nil, fmt.Errorf("fs.write() takes 2 arguments (path, content), got %d", len(args))
					}
					path, ok := args[0].(*vm.StringValue)
					if !ok {
						return nil, fmt.Errorf("fs.write() path must be a string")
					}
					content, ok := args[1].(*vm.StringValue)
					if !ok {
						return nil, fmt.Errorf("fs.write() content must be a string")
					}
					if err := os.WriteFile(path.Val, []byte(content.Val), 0644); err != nil {
						return nil, fmt.Errorf("fs.write() failed: %s", err.Error())
					}
					return &vm.NilValue{}, nil
				},
			},
		},
	}
}


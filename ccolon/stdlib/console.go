package stdlib

import (
	"bufio"
	"fmt"
	"os"

	"github.com/TRC-Loop/ccolon/vm"
)

var stdinReader = bufio.NewReader(os.Stdin)

func NewConsoleModule() *vm.ModuleValue {
	return &vm.ModuleValue{
		Name: "console",
		Methods: map[string]*vm.NativeFuncValue{
			"println": {
				Name: "console.println",
				Fn: func(args []vm.Value) (vm.Value, error) {
					parts := make([]interface{}, len(args))
					for i, arg := range args {
						parts[i] = arg.String()
					}
					fmt.Println(parts...)
					return &vm.NilValue{}, nil
				},
			},
			"print": {
				Name: "console.print",
				Fn: func(args []vm.Value) (vm.Value, error) {
					for _, arg := range args {
						fmt.Print(arg.String())
					}
					return &vm.NilValue{}, nil
				},
			},
			"scanp": {
				Name: "console.scanp",
				Fn: func(args []vm.Value) (vm.Value, error) {
					if len(args) > 0 {
						fmt.Print(args[0].String())
					}
					line, err := stdinReader.ReadString('\n')
					if err != nil {
						return &vm.StringValue{Val: ""}, nil
					}
					// trim trailing newline
					if len(line) > 0 && line[len(line)-1] == '\n' {
						line = line[:len(line)-1]
					}
					if len(line) > 0 && line[len(line)-1] == '\r' {
						line = line[:len(line)-1]
					}
					return &vm.StringValue{Val: line}, nil
				},
			},
		},
	}
}

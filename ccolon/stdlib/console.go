package stdlib

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"syscall"
	"unsafe"

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
					os.Stdout.Sync()
					return &vm.NilValue{}, nil
				},
			},
			"flush": {
				Name: "console.flush",
				Fn: func(args []vm.Value) (vm.Value, error) {
					os.Stdout.Sync()
					return &vm.NilValue{}, nil
				},
			},
			"clearLine": {
				Name: "console.clearLine",
				Fn: func(args []vm.Value) (vm.Value, error) {
					fmt.Print("\r\033[K")
					os.Stdout.Sync()
					return &vm.NilValue{}, nil
				},
			},
			"setCursor": {
				Name: "console.setCursor",
				Fn: func(args []vm.Value) (vm.Value, error) {
					if len(args) < 2 {
						return nil, fmt.Errorf("console.setCursor() requires 2 arguments (row, col)")
					}
					row, ok := args[0].(*vm.IntValue)
					if !ok {
						return nil, fmt.Errorf("console.setCursor() row must be an int")
					}
					col, ok := args[1].(*vm.IntValue)
					if !ok {
						return nil, fmt.Errorf("console.setCursor() col must be an int")
					}
					fmt.Printf("\033[%d;%dH", row.Val, col.Val)
					os.Stdout.Sync()
					return &vm.NilValue{}, nil
				},
			},
			"getCursor": {
				Name: "console.getCursor",
				Fn: func(args []vm.Value) (vm.Value, error) {
					fmt.Print("\033[6n")
					os.Stdout.Sync()
					res, err := stdinReader.ReadString('R')
					if err != nil {
						return &vm.DictValue{
							Entries: map[string]vm.Value{
								"row": &vm.IntValue{Val: 0},
								"col": &vm.IntValue{Val: 0},
							},
							Order: []string{"row", "col"},
						}, nil
					}
					res = strings.TrimPrefix(res, "\x1b[")
					res = strings.TrimSuffix(res, "R")
					parts := strings.SplitN(res, ";", 2)
					row := int64(0)
					col := int64(0)
					if len(parts) == 2 {
						fmt.Sscanf(parts[0], "%d", &row)
						fmt.Sscanf(parts[1], "%d", &col)
					}
					return &vm.DictValue{
						Entries: map[string]vm.Value{
							"row": &vm.IntValue{Val: row},
							"col": &vm.IntValue{Val: col},
						},
						Order: []string{"row", "col"},
					}, nil
				},
			},
			"getSize": {
				Name: "console.getSize",
				Fn: func(args []vm.Value) (vm.Value, error) {
					type winsize struct {
						Row    uint16
						Col    uint16
						Xpixel uint16
						Ypixel uint16
					}
					ws := &winsize{}
					_, _, err := syscall.Syscall(syscall.SYS_IOCTL,
						uintptr(syscall.Stdout),
						uintptr(syscall.TIOCGWINSZ),
						uintptr(unsafe.Pointer(ws)))

					width := int64(80)
					height := int64(24)
					if err == 0 {
						width = int64(ws.Col)
						height = int64(ws.Row)
					}
					return &vm.DictValue{
						Entries: map[string]vm.Value{
							"width":  &vm.IntValue{Val: width},
							"height": &vm.IntValue{Val: height},
						},
						Order: []string{"width", "height"},
					}, nil
				},
			},
			"scanp": {
				Name: "console.scanp",
				Fn: func(args []vm.Value) (vm.Value, error) {
					if len(args) > 0 {
						fmt.Print(args[0].String())
						os.Stdout.Sync()
					}
					line, err := stdinReader.ReadString('\n')
					if err != nil {
						return &vm.StringValue{Val: ""}, nil
					}
					line = strings.TrimRight(line, "\r\n")
					return &vm.StringValue{Val: line}, nil
				},
			},
		},
	}
}

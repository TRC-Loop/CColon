package stdlib

import (
	"fmt"
	"math/rand"

	"github.com/TRC-Loop/ccolon/vm"
)

func NewRandomModule() *vm.ModuleValue {
	return &vm.ModuleValue{
		Name: "random",
		Methods: map[string]*vm.NativeFuncValue{
			"randint": {
				Name: "random.randint",
				Fn: func(args []vm.Value) (vm.Value, error) {
					if len(args) != 2 {
						return nil, fmt.Errorf("random.int() takes 2 arguments (min, max), got %d", len(args))
					}
					minVal, ok1 := args[0].(*vm.IntValue)
					maxVal, ok2 := args[1].(*vm.IntValue)
					if !ok1 || !ok2 {
						return nil, fmt.Errorf("random.int() requires integer arguments")
					}
					if minVal.Val > maxVal.Val {
						return nil, fmt.Errorf("random.int() min must be <= max")
					}
					r := minVal.Val + rand.Int63n(maxVal.Val-minVal.Val+1)
					return &vm.IntValue{Val: r}, nil
				},
			},
			"randfloat": {
				Name: "random.randfloat",
				Fn: func(args []vm.Value) (vm.Value, error) {
					return &vm.FloatValue{Val: rand.Float64()}, nil
				},
			},
			"choice": {
				Name: "random.choice",
				Fn: func(args []vm.Value) (vm.Value, error) {
					if len(args) != 1 {
						return nil, fmt.Errorf("random.choice() takes 1 argument, got %d", len(args))
					}
					list, ok := args[0].(*vm.ListValue)
					if !ok {
						return nil, fmt.Errorf("random.choice() requires a list")
					}
					if len(list.Elements) == 0 {
						return nil, fmt.Errorf("random.choice() cannot choose from empty list")
					}
					idx := rand.Intn(len(list.Elements))
					return list.Elements[idx], nil
				},
			},
			"char": {
				Name: "random.char",
				Fn: func(args []vm.Value) (vm.Value, error) {
					if len(args) != 1 {
						return nil, fmt.Errorf("random.char() takes 1 argument, got %d", len(args))
					}
					str, ok := args[0].(*vm.StringValue)
					if !ok {
						return nil, fmt.Errorf("random.char() requires a string")
					}
					if len(str.Val) == 0 {
						return nil, fmt.Errorf("random.char() cannot choose from empty string")
					}
					runes := []rune(str.Val)
					idx := rand.Intn(len(runes))
					return &vm.StringValue{Val: string(runes[idx])}, nil
				},
			},
			"shuffle": {
				Name: "random.shuffle",
				Fn: func(args []vm.Value) (vm.Value, error) {
					if len(args) != 1 {
						return nil, fmt.Errorf("random.shuffle() takes 1 argument, got %d", len(args))
					}
					list, ok := args[0].(*vm.ListValue)
					if !ok {
						return nil, fmt.Errorf("random.shuffle() requires a list")
					}
					rand.Shuffle(len(list.Elements), func(i, j int) {
						list.Elements[i], list.Elements[j] = list.Elements[j], list.Elements[i]
					})
					return &vm.NilValue{}, nil
				},
			},
		},
	}
}

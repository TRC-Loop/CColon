package stdlib

import (
	"fmt"
	"math"

	"github.com/TRC-Loop/ccolon/vm"
)

func toFloat(v vm.Value) (float64, bool) {
	switch val := v.(type) {
	case *vm.IntValue:
		return float64(val.Val), true
	case *vm.FloatValue:
		return val.Val, true
	default:
		return 0, false
	}
}

func NewMathModule() *vm.ModuleValue {
	return &vm.ModuleValue{
		Name: "math",
		Methods: map[string]*vm.NativeFuncValue{
			"pi": {
				Name: "math.pi",
				Fn: func(args []vm.Value) (vm.Value, error) {
					return &vm.FloatValue{Val: math.Pi}, nil
				},
			},
			"e": {
				Name: "math.e",
				Fn: func(args []vm.Value) (vm.Value, error) {
					return &vm.FloatValue{Val: math.E}, nil
				},
			},
			"inf": {
				Name: "math.inf",
				Fn: func(args []vm.Value) (vm.Value, error) {
					return &vm.FloatValue{Val: math.Inf(1)}, nil
				},
			},
			"sqrt": {
				Name: "math.sqrt",
				Fn: func(args []vm.Value) (vm.Value, error) {
					if len(args) != 1 {
						return nil, fmt.Errorf("math.sqrt() takes 1 argument, got %d", len(args))
					}
					f, ok := toFloat(args[0])
					if !ok {
						return nil, fmt.Errorf("math.sqrt() requires a number")
					}
					return &vm.FloatValue{Val: math.Sqrt(f)}, nil
				},
			},
			"abs": {
				Name: "math.abs",
				Fn: func(args []vm.Value) (vm.Value, error) {
					if len(args) != 1 {
						return nil, fmt.Errorf("math.abs() takes 1 argument, got %d", len(args))
					}
					switch v := args[0].(type) {
					case *vm.IntValue:
						if v.Val < 0 {
							return &vm.IntValue{Val: -v.Val}, nil
						}
						return v, nil
					case *vm.FloatValue:
						return &vm.FloatValue{Val: math.Abs(v.Val)}, nil
					default:
						return nil, fmt.Errorf("math.abs() requires a number")
					}
				},
			},
			"floor": {
				Name: "math.floor",
				Fn: func(args []vm.Value) (vm.Value, error) {
					if len(args) != 1 {
						return nil, fmt.Errorf("math.floor() takes 1 argument, got %d", len(args))
					}
					f, ok := toFloat(args[0])
					if !ok {
						return nil, fmt.Errorf("math.floor() requires a number")
					}
					return &vm.IntValue{Val: int64(math.Floor(f))}, nil
				},
			},
			"ceil": {
				Name: "math.ceil",
				Fn: func(args []vm.Value) (vm.Value, error) {
					if len(args) != 1 {
						return nil, fmt.Errorf("math.ceil() takes 1 argument, got %d", len(args))
					}
					f, ok := toFloat(args[0])
					if !ok {
						return nil, fmt.Errorf("math.ceil() requires a number")
					}
					return &vm.IntValue{Val: int64(math.Ceil(f))}, nil
				},
			},
			"round": {
				Name: "math.round",
				Fn: func(args []vm.Value) (vm.Value, error) {
					if len(args) != 1 {
						return nil, fmt.Errorf("math.round() takes 1 argument, got %d", len(args))
					}
					f, ok := toFloat(args[0])
					if !ok {
						return nil, fmt.Errorf("math.round() requires a number")
					}
					return &vm.IntValue{Val: int64(math.Round(f))}, nil
				},
			},
			"pow": {
				Name: "math.pow",
				Fn: func(args []vm.Value) (vm.Value, error) {
					if len(args) != 2 {
						return nil, fmt.Errorf("math.pow() takes 2 arguments, got %d", len(args))
					}
					base, ok1 := toFloat(args[0])
					exp, ok2 := toFloat(args[1])
					if !ok1 || !ok2 {
						return nil, fmt.Errorf("math.pow() requires numbers")
					}
					return &vm.FloatValue{Val: math.Pow(base, exp)}, nil
				},
			},
			"sin": {
				Name: "math.sin",
				Fn: func(args []vm.Value) (vm.Value, error) {
					if len(args) != 1 {
						return nil, fmt.Errorf("math.sin() takes 1 argument, got %d", len(args))
					}
					f, ok := toFloat(args[0])
					if !ok {
						return nil, fmt.Errorf("math.sin() requires a number")
					}
					return &vm.FloatValue{Val: math.Sin(f)}, nil
				},
			},
			"cos": {
				Name: "math.cos",
				Fn: func(args []vm.Value) (vm.Value, error) {
					if len(args) != 1 {
						return nil, fmt.Errorf("math.cos() takes 1 argument, got %d", len(args))
					}
					f, ok := toFloat(args[0])
					if !ok {
						return nil, fmt.Errorf("math.cos() requires a number")
					}
					return &vm.FloatValue{Val: math.Cos(f)}, nil
				},
			},
			"tan": {
				Name: "math.tan",
				Fn: func(args []vm.Value) (vm.Value, error) {
					if len(args) != 1 {
						return nil, fmt.Errorf("math.tan() takes 1 argument, got %d", len(args))
					}
					f, ok := toFloat(args[0])
					if !ok {
						return nil, fmt.Errorf("math.tan() requires a number")
					}
					return &vm.FloatValue{Val: math.Tan(f)}, nil
				},
			},
			"log": {
				Name: "math.log",
				Fn: func(args []vm.Value) (vm.Value, error) {
					if len(args) != 1 {
						return nil, fmt.Errorf("math.log() takes 1 argument, got %d", len(args))
					}
					f, ok := toFloat(args[0])
					if !ok {
						return nil, fmt.Errorf("math.log() requires a number")
					}
					return &vm.FloatValue{Val: math.Log(f)}, nil
				},
			},
			"log10": {
				Name: "math.log10",
				Fn: func(args []vm.Value) (vm.Value, error) {
					if len(args) != 1 {
						return nil, fmt.Errorf("math.log10() takes 1 argument, got %d", len(args))
					}
					f, ok := toFloat(args[0])
					if !ok {
						return nil, fmt.Errorf("math.log10() requires a number")
					}
					return &vm.FloatValue{Val: math.Log10(f)}, nil
				},
			},
			"min": {
				Name: "math.min",
				Fn: func(args []vm.Value) (vm.Value, error) {
					if len(args) != 2 {
						return nil, fmt.Errorf("math.min() takes 2 arguments, got %d", len(args))
					}
					// preserve int type when both are int
					ai, aIsInt := args[0].(*vm.IntValue)
					bi, bIsInt := args[1].(*vm.IntValue)
					if aIsInt && bIsInt {
						if ai.Val < bi.Val {
							return ai, nil
						}
						return bi, nil
					}
					a, ok1 := toFloat(args[0])
					b, ok2 := toFloat(args[1])
					if !ok1 || !ok2 {
						return nil, fmt.Errorf("math.min() requires numbers")
					}
					return &vm.FloatValue{Val: math.Min(a, b)}, nil
				},
			},
			"max": {
				Name: "math.max",
				Fn: func(args []vm.Value) (vm.Value, error) {
					if len(args) != 2 {
						return nil, fmt.Errorf("math.max() takes 2 arguments, got %d", len(args))
					}
					ai, aIsInt := args[0].(*vm.IntValue)
					bi, bIsInt := args[1].(*vm.IntValue)
					if aIsInt && bIsInt {
						if ai.Val > bi.Val {
							return ai, nil
						}
						return bi, nil
					}
					a, ok1 := toFloat(args[0])
					b, ok2 := toFloat(args[1])
					if !ok1 || !ok2 {
						return nil, fmt.Errorf("math.max() requires numbers")
					}
					return &vm.FloatValue{Val: math.Max(a, b)}, nil
				},
			},
		},
	}
}

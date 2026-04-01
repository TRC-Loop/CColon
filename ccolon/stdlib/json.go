package stdlib

import (
	"encoding/json"
	"fmt"

	"github.com/TRC-Loop/ccolon/vm"
)

func goToValue(v interface{}) vm.Value {
	switch val := v.(type) {
	case nil:
		return &vm.NilValue{}
	case bool:
		return &vm.BoolValue{Val: val}
	case float64:
		// JSON numbers are always float64 in Go
		if val == float64(int64(val)) {
			return &vm.IntValue{Val: int64(val)}
		}
		return &vm.FloatValue{Val: val}
	case string:
		return &vm.StringValue{Val: val}
	case []interface{}:
		elements := make([]vm.Value, len(val))
		for i, elem := range val {
			elements[i] = goToValue(elem)
		}
		return &vm.ListValue{Elements: elements}
	case map[string]interface{}:
		entries := make(map[string]vm.Value)
		order := make([]string, 0, len(val))
		for k, v := range val {
			order = append(order, k)
			entries[k] = goToValue(v)
		}
		return &vm.DictValue{Entries: entries, Order: order}
	default:
		return &vm.StringValue{Val: fmt.Sprintf("%v", val)}
	}
}

func valueToGo(v vm.Value) interface{} {
	switch val := v.(type) {
	case *vm.NilValue:
		return nil
	case *vm.BoolValue:
		return val.Val
	case *vm.IntValue:
		return val.Val
	case *vm.FloatValue:
		return val.Val
	case *vm.StringValue:
		return val.Val
	case *vm.ListValue:
		result := make([]interface{}, len(val.Elements))
		for i, elem := range val.Elements {
			result[i] = valueToGo(elem)
		}
		return result
	case *vm.DictValue:
		result := make(map[string]interface{})
		for k, v := range val.Entries {
			result[k] = valueToGo(v)
		}
		return result
	default:
		return v.String()
	}
}

func NewJsonModule() *vm.ModuleValue {
	return &vm.ModuleValue{
		Name: "json",
		Methods: map[string]*vm.NativeFuncValue{
			"parse": {
				Name: "json.parse",
				Fn: func(args []vm.Value) (vm.Value, error) {
					if len(args) != 1 {
						return nil, fmt.Errorf("json.parse() takes 1 argument, got %d", len(args))
					}
					str, ok := args[0].(*vm.StringValue)
					if !ok {
						return nil, fmt.Errorf("json.parse() requires a string")
					}
					var result interface{}
					if err := json.Unmarshal([]byte(str.Val), &result); err != nil {
						return nil, fmt.Errorf("json.parse() failed: %s", err.Error())
					}
					return goToValue(result), nil
				},
			},
			"stringify": {
				Name: "json.stringify",
				Fn: func(args []vm.Value) (vm.Value, error) {
					if len(args) != 1 {
						return nil, fmt.Errorf("json.stringify() takes 1 argument, got %d", len(args))
					}
					goVal := valueToGo(args[0])
					bytes, err := json.Marshal(goVal)
					if err != nil {
						return nil, fmt.Errorf("json.stringify() failed: %s", err.Error())
					}
					return &vm.StringValue{Val: string(bytes)}, nil
				},
			},
		},
	}
}

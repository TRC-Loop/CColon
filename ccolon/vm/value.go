package vm

import (
	"fmt"
	"strings"

	"github.com/TRC-Loop/ccolon/compiler"
)

type ValueType uint8

const (
	VAL_NIL ValueType = iota
	VAL_INT
	VAL_FLOAT
	VAL_BOOL
	VAL_STRING
	VAL_LIST
	VAL_ARRAY
	VAL_FUNC
	VAL_NATIVE_FUNC
	VAL_MODULE
)

type Value interface {
	Type() ValueType
	String() string
}

type NilValue struct{}
type IntValue struct{ Val int64 }
type FloatValue struct{ Val float64 }
type BoolValue struct{ Val bool }
type StringValue struct{ Val string }
type ListValue struct{ Elements []Value }
type ArrayValue struct{ Elements []Value }
type FuncValue struct{ Obj *compiler.FuncObject }
type NativeFuncValue struct {
	Name string
	Fn   func(args []Value) (Value, error)
}
type ModuleValue struct {
	Name    string
	Methods map[string]*NativeFuncValue
}

func (v *NilValue) Type() ValueType        { return VAL_NIL }
func (v *IntValue) Type() ValueType         { return VAL_INT }
func (v *FloatValue) Type() ValueType       { return VAL_FLOAT }
func (v *BoolValue) Type() ValueType        { return VAL_BOOL }
func (v *StringValue) Type() ValueType      { return VAL_STRING }
func (v *ListValue) Type() ValueType        { return VAL_LIST }
func (v *ArrayValue) Type() ValueType       { return VAL_ARRAY }
func (v *FuncValue) Type() ValueType        { return VAL_FUNC }
func (v *NativeFuncValue) Type() ValueType  { return VAL_NATIVE_FUNC }
func (v *ModuleValue) Type() ValueType      { return VAL_MODULE }

func (v *NilValue) String() string    { return "nil" }
func (v *IntValue) String() string    { return fmt.Sprintf("%d", v.Val) }
func (v *FloatValue) String() string  { return fmt.Sprintf("%g", v.Val) }
func (v *BoolValue) String() string   { return fmt.Sprintf("%t", v.Val) }
func (v *StringValue) String() string { return v.Val }
func (v *FuncValue) String() string   { return fmt.Sprintf("<function %s>", v.Obj.Name) }
func (v *NativeFuncValue) String() string { return fmt.Sprintf("<native %s>", v.Name) }
func (v *ModuleValue) String() string { return fmt.Sprintf("<module %s>", v.Name) }

func (v *ListValue) String() string {
	parts := make([]string, len(v.Elements))
	for i, e := range v.Elements {
		if s, ok := e.(*StringValue); ok {
			parts[i] = fmt.Sprintf("%q", s.Val)
		} else {
			parts[i] = e.String()
		}
	}
	return "[" + strings.Join(parts, ", ") + "]"
}

func (v *ArrayValue) String() string {
	parts := make([]string, len(v.Elements))
	for i, e := range v.Elements {
		if s, ok := e.(*StringValue); ok {
			parts[i] = fmt.Sprintf("%q", s.Val)
		} else {
			parts[i] = e.String()
		}
	}
	return "fixed([" + strings.Join(parts, ", ") + "])"
}

func IsTruthy(v Value) bool {
	switch val := v.(type) {
	case *NilValue:
		return false
	case *BoolValue:
		return val.Val
	case *IntValue:
		return val.Val != 0
	case *FloatValue:
		return val.Val != 0
	case *StringValue:
		return val.Val != ""
	case *ListValue:
		return len(val.Elements) > 0
	case *ArrayValue:
		return len(val.Elements) > 0
	default:
		return true
	}
}

func ValuesEqual(a, b Value) bool {
	if a.Type() != b.Type() {
		// int/float cross-comparison
		if ai, ok := a.(*IntValue); ok {
			if bf, ok := b.(*FloatValue); ok {
				return float64(ai.Val) == bf.Val
			}
		}
		if af, ok := a.(*FloatValue); ok {
			if bi, ok := b.(*IntValue); ok {
				return af.Val == float64(bi.Val)
			}
		}
		return false
	}
	switch av := a.(type) {
	case *NilValue:
		return true
	case *IntValue:
		return av.Val == b.(*IntValue).Val
	case *FloatValue:
		return av.Val == b.(*FloatValue).Val
	case *BoolValue:
		return av.Val == b.(*BoolValue).Val
	case *StringValue:
		return av.Val == b.(*StringValue).Val
	default:
		return a == b // pointer equality for reference types
	}
}

package vm

import (
	"bufio"
	"fmt"
	"math/big"
	"os"
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
	VAL_CLASS
	VAL_INSTANCE
	VAL_DICT
	VAL_FILE
	VAL_SINT
)

type Value interface {
	Type() ValueType
	String() string
}

type NilValue struct{}
type IntValue struct{ Val int64 }
type SintValue struct{ Val *big.Int }
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

type FieldDef struct {
	Visibility string
	TypeName   string
	Default    Value
}

type MethodDef struct {
	Visibility string
	Fn         *compiler.FuncObject
}

type ClassValue struct {
	Name      string
	Super     *ClassValue
	Fields    map[string]FieldDef
	Methods   map[string]*MethodDef
	InitArity int
	MaxArity  int
	InitDefs  []interface{}
}

type InstanceValue struct {
	Class  *ClassValue
	Fields map[string]Value
}

type DictValue struct {
	Entries map[string]Value
	Order   []string
}

type FileValue struct {
	Path    string
	Mode    string
	File    *os.File
	Scanner *bufio.Scanner
	Closed  bool
}

func (v *NilValue) Type() ValueType        { return VAL_NIL }
func (v *IntValue) Type() ValueType         { return VAL_INT }
func (v *SintValue) Type() ValueType       { return VAL_SINT }
func (v *FloatValue) Type() ValueType       { return VAL_FLOAT }
func (v *BoolValue) Type() ValueType        { return VAL_BOOL }
func (v *StringValue) Type() ValueType      { return VAL_STRING }
func (v *ListValue) Type() ValueType        { return VAL_LIST }
func (v *ArrayValue) Type() ValueType       { return VAL_ARRAY }
func (v *FuncValue) Type() ValueType        { return VAL_FUNC }
func (v *NativeFuncValue) Type() ValueType  { return VAL_NATIVE_FUNC }
func (v *ModuleValue) Type() ValueType      { return VAL_MODULE }
func (v *ClassValue) Type() ValueType       { return VAL_CLASS }
func (v *InstanceValue) Type() ValueType    { return VAL_INSTANCE }
func (v *DictValue) Type() ValueType        { return VAL_DICT }
func (v *FileValue) Type() ValueType        { return VAL_FILE }

func (v *NilValue) String() string    { return "nil" }
func (v *IntValue) String() string    { return fmt.Sprintf("%d", v.Val) }
func (v *SintValue) String() string   { return v.Val.String() }
func (v *FloatValue) String() string  { return fmt.Sprintf("%g", v.Val) }
func (v *BoolValue) String() string   { return fmt.Sprintf("%t", v.Val) }
func (v *StringValue) String() string { return v.Val }
func (v *FuncValue) String() string   { return fmt.Sprintf("<function %s>", v.Obj.Name) }
func (v *NativeFuncValue) String() string { return fmt.Sprintf("<native %s>", v.Name) }
func (v *ModuleValue) String() string  { return fmt.Sprintf("<module %s>", v.Name) }
func (v *ClassValue) String() string   { return fmt.Sprintf("<class %s>", v.Name) }

func (v *InstanceValue) String() string {
	parts := []string{}
	for name, val := range v.Fields {
		def, hasDef := v.Class.Fields[name]
		if hasDef && def.Visibility == "private" {
			continue
		}
		if s, ok := val.(*StringValue); ok {
			parts = append(parts, fmt.Sprintf("%s=%q", name, s.Val))
		} else {
			parts = append(parts, fmt.Sprintf("%s=%s", name, val.String()))
		}
	}
	methods := []string{}
	cls := v.Class
	for cls != nil {
		for name, m := range cls.Methods {
			if m.Visibility == "public" {
				methods = append(methods, name+"()")
			}
		}
		cls = cls.Super
	}
	result := fmt.Sprintf("<%s", v.Class.Name)
	if len(parts) > 0 {
		result += " " + strings.Join(parts, ", ")
	}
	if len(methods) > 0 {
		result += " | " + strings.Join(methods, ", ")
	}
	result += ">"
	return result
}

func (v *DictValue) String() string {
	parts := make([]string, len(v.Order))
	for i, key := range v.Order {
		val := v.Entries[key]
		if s, ok := val.(*StringValue); ok {
			parts[i] = fmt.Sprintf("%q: %q", key, s.Val)
		} else {
			parts[i] = fmt.Sprintf("%q: %s", key, val.String())
		}
	}
	return "{" + strings.Join(parts, ", ") + "}"
}

func (v *FileValue) String() string {
	state := "open"
	if v.Closed {
		state = "closed"
	}
	return fmt.Sprintf("<file '%s' mode='%s' %s>", v.Path, v.Mode, state)
}

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
	case *SintValue:
		return val.Val.Sign() != 0
	case *FloatValue:
		return val.Val != 0
	case *StringValue:
		return val.Val != ""
	case *ListValue:
		return len(val.Elements) > 0
	case *ArrayValue:
		return len(val.Elements) > 0
	case *DictValue:
		return len(val.Entries) > 0
	case *InstanceValue:
		return true
	case *ClassValue:
		return true
	default:
		return true
	}
}

func ValuesEqual(a, b Value) bool {
	if a.Type() != b.Type() {
		// int/float/sint cross-comparison
		if ai, ok := a.(*IntValue); ok {
			if bf, ok := b.(*FloatValue); ok {
				return float64(ai.Val) == bf.Val
			}
			if bs, ok := b.(*SintValue); ok {
				return bs.Val.Cmp(big.NewInt(ai.Val)) == 0
			}
		}
		if af, ok := a.(*FloatValue); ok {
			if bi, ok := b.(*IntValue); ok {
				return af.Val == float64(bi.Val)
			}
		}
		if as, ok := a.(*SintValue); ok {
			if bi, ok := b.(*IntValue); ok {
				return as.Val.Cmp(big.NewInt(bi.Val)) == 0
			}
		}
		return false
	}
	switch av := a.(type) {
	case *NilValue:
		return true
	case *IntValue:
		return av.Val == b.(*IntValue).Val
	case *SintValue:
		return av.Val.Cmp(b.(*SintValue).Val) == 0
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

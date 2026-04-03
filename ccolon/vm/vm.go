package vm

import (
	"bufio"
	"fmt"
	"io"
	"math"
	"math/big"
	"strings"

	"github.com/TRC-Loop/ccolon/compiler"
)

func addOverflows(a, b int64) bool {
	r := a + b
	return (a > 0 && b > 0 && r < 0) || (a < 0 && b < 0 && r > 0)
}

func subOverflows(a, b int64) bool {
	r := a - b
	return (a > 0 && b < 0 && r < 0) || (a < 0 && b > 0 && r > 0)
}

func mulOverflows(a, b int64) bool {
	if a == 0 || b == 0 {
		return false
	}
	r := a * b
	return r/a != b
}

type CallFrame struct {
	function *compiler.FuncObject
	ip       int
	basePtr  int
	isInit   bool
	instance *InstanceValue
}

type ExceptionHandler struct {
	CatchIP  int
	FramePtr int
	StackPtr int
	BasePtr  int
	FuncObj  *compiler.FuncObject
}

type VM struct {
	stack       []Value
	sp          int
	frames      []CallFrame
	fp          int
	globals       map[string]Value
	constants     map[string]bool
	modules       map[string]*ModuleValue
	imported      map[string]bool
	handlers      []ExceptionHandler
	FileLoader    func(path string) (*compiler.FuncObject, error)
	importedFiles map[string]bool
}

func New() *VM {
	return &VM{
		stack:         make([]Value, 512),
		sp:            0,
		frames:        make([]CallFrame, 0, 64),
		globals:       make(map[string]Value),
		constants:     make(map[string]bool),
		modules:       make(map[string]*ModuleValue),
		imported:      make(map[string]bool),
		importedFiles: make(map[string]bool),
	}
}

func (vm *VM) RegisterModule(name string, mod *ModuleValue) {
	vm.modules[name] = mod
}

func (vm *VM) GetModule(name string) *ModuleValue {
	return vm.modules[name]
}

func (vm *VM) GlobalNames() []string {
	names := make([]string, 0, len(vm.globals))
	for name := range vm.globals {
		names = append(names, name)
	}
	return names
}

func (vm *VM) GetGlobal(name string) Value {
	return vm.globals[name]
}

func (vm *VM) DeleteGlobal(name string) {
	delete(vm.globals, name)
}

func (vm *VM) push(v Value) {
	if vm.sp >= len(vm.stack) {
		vm.stack = append(vm.stack, make([]Value, 256)...)
	}
	vm.stack[vm.sp] = v
	vm.sp++
}

func (vm *VM) pop() Value {
	vm.sp--
	return vm.stack[vm.sp]
}

func (vm *VM) peek(distance int) Value {
	return vm.stack[vm.sp-1-distance]
}

func (vm *VM) frame() *CallFrame {
	return &vm.frames[vm.fp-1]
}

func (vm *VM) readByte() byte {
	f := vm.frame()
	if f.ip >= len(f.function.Code) {
		return byte(compiler.OP_HALT)
	}
	b := f.function.Code[f.ip]
	f.ip++
	return b
}

func (vm *VM) readUint16() int {
	hi := vm.readByte()
	lo := vm.readByte()
	return int(hi)<<8 | int(lo)
}

func (vm *VM) getConstant(idx int) interface{} {
	return vm.frame().function.Constants[idx]
}

func (vm *VM) runtimeError(format string, args ...interface{}) error {
	f := vm.frame()
	line := 0
	if f.ip > 0 && f.ip-1 < len(f.function.Lines) {
		line = f.function.Lines[f.ip-1]
	}
	msg := fmt.Sprintf(format, args...)
	return fmt.Errorf("runtime error at line %d: %s", line, msg)
}

func (vm *VM) Run(fn *compiler.FuncObject) error {
	vm.frames = append(vm.frames, CallFrame{
		function: fn,
		ip:       0,
		basePtr:  0,
	})
	vm.fp = len(vm.frames)
	return vm.execute()
}

// CallFunc calls a CColon function from Go code with the given arguments.
// This enables native-to-CColon callbacks (e.g. HTTP server handlers).
func (vm *VM) CallFunc(fn Value, args []Value) (Value, error) {
	funcVal, ok := fn.(*FuncValue)
	if !ok {
		return nil, fmt.Errorf("CallFunc: expected a function, got %s", fn.String())
	}

	obj := funcVal.Obj
	maxArity := obj.MaxArity
	if maxArity == 0 {
		maxArity = obj.Arity
	}

	// Push the function and args onto the stack
	vm.push(funcVal)
	for _, arg := range args {
		vm.push(arg)
	}
	argCount := len(args)

	// Fill defaults for missing optional args
	for i := argCount; i < maxArity; i++ {
		if obj.Defaults != nil && i < len(obj.Defaults) && obj.Defaults[i] != nil {
			vm.push(vm.wrapConstant(obj.Defaults[i]))
		} else {
			vm.push(&NilValue{})
		}
	}
	totalArgs := maxArity
	if totalArgs < argCount {
		totalArgs = argCount
	}

	// Set up the call frame (same logic as callFunc)
	basePtr := vm.sp - totalArgs
	for i := 0; i < totalArgs; i++ {
		vm.stack[basePtr-1+i] = vm.stack[basePtr+i]
	}
	vm.sp = basePtr - 1 + totalArgs
	newBase := basePtr - 1

	savedFP := vm.fp
	savedFrames := len(vm.frames)

	vm.frames = append(vm.frames, CallFrame{
		function: obj,
		ip:       0,
		basePtr:  newBase,
	})
	vm.fp = len(vm.frames)

	// Run until the frame returns
	if err := vm.execute(); err != nil {
		vm.frames = vm.frames[:savedFrames]
		vm.fp = savedFP
		return nil, err
	}

	// Restore frame state so we don't fall into the caller's bytecode
	vm.frames = vm.frames[:savedFrames]
	vm.fp = savedFP

	// The return value is on top of the stack
	if vm.sp > 0 {
		return vm.pop(), nil
	}
	return &NilValue{}, nil
}

func (vm *VM) execute() error {
	for {
		op := compiler.OpCode(vm.readByte())

		switch op {
		case compiler.OP_HALT:
			return nil

		case compiler.OP_CONST:
			idx := vm.readUint16()
			c := vm.getConstant(idx)
			vm.push(vm.wrapConstant(c))

		case compiler.OP_TRUE:
			vm.push(&BoolValue{Val: true})

		case compiler.OP_FALSE:
			vm.push(&BoolValue{Val: false})

		case compiler.OP_NIL:
			vm.push(&NilValue{})

		case compiler.OP_POP:
			vm.pop()

		case compiler.OP_DUP:
			vm.push(vm.peek(0))

		case compiler.OP_ADD:
			b := vm.pop()
			a := vm.pop()
			result, err := vm.opAdd(a, b)
			if err != nil {
				if e := vm.tryThrowError(err); e != nil {
					return e
				}
				continue
			}
			vm.push(result)

		case compiler.OP_SUB:
			b := vm.pop()
			a := vm.pop()
			result, err := vm.opArith(a, b, "-")
			if err != nil {
				if e := vm.tryThrowError(err); e != nil {
					return e
				}
				continue
			}
			vm.push(result)

		case compiler.OP_MUL:
			b := vm.pop()
			a := vm.pop()
			result, err := vm.opArith(a, b, "*")
			if err != nil {
				if e := vm.tryThrowError(err); e != nil {
					return e
				}
				continue
			}
			vm.push(result)

		case compiler.OP_DIV:
			b := vm.pop()
			a := vm.pop()
			result, err := vm.opDiv(a, b)
			if err != nil {
				if e := vm.tryThrowError(err); e != nil {
					return e
				}
				continue
			}
			vm.push(result)

		case compiler.OP_MOD:
			b := vm.pop()
			a := vm.pop()
			result, err := vm.opMod(a, b)
			if err != nil {
				if e := vm.tryThrowError(err); e != nil {
					return e
				}
				continue
			}
			vm.push(result)

		case compiler.OP_NEG:
			v := vm.pop()
			switch val := v.(type) {
			case *IntValue:
				if val.Val == math.MinInt64 {
					if e := vm.tryThrowError(vm.runtimeError("integer overflow")); e != nil {
						return e
					}
					continue
				}
				vm.push(&IntValue{Val: -val.Val})
			case *SintValue:
				vm.push(&SintValue{Val: new(big.Int).Neg(val.Val)})
			case *FloatValue:
				vm.push(&FloatValue{Val: -val.Val})
			default:
				return vm.runtimeError("cannot negate %s", v.String())
			}

		case compiler.OP_NOT:
			v := vm.pop()
			vm.push(&BoolValue{Val: !IsTruthy(v)})

		case compiler.OP_EQ:
			b := vm.pop()
			a := vm.pop()
			vm.push(&BoolValue{Val: ValuesEqual(a, b)})

		case compiler.OP_NEQ:
			b := vm.pop()
			a := vm.pop()
			vm.push(&BoolValue{Val: !ValuesEqual(a, b)})

		case compiler.OP_LT, compiler.OP_GT, compiler.OP_LTE, compiler.OP_GTE:
			b := vm.pop()
			a := vm.pop()
			result, err := vm.opCompare(a, b, op)
			if err != nil {
				return err
			}
			vm.push(result)

		case compiler.OP_JUMP:
			offset := vm.readUint16()
			vm.frame().ip += offset

		case compiler.OP_JUMP_IF_FALSE:
			offset := vm.readUint16()
			v := vm.pop()
			if !IsTruthy(v) {
				vm.frame().ip += offset
			}

		case compiler.OP_LOOP:
			offset := vm.readUint16()
			vm.frame().ip -= offset

		case compiler.OP_LOAD_LOCAL:
			slot := int(vm.readByte())
			vm.push(vm.stack[vm.frame().basePtr+slot])

		case compiler.OP_STORE_LOCAL:
			slot := int(vm.readByte())
			vm.stack[vm.frame().basePtr+slot] = vm.peek(0)
			vm.pop()

		case compiler.OP_LOAD_GLOBAL:
			idx := vm.readUint16()
			name := vm.getConstant(idx).(string)
			val, ok := vm.globals[name]
			if !ok {
				if vm.imported[name] {
					if mod, ok := vm.modules[name]; ok {
						vm.push(mod)
						continue
					}
				}
				// friendly hint for common modules
			if name == "console" || name == "math" || name == "random" || name == "json" || name == "fs" {
				return vm.runtimeError("undefined variable '%s' -- did you forget 'import %s'?", name, name)
			}
			return vm.runtimeError("undefined variable '%s'", name)
			}
			vm.push(val)

		case compiler.OP_STORE_GLOBAL:
			idx := vm.readUint16()
			name := vm.getConstant(idx).(string)
			if vm.constants[name] {
				vm.pop() // discard the value
				return vm.runtimeError("cannot reassign constant '%s'", name)
			}
			vm.globals[name] = vm.pop()

		case compiler.OP_MARK_CONST:
			idx := vm.readUint16()
			name := vm.getConstant(idx).(string)
			vm.constants[name] = true

		case compiler.OP_CALL:
			argCount := int(vm.readByte())
			callee := vm.stack[vm.sp-1-argCount]
			if err := vm.callValue(callee, argCount); err != nil {
				return err
			}

		case compiler.OP_CALL_KW:
			posCount := int(vm.readByte())
			namedCount := int(vm.readByte())
			namesIdx := vm.readUint16()
			names := vm.getConstant(namesIdx).([]string)

			// pop named arg values (in order)
			namedValues := make([]Value, namedCount)
			for i := namedCount - 1; i >= 0; i-- {
				namedValues[i] = vm.pop()
			}
			// pop positional args
			posValues := make([]Value, posCount)
			for i := posCount - 1; i >= 0; i-- {
				posValues[i] = vm.pop()
			}
			callee := vm.pop() // pop the function

			// resolve the function to get param names
			funcVal, ok := callee.(*FuncValue)
			if !ok {
				return vm.runtimeError("keyword arguments are only supported on user-defined functions")
			}
			paramNames := funcVal.Obj.ParamNames
			maxArity := funcVal.Obj.MaxArity
			if maxArity == 0 {
				maxArity = funcVal.Obj.Arity
			}

			// build the full arg list
			fullArgs := make([]Value, maxArity)
			// fill positional first
			for i := 0; i < posCount && i < maxArity; i++ {
				fullArgs[i] = posValues[i]
			}
			// fill named args by matching param names
			for i, name := range names {
				found := false
				for j, pn := range paramNames {
					if pn == name {
						if fullArgs[j] != nil {
							return vm.runtimeError("duplicate argument for parameter '%s'", name)
						}
						fullArgs[j] = namedValues[i]
						found = true
						break
					}
				}
				if !found {
					return vm.runtimeError("unknown keyword argument '%s'", name)
				}
			}
			// fill defaults for any remaining nil slots
			for i := 0; i < maxArity; i++ {
				if fullArgs[i] == nil {
					if funcVal.Obj.Defaults != nil && i < len(funcVal.Obj.Defaults) && funcVal.Obj.Defaults[i] != nil {
						fullArgs[i] = vm.wrapConstant(funcVal.Obj.Defaults[i])
					} else if i >= funcVal.Obj.Arity {
						fullArgs[i] = &NilValue{}
					} else {
						return vm.runtimeError("missing required argument '%s'", paramNames[i])
					}
				}
			}

			// push function and resolved args back onto the stack
			// callValue expects stack: [..., callee, arg0, arg1, ...]
			vm.push(callee)
			for _, a := range fullArgs {
				vm.push(a)
			}
			totalArgs := len(fullArgs)
			calleeOnStack := vm.stack[vm.sp-1-totalArgs]
			if err := vm.callValue(calleeOnStack, totalArgs); err != nil {
				return err
			}

		case compiler.OP_RETURN:
			result := vm.pop()
			frame := vm.frame()
			isInit := frame.isInit
			instance := frame.instance
			vm.sp = frame.basePtr
			vm.fp--
			vm.frames = vm.frames[:vm.fp]
			if vm.fp == 0 {
				return nil
			}
			if isInit {
				vm.push(instance)
			} else {
				vm.push(result)
			}

		case compiler.OP_LIST_NEW:
			count := vm.readUint16()
			elements := make([]Value, count)
			for i := count - 1; i >= 0; i-- {
				elements[i] = vm.pop()
			}
			vm.push(&ListValue{Elements: elements})

		case compiler.OP_ARRAY_NEW:
			count := vm.readUint16()
			elements := make([]Value, count)
			for i := count - 1; i >= 0; i-- {
				elements[i] = vm.pop()
			}
			vm.push(&ArrayValue{Elements: elements})

		case compiler.OP_INDEX_GET:
			index := vm.pop()
			object := vm.pop()
			result, err := vm.indexGet(object, index)
			if err != nil {
				return err
			}
			vm.push(result)

		case compiler.OP_INDEX_SET:
			value := vm.pop()
			index := vm.pop()
			object := vm.pop()
			if err := vm.indexSet(object, index, value); err != nil {
				return err
			}

		case compiler.OP_METHOD_CALL:
			methodIdx := vm.readUint16()
			argCount := int(vm.readByte())
			methodName := vm.frame().function.Constants[methodIdx].(string)
			// object is below the args on the stack
			object := vm.stack[vm.sp-1-argCount]

			// for instance method calls, push a call frame instead of calling synchronously
			if inst, ok := object.(*InstanceValue); ok {
				method := vm.resolveInstanceMethod(inst, methodName)
				if method != nil {
					fn := method.Fn
					maxArity := fn.MaxArity
					if maxArity == 0 {
						maxArity = fn.Arity
					}

					// instance position (before any args)
					instancePos := vm.sp - 1 - argCount

					// fill defaults for missing optional params (self + user args)
					for i := argCount + 1; i < maxArity; i++ {
						if fn.Defaults != nil && i < len(fn.Defaults) && fn.Defaults[i] != nil {
							vm.push(vm.wrapConstant(fn.Defaults[i]))
						} else {
							vm.push(&NilValue{})
						}
					}

					newBase := instancePos

					vm.frames = append(vm.frames, CallFrame{
						function: fn,
						ip:       0,
						basePtr:  newBase,
					})
					vm.fp = len(vm.frames)
					continue
				}

				// fallback: built-in methods like tostring
				args := make([]Value, argCount)
				for i := argCount - 1; i >= 0; i-- {
					args[i] = vm.pop()
				}
				vm.pop()
				result, err := vm.instanceBuiltinMethod(inst, methodName, args)
				if err != nil {
					return err
				}
				vm.push(result)
				continue
			}

			args := make([]Value, argCount)
			for i := argCount - 1; i >= 0; i-- {
				args[i] = vm.pop()
			}
			vm.pop() // pop the object
			result, err := vm.invokeMethod(object, methodName, args)
			if err != nil {
				return err
			}
			vm.push(result)

		case compiler.OP_IMPORT:
			idx := vm.readUint16()
			moduleName := vm.getConstant(idx).(string)
			if _, ok := vm.modules[moduleName]; !ok {
				return vm.runtimeError("unknown module '%s'", moduleName)
			}
			vm.imported[moduleName] = true

		case compiler.OP_FROM_IMPORT:
			moduleIdx := vm.readUint16()
			namesIdx := vm.readUint16()
			moduleName := vm.getConstant(moduleIdx).(string)
			names := vm.getConstant(namesIdx).([]string)
			mod, ok := vm.modules[moduleName]
			if !ok {
				return vm.runtimeError("unknown module '%s'", moduleName)
			}
			vm.imported[moduleName] = true
			if len(names) == 1 && names[0] == "*" {
				// import all methods as globals
				for name, fn := range mod.Methods {
					vm.globals[name] = fn
				}
				if mod.Properties != nil {
					for name, val := range mod.Properties {
						vm.globals[name] = val
					}
				}
			} else {
				for _, name := range names {
					if fn, ok := mod.Methods[name]; ok {
						vm.globals[name] = fn
					} else if mod.Properties != nil {
						if val, ok := mod.Properties[name]; ok {
							vm.globals[name] = val
						} else {
							return vm.runtimeError("module '%s' has no member '%s'", moduleName, name)
						}
					} else {
						return vm.runtimeError("module '%s' has no member '%s'", moduleName, name)
					}
				}
			}

		case compiler.OP_IMPORT_FILE:
			idx := vm.readUint16()
			filePath := vm.getConstant(idx).(string)
			if !vm.importedFiles[filePath] {
				vm.importedFiles[filePath] = true
				if vm.FileLoader == nil {
					return vm.runtimeError("file imports are not supported in this context")
				}
				fn, err := vm.FileLoader(filePath)
				if err != nil {
					return vm.runtimeError("import '%s': %s", filePath, err.Error())
				}
				// Save frame state, run imported file, restore
				savedFP := vm.fp
				savedFrames := len(vm.frames)
				vm.frames = append(vm.frames, CallFrame{
					function: fn,
					ip:       0,
					basePtr:  vm.sp,
				})
				vm.fp = len(vm.frames)
				if err := vm.execute(); err != nil {
					return err
				}
				vm.frames = vm.frames[:savedFrames]
				vm.fp = savedFP
			}

		case compiler.OP_GET_FIELD:
			idx := vm.readUint16()
			fieldName := vm.getConstant(idx).(string)
			object := vm.pop()
			switch obj := object.(type) {
			case *InstanceValue:
				val, exists := obj.Fields[fieldName]
				if !exists {
					return vm.runtimeError("instance of '%s' has no field '%s'", obj.Class.Name, fieldName)
				}
				vm.push(val)
			case *ModuleValue:
				if obj.Properties != nil {
					if val, exists := obj.Properties[fieldName]; exists {
						vm.push(val)
						break
					}
				}
				return vm.runtimeError("module '%s' has no property '%s'", obj.Name, fieldName)
			default:
				return vm.runtimeError("cannot access field '%s' on %s", fieldName, object.String())
			}

		case compiler.OP_SET_FIELD:
			idx := vm.readUint16()
			fieldName := vm.getConstant(idx).(string)
			value := vm.pop()
			object := vm.pop()
			inst, ok := object.(*InstanceValue)
			if !ok {
				return vm.runtimeError("cannot set field '%s' on %s", fieldName, object.String())
			}
			inst.Fields[fieldName] = value

		case compiler.OP_INHERIT:
			superVal := vm.pop()
			classDefVal := vm.pop()
			superClass, ok := superVal.(*ClassValue)
			if !ok {
				return vm.runtimeError("superclass must be a class")
			}
			classDef, ok := classDefVal.(*ClassValue)
			if !ok {
				return vm.runtimeError("expected class for inheritance")
			}
			classDef.Super = superClass
			// copy parent fields not already defined
			for name, field := range superClass.Fields {
				if _, exists := classDef.Fields[name]; !exists {
					classDef.Fields[name] = field
				}
			}
			// copy parent methods not already defined
			for name, method := range superClass.Methods {
				if _, exists := classDef.Methods[name]; !exists {
					classDef.Methods[name] = method
				}
			}
			vm.push(classDef)

		case compiler.OP_DICT_NEW:
			count := vm.readUint16()
			entries := make(map[string]Value)
			order := make([]string, count)
			// stack has: key0, val0, key1, val1, ... (count pairs)
			items := make([]Value, count*2)
			for i := count*2 - 1; i >= 0; i-- {
				items[i] = vm.pop()
			}
			for i := 0; i < count; i++ {
				key, ok := items[i*2].(*StringValue)
				if !ok {
					return vm.runtimeError("dict key must be a string, got %s", items[i*2].String())
				}
				order[i] = key.Val
				entries[key.Val] = items[i*2+1]
			}
			vm.push(&DictValue{Entries: entries, Order: order})

		case compiler.OP_TRY_BEGIN:
			offset := vm.readUint16()
			f := vm.frame()
			vm.handlers = append(vm.handlers, ExceptionHandler{
				CatchIP:  f.ip + offset,
				FramePtr: vm.fp,
				StackPtr: vm.sp,
				BasePtr:  f.basePtr,
				FuncObj:  f.function,
			})

		case compiler.OP_TRY_END:
			if len(vm.handlers) > 0 {
				vm.handlers = vm.handlers[:len(vm.handlers)-1]
			}

		case compiler.OP_THROW:
			thrown := vm.pop()
			if err := vm.throwValue(thrown); err != nil {
				return err
			}

		default:
			return vm.runtimeError("unknown opcode %d", op)
		}
	}
}

func (vm *VM) wrapConstant(c interface{}) Value {
	switch v := c.(type) {
	case int64:
		return &IntValue{Val: v}
	case *big.Int:
		return &SintValue{Val: v}
	case float64:
		return &FloatValue{Val: v}
	case string:
		return &StringValue{Val: v}
	case bool:
		if v {
			return &BoolValue{Val: true}
		}
		return &BoolValue{Val: false}
	case *compiler.FuncObject:
		return &FuncValue{Obj: v}
	case *compiler.ClassDef:
		return vm.wrapClassDef(v)
	default:
		return &NilValue{}
	}
}

func (vm *VM) callValue(callee Value, argCount int) error {
	switch fn := callee.(type) {
	case *FuncValue:
		return vm.callFunc(fn.Obj, argCount)

	case *NativeFuncValue:
		args := make([]Value, argCount)
		for i := argCount - 1; i >= 0; i-- {
			args[i] = vm.pop()
		}
		vm.pop() // pop the function itself
		result, err := fn.Fn(args)
		if err != nil {
			return vm.runtimeError("%s", err.Error())
		}
		vm.push(result)
		return nil

	case *ClassValue:
		return vm.instantiateClass(fn, argCount)

	default:
		return vm.runtimeError("cannot call %s", callee.String())
	}
}

func (vm *VM) callFunc(obj *compiler.FuncObject, argCount int) error {
	maxArity := obj.MaxArity
	if maxArity == 0 {
		maxArity = obj.Arity
	}

	if argCount < obj.Arity || argCount > maxArity {
		if obj.Arity == maxArity {
			return vm.runtimeError("function '%s' expects %d arguments, got %d",
				obj.Name, obj.Arity, argCount)
		}
		return vm.runtimeError("function '%s' expects %d to %d arguments, got %d",
			obj.Name, obj.Arity, maxArity, argCount)
	}

	// fill in default values for missing optional args
	for i := argCount; i < maxArity; i++ {
		if obj.Defaults != nil && i < len(obj.Defaults) && obj.Defaults[i] != nil {
			vm.push(vm.wrapConstant(obj.Defaults[i]))
		} else {
			vm.push(&NilValue{})
		}
	}
	totalArgs := maxArity

	// stack: [..., callee, arg0, arg1, ..., argN]
	basePtr := vm.sp - totalArgs
	for i := 0; i < totalArgs; i++ {
		vm.stack[basePtr-1+i] = vm.stack[basePtr+i]
	}
	vm.sp = basePtr - 1 + totalArgs
	newBase := basePtr - 1

	vm.frames = append(vm.frames, CallFrame{
		function: obj,
		ip:       0,
		basePtr:  newBase,
	})
	vm.fp = len(vm.frames)
	return nil
}

func (vm *VM) instantiateClass(class *ClassValue, argCount int) error {
	// create instance with default field values
	instance := &InstanceValue{
		Class:  class,
		Fields: make(map[string]Value),
	}

	// walk class hierarchy to collect all fields with defaults
	cls := class
	for cls != nil {
		for name, field := range cls.Fields {
			if _, exists := instance.Fields[name]; !exists {
				if field.Default != nil {
					instance.Fields[name] = field.Default
				} else {
					instance.Fields[name] = &NilValue{}
				}
			}
		}
		cls = cls.Super
	}

	// find init method
	initMethod := vm.findMethod(class, "init")
	if initMethod != nil {
		fn := initMethod.Fn
		maxArity := fn.MaxArity
		if maxArity == 0 {
			maxArity = fn.Arity
		}

		// stack: [..., classValue, arg0, ..., argN]
		// Replace classValue with instance (self = slot 0)
		vm.stack[vm.sp-1-argCount] = instance

		// fill defaults for missing optional params (self + user args)
		for i := argCount + 1; i < maxArity; i++ {
			if fn.Defaults != nil && i < len(fn.Defaults) && fn.Defaults[i] != nil {
				vm.push(vm.wrapConstant(fn.Defaults[i]))
			} else {
				vm.push(&NilValue{})
			}
		}

		newBase := vm.sp - 1 - argCount
		if maxArity > argCount+1 {
			newBase = vm.sp - maxArity
		}

		vm.frames = append(vm.frames, CallFrame{
			function: fn,
			ip:       0,
			basePtr:  newBase,
			isInit:   true,
			instance: instance,
		})
		vm.fp = len(vm.frames)
		return nil
	}

	// no init: pop args and class, push instance
	if argCount > 0 {
		return vm.runtimeError("class '%s' has no init and takes no arguments", class.Name)
	}
	vm.pop() // pop class
	vm.push(instance)
	return nil
}

// toBigInt converts int or sint to *big.Int, returns nil if not numeric integer.
func toBigInt(v Value) *big.Int {
	switch val := v.(type) {
	case *IntValue:
		return big.NewInt(val.Val)
	case *SintValue:
		return new(big.Int).Set(val.Val)
	}
	return nil
}

func (vm *VM) opAdd(a, b Value) (Value, error) {
	// sint promotion: if either side is sint, promote both
	if _, ok := a.(*SintValue); ok {
		if bb := toBigInt(b); bb != nil {
			return &SintValue{Val: new(big.Int).Add(toBigInt(a), bb)}, nil
		}
	}
	if _, ok := b.(*SintValue); ok {
		if ab := toBigInt(a); ab != nil {
			return &SintValue{Val: new(big.Int).Add(ab, toBigInt(b))}, nil
		}
	}
	switch av := a.(type) {
	case *IntValue:
		switch bv := b.(type) {
		case *IntValue:
			if addOverflows(av.Val, bv.Val) {
				return nil, vm.runtimeError("integer overflow")
			}
			return &IntValue{Val: av.Val + bv.Val}, nil
		case *FloatValue:
			return &FloatValue{Val: float64(av.Val) + bv.Val}, nil
		}
	case *FloatValue:
		switch bv := b.(type) {
		case *IntValue:
			return &FloatValue{Val: av.Val + float64(bv.Val)}, nil
		case *FloatValue:
			return &FloatValue{Val: av.Val + bv.Val}, nil
		}
	case *StringValue:
		if bv, ok := b.(*StringValue); ok {
			return &StringValue{Val: av.Val + bv.Val}, nil
		}
		return &StringValue{Val: av.Val + b.String()}, nil
	}
	// If right side is string, coerce left to string
	if bv, ok := b.(*StringValue); ok {
		return &StringValue{Val: a.String() + bv.Val}, nil
	}
	return nil, vm.runtimeError("cannot add %s and %s", a.String(), b.String())
}

func (vm *VM) opArith(a, b Value, op string) (Value, error) {
	// sint promotion
	ab, bb := toBigInt(a), toBigInt(b)
	if ab != nil && bb != nil {
		if _, isSint := a.(*SintValue); isSint {
			switch op {
			case "-":
				return &SintValue{Val: new(big.Int).Sub(ab, bb)}, nil
			case "*":
				return &SintValue{Val: new(big.Int).Mul(ab, bb)}, nil
			}
		}
		if _, isSint := b.(*SintValue); isSint {
			switch op {
			case "-":
				return &SintValue{Val: new(big.Int).Sub(ab, bb)}, nil
			case "*":
				return &SintValue{Val: new(big.Int).Mul(ab, bb)}, nil
			}
		}
	}
	switch av := a.(type) {
	case *IntValue:
		switch bv := b.(type) {
		case *IntValue:
			switch op {
			case "-":
				if subOverflows(av.Val, bv.Val) {
					return nil, vm.runtimeError("integer overflow")
				}
				return &IntValue{Val: av.Val - bv.Val}, nil
			case "*":
				if mulOverflows(av.Val, bv.Val) {
					return nil, vm.runtimeError("integer overflow")
				}
				return &IntValue{Val: av.Val * bv.Val}, nil
			}
		case *FloatValue:
			af := float64(av.Val)
			switch op {
			case "-":
				return &FloatValue{Val: af - bv.Val}, nil
			case "*":
				return &FloatValue{Val: af * bv.Val}, nil
			}
		}
	case *FloatValue:
		switch bv := b.(type) {
		case *IntValue:
			bf := float64(bv.Val)
			switch op {
			case "-":
				return &FloatValue{Val: av.Val - bf}, nil
			case "*":
				return &FloatValue{Val: av.Val * bf}, nil
			}
		case *FloatValue:
			switch op {
			case "-":
				return &FloatValue{Val: av.Val - bv.Val}, nil
			case "*":
				return &FloatValue{Val: av.Val * bv.Val}, nil
			}
		}
	}
	return nil, vm.runtimeError("cannot %s %s and %s", op, a.String(), b.String())
}

func (vm *VM) opDiv(a, b Value) (Value, error) {
	// sint division
	ab, bb := toBigInt(a), toBigInt(b)
	if ab != nil && bb != nil {
		if _, ok := a.(*SintValue); ok {
			if bb.Sign() == 0 {
				return nil, vm.runtimeError("division by zero")
			}
			return &SintValue{Val: new(big.Int).Div(ab, bb)}, nil
		}
		if _, ok := b.(*SintValue); ok {
			if bb.Sign() == 0 {
				return nil, vm.runtimeError("division by zero")
			}
			return &SintValue{Val: new(big.Int).Div(ab, bb)}, nil
		}
	}
	switch av := a.(type) {
	case *IntValue:
		switch bv := b.(type) {
		case *IntValue:
			if bv.Val == 0 {
				return nil, vm.runtimeError("division by zero")
			}
			return &IntValue{Val: av.Val / bv.Val}, nil
		case *FloatValue:
			if bv.Val == 0 {
				return nil, vm.runtimeError("division by zero")
			}
			return &FloatValue{Val: float64(av.Val) / bv.Val}, nil
		}
	case *FloatValue:
		switch bv := b.(type) {
		case *IntValue:
			if bv.Val == 0 {
				return nil, vm.runtimeError("division by zero")
			}
			return &FloatValue{Val: av.Val / float64(bv.Val)}, nil
		case *FloatValue:
			if bv.Val == 0 {
				return nil, vm.runtimeError("division by zero")
			}
			return &FloatValue{Val: av.Val / bv.Val}, nil
		}
	}
	return nil, vm.runtimeError("cannot divide %s by %s", a.String(), b.String())
}

func (vm *VM) opMod(a, b Value) (Value, error) {
	// sint modulo
	ab, bb := toBigInt(a), toBigInt(b)
	if ab != nil && bb != nil {
		_, aIsSint := a.(*SintValue)
		_, bIsSint := b.(*SintValue)
		if aIsSint || bIsSint {
			if bb.Sign() == 0 {
				return nil, vm.runtimeError("modulo by zero")
			}
			return &SintValue{Val: new(big.Int).Mod(ab, bb)}, nil
		}
	}
	ai, aOk := a.(*IntValue)
	bi, bOk := b.(*IntValue)
	if aOk && bOk {
		if bi.Val == 0 {
			return nil, vm.runtimeError("modulo by zero")
		}
		return &IntValue{Val: ai.Val % bi.Val}, nil
	}
	return nil, vm.runtimeError("modulo requires integers, got %s and %s", a.String(), b.String())
}

func (vm *VM) opCompare(a, b Value, op compiler.OpCode) (*BoolValue, error) {
	// sint comparison
	ab, bb := toBigInt(a), toBigInt(b)
	_, aIsSint := a.(*SintValue)
	_, bIsSint := b.(*SintValue)
	if (aIsSint || bIsSint) && ab != nil && bb != nil {
		cmp := ab.Cmp(bb)
		var result bool
		switch op {
		case compiler.OP_LT:
			result = cmp < 0
		case compiler.OP_GT:
			result = cmp > 0
		case compiler.OP_LTE:
			result = cmp <= 0
		case compiler.OP_GTE:
			result = cmp >= 0
		}
		return &BoolValue{Val: result}, nil
	}

	var af, bf float64
	switch av := a.(type) {
	case *IntValue:
		af = float64(av.Val)
	case *FloatValue:
		af = av.Val
	default:
		return nil, vm.runtimeError("cannot compare %s", a.String())
	}
	switch bv := b.(type) {
	case *IntValue:
		bf = float64(bv.Val)
	case *FloatValue:
		bf = bv.Val
	default:
		return nil, vm.runtimeError("cannot compare %s", b.String())
	}

	var result bool
	switch op {
	case compiler.OP_LT:
		result = af < bf
	case compiler.OP_GT:
		result = af > bf
	case compiler.OP_LTE:
		result = af <= bf
	case compiler.OP_GTE:
		result = af >= bf
	}
	return &BoolValue{Val: result}, nil
}

func (vm *VM) indexGet(object, index Value) (Value, error) {
	// dict indexing with string key
	if dict, ok := object.(*DictValue); ok {
		if strKey, ok := index.(*StringValue); ok {
			val, exists := dict.Entries[strKey.Val]
			if !exists {
				return nil, vm.runtimeError("key '%s' not found in dict", strKey.Val)
			}
			return val, nil
		}
		// integer index for dict returns nth key (for for-in iteration)
		if intKey, ok := index.(*IntValue); ok {
			i := int(intKey.Val)
			if i < 0 || i >= len(dict.Order) {
				return nil, vm.runtimeError("dict index %d out of range (length %d)", i, len(dict.Order))
			}
			return &StringValue{Val: dict.Order[i]}, nil
		}
		return nil, vm.runtimeError("dict index must be a string or integer")
	}

	idx, ok := index.(*IntValue)
	if !ok {
		return nil, vm.runtimeError("index must be an integer, got %s", index.String())
	}
	i := int(idx.Val)

	switch obj := object.(type) {
	case *ListValue:
		if i < 0 || i >= len(obj.Elements) {
			return nil, vm.runtimeError("list index %d out of range (length %d)", i, len(obj.Elements))
		}
		return obj.Elements[i], nil
	case *ArrayValue:
		if i < 0 || i >= len(obj.Elements) {
			return nil, vm.runtimeError("array index %d out of range (length %d)", i, len(obj.Elements))
		}
		return obj.Elements[i], nil
	case *StringValue:
		if i < 0 || i >= len(obj.Val) {
			return nil, vm.runtimeError("string index %d out of range (length %d)", i, len(obj.Val))
		}
		return &StringValue{Val: string(obj.Val[i])}, nil
	default:
		return nil, vm.runtimeError("cannot index %s", object.String())
	}
}

func (vm *VM) indexSet(object, index, value Value) error {
	// dict indexing with string key
	if dict, ok := object.(*DictValue); ok {
		strKey, ok := index.(*StringValue)
		if !ok {
			return vm.runtimeError("dict key must be a string")
		}
		if _, exists := dict.Entries[strKey.Val]; !exists {
			dict.Order = append(dict.Order, strKey.Val)
		}
		dict.Entries[strKey.Val] = value
		return nil
	}

	idx, ok := index.(*IntValue)
	if !ok {
		return vm.runtimeError("index must be an integer")
	}
	i := int(idx.Val)

	switch obj := object.(type) {
	case *ListValue:
		if i < 0 || i >= len(obj.Elements) {
			return vm.runtimeError("list index %d out of range (length %d)", i, len(obj.Elements))
		}
		obj.Elements[i] = value
		return nil
	case *ArrayValue:
		if i < 0 || i >= len(obj.Elements) {
			return vm.runtimeError("array index %d out of range (length %d)", i, len(obj.Elements))
		}
		obj.Elements[i] = value
		return nil
	default:
		return vm.runtimeError("cannot set index on %s", object.String())
	}
}

func (vm *VM) invokeMethod(object Value, method string, args []Value) (Value, error) {
	// module methods
	if mod, ok := object.(*ModuleValue); ok {
		fn, exists := mod.Methods[method]
		if !exists {
			return nil, vm.runtimeError("module '%s' has no method '%s'", mod.Name, method)
		}
		return fn.Fn(args)
	}

	// built-in methods on value types
	switch obj := object.(type) {
	case *IntValue:
		return vm.intMethod(obj, method, args)
	case *SintValue:
		return vm.sintMethod(obj, method, args)
	case *FloatValue:
		return vm.floatMethod(obj, method, args)
	case *StringValue:
		return vm.stringMethod(obj, method, args)
	case *BoolValue:
		return vm.boolMethod(obj, method, args)
	case *ListValue:
		return vm.listMethod(obj, method, args)
	case *ArrayValue:
		return vm.arrayMethod(obj, method, args)
	case *DictValue:
		return vm.dictMethod(obj, method, args)
	case *FileValue:
		return vm.fileMethod(obj, method, args)
	default:
		return nil, vm.runtimeError("cannot call method '%s' on %s", method, object.String())
	}
}

func (vm *VM) intMethod(v *IntValue, method string, args []Value) (Value, error) {
	switch method {
	case "tostring":
		return &StringValue{Val: v.String()}, nil
	case "tofloat":
		return &FloatValue{Val: float64(v.Val)}, nil
	case "tosint":
		return &SintValue{Val: big.NewInt(v.Val)}, nil
	default:
		return nil, vm.runtimeError("int has no method '%s'", method)
	}
}

func (vm *VM) sintMethod(v *SintValue, method string, args []Value) (Value, error) {
	switch method {
	case "tostring":
		return &StringValue{Val: v.Val.String()}, nil
	case "toint":
		if v.Val.IsInt64() {
			return &IntValue{Val: v.Val.Int64()}, nil
		}
		return nil, vm.runtimeError("sint value too large for int")
	case "tofloat":
		f, _ := new(big.Float).SetInt(v.Val).Float64()
		return &FloatValue{Val: f}, nil
	case "abs":
		return &SintValue{Val: new(big.Int).Abs(v.Val)}, nil
	case "pow":
		if len(args) != 1 {
			return nil, vm.runtimeError("sint.pow() takes 1 argument")
		}
		exp := toBigInt(args[0])
		if exp == nil {
			return nil, vm.runtimeError("sint.pow() requires an integer exponent")
		}
		return &SintValue{Val: new(big.Int).Exp(v.Val, exp, nil)}, nil
	default:
		return nil, vm.runtimeError("sint has no method '%s'", method)
	}
}

func (vm *VM) floatMethod(v *FloatValue, method string, args []Value) (Value, error) {
	switch method {
	case "tostring":
		return &StringValue{Val: v.String()}, nil
	case "toint":
		return &IntValue{Val: int64(v.Val)}, nil
	default:
		return nil, vm.runtimeError("float has no method '%s'", method)
	}
}

func (vm *VM) stringMethod(v *StringValue, method string, args []Value) (Value, error) {
	switch method {
	case "length":
		return &IntValue{Val: int64(len(v.Val))}, nil
	case "tostring":
		return v, nil
	case "toint":
		var n int64
		_, err := fmt.Sscanf(v.Val, "%d", &n)
		if err != nil {
			return nil, vm.runtimeError("cannot convert '%s' to int", v.Val)
		}
		return &IntValue{Val: n}, nil
	case "tofloat":
		var f float64
		_, err := fmt.Sscanf(v.Val, "%f", &f)
		if err != nil {
			return nil, vm.runtimeError("cannot convert '%s' to float", v.Val)
		}
		return &FloatValue{Val: f}, nil
	case "tosint":
		bi := new(big.Int)
		if _, ok := bi.SetString(v.Val, 10); !ok {
			return nil, vm.runtimeError("cannot convert '%s' to sint", v.Val)
		}
		return &SintValue{Val: bi}, nil
	case "split":
		sep := ""
		if len(args) > 0 {
			s, ok := args[0].(*StringValue)
			if !ok {
				return nil, vm.runtimeError("split() separator must be a string")
			}
			sep = s.Val
		}
		var parts []string
		if sep == "" {
			// split on whitespace
			parts = strings.Fields(v.Val)
		} else {
			parts = strings.Split(v.Val, sep)
		}
		elems := make([]Value, len(parts))
		for i, p := range parts {
			elems[i] = &StringValue{Val: p}
		}
		return &ListValue{Elements: elems}, nil
	case "reverse":
		runes := []rune(v.Val)
		for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
			runes[i], runes[j] = runes[j], runes[i]
		}
		return &StringValue{Val: string(runes)}, nil
	case "upper":
		return &StringValue{Val: strings.ToUpper(v.Val)}, nil
	case "lower":
		return &StringValue{Val: strings.ToLower(v.Val)}, nil
	case "trim":
		return &StringValue{Val: strings.TrimSpace(v.Val)}, nil
	case "contains":
		if len(args) != 1 {
			return nil, vm.runtimeError("contains() takes 1 argument")
		}
		s, ok := args[0].(*StringValue)
		if !ok {
			return nil, vm.runtimeError("contains() argument must be a string")
		}
		return &BoolValue{Val: strings.Contains(v.Val, s.Val)}, nil
	case "startswith":
		if len(args) != 1 {
			return nil, vm.runtimeError("startswith() takes 1 argument")
		}
		s, ok := args[0].(*StringValue)
		if !ok {
			return nil, vm.runtimeError("startswith() argument must be a string")
		}
		return &BoolValue{Val: strings.HasPrefix(v.Val, s.Val)}, nil
	case "endswith":
		if len(args) != 1 {
			return nil, vm.runtimeError("endswith() takes 1 argument")
		}
		s, ok := args[0].(*StringValue)
		if !ok {
			return nil, vm.runtimeError("endswith() argument must be a string")
		}
		return &BoolValue{Val: strings.HasSuffix(v.Val, s.Val)}, nil
	case "replace":
		if len(args) != 2 {
			return nil, vm.runtimeError("replace() takes 2 arguments (old, new)")
		}
		old, ok1 := args[0].(*StringValue)
		new_, ok2 := args[1].(*StringValue)
		if !ok1 || !ok2 {
			return nil, vm.runtimeError("replace() arguments must be strings")
		}
		return &StringValue{Val: strings.ReplaceAll(v.Val, old.Val, new_.Val)}, nil
	case "index":
		if len(args) != 1 {
			return nil, vm.runtimeError("index() takes 1 argument")
		}
		s, ok := args[0].(*StringValue)
		if !ok {
			return nil, vm.runtimeError("index() argument must be a string")
		}
		idx := strings.Index(v.Val, s.Val)
		return &IntValue{Val: int64(idx)}, nil
	case "repeat":
		if len(args) != 1 {
			return nil, vm.runtimeError("repeat() takes 1 argument")
		}
		n, ok := args[0].(*IntValue)
		if !ok {
			return nil, vm.runtimeError("repeat() argument must be an int")
		}
		return &StringValue{Val: strings.Repeat(v.Val, int(n.Val))}, nil
	case "join":
		if len(args) != 1 {
			return nil, vm.runtimeError("join() takes 1 argument (list)")
		}
		lst, ok := args[0].(*ListValue)
		if !ok {
			return nil, vm.runtimeError("join() argument must be a list")
		}
		parts := make([]string, len(lst.Elements))
		for i, el := range lst.Elements {
			parts[i] = el.String()
		}
		return &StringValue{Val: strings.Join(parts, v.Val)}, nil
	default:
		return nil, vm.runtimeError("string has no method '%s'", method)
	}
}

func (vm *VM) boolMethod(v *BoolValue, method string, args []Value) (Value, error) {
	switch method {
	case "tostring":
		return &StringValue{Val: v.String()}, nil
	default:
		return nil, vm.runtimeError("bool has no method '%s'", method)
	}
}

func (vm *VM) listMethod(v *ListValue, method string, args []Value) (Value, error) {
	switch method {
	case "length":
		return &IntValue{Val: int64(len(v.Elements))}, nil
	case "append":
		if len(args) != 1 {
			return nil, vm.runtimeError("list.append() takes 1 argument, got %d", len(args))
		}
		v.Elements = append(v.Elements, args[0])
		return &NilValue{}, nil
	case "pop":
		if len(v.Elements) == 0 {
			return nil, vm.runtimeError("cannot pop from empty list")
		}
		last := v.Elements[len(v.Elements)-1]
		v.Elements = v.Elements[:len(v.Elements)-1]
		return last, nil
	case "tostring":
		return &StringValue{Val: v.String()}, nil
	default:
		return nil, vm.runtimeError("list has no method '%s'", method)
	}
}

func (vm *VM) arrayMethod(v *ArrayValue, method string, args []Value) (Value, error) {
	switch method {
	case "length":
		return &IntValue{Val: int64(len(v.Elements))}, nil
	case "tostring":
		return &StringValue{Val: v.String()}, nil
	default:
		return nil, vm.runtimeError("array has no method '%s'", method)
	}
}

func (vm *VM) dictMethod(v *DictValue, method string, args []Value) (Value, error) {
	switch method {
	case "length":
		return &IntValue{Val: int64(len(v.Entries))}, nil
	case "keys":
		keys := make([]Value, len(v.Order))
		for i, k := range v.Order {
			keys[i] = &StringValue{Val: k}
		}
		return &ListValue{Elements: keys}, nil
	case "values":
		vals := make([]Value, len(v.Order))
		for i, k := range v.Order {
			vals[i] = v.Entries[k]
		}
		return &ListValue{Elements: vals}, nil
	case "get":
		if len(args) != 1 {
			return nil, vm.runtimeError("dict.get() takes 1 argument, got %d", len(args))
		}
		key, ok := args[0].(*StringValue)
		if !ok {
			return nil, vm.runtimeError("dict.get() key must be a string")
		}
		val, exists := v.Entries[key.Val]
		if !exists {
			return &NilValue{}, nil
		}
		return val, nil
	case "set":
		if len(args) != 2 {
			return nil, vm.runtimeError("dict.set() takes 2 arguments, got %d", len(args))
		}
		key, ok := args[0].(*StringValue)
		if !ok {
			return nil, vm.runtimeError("dict.set() key must be a string")
		}
		if _, exists := v.Entries[key.Val]; !exists {
			v.Order = append(v.Order, key.Val)
		}
		v.Entries[key.Val] = args[1]
		return &NilValue{}, nil
	case "has":
		if len(args) != 1 {
			return nil, vm.runtimeError("dict.has() takes 1 argument, got %d", len(args))
		}
		key, ok := args[0].(*StringValue)
		if !ok {
			return nil, vm.runtimeError("dict.has() key must be a string")
		}
		_, exists := v.Entries[key.Val]
		return &BoolValue{Val: exists}, nil
	case "tostring":
		return &StringValue{Val: v.String()}, nil
	default:
		return nil, vm.runtimeError("dict has no method '%s'", method)
	}
}


func (vm *VM) resolveInstanceMethod(inst *InstanceValue, method string) *MethodDef {
	// super call
	if strings.HasPrefix(method, "$super.") {
		actualMethod := method[7:]
		if inst.Class.Super != nil {
			return vm.findMethod(inst.Class.Super, actualMethod)
		}
		return nil
	}

	m := vm.findMethod(inst.Class, method)
	if m == nil {
		return nil
	}
	// private methods can only be called from within the same class (via self)
	// but we allow them here since the compiler handles visibility for external calls
	return m
}

func (vm *VM) fileMethod(f *FileValue, method string, args []Value) (Value, error) {
	if f.Closed && method != "close" {
		return nil, vm.runtimeError("cannot call %s() on a closed file", method)
	}
	switch method {
	case "read":
		if f.Scanner != nil {
			f.File.Seek(0, io.SeekStart)
			f.Scanner = nil
		}
		data, err := io.ReadAll(f.File)
		if err != nil {
			return nil, vm.runtimeError("file.read() failed: %s", err.Error())
		}
		return &StringValue{Val: string(data)}, nil
	case "readline":
		if f.Scanner == nil {
			f.Scanner = bufio.NewScanner(f.File)
		}
		if f.Scanner.Scan() {
			return &StringValue{Val: f.Scanner.Text()}, nil
		}
		return &NilValue{}, nil
	case "readlines":
		if f.Scanner != nil {
			f.File.Seek(0, io.SeekStart)
			f.Scanner = nil
		}
		data, err := io.ReadAll(f.File)
		if err != nil {
			return nil, vm.runtimeError("file.readlines() failed: %s", err.Error())
		}
		text := string(data)
		lines := strings.Split(text, "\n")
		if len(lines) > 0 && lines[len(lines)-1] == "" {
			lines = lines[:len(lines)-1]
		}
		elements := make([]Value, len(lines))
		for i, line := range lines {
			elements[i] = &StringValue{Val: line}
		}
		return &ListValue{Elements: elements}, nil
	case "write":
		if len(args) != 1 {
			return nil, vm.runtimeError("file.write() takes 1 argument, got %d", len(args))
		}
		str, ok := args[0].(*StringValue)
		if !ok {
			return nil, vm.runtimeError("file.write() requires a string")
		}
		_, err := f.File.WriteString(str.Val)
		if err != nil {
			return nil, vm.runtimeError("file.write() failed: %s", err.Error())
		}
		return &NilValue{}, nil
	case "close":
		if !f.Closed {
			f.Closed = true
			f.File.Close()
		}
		return &NilValue{}, nil
	case "tostring":
		return &StringValue{Val: f.String()}, nil
	default:
		return nil, vm.runtimeError("file has no method '%s'", method)
	}
}

func (vm *VM) instanceBuiltinMethod(inst *InstanceValue, method string, args []Value) (Value, error) {
	switch method {
	case "tostring":
		return &StringValue{Val: inst.String()}, nil
	default:
		return nil, vm.runtimeError("'%s' has no method '%s'", inst.Class.Name, method)
	}
}

func (vm *VM) findMethod(class *ClassValue, name string) *MethodDef {
	cls := class
	for cls != nil {
		if m, ok := cls.Methods[name]; ok {
			return m
		}
		cls = cls.Super
	}
	return nil
}


func (vm *VM) wrapClassDef(def *compiler.ClassDef) *ClassValue {
	fields := make(map[string]FieldDef)
	for name, f := range def.Fields {
		var defaultVal Value
		if f.Default != nil {
			defaultVal = vm.wrapConstant(f.Default)
		}
		fields[name] = FieldDef{
			Visibility: f.Visibility,
			TypeName:   f.TypeName,
			Default:    defaultVal,
		}
	}

	methods := make(map[string]*MethodDef)
	for name, m := range def.Methods {
		methods[name] = &MethodDef{
			Visibility: m.Visibility,
			Fn:         m.Fn,
		}
	}

	return &ClassValue{
		Name:      def.Name,
		Fields:    fields,
		Methods:   methods,
		InitArity: def.InitArity,
		MaxArity:  def.MaxArity,
		InitDefs:  def.InitDefs,
	}
}

// tryThrowError attempts to throw a Go error through the CColon exception
// handler stack. If there are active handlers, it creates an Error instance
// and throws it (returns nil on success). If no handlers, returns the original error.
func (vm *VM) tryThrowError(goErr error) error {
	if len(vm.handlers) == 0 {
		return goErr
	}
	msg := goErr.Error()
	// strip "runtime error at line N: " prefix if present
	if idx := strings.Index(msg, ": "); idx != -1 && strings.HasPrefix(msg, "runtime error") {
		msg = msg[idx+2:]
	}
	errorClass, ok := vm.globals["Error"].(*ClassValue)
	if !ok {
		return goErr
	}
	inst := &InstanceValue{
		Class:  errorClass,
		Fields: make(map[string]Value),
	}
	for name, fd := range errorClass.Fields {
		if fd.Default != nil {
			inst.Fields[name] = fd.Default
		} else {
			inst.Fields[name] = &NilValue{}
		}
	}
	inst.Fields["message"] = &StringValue{Val: msg}
	return vm.throwValue(inst)
}

// throwRuntimeErr creates an Error instance and throws it through the exception
// handler stack if handlers exist. Otherwise returns a Go error.
func (vm *VM) throwRuntimeErr(format string, args ...interface{}) error {
	msg := fmt.Sprintf(format, args...)
	if len(vm.handlers) > 0 {
		// create an Error instance
		errorClass, ok := vm.globals["Error"].(*ClassValue)
		if ok {
			inst := &InstanceValue{
				Class:  errorClass,
				Fields: make(map[string]Value),
			}
			for name, fd := range errorClass.Fields {
				if fd.Default != nil {
					inst.Fields[name] = fd.Default
				} else {
					inst.Fields[name] = &NilValue{}
				}
			}
			inst.Fields["message"] = &StringValue{Val: msg}
			return vm.throwValue(inst)
		}
	}
	f := vm.frame()
	line := 0
	if f.ip > 0 && f.ip-1 < len(f.function.Lines) {
		line = f.function.Lines[f.ip-1]
	}
	return fmt.Errorf("runtime error at line %d: %s", line, msg)
}

func (vm *VM) throwValue(thrown Value) error {
	for len(vm.handlers) > 0 {
		handler := vm.handlers[len(vm.handlers)-1]
		vm.handlers = vm.handlers[:len(vm.handlers)-1]

		// unwind to handler's frame
		vm.fp = handler.FramePtr
		vm.frames = vm.frames[:vm.fp]
		vm.sp = handler.StackPtr

		// push the thrown value for the catch block
		vm.push(thrown)

		// jump to catch handler
		vm.frame().ip = handler.CatchIP
		return nil
	}

	// no handler found
	if inst, ok := thrown.(*InstanceValue); ok {
		msg, exists := inst.Fields["message"]
		if exists {
			return vm.runtimeError("uncaught %s: %s", inst.Class.Name, msg.String())
		}
		return vm.runtimeError("uncaught %s", inst.Class.Name)
	}
	return vm.runtimeError("uncaught exception: %s", thrown.String())
}

func (vm *VM) isInstanceOf(inst *InstanceValue, className string) bool {
	cls := inst.Class
	for cls != nil {
		if cls.Name == className {
			return true
		}
		cls = cls.Super
	}
	return false
}

func (vm *VM) RegisterBuiltinError() {
	errorClass := &ClassValue{
		Name: "Error",
		Fields: map[string]FieldDef{
			"message": {Visibility: "public", TypeName: "string", Default: &StringValue{Val: ""}},
		},
		Methods: map[string]*MethodDef{},
		InitArity: 1,
		MaxArity:  1,
	}

	// compile a simple init function for Error
	initFn := &compiler.FuncObject{
		Name:       "Error.init",
		Arity:      2, // self + message
		MaxArity:   2,
		Code:       []byte{},
		Constants:  []interface{}{},
		Lines:      []int{},
		LocalCount: 2,
	}

	// bytecode: self.message = message; return nil
	// LOAD_LOCAL 0 (self), LOAD_LOCAL 1 (message), SET_FIELD "message", POP, NIL, RETURN
	msgIdx := 0
	initFn.Constants = append(initFn.Constants, "message")
	initFn.Code = append(initFn.Code,
		byte(compiler.OP_LOAD_LOCAL), 0,  // self
		byte(compiler.OP_LOAD_LOCAL), 1,  // message
		byte(compiler.OP_SET_FIELD), byte(msgIdx>>8), byte(msgIdx&0xFF),
		byte(compiler.OP_NIL),
		byte(compiler.OP_RETURN),
	)
	initFn.Lines = make([]int, len(initFn.Code))

	errorClass.Methods["init"] = &MethodDef{
		Visibility: "private",
		Fn:         initFn,
	}

	vm.globals["Error"] = errorClass
}

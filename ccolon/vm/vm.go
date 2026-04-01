package vm

import (
	"bufio"
	"fmt"
	"io"
	"strings"

	"github.com/TRC-Loop/ccolon/compiler"
)

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
	globals     map[string]Value
	modules     map[string]*ModuleValue
	imported    map[string]bool
	handlers    []ExceptionHandler
	FileLoader  func(path string) (*compiler.FuncObject, error)
	importedFiles map[string]bool
}

func New() *VM {
	return &VM{
		stack:         make([]Value, 512),
		sp:            0,
		frames:        make([]CallFrame, 0, 64),
		globals:       make(map[string]Value),
		modules:       make(map[string]*ModuleValue),
		imported:      make(map[string]bool),
		importedFiles: make(map[string]bool),
	}
}

func (vm *VM) RegisterModule(name string, mod *ModuleValue) {
	vm.modules[name] = mod
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
				return err
			}
			vm.push(result)

		case compiler.OP_SUB:
			b := vm.pop()
			a := vm.pop()
			result, err := vm.opArith(a, b, "-")
			if err != nil {
				return err
			}
			vm.push(result)

		case compiler.OP_MUL:
			b := vm.pop()
			a := vm.pop()
			result, err := vm.opArith(a, b, "*")
			if err != nil {
				return err
			}
			vm.push(result)

		case compiler.OP_DIV:
			b := vm.pop()
			a := vm.pop()
			result, err := vm.opDiv(a, b)
			if err != nil {
				return err
			}
			vm.push(result)

		case compiler.OP_MOD:
			b := vm.pop()
			a := vm.pop()
			result, err := vm.opMod(a, b)
			if err != nil {
				return err
			}
			vm.push(result)

		case compiler.OP_NEG:
			v := vm.pop()
			switch val := v.(type) {
			case *IntValue:
				vm.push(&IntValue{Val: -val.Val})
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
			vm.globals[name] = vm.pop()

		case compiler.OP_CALL:
			argCount := int(vm.readByte())
			callee := vm.stack[vm.sp-1-argCount]
			if err := vm.callValue(callee, argCount); err != nil {
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
			inst, ok := object.(*InstanceValue)
			if !ok {
				return vm.runtimeError("cannot access field '%s' on %s", fieldName, object.String())
			}
			val, exists := inst.Fields[fieldName]
			if !exists {
				return vm.runtimeError("instance of '%s' has no field '%s'", inst.Class.Name, fieldName)
			}
			vm.push(val)

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

func (vm *VM) opAdd(a, b Value) (Value, error) {
	switch av := a.(type) {
	case *IntValue:
		switch bv := b.(type) {
		case *IntValue:
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
	}
	return nil, vm.runtimeError("cannot add %s and %s", a.String(), b.String())
}

func (vm *VM) opArith(a, b Value, op string) (Value, error) {
	switch av := a.(type) {
	case *IntValue:
		switch bv := b.(type) {
		case *IntValue:
			switch op {
			case "-":
				return &IntValue{Val: av.Val - bv.Val}, nil
			case "*":
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
	default:
		return nil, vm.runtimeError("int has no method '%s'", method)
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

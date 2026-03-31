package vm

import (
	"fmt"

	"github.com/TRC-Loop/ccolon/compiler"
)

type CallFrame struct {
	function *compiler.FuncObject
	ip       int
	basePtr  int
}

type VM struct {
	stack    []Value
	sp       int
	frames   []CallFrame
	fp       int
	globals  map[string]Value
	modules  map[string]*ModuleValue
	imported map[string]bool
}

func New() *VM {
	return &VM{
		stack:    make([]Value, 512),
		sp:       0,
		frames:   make([]CallFrame, 0, 64),
		globals:  make(map[string]Value),
		modules:  make(map[string]*ModuleValue),
		imported: make(map[string]bool),
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
			vm.sp = frame.basePtr
			vm.fp--
			vm.frames = vm.frames[:vm.fp]
			if vm.fp == 0 {
				return nil
			}
			vm.push(result)

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
	default:
		return &NilValue{}
	}
}

func (vm *VM) callValue(callee Value, argCount int) error {
	switch fn := callee.(type) {
	case *FuncValue:
		if argCount != fn.Obj.Arity {
			return vm.runtimeError("function '%s' expects %d arguments, got %d",
				fn.Obj.Name, fn.Obj.Arity, argCount)
		}
		// stack: [..., callee, arg0, arg1, ..., argN]
		// move args down by 1 to overwrite the callee slot
		basePtr := vm.sp - argCount
		for i := 0; i < argCount; i++ {
			vm.stack[basePtr-1+i] = vm.stack[basePtr+i]
		}
		vm.sp = basePtr - 1 + argCount
		newBase := basePtr - 1

		vm.frames = append(vm.frames, CallFrame{
			function: fn.Obj,
			ip:       0,
			basePtr:  newBase,
		})
		vm.fp = len(vm.frames)
		return nil

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

	default:
		return vm.runtimeError("cannot call %s", callee.String())
	}
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

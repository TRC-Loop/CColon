package compiler

import (
	"fmt"

	"github.com/TRC-Loop/ccolon/lexer"
	"github.com/TRC-Loop/ccolon/parser"
)

type Local struct {
	Name  string
	Depth int
	Slot  int
}

type LoopContext struct {
	StartOffset      int
	BreakPatches     []int
	ContinuePatches  []int
	HasContinueTarget bool
	ContinueTarget   int
}

type Compiler struct {
	function    *FuncObject
	locals      []Local
	scopeDepth  int
	loops       []LoopContext
	constLocals map[string]bool // tracks const local variable names
}

func New() *Compiler {
	fn := &FuncObject{Name: "<script>"}
	return &Compiler{
		function:    fn,
		constLocals: make(map[string]bool),
	}
}

func (c *Compiler) Compile(program *parser.Program) (*FuncObject, error) {
	for _, stmt := range program.Stmts {
		if err := c.compileStmt(stmt); err != nil {
			return nil, err
		}
	}
	c.emit(byte(OP_HALT), 0)
	c.function.LocalCount = len(c.locals)
	return c.function, nil
}

func (c *Compiler) emit(b byte, line int) {
	c.function.Code = append(c.function.Code, b)
	c.function.Lines = append(c.function.Lines, line)
}

func (c *Compiler) emitOp(op OpCode, line int) {
	c.emit(byte(op), line)
}

func (c *Compiler) emitBytes(line int, bytes ...byte) {
	for _, b := range bytes {
		c.emit(b, line)
	}
}

func (c *Compiler) addConstant(val interface{}) int {
	// dedup strings and numbers (skip non-comparable types)
	switch val.(type) {
	case string, int64, float64, bool:
		for i, existing := range c.function.Constants {
			if existing == val {
				return i
			}
		}
	}
	c.function.Constants = append(c.function.Constants, val)
	return len(c.function.Constants) - 1
}

func (c *Compiler) emitConstant(val interface{}, line int) {
	idx := c.addConstant(val)
	c.emitOp(OP_CONST, line)
	c.emitUint16(idx, line)
}

func (c *Compiler) emitUint16(val int, line int) {
	c.emit(byte(val>>8), line)
	c.emit(byte(val&0xFF), line)
}

func (c *Compiler) emitJump(op OpCode, line int) int {
	c.emitOp(op, line)
	c.emit(0xFF, line)
	c.emit(0xFF, line)
	return len(c.function.Code) - 2
}

func (c *Compiler) patchJump(offset int) {
	jump := len(c.function.Code) - offset - 2
	c.function.Code[offset] = byte(jump >> 8)
	c.function.Code[offset+1] = byte(jump & 0xFF)
}

func (c *Compiler) emitLoop(loopStart int, line int) {
	c.emitOp(OP_LOOP, line)
	offset := len(c.function.Code) - loopStart + 2
	c.emit(byte(offset>>8), line)
	c.emit(byte(offset&0xFF), line)
}

func (c *Compiler) beginScope() {
	c.scopeDepth++
}

func (c *Compiler) endScope(line int) {
	c.scopeDepth--
	for len(c.locals) > 0 && c.locals[len(c.locals)-1].Depth > c.scopeDepth {
		c.emitOp(OP_POP, line)
		c.locals = c.locals[:len(c.locals)-1]
	}
}

func (c *Compiler) addLocal(name string) int {
	slot := len(c.locals)
	c.locals = append(c.locals, Local{Name: name, Depth: c.scopeDepth, Slot: slot})
	return slot
}

func (c *Compiler) resolveLocal(name string) (int, bool) {
	for i := len(c.locals) - 1; i >= 0; i-- {
		if c.locals[i].Name == name {
			return c.locals[i].Slot, true
		}
	}
	return 0, false
}

func (c *Compiler) compileStmt(stmt parser.Stmt) error {
	switch s := stmt.(type) {
	case *parser.ImportStmt:
		return c.compileImport(s)
	case *parser.VarDecl:
		return c.compileVarDecl(s)
	case *parser.FuncDecl:
		return c.compileFuncDecl(s)
	case *parser.ExprStmt:
		if err := c.compileExpr(s.Expression); err != nil {
			return err
		}
		c.emitOp(OP_POP, s.P.Line)
		return nil
	case *parser.AssignStmt:
		return c.compileAssign(s)
	case *parser.IfStmt:
		return c.compileIf(s)
	case *parser.WhileStmt:
		return c.compileWhile(s)
	case *parser.ForInStmt:
		return c.compileForIn(s)
	case *parser.ReturnStmt:
		return c.compileReturn(s)
	case *parser.BreakStmt:
		return c.compileBreak(s)
	case *parser.ContinueStmt:
		return c.compileContinue(s)
	case *parser.ClassDecl:
		return c.compileClassDecl(s)
	case *parser.TryCatchStmt:
		return c.compileTryCatch(s)
	case *parser.ThrowStmt:
		return c.compileThrow(s)
	case *parser.WithStmt:
		return c.compileWith(s)
	default:
		return fmt.Errorf("unknown statement type %T", stmt)
	}
}

func (c *Compiler) compileImport(s *parser.ImportStmt) error {
	idx := c.addConstant(s.Module)
	if s.IsFile {
		c.emitOp(OP_IMPORT_FILE, s.P.Line)
	} else {
		c.emitOp(OP_IMPORT, s.P.Line)
	}
	c.emitUint16(idx, s.P.Line)
	return nil
}

func (c *Compiler) compileVarDecl(s *parser.VarDecl) error {
	if err := c.compileExpr(s.Value); err != nil {
		return err
	}

	if c.scopeDepth > 0 {
		c.addLocal(s.Name)
		if s.IsConst {
			c.constLocals[s.Name] = true
		}
		return nil
	}

	idx := c.addConstant(s.Name)
	c.emitOp(OP_STORE_GLOBAL, s.P.Line)
	c.emitUint16(idx, s.P.Line)
	if s.IsConst {
		c.emitOp(OP_MARK_CONST, s.P.Line)
		c.emitUint16(idx, s.P.Line)
	}
	return nil
}

func (c *Compiler) compileFuncDecl(s *parser.FuncDecl) error {
	// count required and optional params
	requiredCount := 0
	for _, p := range s.Params {
		if p.Default == nil {
			requiredCount++
		}
	}

	fnCompiler := &Compiler{
		function:    &FuncObject{Name: s.Name, Arity: requiredCount, MaxArity: len(s.Params)},
		scopeDepth:  1,
		constLocals: make(map[string]bool),
	}

	// build defaults list
	defaults := make([]interface{}, len(s.Params))
	for i, p := range s.Params {
		if p.Default != nil {
			switch d := p.Default.(type) {
			case *parser.IntLiteral:
				defaults[i] = d.Value
			case *parser.SintLiteral:
				defaults[i] = d.Value
			case *parser.FloatLiteral:
				defaults[i] = d.Value
			case *parser.StringLiteral:
				defaults[i] = d.Value
			case *parser.BoolLiteral:
				defaults[i] = d.Value
			default:
				defaults[i] = nil
			}
		}
	}
	fnCompiler.function.Defaults = defaults

	paramNames := make([]string, len(s.Params))
	for i, param := range s.Params {
		fnCompiler.addLocal(param.Name)
		paramNames[i] = param.Name
	}
	fnCompiler.function.ParamNames = paramNames

	for _, stmt := range s.Body {
		if err := fnCompiler.compileStmt(stmt); err != nil {
			return err
		}
	}

	// implicit nil return
	fnCompiler.emitOp(OP_NIL, s.P.Line)
	fnCompiler.emitOp(OP_RETURN, s.P.Line)
	fnCompiler.function.LocalCount = len(fnCompiler.locals)

	idx := c.addConstant(fnCompiler.function)
	c.emitOp(OP_CONST, s.P.Line)
	c.emitUint16(idx, s.P.Line)

	nameIdx := c.addConstant(s.Name)
	c.emitOp(OP_STORE_GLOBAL, s.P.Line)
	c.emitUint16(nameIdx, s.P.Line)
	return nil
}

func (c *Compiler) compileAssign(s *parser.AssignStmt) error {
	switch target := s.Target.(type) {
	case *parser.Identifier:
		// check for const reassignment (local)
		if c.constLocals[target.Name] {
			return fmt.Errorf("line %d:%d: cannot reassign constant '%s'", s.P.Line, s.P.Col, target.Name)
		}
		if err := c.compileExpr(s.Value); err != nil {
			return err
		}
		if slot, ok := c.resolveLocal(target.Name); ok {
			c.emitOp(OP_STORE_LOCAL, s.P.Line)
			c.emit(byte(slot), s.P.Line)
			return nil
		}
		idx := c.addConstant(target.Name)
		c.emitOp(OP_STORE_GLOBAL, s.P.Line)
		c.emitUint16(idx, s.P.Line)
		return nil
	case *parser.IndexExpr:
		return c.compileIndexAssign(target, s.Value, s.P.Line)
	case *parser.FieldAccessExpr:
		return c.compileFieldAssign(target, s.Value, s.P.Line)
	default:
		return fmt.Errorf("line %d:%d: invalid assignment target", s.P.Line, s.P.Col)
	}
}

func (c *Compiler) compileIndexAssign(target *parser.IndexExpr, value parser.Expr, line int) error {
	if err := c.compileExpr(target.Object); err != nil {
		return err
	}
	if err := c.compileExpr(target.Index); err != nil {
		return err
	}
	if err := c.compileExpr(value); err != nil {
		return err
	}
	c.emitOp(OP_INDEX_SET, line)
	return nil
}

func (c *Compiler) compileIf(s *parser.IfStmt) error {
	if err := c.compileExpr(s.Cond); err != nil {
		return err
	}

	falseJump := c.emitJump(OP_JUMP_IF_FALSE, s.P.Line)

	c.beginScope()
	for _, stmt := range s.Body {
		if err := c.compileStmt(stmt); err != nil {
			return err
		}
	}
	c.endScope(s.P.Line)

	if len(s.ElseBody) > 0 {
		endJump := c.emitJump(OP_JUMP, s.P.Line)
		c.patchJump(falseJump)
		c.beginScope()
		for _, stmt := range s.ElseBody {
			if err := c.compileStmt(stmt); err != nil {
				return err
			}
		}
		c.endScope(s.P.Line)
		c.patchJump(endJump)
	} else {
		c.patchJump(falseJump)
	}

	return nil
}

func (c *Compiler) compileWhile(s *parser.WhileStmt) error {
	loopStart := len(c.function.Code)

	loop := LoopContext{StartOffset: loopStart, HasContinueTarget: true, ContinueTarget: loopStart}
	c.loops = append(c.loops, loop)

	if err := c.compileExpr(s.Cond); err != nil {
		return err
	}
	exitJump := c.emitJump(OP_JUMP_IF_FALSE, s.P.Line)

	c.beginScope()
	for _, stmt := range s.Body {
		if err := c.compileStmt(stmt); err != nil {
			return err
		}
	}
	c.endScope(s.P.Line)

	c.emitLoop(loopStart, s.P.Line)
	c.patchJump(exitJump)

	lctx := c.loops[len(c.loops)-1]
	for _, bp := range lctx.BreakPatches {
		c.patchJump(bp)
	}
	c.loops = c.loops[:len(c.loops)-1]

	return nil
}

func (c *Compiler) compileForIn(s *parser.ForInStmt) error {
	rangeExpr, ok := s.Iterable.(*parser.RangeExpr)
	if !ok {
		// for-in over a list: compile differently
		return c.compileForInList(s)
	}

	// for i in range(start, end) -> desugar to while
	c.beginScope()

	// compile start value and create loop variable
	if err := c.compileExpr(rangeExpr.Start); err != nil {
		return err
	}
	iterSlot := c.addLocal(s.VarName)

	// compile end value into a hidden local
	if err := c.compileExpr(rangeExpr.End); err != nil {
		return err
	}
	limitSlot := c.addLocal("$limit")

	loopStart := len(c.function.Code)

	loop := LoopContext{StartOffset: loopStart}
	c.loops = append(c.loops, loop)

	// condition: i < limit
	c.emitOp(OP_LOAD_LOCAL, s.P.Line)
	c.emit(byte(iterSlot), s.P.Line)
	c.emitOp(OP_LOAD_LOCAL, s.P.Line)
	c.emit(byte(limitSlot), s.P.Line)
	c.emitOp(OP_LT, s.P.Line)
	exitJump := c.emitJump(OP_JUMP_IF_FALSE, s.P.Line)

	// body
	for _, stmt := range s.Body {
		if err := c.compileStmt(stmt); err != nil {
			return err
		}
	}

	// patch continue jumps to here (the increment step)
	lctx := &c.loops[len(c.loops)-1]
	for _, cp := range lctx.ContinuePatches {
		c.patchJump(cp)
	}

	// increment: i = i + 1
	c.emitOp(OP_LOAD_LOCAL, s.P.Line)
	c.emit(byte(iterSlot), s.P.Line)
	c.emitConstant(int64(1), s.P.Line)
	c.emitOp(OP_ADD, s.P.Line)
	c.emitOp(OP_STORE_LOCAL, s.P.Line)
	c.emit(byte(iterSlot), s.P.Line)

	c.emitLoop(loopStart, s.P.Line)
	c.patchJump(exitJump)

	for _, bp := range lctx.BreakPatches {
		c.patchJump(bp)
	}
	c.loops = c.loops[:len(c.loops)-1]

	c.endScope(s.P.Line)
	return nil
}

func (c *Compiler) compileForInList(s *parser.ForInStmt) error {
	c.beginScope()

	// compile the list
	if err := c.compileExpr(s.Iterable); err != nil {
		return err
	}
	listSlot := c.addLocal("$list")

	// index counter = 0
	c.emitConstant(int64(0), s.P.Line)
	idxSlot := c.addLocal("$idx")

	// the iteration variable
	c.emitOp(OP_NIL, s.P.Line)
	iterSlot := c.addLocal(s.VarName)

	loopStart := len(c.function.Code)

	loop := LoopContext{StartOffset: loopStart}
	c.loops = append(c.loops, loop)

	// condition: $idx < $list.length()
	c.emitOp(OP_LOAD_LOCAL, s.P.Line)
	c.emit(byte(idxSlot), s.P.Line)
	c.emitOp(OP_LOAD_LOCAL, s.P.Line)
	c.emit(byte(listSlot), s.P.Line)
	methodIdx := c.addConstant("length")
	c.emitOp(OP_METHOD_CALL, s.P.Line)
	c.emitUint16(methodIdx, s.P.Line)
	c.emit(0, s.P.Line) // 0 args
	c.emitOp(OP_LT, s.P.Line)
	exitJump := c.emitJump(OP_JUMP_IF_FALSE, s.P.Line)

	// iter = list[idx]
	c.emitOp(OP_LOAD_LOCAL, s.P.Line)
	c.emit(byte(listSlot), s.P.Line)
	c.emitOp(OP_LOAD_LOCAL, s.P.Line)
	c.emit(byte(idxSlot), s.P.Line)
	c.emitOp(OP_INDEX_GET, s.P.Line)
	c.emitOp(OP_STORE_LOCAL, s.P.Line)
	c.emit(byte(iterSlot), s.P.Line)

	for _, stmt := range s.Body {
		if err := c.compileStmt(stmt); err != nil {
			return err
		}
	}

	// patch continue jumps to here (the increment step)
	lctx := &c.loops[len(c.loops)-1]
	for _, cp := range lctx.ContinuePatches {
		c.patchJump(cp)
	}

	// idx = idx + 1
	c.emitOp(OP_LOAD_LOCAL, s.P.Line)
	c.emit(byte(idxSlot), s.P.Line)
	c.emitConstant(int64(1), s.P.Line)
	c.emitOp(OP_ADD, s.P.Line)
	c.emitOp(OP_STORE_LOCAL, s.P.Line)
	c.emit(byte(idxSlot), s.P.Line)

	c.emitLoop(loopStart, s.P.Line)
	c.patchJump(exitJump)

	for _, bp := range lctx.BreakPatches {
		c.patchJump(bp)
	}
	c.loops = c.loops[:len(c.loops)-1]

	c.endScope(s.P.Line)
	return nil
}

func (c *Compiler) compileReturn(s *parser.ReturnStmt) error {
	if s.Value != nil {
		if err := c.compileExpr(s.Value); err != nil {
			return err
		}
	} else {
		c.emitOp(OP_NIL, s.P.Line)
	}
	c.emitOp(OP_RETURN, s.P.Line)
	return nil
}

func (c *Compiler) compileBreak(s *parser.BreakStmt) error {
	if len(c.loops) == 0 {
		return fmt.Errorf("line %d:%d: break outside of loop", s.P.Line, s.P.Col)
	}
	// pop locals in loop scope
	bp := c.emitJump(OP_JUMP, s.P.Line)
	c.loops[len(c.loops)-1].BreakPatches = append(c.loops[len(c.loops)-1].BreakPatches, bp)
	return nil
}

func (c *Compiler) compileContinue(s *parser.ContinueStmt) error {
	if len(c.loops) == 0 {
		return fmt.Errorf("line %d:%d: continue outside of loop", s.P.Line, s.P.Col)
	}
	loop := &c.loops[len(c.loops)-1]
	if loop.HasContinueTarget {
		c.emitLoop(loop.ContinueTarget, s.P.Line)
	} else {
		// forward jump to be patched later (for-in loops where increment is after body)
		bp := c.emitJump(OP_JUMP, s.P.Line)
		loop.ContinuePatches = append(loop.ContinuePatches, bp)
	}
	return nil
}

func (c *Compiler) compileExpr(expr parser.Expr) error {
	switch e := expr.(type) {
	case *parser.IntLiteral:
		c.emitConstant(e.Value, e.P.Line)
	case *parser.SintLiteral:
		c.emitConstant(e.Value, e.P.Line)
	case *parser.FloatLiteral:
		c.emitConstant(e.Value, e.P.Line)
	case *parser.StringLiteral:
		c.emitConstant(e.Value, e.P.Line)
	case *parser.FStringExpr:
		if err := c.compileFString(e); err != nil {
			return err
		}
	case *parser.BoolLiteral:
		if e.Value {
			c.emitOp(OP_TRUE, e.P.Line)
		} else {
			c.emitOp(OP_FALSE, e.P.Line)
		}
	case *parser.Identifier:
		return c.compileIdentifier(e)
	case *parser.BinaryExpr:
		return c.compileBinary(e)
	case *parser.UnaryExpr:
		return c.compileUnary(e)
	case *parser.CallExpr:
		return c.compileCall(e)
	case *parser.MethodCallExpr:
		return c.compileMethodCall(e)
	case *parser.IndexExpr:
		return c.compileIndex(e)
	case *parser.ListLiteral:
		return c.compileList(e)
	case *parser.FixedArrayLiteral:
		return c.compileFixedArray(e)
	case *parser.SelfExpr:
		c.emitOp(OP_LOAD_LOCAL, e.P.Line)
		c.emit(0, e.P.Line) // self is always slot 0 in methods
	case *parser.FieldAccessExpr:
		return c.compileFieldAccess(e)
	case *parser.SuperCallExpr:
		return c.compileSuperCall(e)
	case *parser.DictLiteral:
		return c.compileDictLiteral(e)
	case *parser.RangeExpr:
		// range as an expression (used in for-in, but could be standalone)
		return fmt.Errorf("line %d:%d: range() can only be used in for loops", e.P.Line, e.P.Col)
	default:
		return fmt.Errorf("unknown expression type %T", expr)
	}
	return nil
}

func (c *Compiler) compileIdentifier(e *parser.Identifier) error {
	if e.Name == "nil" || e.Name == "None" {
		c.emitOp(OP_NIL, e.P.Line)
		return nil
	}
	if slot, ok := c.resolveLocal(e.Name); ok {
		c.emitOp(OP_LOAD_LOCAL, e.P.Line)
		c.emit(byte(slot), e.P.Line)
		return nil
	}
	idx := c.addConstant(e.Name)
	c.emitOp(OP_LOAD_GLOBAL, e.P.Line)
	c.emitUint16(idx, e.P.Line)
	return nil
}

func (c *Compiler) compileBinary(e *parser.BinaryExpr) error {
	// short-circuit for and/or
	if e.Op == lexer.TOKEN_AND {
		if err := c.compileExpr(e.Left); err != nil {
			return err
		}
		falseJump := c.emitJump(OP_JUMP_IF_FALSE, e.P.Line)
		if err := c.compileExpr(e.Right); err != nil {
			return err
		}
		endJump := c.emitJump(OP_JUMP, e.P.Line)
		c.patchJump(falseJump)
		c.emitOp(OP_FALSE, e.P.Line)
		c.patchJump(endJump)
		return nil
	}
	if e.Op == lexer.TOKEN_OR {
		if err := c.compileExpr(e.Left); err != nil {
			return err
		}
		// if true, skip right side
		trueJump := c.emitJump(OP_JUMP_IF_FALSE, e.P.Line)
		c.emitOp(OP_TRUE, e.P.Line)
		endJump := c.emitJump(OP_JUMP, e.P.Line)
		c.patchJump(trueJump)
		if err := c.compileExpr(e.Right); err != nil {
			return err
		}
		c.patchJump(endJump)
		return nil
	}

	if err := c.compileExpr(e.Left); err != nil {
		return err
	}
	if err := c.compileExpr(e.Right); err != nil {
		return err
	}

	switch e.Op {
	case lexer.TOKEN_PLUS:
		c.emitOp(OP_ADD, e.P.Line)
	case lexer.TOKEN_MINUS:
		c.emitOp(OP_SUB, e.P.Line)
	case lexer.TOKEN_STAR:
		c.emitOp(OP_MUL, e.P.Line)
	case lexer.TOKEN_SLASH:
		c.emitOp(OP_DIV, e.P.Line)
	case lexer.TOKEN_PERCENT:
		c.emitOp(OP_MOD, e.P.Line)
	case lexer.TOKEN_EQ:
		c.emitOp(OP_EQ, e.P.Line)
	case lexer.TOKEN_NEQ:
		c.emitOp(OP_NEQ, e.P.Line)
	case lexer.TOKEN_LT:
		c.emitOp(OP_LT, e.P.Line)
	case lexer.TOKEN_GT:
		c.emitOp(OP_GT, e.P.Line)
	case lexer.TOKEN_LTE:
		c.emitOp(OP_LTE, e.P.Line)
	case lexer.TOKEN_GTE:
		c.emitOp(OP_GTE, e.P.Line)
	default:
		return fmt.Errorf("unknown binary operator %s", e.Op.String())
	}
	return nil
}

func (c *Compiler) compileUnary(e *parser.UnaryExpr) error {
	if err := c.compileExpr(e.Operand); err != nil {
		return err
	}
	switch e.Op {
	case lexer.TOKEN_MINUS:
		c.emitOp(OP_NEG, e.P.Line)
	case lexer.TOKEN_NOT:
		c.emitOp(OP_NOT, e.P.Line)
	default:
		return fmt.Errorf("unknown unary operator %s", e.Op.String())
	}
	return nil
}

func (c *Compiler) compileCall(e *parser.CallExpr) error {
	if err := c.compileExpr(e.Callee); err != nil {
		return err
	}
	for _, arg := range e.Args {
		if err := c.compileExpr(arg); err != nil {
			return err
		}
	}
	if len(e.NamedArgs) > 0 {
		// emit named arg values, then emit their names as a constant list
		for _, na := range e.NamedArgs {
			if err := c.compileExpr(na.Value); err != nil {
				return err
			}
		}
		// build names list as a constant
		names := make([]string, len(e.NamedArgs))
		for i, na := range e.NamedArgs {
			names[i] = na.Name
		}
		namesIdx := c.addConstant(names)
		c.emitOp(OP_CALL_KW, e.P.Line)
		c.emit(byte(len(e.Args)), e.P.Line)
		c.emit(byte(len(e.NamedArgs)), e.P.Line)
		c.emitUint16(namesIdx, e.P.Line)
	} else {
		c.emitOp(OP_CALL, e.P.Line)
		c.emit(byte(len(e.Args)), e.P.Line)
	}
	return nil
}

func (c *Compiler) compileMethodCall(e *parser.MethodCallExpr) error {
	if err := c.compileExpr(e.Object); err != nil {
		return err
	}
	for _, arg := range e.Args {
		if err := c.compileExpr(arg); err != nil {
			return err
		}
	}
	methodIdx := c.addConstant(e.Method)
	c.emitOp(OP_METHOD_CALL, e.P.Line)
	c.emitUint16(methodIdx, e.P.Line)
	c.emit(byte(len(e.Args)), e.P.Line)
	return nil
}

func (c *Compiler) compileIndex(e *parser.IndexExpr) error {
	if err := c.compileExpr(e.Object); err != nil {
		return err
	}
	if err := c.compileExpr(e.Index); err != nil {
		return err
	}
	c.emitOp(OP_INDEX_GET, e.P.Line)
	return nil
}

func (c *Compiler) compileList(e *parser.ListLiteral) error {
	for _, elem := range e.Elements {
		if err := c.compileExpr(elem); err != nil {
			return err
		}
	}
	c.emitOp(OP_LIST_NEW, e.P.Line)
	c.emitUint16(len(e.Elements), e.P.Line)
	return nil
}

func (c *Compiler) compileFixedArray(e *parser.FixedArrayLiteral) error {
	for _, elem := range e.Elements {
		if err := c.compileExpr(elem); err != nil {
			return err
		}
	}
	c.emitOp(OP_ARRAY_NEW, e.P.Line)
	c.emitUint16(len(e.Elements), e.P.Line)
	return nil
}

func (c *Compiler) compileFieldAccess(e *parser.FieldAccessExpr) error {
	if err := c.compileExpr(e.Object); err != nil {
		return err
	}
	idx := c.addConstant(e.Field)
	c.emitOp(OP_GET_FIELD, e.P.Line)
	c.emitUint16(idx, e.P.Line)
	return nil
}

func (c *Compiler) compileFieldAssign(target *parser.FieldAccessExpr, value parser.Expr, line int) error {
	if err := c.compileExpr(target.Object); err != nil {
		return err
	}
	if err := c.compileExpr(value); err != nil {
		return err
	}
	idx := c.addConstant(target.Field)
	c.emitOp(OP_SET_FIELD, line)
	c.emitUint16(idx, line)
	return nil
}

func (c *Compiler) compileSuperCall(e *parser.SuperCallExpr) error {
	// push self (slot 0)
	c.emitOp(OP_LOAD_LOCAL, e.P.Line)
	c.emit(0, e.P.Line)
	// push args
	for _, arg := range e.Args {
		if err := c.compileExpr(arg); err != nil {
			return err
		}
	}
	// emit as method call with special "$super." prefix so the VM knows to look up the parent
	methodIdx := c.addConstant("$super." + e.Method)
	c.emitOp(OP_METHOD_CALL, e.P.Line)
	c.emitUint16(methodIdx, e.P.Line)
	c.emit(byte(len(e.Args)), e.P.Line)
	return nil
}

func (c *Compiler) compileDictLiteral(e *parser.DictLiteral) error {
	for i := range e.Keys {
		if err := c.compileExpr(e.Keys[i]); err != nil {
			return err
		}
		if err := c.compileExpr(e.Values[i]); err != nil {
			return err
		}
	}
	c.emitOp(OP_DICT_NEW, e.P.Line)
	c.emitUint16(len(e.Keys), e.P.Line)
	return nil
}

func (c *Compiler) compileClassDecl(s *parser.ClassDecl) error {
	// compile each method as a FuncObject with self as implicit first param
	methods := make(map[string]*MethodDef)
	fields := make(map[string]*FieldDef)

	for _, f := range s.Fields {
		var defaultVal interface{}
		if f.Default != nil {
			switch d := f.Default.(type) {
			case *parser.IntLiteral:
				defaultVal = d.Value
			case *parser.SintLiteral:
				defaultVal = d.Value
			case *parser.FloatLiteral:
				defaultVal = d.Value
			case *parser.StringLiteral:
				defaultVal = d.Value
			case *parser.BoolLiteral:
				defaultVal = d.Value
			}
		}
		fields[f.Name] = &FieldDef{
			Visibility: f.Visibility,
			TypeName:   f.TypeName,
			Default:    defaultVal,
		}
	}

	initArity := 0
	initMaxArity := 0
	var initDefaults []interface{}

	for _, m := range s.Methods {
		requiredCount := 0
		for _, p := range m.Params {
			if p.Default == nil {
				requiredCount++
			}
		}

		fnCompiler := &Compiler{
			function: &FuncObject{
				Name:     s.Name + "." + m.Name,
				Arity:    requiredCount + 1, // +1 for self
				MaxArity: len(m.Params) + 1, // +1 for self
			},
			scopeDepth:  1,
			constLocals: make(map[string]bool),
		}

		// build defaults (first entry nil for self)
		defaults := make([]interface{}, len(m.Params)+1)
		for i, p := range m.Params {
			if p.Default != nil {
				switch d := p.Default.(type) {
				case *parser.IntLiteral:
					defaults[i+1] = d.Value
				case *parser.SintLiteral:
					defaults[i+1] = d.Value
				case *parser.FloatLiteral:
					defaults[i+1] = d.Value
				case *parser.StringLiteral:
					defaults[i+1] = d.Value
				case *parser.BoolLiteral:
					defaults[i+1] = d.Value
				}
			}
		}
		fnCompiler.function.Defaults = defaults

		// self is slot 0
		fnCompiler.addLocal("self")
		for _, param := range m.Params {
			fnCompiler.addLocal(param.Name)
		}

		for _, stmt := range m.Body {
			if err := fnCompiler.compileStmt(stmt); err != nil {
				return err
			}
		}

		fnCompiler.emitOp(OP_NIL, m.P.Line)
		fnCompiler.emitOp(OP_RETURN, m.P.Line)
		fnCompiler.function.LocalCount = len(fnCompiler.locals)

		methods[m.Name] = &MethodDef{
			Visibility: m.Visibility,
			Fn:         fnCompiler.function,
		}

		if m.Name == "init" {
			initArity = requiredCount
			initMaxArity = len(m.Params)
			initDefaults = defaults[1:] // strip self
		}
	}

	// store class as constant
	classConst := &ClassDef{
		Name:      s.Name,
		SuperName: s.SuperName,
		Fields:    fields,
		Methods:   methods,
		InitArity: initArity,
		MaxArity:  initMaxArity,
		InitDefs:  initDefaults,
	}

	idx := c.addConstant(classConst)
	c.emitOp(OP_CONST, s.P.Line)
	c.emitUint16(idx, s.P.Line)

	// if extends, load the superclass and emit INHERIT
	if s.SuperName != "" {
		superIdx := c.addConstant(s.SuperName)
		c.emitOp(OP_LOAD_GLOBAL, s.P.Line)
		c.emitUint16(superIdx, s.P.Line)
		c.emitOp(OP_INHERIT, s.P.Line)
	}

	nameIdx := c.addConstant(s.Name)
	c.emitOp(OP_STORE_GLOBAL, s.P.Line)
	c.emitUint16(nameIdx, s.P.Line)
	return nil
}

func (c *Compiler) compileTryCatch(s *parser.TryCatchStmt) error {
	// emit TRY_BEGIN with placeholder offset to catch handler
	tryBegin := c.emitJump(OP_TRY_BEGIN, s.P.Line)

	// compile try body
	c.beginScope()
	for _, stmt := range s.TryBody {
		if err := c.compileStmt(stmt); err != nil {
			return err
		}
	}
	c.endScope(s.P.Line)

	// remove exception handler
	c.emitOp(OP_TRY_END, s.P.Line)

	// jump over catch body
	endJump := c.emitJump(OP_JUMP, s.P.Line)

	// patch TRY_BEGIN to point here (catch handler)
	c.patchJump(tryBegin)

	// the VM will push the error value onto the stack before jumping here
	c.beginScope()
	c.addLocal(s.CatchName)

	for _, stmt := range s.CatchBody {
		if err := c.compileStmt(stmt); err != nil {
			return err
		}
	}
	c.endScope(s.P.Line)

	c.patchJump(endJump)
	return nil
}

func (c *Compiler) compileThrow(s *parser.ThrowStmt) error {
	if err := c.compileExpr(s.Value); err != nil {
		return err
	}
	c.emitOp(OP_THROW, s.P.Line)
	return nil
}

func (c *Compiler) compileWith(s *parser.WithStmt) error {
	// compile the resource expression
	if err := c.compileExpr(s.Expr); err != nil {
		return err
	}

	c.beginScope()
	varSlot := c.addLocal(s.VarName)

	// set up try handler for cleanup
	tryBegin := c.emitJump(OP_TRY_BEGIN, s.P.Line)

	// compile body
	for _, stmt := range s.Body {
		if err := c.compileStmt(stmt); err != nil {
			return err
		}
	}

	// normal exit: remove handler, close resource
	c.emitOp(OP_TRY_END, s.P.Line)

	// call .close() on the resource
	c.emitOp(OP_LOAD_LOCAL, s.P.Line)
	c.emit(byte(varSlot), s.P.Line)
	closeIdx := c.addConstant("close")
	c.emitOp(OP_METHOD_CALL, s.P.Line)
	c.emitUint16(closeIdx, s.P.Line)
	c.emit(0, s.P.Line) // 0 args
	c.emitOp(OP_POP, s.P.Line)

	endJump := c.emitJump(OP_JUMP, s.P.Line)

	// catch handler: close resource and re-throw
	c.patchJump(tryBegin)

	// the error is on the stack, save it in its own scope
	c.beginScope()
	errSlot := c.addLocal("$with_err")

	// call .close()
	c.emitOp(OP_LOAD_LOCAL, s.P.Line)
	c.emit(byte(varSlot), s.P.Line)
	c.emitOp(OP_METHOD_CALL, s.P.Line)
	c.emitUint16(closeIdx, s.P.Line)
	c.emit(0, s.P.Line)
	c.emitOp(OP_POP, s.P.Line)

	// re-throw
	c.emitOp(OP_LOAD_LOCAL, s.P.Line)
	c.emit(byte(errSlot), s.P.Line)
	c.emitOp(OP_THROW, s.P.Line)

	c.endScope(s.P.Line) // end error local scope

	c.patchJump(endJump)
	c.endScope(s.P.Line) // end resource scope
	return nil
}

func (c *Compiler) compileFString(e *parser.FStringExpr) error {
	// Start with an empty string to ensure the result is always a string
	c.emitConstant("", e.P.Line)
	for _, part := range e.Parts {
		if err := c.compileExpr(part); err != nil {
			return err
		}
		c.emitOp(OP_ADD, e.P.Line)
	}
	return nil
}

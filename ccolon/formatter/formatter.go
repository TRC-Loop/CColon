package formatter

import (
	"strings"

	"github.com/TRC-Loop/ccolon/lexer"
	"github.com/TRC-Loop/ccolon/parser"
)

func Format(source string, cfg Config) (string, error) {
	l := lexer.New(source)
	tokens, err := l.Tokenize()
	if err != nil {
		return "", err
	}

	p := parser.New(tokens)
	prog, err := p.Parse()
	if err != nil {
		return "", err
	}

	f := &fmtWriter{
		cfg:    cfg,
		indent: 0,
	}

	for i, stmt := range prog.Stmts {
		f.formatStmt(stmt)
		if i < len(prog.Stmts)-1 {
			// blank line between top-level declarations
			if isDecl(stmt) || isDecl(prog.Stmts[i+1]) {
				f.writeln("")
			}
		}
	}

	result := f.buf.String()
	if !strings.HasSuffix(result, "\n") {
		result += "\n"
	}
	return result, nil
}

func isDecl(s parser.Stmt) bool {
	switch s.(type) {
	case *parser.FuncDecl, *parser.ClassDecl, *parser.ImportStmt:
		return true
	}
	return false
}

type fmtWriter struct {
	cfg    Config
	indent int
	buf    strings.Builder
}

func (f *fmtWriter) indentStr() string {
	if f.cfg.UseTabs {
		return strings.Repeat("\t", f.indent)
	}
	return strings.Repeat(" ", f.indent*f.cfg.IndentSize)
}

func (f *fmtWriter) writeln(s string) {
	if s == "" {
		f.buf.WriteString("\n")
	} else {
		f.buf.WriteString(f.indentStr() + s + "\n")
	}
}

func (f *fmtWriter) formatStmt(stmt parser.Stmt) {
	switch s := stmt.(type) {
	case *parser.ImportStmt:
		if s.IsFile {
			f.writeln("import \"" + s.Module + "\"")
		} else {
			f.writeln("import " + s.Module)
		}
	case *parser.VarDecl:
		f.writeln("var " + s.TypeName + " " + s.Name + " = " + f.formatExpr(s.Value))
	case *parser.FuncDecl:
		f.formatFuncDecl(s)
	case *parser.ClassDecl:
		f.formatClassDecl(s)
	case *parser.IfStmt:
		f.formatIf(s)
	case *parser.WhileStmt:
		f.writeln("while (" + f.formatExpr(s.Cond) + ") {")
		f.indent++
		for _, st := range s.Body {
			f.formatStmt(st)
		}
		f.indent--
		f.writeln("}")
	case *parser.ForInStmt:
		f.writeln("for " + s.VarName + " in " + f.formatExpr(s.Iterable) + " {")
		f.indent++
		for _, st := range s.Body {
			f.formatStmt(st)
		}
		f.indent--
		f.writeln("}")
	case *parser.ReturnStmt:
		if s.Value != nil {
			f.writeln("return " + f.formatExpr(s.Value))
		} else {
			f.writeln("return")
		}
	case *parser.BreakStmt:
		f.writeln("break")
	case *parser.ContinueStmt:
		f.writeln("continue")
	case *parser.ExprStmt:
		f.writeln(f.formatExpr(s.Expression))
	case *parser.AssignStmt:
		f.writeln(f.formatExpr(s.Target) + " = " + f.formatExpr(s.Value))
	case *parser.TryCatchStmt:
		f.writeln("try {")
		f.indent++
		for _, st := range s.TryBody {
			f.formatStmt(st)
		}
		f.indent--
		f.writeln("} catch (" + s.CatchType + " " + s.CatchName + ") {")
		f.indent++
		for _, st := range s.CatchBody {
			f.formatStmt(st)
		}
		f.indent--
		f.writeln("}")
	case *parser.ThrowStmt:
		f.writeln("throw " + f.formatExpr(s.Value))
	case *parser.WithStmt:
		f.writeln("with " + f.formatExpr(s.Expr) + " as " + s.VarName + " {")
		f.indent++
		for _, st := range s.Body {
			f.formatStmt(st)
		}
		f.indent--
		f.writeln("}")
	}
}

func (f *fmtWriter) formatFuncDecl(s *parser.FuncDecl) {
	sig := "function " + s.Name + "(" + f.formatParams(s.Params) + ")"
	if s.ReturnType != "" {
		sig += " " + s.ReturnType
	}
	sig += " {"
	f.writeln(sig)
	f.indent++
	for _, st := range s.Body {
		f.formatStmt(st)
	}
	f.indent--
	f.writeln("}")
}

func (f *fmtWriter) formatClassDecl(s *parser.ClassDecl) {
	line := "class " + s.Name
	if s.SuperName != "" {
		line += " extends " + s.SuperName
	}
	line += " {"
	f.writeln(line)
	f.indent++

	for _, field := range s.Fields {
		line := "var " + field.Visibility + " " + field.TypeName + " " + field.Name
		if field.Default != nil {
			line += " = " + f.formatExpr(field.Default)
		}
		f.writeln(line)
	}

	if len(s.Fields) > 0 && len(s.Methods) > 0 {
		f.writeln("")
	}

	for i, method := range s.Methods {
		sig := method.Visibility + " function " + method.Name + "(" + f.formatParams(method.Params) + ")"
		if method.ReturnType != "" {
			sig += " " + method.ReturnType
		}
		sig += " {"
		f.writeln(sig)
		f.indent++
		for _, st := range method.Body {
			f.formatStmt(st)
		}
		f.indent--
		f.writeln("}")
		if i < len(s.Methods)-1 {
			f.writeln("")
		}
	}

	f.indent--
	f.writeln("}")
}

func (f *fmtWriter) formatParams(params []parser.Param) string {
	parts := make([]string, len(params))
	for i, p := range params {
		s := p.TypeName + " " + p.Name
		if p.Default != nil {
			s += " = " + f.formatExpr(p.Default)
		}
		parts[i] = s
	}
	return strings.Join(parts, ", ")
}

func (f *fmtWriter) formatIf(s *parser.IfStmt) {
	f.writeln("if (" + f.formatExpr(s.Cond) + ") {")
	f.indent++
	for _, st := range s.Body {
		f.formatStmt(st)
	}
	f.indent--
	if len(s.ElseBody) == 0 {
		f.writeln("}")
		return
	}
	if len(s.ElseBody) == 1 {
		if elseIf, ok := s.ElseBody[0].(*parser.IfStmt); ok {
			f.buf.WriteString(f.indentStr() + "} else ")
			// don't add indent for the else-if
			oldIndent := f.indent
			f.indent = 0
			f.formatIf(elseIf)
			f.indent = oldIndent
			return
		}
	}
	f.writeln("} else {")
	f.indent++
	for _, st := range s.ElseBody {
		f.formatStmt(st)
	}
	f.indent--
	f.writeln("}")
}

func (f *fmtWriter) formatExpr(expr parser.Expr) string {
	switch e := expr.(type) {
	case *parser.IntLiteral:
		return intToStr(e.Value)
	case *parser.FloatLiteral:
		return floatToStr(e.Value)
	case *parser.StringLiteral:
		return "\"" + escapeString(e.Value) + "\""
	case *parser.BoolLiteral:
		if e.Value {
			return "true"
		}
		return "false"
	case *parser.Identifier:
		return e.Name
	case *parser.BinaryExpr:
		return f.formatExpr(e.Left) + " " + opStr(e.Op) + " " + f.formatExpr(e.Right)
	case *parser.UnaryExpr:
		op := opStr(e.Op)
		if op == "not" {
			return "not " + f.formatExpr(e.Operand)
		}
		return op + f.formatExpr(e.Operand)
	case *parser.CallExpr:
		args := f.formatArgs(e.Args)
		return f.formatExpr(e.Callee) + "(" + args + ")"
	case *parser.MethodCallExpr:
		args := f.formatArgs(e.Args)
		return f.formatExpr(e.Object) + "." + e.Method + "(" + args + ")"
	case *parser.IndexExpr:
		return f.formatExpr(e.Object) + "[" + f.formatExpr(e.Index) + "]"
	case *parser.FieldAccessExpr:
		return f.formatExpr(e.Object) + "." + e.Field
	case *parser.ListLiteral:
		return "[" + f.formatArgs(e.Elements) + "]"
	case *parser.FixedArrayLiteral:
		return "fixed([" + f.formatArgs(e.Elements) + "])"
	case *parser.RangeExpr:
		if e.End != nil {
			return "range(" + f.formatExpr(e.Start) + ", " + f.formatExpr(e.End) + ")"
		}
		return "range(" + f.formatExpr(e.Start) + ")"
	case *parser.DictLiteral:
		if len(e.Keys) == 0 {
			return "{}"
		}
		pairs := make([]string, len(e.Keys))
		for i := range e.Keys {
			pairs[i] = f.formatExpr(e.Keys[i]) + ": " + f.formatExpr(e.Values[i])
		}
		return "{" + strings.Join(pairs, ", ") + "}"
	case *parser.SelfExpr:
		return "self"
	case *parser.SuperCallExpr:
		return "super." + e.Method + "(" + f.formatArgs(e.Args) + ")"
	default:
		return "<?>"
	}
}

func (f *fmtWriter) formatArgs(args []parser.Expr) string {
	parts := make([]string, len(args))
	for i, a := range args {
		parts[i] = f.formatExpr(a)
	}
	return strings.Join(parts, ", ")
}

func opStr(op lexer.TokenType) string {
	switch op {
	case lexer.TOKEN_PLUS:
		return "+"
	case lexer.TOKEN_MINUS:
		return "-"
	case lexer.TOKEN_STAR:
		return "*"
	case lexer.TOKEN_SLASH:
		return "/"
	case lexer.TOKEN_PERCENT:
		return "%"
	case lexer.TOKEN_EQ:
		return "=="
	case lexer.TOKEN_NEQ:
		return "!="
	case lexer.TOKEN_LT:
		return "<"
	case lexer.TOKEN_GT:
		return ">"
	case lexer.TOKEN_LTE:
		return "<="
	case lexer.TOKEN_GTE:
		return ">="
	case lexer.TOKEN_AND:
		return "and"
	case lexer.TOKEN_OR:
		return "or"
	case lexer.TOKEN_NOT:
		return "not"
	default:
		return "?"
	}
}

func escapeString(s string) string {
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, "\"", "\\\"")
	s = strings.ReplaceAll(s, "\n", "\\n")
	s = strings.ReplaceAll(s, "\t", "\\t")
	return s
}

func intToStr(v int64) string {
	return strings.TrimRight(strings.TrimRight(
		strings.Replace(
			strings.Replace(
				formatInt(v), "", "", 0,
			), "", "", 0,
		), "0"), ".")
}

func formatInt(v int64) string {
	if v < 0 {
		return "-" + formatUint(uint64(-v))
	}
	return formatUint(uint64(v))
}

func formatUint(v uint64) string {
	if v == 0 {
		return "0"
	}
	buf := make([]byte, 0, 20)
	for v > 0 {
		buf = append(buf, byte('0'+v%10))
		v /= 10
	}
	// reverse
	for i, j := 0, len(buf)-1; i < j; i, j = i+1, j-1 {
		buf[i], buf[j] = buf[j], buf[i]
	}
	return string(buf)
}

func floatToStr(v float64) string {
	return strings.TrimRight(strings.TrimRight(
		formatFloat(v), "0"), ".")
}

func formatFloat(v float64) string {
	// simple float formatting
	s := ""
	if v < 0 {
		s = "-"
		v = -v
	}
	intPart := int64(v)
	fracPart := v - float64(intPart)

	s += formatInt(intPart) + "."
	for i := 0; i < 15; i++ {
		fracPart *= 10
		digit := int(fracPart)
		s += string(rune('0' + digit))
		fracPart -= float64(digit)
	}
	return s
}

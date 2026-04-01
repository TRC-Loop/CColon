package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/TRC-Loop/ccolon/compiler"
	"github.com/TRC-Loop/ccolon/lexer"
	"github.com/TRC-Loop/ccolon/parser"
	"github.com/TRC-Loop/ccolon/repl"
	"github.com/TRC-Loop/ccolon/stdlib"
	"github.com/TRC-Loop/ccolon/vm"
)

const version = "0.1.0"

func main() {
	if len(os.Args) < 2 {
		machine := vm.New()
		stdlib.NewRegistry().RegisterAll(machine)
		machine.RegisterBuiltinError()
		repl.Start(os.Stdin, os.Stdout, machine)
		return
	}

	if os.Args[1] == "--version" || os.Args[1] == "-v" {
		fmt.Printf("CColon v%s\n", version)
		return
	}

	if os.Args[1] == "--help" || os.Args[1] == "-h" {
		fmt.Println("Usage: ccolon [file.ccl]")
		fmt.Println("  No arguments: start interactive REPL")
		fmt.Println("  file.ccl:     run a CColon source file")
		fmt.Println("  --version:    print version")
		return
	}

	filePath := os.Args[1]
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %s\n", err)
		os.Exit(1)
	}

	source, err := os.ReadFile(absPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %s\n", err)
		os.Exit(1)
	}

	if err := runFile(string(source), filepath.Dir(absPath)); err != nil {
		fmt.Fprintf(os.Stderr, "error: %s\n", err)
		os.Exit(1)
	}
}

func compileSource(source string) (*compiler.FuncObject, error) {
	l := lexer.New(source)
	tokens, err := l.Tokenize()
	if err != nil {
		return nil, err
	}

	p := parser.New(tokens)
	program, err := p.Parse()
	if err != nil {
		return nil, err
	}

	c := compiler.New()
	return c.Compile(program)
}

func runFile(source string, baseDir string) error {
	l := lexer.New(source)
	tokens, err := l.Tokenize()
	if err != nil {
		return err
	}

	p := parser.New(tokens)
	program, err := p.Parse()
	if err != nil {
		return err
	}

	// main() is required
	hasMain := false
	for _, stmt := range program.Stmts {
		if fd, ok := stmt.(*parser.FuncDecl); ok && fd.Name == "main" {
			hasMain = true
			break
		}
	}
	if !hasMain {
		return fmt.Errorf("missing function main() -- every program needs a main function as entry point")
	}

	c := compiler.New()
	fn, err := c.Compile(program)
	if err != nil {
		return err
	}

	machine := vm.New()
	stdlib.NewRegistry().RegisterAll(machine)
	machine.RegisterBuiltinError()

	machine.FileLoader = func(path string) (*compiler.FuncObject, error) {
		resolved := path
		if !filepath.IsAbs(path) {
			resolved = filepath.Join(baseDir, path)
		}
		data, err := os.ReadFile(resolved)
		if err != nil {
			return nil, fmt.Errorf("cannot read '%s': %s", path, err.Error())
		}
		return compileSource(string(data))
	}

	if err := machine.Run(fn); err != nil {
		return err
	}

	return callMain(machine)
}

func callMain(machine *vm.VM) error {
	// compile a tiny program that just calls main()
	c := compiler.New()
	mainCall := &parser.Program{
		Stmts: []parser.Stmt{
			&parser.ExprStmt{
				Expression: &parser.CallExpr{
					Callee: &parser.Identifier{Name: "main"},
					Args:   nil,
				},
			},
		},
	}
	fn, err := c.Compile(mainCall)
	if err != nil {
		return err
	}
	return machine.Run(fn)
}

package main

import (
	"fmt"
	"os"

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

	source, err := os.ReadFile(os.Args[1])
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %s\n", err)
		os.Exit(1)
	}

	if err := runFile(string(source)); err != nil {
		fmt.Fprintf(os.Stderr, "error: %s\n", err)
		os.Exit(1)
	}
}

func runFile(source string) error {
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

	c := compiler.New()
	fn, err := c.Compile(program)
	if err != nil {
		return err
	}

	// auto-call main() if it exists
	hasMain := false
	for _, stmt := range program.Stmts {
		if fd, ok := stmt.(*parser.FuncDecl); ok && fd.Name == "main" {
			hasMain = true
			break
		}
	}

	machine := vm.New()
	stdlib.NewRegistry().RegisterAll(machine)

	if err := machine.Run(fn); err != nil {
		return err
	}

	if hasMain {
		// main was stored as a global; now call it
		return callMain(machine)
	}

	return nil
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

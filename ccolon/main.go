package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/TRC-Loop/ccolon/bytecode"
	"github.com/TRC-Loop/ccolon/compiler"
	"github.com/TRC-Loop/ccolon/formatter"
	"github.com/TRC-Loop/ccolon/lexer"
	"github.com/TRC-Loop/ccolon/parser"
	"github.com/TRC-Loop/ccolon/pkg"
	"github.com/TRC-Loop/ccolon/repl"
	"github.com/TRC-Loop/ccolon/stdlib"
	"github.com/TRC-Loop/ccolon/vm"
)

const version = "1.0.0"

func main() {
	if len(os.Args) < 2 {
		machine := vm.New()
		stdlib.NewRegistry().RegisterAll(machine)
		machine.RegisterBuiltinError()
		repl.Start(os.Stdin, os.Stdout, machine, version)
		return
	}

	switch os.Args[1] {
	case "--version", "-v":
		fmt.Printf("CColon v%s\n", version)
	case "--help", "-h":
		printHelp()
	case "fmt":
		runFormatter(os.Args[2:])
	case "compile":
		runCompile(os.Args[2:])
	case "pkg":
		runPkg(os.Args[2:])
	default:
		filePath := os.Args[1]
		absPath, err := filepath.Abs(filePath)
		if err != nil {
			fatal(err)
		}

		if strings.HasSuffix(absPath, ".cclb") {
			if err := runBytecodeFile(absPath); err != nil {
				fatal(err)
			}
			return
		}

		source, err := os.ReadFile(absPath)
		if err != nil {
			fatal(err)
		}
		if err := runFile(string(source), filepath.Dir(absPath)); err != nil {
			fatal(err)
		}
	}
}

func fatal(err error) {
	fmt.Fprintf(os.Stderr, "error: %s\n", err)
	os.Exit(1)
}

func printHelp() {
	fmt.Println("Usage: ccolon [command] [options]")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  <file.ccl>              Run a CColon source file")
	fmt.Println("  <file.cclb>             Run a compiled bytecode file")
	fmt.Println("  fmt <file.ccl>          Format a source file")
	fmt.Println("  compile <file>          Compile to bytecode (.cclb)")
	fmt.Println("  pkg install <url[@ver]> Install a package from GitHub")
	fmt.Println("  pkg remove <name>       Remove an installed package")
	fmt.Println("  pkg list                List installed packages")
	fmt.Println("  pkg init                Create a ccolon.json")
	fmt.Println()
	fmt.Println("Options:")
	fmt.Println("  --version, -v    Print version")
	fmt.Println("  --help, -h       Show this help")
	fmt.Println()
	fmt.Println("No arguments starts the interactive REPL.")
}

// --- Formatter ---

func runFormatter(args []string) {
	check := false
	var files []string
	for _, arg := range args {
		if arg == "--check" {
			check = true
		} else {
			files = append(files, arg)
		}
	}
	if len(files) == 0 {
		fmt.Fprintln(os.Stderr, "usage: ccolon fmt [--check] <file.ccl> ...")
		os.Exit(1)
	}

	cfg := formatter.LoadConfig(".")
	exitCode := 0
	for _, f := range files {
		source, err := os.ReadFile(f)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %s\n", err)
			exitCode = 1
			continue
		}
		formatted, err := formatter.Format(string(source), cfg)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error formatting %s: %s\n", f, err)
			exitCode = 1
			continue
		}
		if check {
			if string(source) != formatted {
				fmt.Fprintf(os.Stderr, "%s: not formatted\n", f)
				exitCode = 1
			}
		} else {
			if err := os.WriteFile(f, []byte(formatted), 0644); err != nil {
				fmt.Fprintf(os.Stderr, "error writing %s: %s\n", f, err)
				exitCode = 1
			}
		}
	}
	os.Exit(exitCode)
}

// --- Compiler ---

func runCompile(args []string) {
	platform := ""
	var files []string
	for i := 0; i < len(args); i++ {
		if args[i] == "--platform" {
			platform = fmt.Sprintf("%s/%s", os.Getenv("GOOS"), os.Getenv("GOARCH"))
			if platform == "/" {
				platform = "unknown"
			}
		} else {
			files = append(files, args[i])
		}
	}
	if len(files) == 0 {
		fmt.Fprintln(os.Stderr, "usage: ccolon compile [--platform] <file.ccl> ...")
		os.Exit(1)
	}

	for _, f := range files {
		source, err := os.ReadFile(f)
		if err != nil {
			fatal(err)
		}
		fn, err := compileSource(string(source))
		if err != nil {
			fatal(err)
		}
		data, err := bytecode.Encode(fn, version, platform)
		if err != nil {
			fatal(err)
		}
		outPath := strings.TrimSuffix(f, filepath.Ext(f)) + ".cclb"
		if err := os.WriteFile(outPath, data, 0644); err != nil {
			fatal(err)
		}
		fmt.Printf("compiled %s -> %s\n", f, outPath)
	}
}

// --- Package Manager ---

func runPkg(args []string) {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "usage: ccolon pkg <install|remove|list|init> [args]")
		os.Exit(1)
	}

	switch args[0] {
	case "install":
		if len(args) < 2 {
			fmt.Fprintln(os.Stderr, "usage: ccolon pkg install <github-url[@version]>")
			os.Exit(1)
		}
		repoURL, version := pkg.ParseInstallArg(args[1])
		if err := pkg.Install(repoURL, version); err != nil {
			fatal(err)
		}
	case "remove":
		if len(args) < 2 {
			fmt.Fprintln(os.Stderr, "usage: ccolon pkg remove <name>")
			os.Exit(1)
		}
		if err := pkg.Remove(args[1]); err != nil {
			fatal(err)
		}
	case "list":
		packages, err := pkg.List()
		if err != nil {
			fatal(err)
		}
		if len(packages) == 0 {
			fmt.Println("no packages installed")
			return
		}
		for _, p := range packages {
			fmt.Printf("  %s@%s  (%s)\n", p.Name, p.Version, p.Path)
		}
	case "init":
		if err := pkg.Init(); err != nil {
			fatal(err)
		}
	default:
		fmt.Fprintf(os.Stderr, "unknown pkg command: %s\n", args[0])
		os.Exit(1)
	}
}

// --- Run bytecode ---

func runBytecodeFile(absPath string) error {
	data, err := os.ReadFile(absPath)
	if err != nil {
		return err
	}
	fn, fileVersion, err := bytecode.Decode(data)
	if err != nil {
		return fmt.Errorf("invalid bytecode file: %s", err)
	}
	_ = fileVersion // version info available if needed

	machine := vm.New()
	stdlib.NewRegistry().RegisterAll(machine)
	machine.RegisterBuiltinError()

	machine.FileLoader = func(path string) (*compiler.FuncObject, error) {
		resolved := path
		if !filepath.IsAbs(path) {
			resolved = filepath.Join(filepath.Dir(absPath), path)
		}
		d, err := os.ReadFile(resolved)
		if err != nil {
			return nil, fmt.Errorf("cannot read '%s': %s", path, err.Error())
		}
		return compileSource(string(d))
	}

	if err := machine.Run(fn); err != nil {
		return err
	}
	return callMain(machine)
}

// --- Source execution ---

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

package repl

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/peterh/liner"

	"github.com/TRC-Loop/ccolon/compiler"
	"github.com/TRC-Loop/ccolon/lexer"
	"github.com/TRC-Loop/ccolon/parser"
	"github.com/TRC-Loop/ccolon/vm"
)

const (
	prompt         = "\033[35mc: > \033[0m"
	continuePrompt = "\033[35m...  \033[0m"
)

var keywords = []string{
	"var", "function", "class", "if", "else", "while", "for", "in",
	"return", "break", "continue", "import", "true", "false", "nil",
	"None", "and", "or", "not", "try", "catch", "throw", "with", "as",
	"extends", "self", "super", "fixed", "range",
	"int", "float", "string", "bool", "list", "array", "dict",
	"public", "private",
}

var builtinModules = []string{
	"console", "math", "random", "json", "fs", "datetime", "os", "http",
}

func historyPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".ccolon_history")
}

func Start(in io.Reader, out io.Writer, machine *vm.VM, version string) {
	line := liner.NewLiner()
	defer line.Close()

	line.SetCtrlCAborts(true)
	line.SetMultiLineMode(true)

	// load history
	if hp := historyPath(); hp != "" {
		if f, err := os.Open(hp); err == nil {
			line.ReadHistory(f)
			f.Close()
		}
	}

	// tab completion
	importedModules := map[string]bool{}
	line.SetCompleter(func(input string) []string {
		var candidates []string
		// complete module methods
		for mod := range importedModules {
			if strings.HasPrefix(input, mod+".") {
				prefix := mod + "."
				partial := strings.TrimPrefix(input, prefix)
				if m := machine.GetModule(mod); m != nil {
					for name := range m.Methods {
						if strings.HasPrefix(name, partial) {
							candidates = append(candidates, prefix+name+"(")
						}
					}
				}
				return candidates
			}
		}
		// complete keywords, modules, "exit"
		all := append(keywords, builtinModules...)
		all = append(all, "exit")
		for _, kw := range all {
			if strings.HasPrefix(kw, input) {
				candidates = append(candidates, kw)
			}
		}
		return candidates
	})

	fmt.Fprintf(out, "CColon v%s - Interactive Mode\n", version)
	fmt.Fprintf(out, "Type 'exit' to quit.\n\n")

	for {
		input, err := line.Prompt(prompt)
		if err != nil {
			if err == liner.ErrPromptAborted {
				continue
			}
			fmt.Fprintln(out)
			break
		}

		trimmed := strings.TrimSpace(input)
		if trimmed == "exit" {
			break
		}
		if trimmed == "" {
			continue
		}

		// multi-line input
		source := input
		depth := braceDepth(source)
		for depth > 0 {
			extra, err := line.Prompt(continuePrompt)
			if err != nil {
				break
			}
			source += "\n" + extra
			depth = braceDepth(source)
		}

		line.AppendHistory(source)

		// track imports for completion
		for _, ln := range strings.Split(source, "\n") {
			ln = strings.TrimSpace(ln)
			if strings.HasPrefix(ln, "import ") {
				mod := strings.TrimSpace(strings.TrimPrefix(ln, "import"))
				if !strings.HasPrefix(mod, "\"") {
					importedModules[mod] = true
				}
			}
		}

		if err := runSource(source, machine, out); err != nil {
			fmt.Fprintf(out, "\033[31merror: %s\033[0m\n", err)
		}
	}

	// save history
	if hp := historyPath(); hp != "" {
		if f, err := os.Create(hp); err == nil {
			line.WriteHistory(f)
			f.Close()
		}
	}
}

func braceDepth(source string) int {
	depth := 0
	inString := false
	inLineComment := false
	inBlockComment := false
	runes := []rune(source)
	for i := 0; i < len(runes); i++ {
		ch := runes[i]
		next := rune(0)
		if i+1 < len(runes) {
			next = runes[i+1]
		}

		if inLineComment {
			if ch == '\n' {
				inLineComment = false
			}
			continue
		}
		if inBlockComment {
			if ch == '*' && next == '/' {
				inBlockComment = false
				i++
			}
			continue
		}
		if inString {
			if ch == '\\' {
				i++
			} else if ch == '"' {
				inString = false
			}
			continue
		}

		if ch == '/' && next == '/' {
			inLineComment = true
			i++
		} else if ch == '/' && next == '*' {
			inBlockComment = true
			i++
		} else if ch == '"' {
			inString = true
		} else if ch == '{' {
			depth++
		} else if ch == '}' {
			depth--
		}
	}
	return depth
}

func runSource(source string, machine *vm.VM, out io.Writer) error {
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

	return machine.Run(fn)
}

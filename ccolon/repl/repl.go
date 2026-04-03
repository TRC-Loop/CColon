package repl

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/chzyer/readline"

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
	"var", "const", "function", "class", "if", "else", "while", "for", "in",
	"return", "break", "continue", "import", "from", "true", "false", "nil",
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
	importedModules := map[string]bool{}

	completer := readline.NewPrefixCompleter()
	allWords := append(keywords, builtinModules...)
	allWords = append(allWords, "exit")
	var items []readline.PrefixCompleterInterface
	for _, w := range allWords {
		items = append(items, readline.PcItem(w))
	}
	completer.SetChildren(items)

	rl, err := readline.NewEx(&readline.Config{
		Prompt:            prompt,
		HistoryFile:       historyPath(),
		AutoComplete:      completer,
		InterruptPrompt:   "^C",
		EOFPrompt:         "exit",
		HistorySearchFold: true,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "repl: failed to initialize: %s\n", err)
		return
	}
	defer rl.Close()

	fmt.Fprintf(out, "CColon v%s - Interactive Mode\n", version)
	fmt.Fprintf(out, "Type 'exit' to quit.\n\n")

	for {
		input, err := rl.Readline()
		if err != nil {
			if err == readline.ErrInterrupt {
				continue
			}
			// EOF or other error
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
			rl.SetPrompt(continuePrompt)
			extra, err := rl.Readline()
			if err != nil {
				break
			}
			source += "\n" + extra
			depth = braceDepth(source)
		}
		rl.SetPrompt(prompt)

		// track imports for completion
		for _, ln := range strings.Split(source, "\n") {
			ln = strings.TrimSpace(ln)
			if strings.HasPrefix(ln, "import ") {
				mod := strings.TrimSpace(strings.TrimPrefix(ln, "import"))
				mod = strings.Trim(mod, "\"")
				if mod != "" {
					importedModules[mod] = true
				}
			}
		}
		_ = importedModules

		if err := runSource(source, machine, out); err != nil {
			fmt.Fprintf(out, "\033[31merror: %s\033[0m\n", err)
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

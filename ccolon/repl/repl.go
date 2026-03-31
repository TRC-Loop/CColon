package repl

import (
	"bufio"
	"fmt"
	"io"
	"strings"

	"github.com/TRC-Loop/ccolon/compiler"
	"github.com/TRC-Loop/ccolon/lexer"
	"github.com/TRC-Loop/ccolon/parser"
	"github.com/TRC-Loop/ccolon/vm"
)

const prompt = "c: > "
const continuePrompt = "...  "

func Start(in io.Reader, out io.Writer, machine *vm.VM) {
	scanner := bufio.NewScanner(in)
	fmt.Fprintf(out, "CColon v0.1.0 - Interactive Mode\n")
	fmt.Fprintf(out, "Type 'exit' to quit.\n\n")

	for {
		fmt.Fprint(out, prompt)
		if !scanner.Scan() {
			fmt.Fprintln(out)
			return
		}

		line := scanner.Text()
		if strings.TrimSpace(line) == "exit" {
			return
		}

		// multi-line: track brace depth
		source := line
		depth := braceDepth(source)
		for depth > 0 {
			fmt.Fprint(out, continuePrompt)
			if !scanner.Scan() {
				break
			}
			source += "\n" + scanner.Text()
			depth = braceDepth(source)
		}

		if strings.TrimSpace(source) == "" {
			continue
		}

		if err := runSource(source, machine, out); err != nil {
			fmt.Fprintf(out, "error: %s\n", err)
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
				i++ // skip escaped char
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

package bytecode

import (
	"testing"

	"github.com/TRC-Loop/ccolon/compiler"
	"github.com/TRC-Loop/ccolon/lexer"
	"github.com/TRC-Loop/ccolon/parser"
)

func compile(t *testing.T, source string) *compiler.FuncObject {
	t.Helper()
	l := lexer.New(source)
	tokens, err := l.Tokenize()
	if err != nil {
		t.Fatalf("lex error: %v", err)
	}
	p := parser.New(tokens)
	prog, err := p.Parse()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	c := compiler.New()
	fn, err := c.Compile(prog)
	if err != nil {
		t.Fatalf("compile error: %v", err)
	}
	return fn
}

func TestRoundTrip(t *testing.T) {
	source := `
import console
function greet(string name) {
    console.println("Hello " + name)
}
greet("world")
`
	fn := compile(t, source)

	data, err := Encode(fn, "0.2.3", "")
	if err != nil {
		t.Fatalf("encode error: %v", err)
	}

	decoded, ver, err := Decode(data)
	if err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if ver != "0.2.3" {
		t.Errorf("version mismatch: got %q, want %q", ver, "0.2.3")
	}
	if decoded.Name != fn.Name {
		t.Errorf("name mismatch: got %q, want %q", decoded.Name, fn.Name)
	}
	if len(decoded.Code) != len(fn.Code) {
		t.Errorf("code length mismatch: got %d, want %d", len(decoded.Code), len(fn.Code))
	}
	for i := range fn.Code {
		if decoded.Code[i] != fn.Code[i] {
			t.Errorf("code byte %d mismatch: got %d, want %d", i, decoded.Code[i], fn.Code[i])
			break
		}
	}
	if len(decoded.Constants) != len(fn.Constants) {
		t.Errorf("constants length mismatch: got %d, want %d", len(decoded.Constants), len(fn.Constants))
	}
}

func TestRoundTripWithClass(t *testing.T) {
	source := `
class Dog {
    var public string name = ""
    public function init(string name) {
        self.name = name
    }
    public function speak() string {
        return self.name + " barks!"
    }
}
`
	fn := compile(t, source)

	data, err := Encode(fn, "0.2.3", "darwin/arm64")
	if err != nil {
		t.Fatalf("encode error: %v", err)
	}

	decoded, _, err := Decode(data)
	if err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if len(decoded.Code) != len(fn.Code) {
		t.Errorf("code length mismatch: got %d, want %d", len(decoded.Code), len(fn.Code))
	}
}

func TestInvalidMagic(t *testing.T) {
	_, _, err := Decode([]byte("XXXX"))
	if err == nil {
		t.Fatal("expected error for invalid magic")
	}
}

func TestTruncatedData(t *testing.T) {
	fn := compile(t, `var int x = 42`)
	data, _ := Encode(fn, "0.2.3", "")
	_, _, err := Decode(data[:10])
	if err == nil {
		t.Fatal("expected error for truncated data")
	}
}

func TestOptionalArgs(t *testing.T) {
	source := `
function greet(string name, string prefix = "Hello") string {
    return prefix + " " + name
}
`
	fn := compile(t, source)
	data, err := Encode(fn, "0.2.3", "")
	if err != nil {
		t.Fatalf("encode error: %v", err)
	}
	decoded, _, err := Decode(data)
	if err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if len(decoded.Constants) != len(fn.Constants) {
		t.Errorf("constants mismatch: got %d, want %d", len(decoded.Constants), len(fn.Constants))
	}
}

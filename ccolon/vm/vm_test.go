package vm_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/TRC-Loop/ccolon/compiler"
	"github.com/TRC-Loop/ccolon/lexer"
	"github.com/TRC-Loop/ccolon/parser"
	"github.com/TRC-Loop/ccolon/stdlib"
	"github.com/TRC-Loop/ccolon/vm"
)

func compileSource(t *testing.T, source string) *compiler.FuncObject {
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

func runAndCapture(t *testing.T, source string) (string, error) {
	t.Helper()
	fn := compileSource(t, source)
	machine := vm.New()
	stdlib.NewRegistry().RegisterAll(machine)
	machine.RegisterBuiltinError()

	// redirect console output by replacing module methods
	var buf bytes.Buffer
	if mod := machine.GetModule("console"); mod != nil {
		mod.Methods["println"] = &vm.NativeFuncValue{
			Name: "println",
			Fn: func(args []vm.Value) (vm.Value, error) {
				parts := make([]string, len(args))
				for i, a := range args {
					parts[i] = a.String()
				}
				buf.WriteString(strings.Join(parts, " ") + "\n")
				return &vm.NilValue{}, nil
			},
		}
		mod.Methods["print"] = &vm.NativeFuncValue{
			Name: "print",
			Fn: func(args []vm.Value) (vm.Value, error) {
				parts := make([]string, len(args))
				for i, a := range args {
					parts[i] = a.String()
				}
				buf.WriteString(strings.Join(parts, " "))
				return &vm.NilValue{}, nil
			},
		}
	}

	err := machine.Run(fn)
	return buf.String(), err
}

func runExpect(t *testing.T, source, expected string) {
	t.Helper()
	out, err := runAndCapture(t, source)
	if err != nil {
		t.Fatalf("runtime error: %v", err)
	}
	got := strings.TrimRight(out, "\n")
	exp := strings.TrimRight(expected, "\n")
	if got != exp {
		t.Errorf("output mismatch:\ngot:  %q\nwant: %q", got, exp)
	}
}

func runExpectError(t *testing.T, source, errContains string) {
	t.Helper()
	_, err := runAndCapture(t, source)
	if err == nil {
		t.Fatalf("expected error containing %q, got none", errContains)
	}
	if !strings.Contains(err.Error(), errContains) {
		t.Errorf("error %q does not contain %q", err.Error(), errContains)
	}
}

// --- Arithmetic ---

func TestBasicArithmetic(t *testing.T) {
	runExpect(t, `
import console
console.println((2 + 3).tostring())
console.println((10 - 4).tostring())
console.println((3 * 7).tostring())
console.println((15 / 4).tostring())
console.println((17 % 5).tostring())
`, "5\n6\n21\n3\n2")
}

func TestFloatArithmetic(t *testing.T) {
	runExpect(t, `
import console
console.println((1.5 + 2.5).tostring())
console.println((10.0 / 3.0).tostring())
`, "4\n3.3333333333333335")
}

func TestIntOverflowAdd(t *testing.T) {
	runExpectError(t, `
var int a = 9223372036854775807
var int b = 1
var int c = a + b
`, "integer overflow")
}

func TestIntOverflowMul(t *testing.T) {
	runExpectError(t, `
var int a = 9223372036854775807
var int b = 2
var int c = a * b
`, "integer overflow")
}

func TestIntOverflowSub(t *testing.T) {
	runExpectError(t, `
var int a = -9223372036854775807
var int b = 9223372036854775807
var int c = a - b
`, "integer overflow")
}

func TestOverflowCatchable(t *testing.T) {
	runExpect(t, `
import console
try {
    var int a = 9223372036854775807
    var int b = a + 1
} catch (Error e) {
    console.println("caught: " + e.message)
}
`, "caught: integer overflow")
}

func TestDivisionByZero(t *testing.T) {
	runExpectError(t, `var int x = 10 / 0`, "division by zero")
}

func TestModuloByZero(t *testing.T) {
	runExpectError(t, `var int x = 10 % 0`, "modulo by zero")
}

// --- Strings ---

func TestStringConcat(t *testing.T) {
	runExpect(t, `
import console
console.println("hello" + " " + "world")
`, "hello world")
}

func TestStringLength(t *testing.T) {
	runExpect(t, `
import console
console.println("hello".length().tostring())
`, "5")
}

func TestStringIndex(t *testing.T) {
	runExpect(t, `
import console
var string s = "hello"
console.println(s[0])
console.println(s[4])
`, "h\no")
}

// --- Variables ---

func TestVariableTypes(t *testing.T) {
	runExpect(t, `
import console
var int a = 42
var float b = 3.14
var string c = "hi"
var bool d = true
console.println(a.tostring())
console.println(b.tostring())
console.println(c)
console.println(d.tostring())
`, "42\n3.14\nhi\ntrue")
}

// --- Control Flow ---

func TestIfElse(t *testing.T) {
	runExpect(t, `
import console
var int x = 10
if (x > 5) {
    console.println("big")
} else {
    console.println("small")
}
`, "big")
}

func TestWhileLoop(t *testing.T) {
	runExpect(t, `
import console
var int i = 0
var int sum = 0
while (i < 5) {
    sum = sum + i
    i = i + 1
}
console.println(sum.tostring())
`, "10")
}

func TestForInRange(t *testing.T) {
	runExpect(t, `
import console
var int sum = 0
for i in range(5) {
    sum = sum + i
}
console.println(sum.tostring())
`, "10")
}

func TestForInList(t *testing.T) {
	runExpect(t, `
import console
var list items = ["a", "b", "c"]
var string result = ""
for item in items {
    result = result + item
}
console.println(result)
`, "abc")
}

func TestBreakContinue(t *testing.T) {
	runExpect(t, `
import console
var int sum = 0
for i in range(10) {
    if (i == 5) { break }
    if (i % 2 == 0) { continue }
    sum = sum + i
}
console.println(sum.tostring())
`, "4")
}

// --- Functions ---

func TestFunctionReturn(t *testing.T) {
	runExpect(t, `
import console
function add(int a, int b) int {
    return a + b
}
console.println(add(3, 4).tostring())
`, "7")
}

func TestRecursion(t *testing.T) {
	runExpect(t, `
import console
function fib(int n) int {
    if (n <= 1) { return n }
    return fib(n - 1) + fib(n - 2)
}
console.println(fib(10).tostring())
`, "55")
}

func TestOptionalArgs(t *testing.T) {
	runExpect(t, `
import console
function greet(string name, string prefix = "Hello") string {
    return prefix + " " + name
}
console.println(greet("world"))
console.println(greet("world", "Hi"))
`, "Hello world\nHi world")
}

// --- Lists and Arrays ---

func TestListOps(t *testing.T) {
	runExpect(t, `
import console
var list l = [1, 2, 3]
l.append(4)
console.println(l.length().tostring())
var int last = l.pop()
console.println(last.tostring())
console.println(l.tostring())
`, "4\n4\n[1, 2, 3]")
}

func TestFixedArray(t *testing.T) {
	runExpect(t, `
import console
var array a = fixed([10, 20, 30])
console.println(a[1].tostring())
console.println(a.length().tostring())
`, "20\n3")
}

// --- Dicts ---

func TestDictBasic(t *testing.T) {
	runExpect(t, `
import console
var dict d = {"a": 1, "b": 2}
console.println(d["a"].tostring())
console.println(d.has("b").tostring())
console.println(d.length().tostring())
d["c"] = 3
console.println(d.length().tostring())
`, "1\ntrue\n2\n3")
}

func TestDictIteration(t *testing.T) {
	runExpect(t, `
import console
var dict d = {"x": 10, "y": 20}
var int sum = 0
for k in d {
    sum = sum + d[k]
}
console.println(sum.tostring())
`, "30")
}

func TestDictMethods(t *testing.T) {
	runExpect(t, `
import console
var dict d = {"a": 1, "b": 2}
console.println(d.keys().tostring())
console.println(d.values().tostring())
console.println(d.has("a").tostring())
console.println(d.has("c").tostring())
`, "[\"a\", \"b\"]\n[1, 2]\ntrue\nfalse")
}

// --- Classes ---

func TestClassBasic(t *testing.T) {
	runExpect(t, `
import console
class Dog {
    var public string name = ""
    public function init(string name) {
        self.name = name
    }
    public function speak() string {
        return self.name + " barks!"
    }
}
var Dog d = Dog("Rex")
console.println(d.speak())
`, "Rex barks!")
}

func TestClassInheritance(t *testing.T) {
	runExpect(t, `
import console
class Animal {
    var public string name = ""
    public function init(string name) {
        self.name = name
    }
    public function speak() string {
        return self.name + " makes a sound"
    }
}
class Cat extends Animal {
    public function init(string name) {
        super.init(name)
    }
    public function speak() string {
        return self.name + " meows"
    }
}
var Cat c = Cat("Whiskers")
console.println(c.speak())
console.println(c.name)
`, "Whiskers meows\nWhiskers")
}

// --- Error Handling ---

func TestTryCatch(t *testing.T) {
	runExpect(t, `
import console
try {
    throw Error("oops")
} catch (Error e) {
    console.println("caught: " + e.message)
}
`, "caught: oops")
}

func TestCustomError(t *testing.T) {
	runExpect(t, `
import console
class MyError extends Error {
    var public int code = 0
    public function init(string message, int code) {
        super.init(message)
        self.code = code
    }
}
try {
    throw MyError("not found", 404)
} catch (Error e) {
    console.println("caught: " + e.message)
}
`, "caught: not found")
}

func TestUncaughtError(t *testing.T) {
	runExpectError(t, `throw Error("boom")`, "uncaught Error")
}

// --- Imports ---

func TestModuleImport(t *testing.T) {
	runExpect(t, `
import console
import math
console.println(math.abs(-42).tostring())
`, "42")
}

func TestUnimportedModuleHint(t *testing.T) {
	runExpectError(t, `console.println("hi")`, "did you forget")
}

// --- Type Conversions ---

func TestTypeConversions(t *testing.T) {
	runExpect(t, `
import console
console.println("42".toint().tostring())
console.println("3.14".tofloat().tostring())
var int n = 7
console.println(n.tofloat().tostring())
`, "42\n3.14\n7")
}

// --- Edge Cases ---

func TestEmptyList(t *testing.T) {
	runExpect(t, `
import console
var list l = []
console.println(l.length().tostring())
l.append(1)
console.println(l.length().tostring())
`, "0\n1")
}

func TestNestedMethodCalls(t *testing.T) {
	runExpect(t, `
import console
var list l = [1, 2, 3]
console.println(l.length().tostring())
`, "3")
}

func TestNegation(t *testing.T) {
	runExpect(t, `
import console
var int x = 42
console.println((-x).tostring())
console.println((-(- x)).tostring())
`, "-42\n42")
}

func TestBooleanLogic(t *testing.T) {
	runExpect(t, `
import console
console.println((true and false).tostring())
console.println((true or false).tostring())
console.println((not true).tostring())
`, "false\ntrue\nfalse")
}

// --- Sint ---

func TestSintLiteral(t *testing.T) {
	runExpect(t, `
import console
var sint x = 99999999999999999999999999999999
console.println(x.tostring())
`, "99999999999999999999999999999999")
}

func TestSintArithmetic(t *testing.T) {
	runExpect(t, `
import console
var sint a = 99999999999999999999
var sint b = 1
console.println((a + b).tostring())
console.println((a * b).tostring())
console.println((a - b).tostring())
`, "100000000000000000000\n99999999999999999999\n99999999999999999998")
}

func TestSintIntPromotion(t *testing.T) {
	runExpect(t, `
import console
var sint big = 99999999999999999999
var int small = 1
console.println((big + small).tostring())
`, "100000000000000000000")
}

func TestSintComparison(t *testing.T) {
	runExpect(t, `
import console
var sint a = 99999999999999999999
var sint b = 99999999999999999998
console.println((a > b).tostring())
console.println((a == b).tostring())
`, "true\nfalse")
}

func TestSintMethods(t *testing.T) {
	runExpect(t, `
import console
var sint x = 42.tosint()
console.println(x.tostring())
console.println(x.toint().tostring())
console.println(x.abs().tostring())
var sint neg = -99999999999999999999
console.println(neg.abs().tostring())
`, "42\n42\n42\n99999999999999999999")
}

func TestSintFromString(t *testing.T) {
	runExpect(t, `
import console
var sint x = "123456789012345678901234567890".tosint()
console.println(x.tostring())
`, "123456789012345678901234567890")
}

func TestSintPow(t *testing.T) {
	runExpect(t, `
import console
var sint base = 2.tosint()
var sint result = base.pow(100.tosint())
console.println(result.tostring())
`, "1267650600228229401496703205376")
}

func TestStringEscapes(t *testing.T) {
	runExpect(t, `
import console
console.println("hello\tworld")
console.println("line1\nline2")
`, "hello\tworld\nline1\nline2")
}

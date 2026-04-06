<p align="center">
  <img src="ccolon.svg" alt="CColon Logo" width="200">
</p>

<h1 align="center">CColon (C:)</h1>

<p align="center">
  A bytecode-compiled programming language built in Go.
  <br>
  Clean like Python, structured like Go, expressive like Kotlin.
</p>

<p align="center">
  <a href="https://ccolon.arne.sh/">Documentation</a> &bull;
  <a href="#installation">Installation</a> &bull;
  <a href="#quick-start">Quick Start</a> &bull;
  <a href="ccolon/examples/demo.ccl">Full Example</a>
</p>

---

## Installation

### Linux / macOS

```sh
curl -fsSL https://raw.githubusercontent.com/TRC-Loop/CColon/main/install.sh | sh
```

Or specify a version:

```sh
curl -fsSL https://raw.githubusercontent.com/TRC-Loop/CColon/main/install.sh | sh -s v1.0.0
```

### Windows (PowerShell)

```powershell
irm https://raw.githubusercontent.com/TRC-Loop/CColon/main/install.ps1 | iex
```

### Build from source

Requires Go 1.22 or later.

```sh
git clone https://github.com/TRC-Loop/CColon.git
cd CColon/ccolon
go build -o ccolon .
```

Move the binary somewhere on your `PATH`:

```sh
sudo mv ccolon /usr/local/bin/
```

## Quick Start

### Interactive Shell

Run `ccolon` with no arguments to start the REPL (with readline, history, and tab completion):

```
$ ccolon
CColon v1.0.0 - Interactive Mode
Type 'exit' to quit.

c: > import console
c: > console.println("Hello from C:!")
Hello from C:!
c: >
```

### Running a File

```sh
ccolon hello.ccl
```

**hello.ccl**:
```
import console

function main() {
    console.println("Hello, World!")
}
```

### LSP

Find LSP Instructions here:
https://github.com/TRC-Loop/ccl-lsp

## Features

| Feature | Description |
|---|---|
| **Bytecode compiled** | Compiles to bytecode, runs on a stack-based VM |
| **Static types** | `int`, `float`, `string`, `bool`, `list`, `array`, `dict`, `sint` |
| **Classes** | Inheritance, constructors, public/private fields and methods |
| **Error handling** | `try/catch/throw` with custom error classes |
| **File imports** | `import "file.ccl"` for multi-file projects |
| **Arbitrary precision** | `sint` type for unlimited-size integers (like Python) |
| **Code formatter** | `ccolon fmt` with configurable `.ccolonfmt` |
| **Bytecode files** | `ccolon compile` to `.cclb`, `ccolon file.cclb` to run |
| **Package manager** | `ccolon pkg install` from any GitHub repo |
| **Rich REPL** | Readline, history, tab completion, colored output |
| **Standard library** | console, math, random, json, fs, datetime, os, http |

## Language Overview

CColon source files use the `.ccl` extension.

### Types

| Type | Description | Example |
|---|---|---|
| `int` | 64-bit integer | `var int x = 42` |
| `sint` | Arbitrary precision integer | `var sint x = 99999999999999999999` |
| `float` | 64-bit float | `var float pi = 3.14` |
| `string` | Text | `var string s = "hello"` |
| `bool` | Boolean | `var bool ok = true` |
| `list` | Dynamic list | `var list a = [1, 2, 3]` |
| `array` | Fixed-size array | `var array a = fixed([1, 2])` |
| `dict` | Dictionary | `var dict d = {"a": 1}` |

### Functions

```
function add(int a, int b) int {
    return a + b
}

function greet(string name, string prefix = "Hello") {
    console.println(prefix + ", " + name + "!")
}
```

### Classes

```
class Dog {
    var public string name = ""
    public function init(string name) {
        self.name = name
    }
    public function speak() string {
        return self.name + " barks!"
    }
}
```

### Error Handling

```
try {
    throw Error("something went wrong")
} catch (Error e) {
    console.println("caught: " + e.message)
}
```

### Control Flow

```
if (x > 0) {
    console.println("positive")
} else {
    console.println("not positive")
}

while (count < 10) {
    count = count + 1
}

for i in range(5) {
    console.println(i.tostring())
}

for item in items {
    console.println(item)
}
```

## Tools

```sh
ccolon file.ccl              # Run a source file
ccolon file.cclb             # Run a compiled bytecode file
ccolon fmt file.ccl           # Format source code
ccolon compile file.ccl       # Compile to bytecode
ccolon pkg install <url>      # Install a package
ccolon pkg list               # List installed packages
```

For the full language reference, see the [documentation](https://ccolon.arne.sh/).

## License

[MIT](LICENSE)

## Logo

You might want to rotate the logo by 90.

This is a happy language indeed.

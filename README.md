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
curl -fsSL https://raw.githubusercontent.com/TRC-Loop/CColon/main/install.sh | sh -s v0.1.0
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

Run `ccolon` with no arguments to start the REPL:

```
$ ccolon
CColon v0.1.0 - Interactive Mode
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

## Language Overview

CColon source files use the `.ccl` extension.

### Types

| Type    | Description            | Example                        |
|---------|------------------------|--------------------------------|
| `int`   | 64-bit integer         | `var int x = 42`               |
| `float` | 64-bit float           | `var float pi = 3.14`          |
| `string`| Text                   | `var string s = "hello"`       |
| `bool`  | Boolean                | `var bool ok = true`           |
| `list`  | Dynamic list           | `var list a = [1, 2, 3]`      |
| `array` | Fixed-size array       | `var array a = fixed([1, 2])` |

### Variables

```
var int count = 0
var string name = "CColon"
count = count + 1
```

### Functions

```
function add(int a, int b) int {
    return a + b
}

function greet(string name) {
    console.println("Hello, " + name + "!")
}
```

The `main()` function is automatically called as the entry point.

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

for i in range(2, 8) {
    console.println(i.tostring())
}
```

### Lists and Arrays

```
var list items = [1, 2, 3]
items.append(4)
console.println(items[0].tostring())
console.println(items.length().tostring())

var array coords = fixed([10, 20, 30])
console.println(coords[1].tostring())
```

### Methods on Values

```
var int n = 42
console.println(n.tostring())
console.println(n.tofloat().tostring())

var string s = "hello"
console.println(s.length().tostring())
```

### Imports

```
import console
```

The `console` module provides `println`, `print`, and `scanp` for terminal I/O.

For the full language reference, see the [documentation](https://ccolon.arne.sh/).

## License

[MIT](LICENSE)

## Logo

You might want to rotate the logo by 90°.

This is a happy language indeed.

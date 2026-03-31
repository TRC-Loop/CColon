# Interactive Shell (REPL)

Run `ccolon` with no arguments to start the interactive shell:

```
$ ccolon
CColon v0.1.0 - Interactive Mode
Type 'exit' to quit.

c: >
```

## Basics

You can type any CColon statement or expression directly:

```
c: > import console
c: > var int x = 42
c: > console.println(x.tostring())
42
c: > x = x + 8
c: > console.println(x.tostring())
50
```

## Multi-line input

When you type a line that opens a block with `{`, the REPL waits for the closing `}` before executing:

```
c: > function square(int n) int {
...      return n * n
...  }
c: > console.println(square(7).tostring())
49
```

The `...` prompt indicates that the REPL is waiting for more input.

## Persistent state

Variables, functions, and imports persist across lines within the same session. Define a function once and call it as many times as you want.

## Exiting

Type `exit` or press `Ctrl+D` to quit.

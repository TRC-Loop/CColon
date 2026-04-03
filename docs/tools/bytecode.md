# Bytecode Compiler

CColon can compile source files to bytecode (`.cclb`) for faster loading and distribution without source code.

## Usage

```bash
// Compile a source file to bytecode
ccolon compile file.ccl
// Produces file.cclb

// Custom output file name
ccolon compile -o output.cclb file.ccl

// Compile with platform tag
ccolon compile --platform file.ccl

// Run a compiled file
ccolon file.cclb
```

### Flags

| Flag | Description |
|------|-------------|
| `-o`, `--output` | Set the output file path |
| `--platform` | Embed OS/architecture info |
| `--bundle` | Bundle imported libraries (reserved) |

## How it works

1. **Compile**: `ccolon compile` runs the lexer, parser, and compiler to produce a `FuncObject`, then serializes it to the CCLB binary format.
2. **Run**: When you run a `.cclb` file, CColon skips lexing/parsing/compiling and directly loads the bytecode into the VM.

## CCLB Format

The `.cclb` binary format has this structure:

| Offset | Size | Field |
|---|---|---|
| 0 | 4 | Magic bytes: `CCLB` |
| 4 | 2 | Format version (uint16, little-endian) |
| 6 | var | Language version (length-prefixed string) |
| var | var | Platform string (empty = portable) |
| var | var | Serialized FuncObject tree |

Constants use type tags: `0` = nil, `1` = int64, `2` = float64, `3` = string, `4` = function, `5` = class.

## Platform tags

By default, compiled bytecode is portable across platforms. Use `--platform` to embed the current OS/architecture for documentation purposes. The runtime does not enforce platform matching.

## Limitations

- File imports (`import "file.ccl"`) still require the imported source files to be available at runtime, since they are compiled on demand.
- The bytecode format may change between CColon versions. Compiled files from one version are not guaranteed to work with another.

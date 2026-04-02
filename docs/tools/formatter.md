# Formatter

CColon includes a built-in code formatter that enforces consistent style across your codebase.

## Usage

```bash
# Format a file in place
ccolon fmt file.ccl

# Format multiple files
ccolon fmt src/*.ccl

# Check formatting without modifying (exits with code 1 if unformatted)
ccolon fmt --check file.ccl
```

## Configuration

The formatter looks for a `.ccolonfmt` file in the current directory and parent directories. If none is found, it uses the defaults.

### `.ccolonfmt` format

```
# CColon formatter config
indent_size = 4
use_tabs = false
max_width = 100
```

### Options

| Option | Default | Description |
|---|---|---|
| `indent_size` | `4` | Number of spaces per indent level |
| `use_tabs` | `false` | Use tabs instead of spaces |
| `max_width` | `100` | Maximum line width (reserved for future use) |

## What the formatter does

- Normalizes indentation to your configured style
- Adds blank lines between top-level declarations (functions, classes, imports)
- Formats expressions with consistent spacing around operators
- Normalizes string literals to use double quotes
- Ensures files end with a newline

## CI Integration

Use `--check` mode in CI to enforce formatting:

```bash
ccolon fmt --check src/*.ccl || echo "Code is not formatted"
```

# CColon v1.2.0 Release Notes

## Bug Fixes

- **Fix HTTP handler panic**: Fixed an index out of range crash when VM callbacks (like HTTP handlers) executed. The VM now properly isolates callback frames and includes a bounds check in bytecode reading as a safety net.
- **Fix REPL instant quit**: Replaced the `peterh/liner` readline library with `chzyer/readline`, fixing the REPL immediately exiting on certain terminals.

## New Language Features

- **F-strings**: String interpolation with `f"text {expression}"`. Expressions inside `{}` are evaluated and automatically converted to strings.
- **String auto-coercion with `+`**: Using `+` with a string and any other type now automatically converts the non-string side to its string representation (`"count: " + 42` produces `"count: 42"`).
- **Constants**: Immutable variables with `const <type> <name> = <value>`. Reassignment is caught at compile time for local variables and at runtime for globals.
- **Keyword arguments**: Call functions with named parameters: `greet(name="World", greeting="Hello")`. Positional and keyword args can be mixed.
- **`from ... import` syntax**: Selectively import specific functions or properties from a module: `from math import sqrt, pi` or `from math import *`.
- **Math module properties**: `math.pi`, `math.e`, and `math.inf` are now accessed as properties (without parentheses) instead of function calls.

## Package Manager

- **`--local` flag**: Install packages into `./ccolon_packages/` for project-specific dependencies with `ccolon pkg install --local <url>`.
- **`pkg upgrade` command**: Re-install a package at its latest version.
- **Repository URL in `pkg list`**: The source GitHub URL is now shown when listing installed packages.
- **Improved progress bar**: Package downloads now show a nicer progress bar with block characters.
- **Packages as modules**: Installed packages are now registered as importable modules. Use `import <package>` or `from <package> import <function>`.

## Compile Command

- **`-o` / `--output` flag**: Set a custom output file path for compiled bytecode.
- **`--bundle` flag**: Reserved for future library bundling support.

## Other Changes

- Bytecode format bumped to version 2 (adds ParamNames and string list constant support).
- Updated all documentation with new features.

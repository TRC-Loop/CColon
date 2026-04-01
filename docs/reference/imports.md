# Imports and Modules

## Importing modules

Use the `import` statement to bring a module into scope:

```
import console
```

After importing, you can access the module's functions using dot notation:

```
console.println("Hello!")
```

Import statements should be placed at the top of the file, before any function definitions or other code.

## Available modules

| Module    | Description                              |
|-----------|------------------------------------------|
| `console` | Terminal output and user input           |
| `math`    | Mathematical functions and constants     |
| `random`  | Random number generation and selection   |
| `json`    | JSON parsing and serialization           |
| `fs`      | File system operations                   |

## How modules work

Modules in CColon are built-in packages provided by the runtime. Each module exposes a set of functions that you call through dot notation after importing.

You must import a module before using it. If you try to use `console.println` without `import console`, you will get a helpful error:

```
// this will fail:
function main() {
    console.println("oops")
    // error: undefined variable 'console' -- did you forget 'import console'?
}
```

## Importing other CColon files

You can import functions and classes from other `.ccl` files using a string path:

```
import "utils.ccl"
import "lib/helpers.ccl"
```

The path is resolved relative to the directory of the file doing the import. All top-level definitions (functions, classes) from the imported file become available in the importing file's global scope.

Each file is imported at most once, even if multiple `import` statements reference it.

### Example

`utils.ccl`:
```
function double(int x) int {
    return x * 2
}
```

`main.ccl`:
```
import console
import "utils.ccl"

function main() {
    console.println(double(5).tostring())    // 10
}
```

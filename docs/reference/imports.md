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

## Import aliases

Use `as` to give an imported module a different name:

```
import math as m

function main() {
    console.println(m.sqrt(16).tostring())  // 4
    console.println(m.pi)                   // 3.141592653589793
}
```

This is useful for shortening long module names or avoiding naming conflicts.

## Selective imports

Use `from ... import` to import specific functions or properties from a module:

```
from math import sqrt, pi

function main() {
    console.println(sqrt(16))  // 4
    console.println(pi)        // 3.141592653589793
}
```

You can also import everything from a module:

```
from math import *

function main() {
    console.println(pi)
    console.println(e)
    console.println(sqrt(25))
}
```

With selective imports, the imported names are available directly without the module prefix.

## Available modules

| Module     | Description                              |
|------------|------------------------------------------|
| `console`  | Terminal output and user input           |
| `math`     | Mathematical functions and constants     |
| `random`   | Random number generation and selection   |
| `json`     | JSON parsing and serialization           |
| `fs`       | File system operations                   |
| `datetime` | Date and time handling                   |
| `os`       | OS operations and environment            |
| `http`     | HTTP client and server                   |

## How modules work

Modules in CColon are built-in packages provided by the runtime. Each module exposes a set of functions that you call through dot notation after importing.

You must import a module before using it:

```
// this will fail:
function main() {
    console.println("oops")
    // error: undefined variable 'console'
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
    console.println(double(5).tostring())  // 10
}
```

## Importing packages

Installed packages are automatically available as modules. After installing a package with `ccolon pkg install`, you can import it like any other module:

```
import ccl_testpkg

function main() {
    ccl_testpkg.test()
}
```

Or use selective imports:

```
from ccl_testpkg import test

function main() {
    test()
}
```

See [Package Manager](../tools/packages.md) for details on installing packages.

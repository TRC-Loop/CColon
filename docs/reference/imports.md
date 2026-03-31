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

CColon currently ships with one built-in module:

| Module    | Description                              |
|-----------|------------------------------------------|
| `console` | Terminal output and user input           |

More standard library modules will be added in future releases.

## How modules work

Modules in CColon are not files. They are built-in packages provided by the runtime. Each module exposes a set of functions that you call through dot notation after importing.

You must import a module before using it. If you try to use `console.println` without `import console`, you will get a runtime error:

```
// this will fail:
function main() {
    console.println("oops")    // error: undefined variable 'console'
}
```

## Future: file imports

Support for importing other `.ccl` files as modules is planned for a future release.

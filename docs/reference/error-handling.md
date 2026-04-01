# Error Handling

CColon supports structured error handling with `try`, `catch`, and `throw`.

## The Error class

CColon has a built-in `Error` class with a `message` field:

```
throw Error("something went wrong")
```

## try/catch

Wrap code that might throw in a `try` block and handle errors in `catch`:

```
try {
    throw Error("oops")
} catch (Error e) {
    console.println("caught: " + e.message)
}
```

The catch block specifies a type and a variable name. The caught error is available as that variable inside the block.

If no error is thrown, the catch block is skipped. If an error is thrown and it matches the catch type (or is a subclass of it), execution jumps to the catch block. If the error does not match, it propagates up to the next enclosing try/catch or crashes the program.

## throw

Use `throw` to raise an error:

```
function divide(int a, int b) int {
    if (b == 0) {
        throw Error("division by zero")
    }
    return a / b
}
```

You can throw any instance of `Error` or a class that extends `Error`.

## Custom error classes

Define your own error types by extending `Error`:

```
class ValidationError extends Error {
    var public string field = ""

    public function init(string message, string field) {
        super.init(message)
        self.field = field
    }
}

function validate(int age) {
    if (age < 0) {
        throw ValidationError("must be positive", "age")
    }
}

function main() {
    try {
        validate(-1)
    } catch (Error e) {
        console.println("error: " + e.message)
    }
}
```

Since `ValidationError` extends `Error`, a `catch (Error e)` block will catch it.

## The with statement

The `with` statement manages resources that need cleanup. It calls `.close()` on the resource when the block exits, even if an error is thrown:

```
import fs

function main() {
    with fs.open("data.txt", "r") as f {
        var string content = f.read()
        console.println(content)
    }
    // f.close() is called automatically
}
```

This is equivalent to:

```
var File f = fs.open("data.txt", "r")
try {
    var string content = f.read()
    console.println(content)
} catch (Error e) {
    f.close()
    throw e
}
f.close()
```

Any object that has a `.close()` method can be used with `with`.

## Full example

```
import console

class AppError extends Error {
    var public int code = 0

    public function init(string message, int code) {
        super.init(message)
        self.code = code
    }
}

function riskyOperation() {
    throw AppError("not found", 404)
}

function main() {
    try {
        riskyOperation()
    } catch (Error e) {
        console.println("caught: " + e.message)
    }
}
```

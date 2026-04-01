# Functions

## Defining functions

Functions are declared with the `function` keyword:

```
function greet() {
    console.println("Hello!")
}
```

## Parameters

Parameters are listed with their types:

```
function add(int a, int b) int {
    return a + b
}
```

Multiple parameters are separated by commas. Each parameter must have a type annotation.

## Return values

To return a value, specify the return type after the parameter list and use `return`:

```
function square(int n) int {
    return n * n
}
```

Functions without a return type implicitly return `nil`.

A bare `return` (with no value) is also valid and returns `nil`:

```
function maybeReturn(bool condition) {
    if (condition) {
        return
    }
    console.println("condition was false")
}
```

## Calling functions

```
var int result = add(3, 4)
greet()
```

## Optional arguments

Parameters can have default values. Optional parameters must come after all required parameters:

```
function repeat(string text, int times = 1) string {
    var string result = ""
    for i in range(times) {
        result = result + text
    }
    return result
}
```

When calling, you can omit optional arguments to use their defaults:

```
repeat("ha", 3)    // "hahaha"
repeat("yo")       // "yo" (times defaults to 1)
```

## The main function

When running a `.ccl` file, CColon looks for a function named `main` and calls it automatically. This is the entry point for every program:

```
import console

function main() {
    console.println("Program starts here")
}
```

Functions defined outside of `main` are available globally and can be called from `main` or from each other.

## Recursion

Functions can call themselves:

```
function factorial(int n) int {
    if (n <= 1) {
        return 1
    }
    return n * factorial(n - 1)
}
```

There is no explicit recursion depth limit beyond available stack space.

## Functions as values

Functions are stored as global values. You can pass them around and call them indirectly:

```
function double(int n) int {
    return n * 2
}

function apply(int n) int {
    var int result = double(n)
    return result
}
```

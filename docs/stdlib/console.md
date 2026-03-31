# console

The `console` module provides functions for terminal output and user input.

```
import console
```

## Functions

### `console.println(...)`

Prints one or more values followed by a newline.

```
console.println("Hello, World!")
console.println(42.tostring())
console.println("Score:", score.tostring())
```

Each argument is converted to its string representation. Multiple arguments are separated by a space.

### `console.print(...)`

Prints one or more values without a trailing newline.

```
console.print("Loading")
console.print(".")
console.print(".")
console.println("")
// Output: Loading..
```

Useful for building output incrementally on a single line.

### `console.scanp(prompt)`

Prints a prompt string and reads a line of input from the user. Returns the input as a `string`.

```
var string name = console.scanp("Enter your name: ")
console.println("Hello, " + name + "!")
```

The returned string does not include the trailing newline.

If you need the input as a number, use `.toint()` or `.tofloat()`:

```
var string input = console.scanp("Enter a number: ")
var int n = input.toint()
```

## Full example

```
import console

function main() {
    var string name = console.scanp("What is your name? ")
    var string ageStr = console.scanp("How old are you? ")
    var int age = ageStr.toint()

    console.println("")
    console.println("Hello, " + name + "!")
    console.println("You will be " + (age + 1).tostring() + " next year.")
}
```

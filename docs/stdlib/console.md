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

### `console.flush()`

Forces the terminal to display any text currently held in the output buffer. This is necessary when using `console.print` for real time updates like progress bars where a newline is not present to trigger a display update.

```
console.print("Processing...")
console.flush()
```

### `console.clearLine()`

Moves the cursor to the beginning of the current line and clears all existing text on that line. This is the primary function used for creating progress bars or status counters that update in place.

```
console.clearLine()
console.print("Progress: 50%")
```

### `console.setCursor(row, col)`

Moves the terminal cursor to a specific coordinate. `row` and `col` should be passed as strings or integers.

```
console.setCursor(1, 1) // Moves cursor to the top left corner
console.print("Status: OK")
```

### `console.getCursor()`

Requests the current cursor position from the terminal. Returns a string in the format `"row,col"`.

```
var string pos = console.getCursor()
console.println("The cursor is at: " + pos)
```

### `console.getSize()`

Returns the current width and height of the terminal window as a string in the format `"width,height"`. This is helpful for dynamically sizing progress bars.

```
var string size = console.getSize()
console.println("Terminal size: " + size)
```

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

    console.println("Starting task...")
    for (var int i = 0; i <= 10; i = i + 1) {
        console.clearLine()
        console.print("Loading: " + (i * 10).tostring() + "%")
        console.flush()
        sleep(200) // Assumes a sleep function exists
    }
    console.println("\nDone!")
}
```

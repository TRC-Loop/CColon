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

### `console.flush()`

Forces the terminal to display any text currently held in the output buffer. Necessary when using `console.print` for real-time updates like progress bars.

```
console.print("Processing...")
console.flush()
```

### `console.clearLine()`

Moves the cursor to the beginning of the current line and clears it. Use this for progress bars or status counters that update in place.

```
console.clearLine()
console.print("Progress: 50%")
```

### `console.setCursor(int row, int col)`

Moves the terminal cursor to a specific position. Both arguments must be integers.

```
console.setCursor(1, 1)
console.print("Status: OK")
```

### `console.getCursor()`

Returns the current cursor position as a dict with `row` and `col` keys (both integers).

```
var dict pos = console.getCursor()
console.println("Row: " + pos["row"].tostring())
console.println("Col: " + pos["col"].tostring())
```

### `console.getSize()`

Returns the terminal dimensions as a dict with `width` and `height` keys (both integers).

```
var dict size = console.getSize()
console.println("Width: " + size["width"].tostring())
console.println("Height: " + size["height"].tostring())
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
import datetime

function main() {
    var string name = console.scanp("What is your name? ")
    console.println("Hello, " + name + "!")

    var dict size = console.getSize()
    console.println("Your terminal is " + size["width"].tostring() + "x" + size["height"].tostring())

    // Progress bar example
    for i in range(11) {
        console.clearLine()
        console.print("Loading: " + (i * 10).tostring() + "%")
        console.flush()
        datetime.sleep(200)
    }
    console.println("\nDone!")
}
```

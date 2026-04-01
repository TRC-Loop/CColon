# random

The `random` module provides random number generation and selection.

```
import random
```

## Functions

### `random.randint(min, max)`

Returns a random integer between `min` and `max` (inclusive).

```
var int roll = random.randint(1, 6)
```

### `random.randfloat()`

Returns a random float between 0.0 (inclusive) and 1.0 (exclusive).

```
var float r = random.randfloat()
```

### `random.choice(list)`

Returns a random element from a list.

```
var list colors = ["red", "green", "blue"]
var string picked = random.choice(colors)
```

Throws an error if the list is empty.

### `random.char(string)`

Returns a random character from a string.

```
var string c = random.char("abcdef")
```

Throws an error if the string is empty.

### `random.shuffle(list)`

Shuffles a list in place. Returns `nil`.

```
var list items = [1, 2, 3, 4, 5]
random.shuffle(items)
// items is now in random order
```

## Full example

```
import console
import random

function main() {
    // simulate rolling two dice
    var int d1 = random.randint(1, 6)
    var int d2 = random.randint(1, 6)
    console.println("Rolled: " + d1.tostring() + " + " + d2.tostring() + " = " + (d1 + d2).tostring())

    // pick a random item
    var list prizes = ["gold", "silver", "bronze"]
    console.println("You won: " + random.choice(prizes))
}
```

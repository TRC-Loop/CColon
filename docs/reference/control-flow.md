# Control Flow

## if / else

```
if (condition) {
    // runs when condition is true
}
```

With an `else` branch:

```
if (x > 0) {
    console.println("positive")
} else {
    console.println("zero or negative")
}
```

Chained conditions using `else if`:

```
if (score >= 90) {
    console.println("A")
} else if (score >= 80) {
    console.println("B")
} else if (score >= 70) {
    console.println("C")
} else {
    console.println("F")
}
```

The condition must be wrapped in parentheses. The body uses curly braces.

### Truthiness

Any value can be used as a condition. The following are considered false:

- `false`
- `0` (int)
- `0.0` (float)
- `""` (empty string)
- `nil`
- `[]` (empty list)
- `fixed([])` (empty array)

Everything else is true.

## while

```
var int i = 0
while (i < 10) {
    console.println(i.tostring())
    i = i + 1
}
```

The condition is checked before each iteration. If the condition is false from the start, the body never runs.

## for ... in range

Iterate over a range of integers.

### range with one argument

`range(n)` produces values from `0` to `n - 1`:

```
for i in range(5) {
    console.println(i.tostring())
}
// prints 0, 1, 2, 3, 4
```

### range with two arguments

`range(start, end)` produces values from `start` to `end - 1`:

```
for i in range(3, 7) {
    console.println(i.tostring())
}
// prints 3, 4, 5, 6
```

## for ... in (list iteration)

Iterate directly over the elements of a list:

```
var list fruits = ["apple", "banana", "cherry"]
for fruit in fruits {
    console.println(fruit)
}
```

## break

Exit a loop early:

```
for i in range(100) {
    if (i == 5) {
        break
    }
    console.println(i.tostring())
}
// prints 0, 1, 2, 3, 4
```

`break` exits only the innermost loop.

## continue

Skip to the next iteration:

```
for i in range(6) {
    if (i % 2 == 0) {
        continue
    }
    console.println(i.tostring())
}
// prints 1, 3, 5
```

`continue` skips the rest of the current iteration and moves to the next one. In `for` loops, the loop variable is still incremented.

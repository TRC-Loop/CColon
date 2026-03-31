# Operators

## Arithmetic

| Operator | Description    | Example         | Result   |
|----------|----------------|-----------------|----------|
| `+`      | Addition       | `3 + 4`         | `7`      |
| `-`      | Subtraction    | `10 - 3`        | `7`      |
| `*`      | Multiplication | `6 * 7`         | `42`     |
| `/`      | Division       | `15 / 4`        | `3`      |
| `%`      | Modulo         | `17 % 5`        | `2`      |
| `-`      | Negation       | `-x`            | negated  |

### Type behavior

When mixing `int` and `float` in arithmetic, the result is always a `float`:

```
var int a = 5
var float b = 2.5
var float c = a + b    // 7.5
```

Integer division truncates toward zero:

```
var int x = 15 / 4     // 3, not 3.75
var float y = 15.0 / 4 // 3.75
```

Division by zero produces a runtime error.

### String concatenation

The `+` operator concatenates two strings:

```
var string full = "Hello" + ", " + "World!"    // "Hello, World!"
```

Both sides must be strings. To concatenate a number with a string, convert it first:

```
console.println("Score: " + score.tostring())
```

## Comparison

| Operator | Description        | Example    |
|----------|--------------------|------------|
| `==`     | Equal              | `x == y`   |
| `!=`     | Not equal          | `x != y`   |
| `<`      | Less than          | `x < y`    |
| `>`      | Greater than       | `x > y`    |
| `<=`     | Less or equal      | `x <= y`   |
| `>=`     | Greater or equal   | `x >= y`   |

Comparison operators return a `bool`. Numeric comparisons work across `int` and `float`:

```
var bool result = 5 == 5.0    // true
```

Equality for strings compares by value:

```
var bool same = "abc" == "abc"    // true
```

## Logical

| Operator | Description | Example           |
|----------|-------------|-------------------|
| `and`    | Logical AND | `x > 0 and x < 10` |
| `or`     | Logical OR  | `x == 0 or x == 1` |
| `not`    | Logical NOT | `not done`          |

`and` and `or` use short-circuit evaluation. If the left side of `and` is false, the right side is not evaluated. If the left side of `or` is true, the right side is not evaluated.

## Operator precedence

From lowest to highest:

| Priority | Operators               |
|----------|-------------------------|
| 1        | `or`                    |
| 2        | `and`                   |
| 3        | `==`, `!=`              |
| 4        | `<`, `>`, `<=`, `>=`    |
| 5        | `+`, `-`                |
| 6        | `*`, `/`, `%`           |
| 7        | `not`, unary `-`        |
| 8        | `.`, `()`, `[]`         |

Use parentheses to override precedence:

```
var int result = 4 + 3 - (2 + 2) * 4    // -9
```

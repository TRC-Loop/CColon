# sint (Safe Integer)

The `sint` type provides arbitrary precision integers with no size limit, similar to Python's `int`. Use `sint` when you need to work with numbers larger than what the regular `int` type (64-bit signed) can hold.

## Declaration

```
var sint x = 12345
var sint big = 99999999999999999999999999999999
```

Integer literals that are too large for a 64-bit int are automatically stored as `sint` values.

## Arithmetic

All standard arithmetic operators work with `sint`:

```
import console

var sint a = 99999999999999999999
var sint b = 88888888888888888888
console.println((a + b).tostring())
console.println((a * b).tostring())
console.println((a - b).tostring())
console.println((a / b).tostring())
console.println((a % b).tostring())
```

When a `sint` is involved in an operation with a regular `int`, the result is promoted to `sint`.

## Comparisons

Comparison operators (`==`, `!=`, `<`, `>`, `<=`, `>=`) work between `sint` and `int`.

```
var sint big = 99999999999999999999
var int small = 42
console.println((big > small).tostring())  // true
```

## Methods

| Method | Return | Description |
|---|---|---|
| `tostring()` | string | String representation |
| `toint()` | int | Convert to int (errors if value too large for 64-bit) |
| `tofloat()` | float | Convert to float (may lose precision) |
| `abs()` | sint | Absolute value |
| `pow(sint exp)` | sint | Raise to a power |

## Conversion

### From int to sint

```
var int x = 42
var sint big = x.tosint()
```

### From string to sint

```
var string s = "123456789012345678901234567890"
var sint big = s.tosint()
```

### From sint to int

```
var sint big = 42.tosint()
var int x = big.toint()  // works if value fits in 64-bit
```

## When to use sint vs int

Use `int` for most purposes. It is faster because it uses native 64-bit arithmetic. Use `sint` when:

- You need numbers larger than 9,223,372,036,854,775,807 (max int64)
- You are doing cryptographic or mathematical computations with large numbers
- You want to avoid integer overflow errors entirely

Note that `int` arithmetic will throw an `Error` on overflow, which can be caught with `try/catch`. With `sint`, overflow is impossible.

# math

The `math` module provides common mathematical functions and constants.

```
import math
```

## Constants

These are called as zero-argument functions:

| Function     | Value                |
|-------------|----------------------|
| `math.pi()` | 3.141592653589793   |
| `math.e()`  | 2.718281828459045   |
| `math.inf()`| Positive infinity    |

```
var float circumference = 2.0 * math.pi() * radius
```

## Functions

### `math.sqrt(x)`

Returns the square root of `x`. Accepts `int` or `float`.

```
math.sqrt(16)    // 4.0
math.sqrt(2.0)   // 1.4142135623730951
```

### `math.abs(x)`

Returns the absolute value. Accepts `int` or `float`, returns the same type.

```
math.abs(-42)    // 42
math.abs(-3.5)   // 3.5
```

### `math.floor(x)`

Rounds down to the nearest integer. Returns `int`.

```
math.floor(3.7)    // 3
math.floor(-1.2)   // -2
```

### `math.ceil(x)`

Rounds up to the nearest integer. Returns `int`.

```
math.ceil(3.2)    // 4
math.ceil(-1.7)   // -1
```

### `math.round(x)`

Rounds to the nearest integer. Returns `int`.

```
math.round(3.5)    // 4
math.round(3.4)    // 3
```

### `math.pow(base, exp)`

Returns `base` raised to the power of `exp`.

```
math.pow(2, 10)    // 1024.0
math.pow(3.0, 2)   // 9.0
```

### `math.sin(x)`, `math.cos(x)`, `math.tan(x)`

Trigonometric functions. Argument is in radians.

```
math.sin(math.pi() / 2.0)    // 1.0
math.cos(0.0)                 // 1.0
```

### `math.log(x)`, `math.log10(x)`

Natural logarithm and base-10 logarithm.

```
math.log(math.e())     // 1.0
math.log10(1000.0)     // 3.0
```

### `math.min(a, b)`, `math.max(a, b)`

Returns the smaller or larger of two values. Accepts `int` or `float`.

```
math.min(5, 3)     // 3
math.max(5, 3)     // 5
```

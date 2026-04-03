# Variables and Types

## Declaring variables

Variables are declared with the `var` keyword, followed by a type and a name:

```
var int count = 0
var string name = "CColon"
var float pi = 3.14159
var bool active = true
```

Every variable must be initialized when declared.

## Reassignment

After a variable is declared, you can reassign it without the `var` keyword:

```
var int x = 10
x = 20
x = x + 5
```

## Types

CColon has eight built-in types:

### `int`

64-bit signed integer.

| Property      | Value                                  |
|---------------|----------------------------------------|
| Size          | 64 bits                                |
| Minimum value | -9,223,372,036,854,775,808             |
| Maximum value | 9,223,372,036,854,775,807              |

```
var int a = 42
var int b = -7
var int c = 1000000
```

### `float`

64-bit floating-point number (IEEE 754 double precision).

| Property      | Value                                  |
|---------------|----------------------------------------|
| Size          | 64 bits                                |
| Precision     | ~15 to 17 significant decimal digits   |
| Minimum value | ~5.0 x 10^-324                         |
| Maximum value | ~1.8 x 10^308                          |

```
var float pi = 3.14159
var float temp = -23.5
var float ratio = 0.75
```

### `string`

A sequence of characters, enclosed in double quotes.

```
var string greeting = "Hello, World!"
var string empty = ""
```

Supported escape sequences:

| Escape | Character       |
|--------|-----------------|
| `\n`   | Newline         |
| `\t`   | Tab             |
| `\\`   | Backslash       |
| `\"`   | Double quote    |

### `bool`

A boolean value, either `true` or `false`.

```
var bool done = false
var bool ready = true
```

### `list`

A dynamic, ordered collection of values. Lists can hold mixed types and grow or shrink at runtime.

```
var list numbers = [1, 2, 3]
var list mixed = ["hello", 42, true, 3.14]
var list empty = []
```

See [Lists and Arrays](collections.md) for more.

### `array`

A fixed-size collection. Created with the `fixed()` function. Once created, the length cannot change.

```
var array coords = fixed([10, 20, 30])
```

See [Lists and Arrays](collections.md) for more.

### `sint`

Arbitrary precision integer with no size limit. Works like Python's `int`. Use this when you need numbers larger than what `int` can hold, or when you want to avoid overflow entirely.

```
var sint big = 99999999999999999999999999999999
var sint x = 42.tosint()
```

Integer literals that are too large for 64-bit are automatically stored as `sint`. See [sint](sint.md) for full details.

### `dict`

A key-value mapping. Keys are strings, values can be any type.

```
var dict person = {"name": "Alice", "age": 30}
console.println(person["name"])
```

See [Dictionaries](dicts.md) for more.

## Constants

Use the `const` keyword to declare variables that cannot be reassigned:

```
const float pi = 3.14159
const string name = "CColon"
const int max = 100
```

Attempting to reassign a constant produces an error:

```
const int x = 5
x = 10  // error: cannot reassign constant 'x'
```

Constants work in both global and local scope. Local constants are checked at compile time, global constants at runtime.

## Type conversions

Values can be converted between types using built-in methods:

```
var int n = 42
var string s = n.tostring()       // "42"
var float f = n.tofloat()         // 42.0

var string numStr = "123"
var int parsed = numStr.toint()   // 123
var float pf = numStr.tofloat()   // 123.0

var float x = 3.7
var int truncated = x.toint()     // 3
```

See [Methods on Values](methods.md) for the full list.

# Methods on Values

Every value type in CColon has built-in methods that can be called with dot notation.

## int methods

| Method       | Description                | Returns  |
|--------------|----------------------------|----------|
| `.tostring()` | Convert to string representation | `string` |
| `.tofloat()`  | Convert to float           | `float`  |
| `.tosint()`   | Convert to arbitrary precision sint | `sint` |

```
var int n = 42
console.println(n.tostring())     // "42"
var float f = n.tofloat()         // 42.0
var sint big = n.tosint()         // 42 as sint
```

## float methods

| Method       | Description                | Returns  |
|--------------|----------------------------|----------|
| `.tostring()` | Convert to string representation | `string` |
| `.toint()`    | Truncate to integer        | `int`    |

```
var float pi = 3.14
console.println(pi.tostring())    // "3.14"
var int truncated = pi.toint()    // 3
```

Note that `.toint()` truncates toward zero, it does not round.

## string methods

| Method              | Description                           | Returns  |
|---------------------|---------------------------------------|----------|
| `.length()`         | Number of characters                  | `int`    |
| `.tostring()`       | Returns the string itself             | `string` |
| `.toint()`          | Parse as integer                      | `int`    |
| `.tofloat()`        | Parse as float                        | `float`  |
| `.split(sep)`       | Split by separator (or whitespace if none) | `list` |
| `.reverse()`        | Reverse the string                    | `string` |
| `.upper()`          | Convert to uppercase                  | `string` |
| `.lower()`          | Convert to lowercase                  | `string` |
| `.trim()`           | Remove leading/trailing whitespace    | `string` |
| `.contains(sub)`    | Check if substring exists             | `bool`   |
| `.startswith(pre)`  | Check if string starts with prefix    | `bool`   |
| `.endswith(suf)`    | Check if string ends with suffix      | `bool`   |
| `.replace(old, new)`| Replace all occurrences               | `string` |
| `.index(sub)`       | Index of first occurrence (-1 if none)| `int`    |
| `.repeat(n)`        | Repeat the string n times             | `string` |
| `.join(list)`       | Join list elements with this string   | `string` |

```
var string s = "hello"
console.println(s.length().tostring())   // 5

var string num = "123"
var int parsed = num.toint()             // 123
```

| `.tosint()`     | Parse as arbitrary precision integer | `sint` |

`.toint()`, `.tofloat()`, and `.tosint()` produce a runtime error if the string is not a valid number.

```
var string s = "hello world"
console.println(s.split(" ").tostring())        // ["hello", "world"]
console.println(s.upper())                       // HELLO WORLD
console.println(s.contains("world").tostring())  // true
console.println(s.replace("world", "there"))     // hello there
console.println(s.reverse())                     // dlrow olleh
console.println(", ".join(["a", "b", "c"]))      // a, b, c
```

### String indexing

Individual characters can be accessed by index:

```
var string word = "hello"
console.println(word[0])    // "h"
console.println(word[4])    // "o"
```

## sint methods

| Method       | Description                        | Returns  |
|--------------|------------------------------------|----------|
| `.tostring()` | Convert to string representation  | `string` |
| `.toint()`    | Convert to int (errors if too large for 64-bit) | `int` |
| `.tofloat()`  | Convert to float (may lose precision) | `float` |
| `.abs()`      | Absolute value                     | `sint`   |
| `.pow(exp)`   | Raise to a power                   | `sint`   |

```
var sint big = 99999999999999999999
console.println(big.tostring())
console.println(big.abs().tostring())
console.println(big.pow(2).tostring())
```

See [sint](sint.md) for full details on the type.

## bool methods

| Method       | Description                | Returns  |
|--------------|----------------------------|----------|
| `.tostring()` | Convert to `"true"` or `"false"` | `string` |

```
var bool flag = true
console.println(flag.tostring())    // "true"
```

## list methods

| Method       | Description                        | Returns  |
|--------------|------------------------------------|----------|
| `.length()`  | Number of elements                 | `int`    |
| `.append(v)` | Add element to end                 | `nil`    |
| `.pop()`     | Remove and return last element     | value    |
| `.tostring()` | String representation             | `string` |

See [Lists and Arrays](collections.md) for details.

## array methods

| Method       | Description            | Returns  |
|--------------|------------------------|----------|
| `.length()`  | Number of elements     | `int`    |
| `.tostring()` | String representation | `string` |

See [Lists and Arrays](collections.md) for details.

## dict methods

| Method       | Description                          | Returns  |
|--------------|--------------------------------------|----------|
| `.keys()`    | List of all keys (insertion order)   | `list`   |
| `.values()`  | List of all values                   | `list`   |
| `.has(key)`  | Check if a key exists                | `bool`   |
| `.length()`  | Number of entries                    | `int`    |
| `.tostring()` | String representation               | `string` |

See [Dictionaries](dicts.md) for details.

## Chaining methods

Methods can be chained:

```
var int n = 42
console.println(n.tofloat().tostring())   // "42"

var list nums = [1, 2, 3]
console.println(nums.length().tostring())  // "3"
```

# Dictionaries

Dictionaries (dicts) are key-value collections where keys are strings.

## Creating dicts

```
var dict colors = {"red": 1, "green": 2, "blue": 3}
var dict empty = {}
```

Keys must be strings. Values can be any type.

## Accessing values

Use bracket notation with a string key:

```
var dict d = {"name": "CColon", "version": 1}
console.println(d["name"])       // "CColon"
```

Accessing a key that does not exist produces a runtime error.

## Setting values

```
var dict d = {"a": 1}
d["b"] = 2
d["a"] = 99
```

If the key exists, its value is updated. If it does not exist, a new entry is added.

## Dict methods

| Method       | Description                          | Returns  |
|--------------|--------------------------------------|----------|
| `.keys()`    | List of all keys (in insertion order)| `list`   |
| `.values()`  | List of all values                   | `list`   |
| `.has(key)`  | Check if a key exists                | `bool`   |
| `.length()`  | Number of entries                    | `int`    |
| `.tostring()` | String representation               | `string` |

```
var dict d = {"x": 10, "y": 20}
console.println(d.keys().tostring())      // ["x", "y"]
console.println(d.has("x").tostring())    // true
console.println(d.length().tostring())    // 2
```

## Iterating over dicts

A `for-in` loop over a dict iterates over the keys:

```
var dict scores = {"alice": 90, "bob": 85}
for key in scores {
    console.println(key + " = " + scores[key].tostring())
}
```

Keys are returned in insertion order.

## Full example

```
import console
import json

function main() {
    var dict config = {"host": "localhost", "port": 8080}
    config["debug"] = true

    for key in config {
        console.println(key + ": " + config[key].tostring())
    }

    console.println("JSON: " + json.stringify(config))
}
```

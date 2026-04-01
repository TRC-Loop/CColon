# json

The `json` module provides JSON parsing and serialization.

```
import json
```

## Functions

### `json.parse(string)`

Parses a JSON string and returns the corresponding CColon value.

```
var dict data = json.parse("{\"name\": \"CColon\", \"version\": 1}")
console.println(data["name"])    // "CColon"
```

Type mapping from JSON to CColon:

| JSON type | CColon type |
|-----------|-------------|
| object    | `dict`      |
| array     | `list`      |
| string    | `string`    |
| number (integer) | `int` |
| number (decimal) | `float` |
| boolean   | `bool`      |
| null      | `nil`       |

Throws an error if the input is not valid JSON.

### `json.stringify(value)`

Converts a CColon value to a JSON string.

```
var dict data = {"lang": "CColon", "year": 2026}
var string s = json.stringify(data)
console.println(s)    // {"lang":"CColon","year":2026}
```

Supported input types: `dict`, `list`, `string`, `int`, `float`, `bool`, and `nil`.

## Full example

```
import console
import json

function main() {
    var string raw = "{\"users\": [{\"name\": \"Alice\"}, {\"name\": \"Bob\"}]}"
    var dict data = json.parse(raw)

    var list users = data["users"]
    for i in range(users.length()) {
        var dict user = users[i]
        console.println(user["name"])
    }

    var dict response = {"status": "ok", "count": users.length()}
    console.println(json.stringify(response))
}
```

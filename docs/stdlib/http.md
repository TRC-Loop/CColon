# http

The `http` module provides functions for making HTTP requests.

```
import http
```

## Functions

| Function | Description |
|---|---|
| `http.get(string url)` | Send a GET request |
| `http.post(string url, string body, dict headers)` | Send a POST request |
| `http.put(string url, string body, dict headers)` | Send a PUT request |
| `http.delete(string url)` | Send a DELETE request |

All request functions return a response dict:

```
{
    "status": 200,
    "body": "...",
    "headers": {"content-type": "application/json", ...}
}
```

## Arguments

- **url** (required): The URL to request
- **body** (optional, POST/PUT): Request body as a string
- **headers** (optional): Dict of HTTP headers

## Examples

### Simple GET request

```
import console
import http

var dict resp = http.get("https://httpbin.org/get")
console.println("Status: " + resp["status"].tostring())
console.println(resp["body"])
```

### POST with JSON

```
import console
import http
import json

var dict data = {"name": "CColon", "version": "0.3.0"}
var string body = json.encode(data)
var dict headers = {"Content-Type": "application/json"}

var dict resp = http.post("https://httpbin.org/post", body, headers)
console.println("Status: " + resp["status"].tostring())
```

### Reading response headers

```
import console
import http

var dict resp = http.get("https://httpbin.org/get")
var dict headers = resp["headers"]
console.println("Content-Type: " + headers["content-type"])
```

### Error handling

```
import console
import http

try {
    var dict resp = http.get("https://nonexistent.example.com")
} catch (Error e) {
    console.println("Request failed: " + e.message)
}
```

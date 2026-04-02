# http

The `http` module provides functions for making HTTP requests and running HTTP servers.

```
import http
```

## Client Functions

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

### Arguments

- **url** (required): The URL to request
- **body** (optional, POST/PUT): Request body as a string
- **headers** (optional): Dict of HTTP headers

## Server Function

| Function | Description |
|---|---|
| `http.listen(int port, function handler)` | Start an HTTP server |

The handler function receives a request dict and should return a response.

### Request dict

```
{
    "method": "GET",
    "path": "/hello",
    "body": "",
    "headers": {"content-type": "text/plain", ...},
    "query": "name=world"
}
```

### Response formats

The handler can return:

- A **string**: sent as `200 OK` with `text/plain` content type
- A **dict** with `status`, `body`, and optionally `headers`
- **nil**: sent as `204 No Content`

## Examples

### Simple GET request

```
import console
import http

function main() {
    var dict resp = http.get("https://httpbin.org/get")
    console.println("Status: " + resp["status"].tostring())
    console.println(resp["body"])
}
```

### POST with JSON

```
import console
import http
import json

function main() {
    var dict data = {"name": "CColon", "version": "1.0.0"}
    var string body = json.encode(data)
    var dict headers = {"Content-Type": "application/json"}

    var dict resp = http.post("https://httpbin.org/post", body, headers)
    console.println("Status: " + resp["status"].tostring())
}
```

### Simple HTTP server

```
import console
import http

function handler(dict req) string {
    console.println(req["method"] + " " + req["path"])
    return "Hello from CColon!"
}

function main() {
    http.listen(8080, handler)
}
```

### Server with JSON responses

```
import http
import json

function handler(dict req) dict {
    if (req["path"] == "/api/hello") {
        var dict body = {"message": "Hello, World!"}
        return {
            "status": 200,
            "body": json.encode(body),
            "headers": {"Content-Type": "application/json"}
        }
    }
    return {"status": 404, "body": "Not Found"}
}

function main() {
    http.listen(3000, handler)
}
```

### Error handling

```
import console
import http

function main() {
    try {
        var dict resp = http.get("https://nonexistent.example.com")
    } catch (Error e) {
        console.println("Request failed: " + e.message)
    }
}
```

# os

The `os` module provides functions for interacting with the operating system.

```
import os
```

## Functions

| Function | Description |
|---|---|
| `os.exec(string cmd)` | Run a shell command, returns `{output, exitcode}` |
| `os.env(string key)` | Get environment variable (returns nil if not set) |
| `os.env(string key, default)` | Get environment variable with default |
| `os.setenv(string key, string val)` | Set environment variable |
| `os.envlist()` | Get all environment variables as a dict |
| `os.cwd()` | Get current working directory |
| `os.chdir(string dir)` | Change working directory |
| `os.args()` | Get command-line arguments as a list |
| `os.platform()` | Get OS name (`"darwin"`, `"linux"`, `"windows"`) |
| `os.arch()` | Get architecture (`"amd64"`, `"arm64"`) |
| `os.hostname()` | Get system hostname |
| `os.exit(int code)` | Exit the program |
| `os.mkdir(string path)` | Create directory (and parents) |
| `os.remove(string path)` | Remove file or directory |
| `os.exists(string path)` | Check if path exists |
| `os.listdir(string path)` | List directory contents |

## Examples

### Run a command

```
import console
import os

var dict result = os.exec("echo hello")
console.println(result["output"])
console.println("Exit code: " + result["exitcode"].tostring())
```

### Environment variables

```
import console
import os

var string home = os.env("HOME", "/tmp")
console.println("Home: " + home)

os.setenv("MY_VAR", "hello")
console.println(os.env("MY_VAR"))
```

### File system operations

```
import console
import os

os.mkdir("testdir")
console.println(os.exists("testdir").tostring())

var list files = os.listdir(".")
for f in files {
    console.println(f)
}

os.remove("testdir")
```

### System information

```
import console
import os

console.println("Platform: " + os.platform())
console.println("Architecture: " + os.arch())
console.println("Hostname: " + os.hostname())
console.println("CWD: " + os.cwd())
```

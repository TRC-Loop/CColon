# fs

The `fs` module provides file system operations: reading, writing, and managing files and directories.

```
import fs
```

## Module functions

### `fs.read(path)`

Reads the entire contents of a file and returns it as a string.

```
var string content = fs.read("data.txt")
```

### `fs.write(path, content)`

Writes a string to a file, creating it if it does not exist and overwriting it if it does.

```
fs.write("output.txt", "Hello, World!")
```

### `fs.open(path, mode)`

Opens a file and returns a file object. The mode must be one of:

| Mode | Description     |
|------|-----------------|
| `"r"` | Read           |
| `"w"` | Write (create/overwrite) |
| `"a"` | Append         |

```
var File f = fs.open("log.txt", "a")
f.write("new entry\n")
f.close()
```

See [File objects](#file-objects) below for the methods available on file objects.

### `fs.exists(path)`

Returns `true` if the file or directory exists, `false` otherwise.

```
if (fs.exists("config.txt")) {
    var string cfg = fs.read("config.txt")
}
```

### `fs.remove(path)`

Deletes a file. Throws an error if the file does not exist.

```
fs.remove("temp.txt")
```

### `fs.mkdir(path)`

Creates a directory. Pass `true` as the second argument to create parent directories and ignore existing ones:

```
fs.mkdir("output")
fs.mkdir("a/b/c", true)    // creates a, a/b, and a/b/c
```

## File objects

File objects are returned by `fs.open()`. They have these methods:

| Method         | Description                              |
|----------------|------------------------------------------|
| `.read()`      | Read all remaining content as a string   |
| `.readline()`  | Read the next line (without newline)     |
| `.readlines()` | Read all lines as a list of strings      |
| `.write(str)`  | Write a string to the file               |
| `.close()`     | Close the file                           |

### Reading line by line

```
import fs
import console

function main() {
    var File f = fs.open("data.txt", "r")
    var list lines = f.readlines()
    f.close()

    for line in lines {
        console.println(line)
    }
}
```

### Using with for automatic cleanup

The `with` statement calls `.close()` automatically when the block exits:

```
import fs
import console

function main() {
    with fs.open("data.txt", "r") as f {
        var string content = f.read()
        console.println(content)
    }
}
```

## Full example

```
import console
import fs

function main() {
    // write a file
    fs.write("greeting.txt", "Hello from CColon!")

    // read it back
    var string content = fs.read("greeting.txt")
    console.println(content)

    // append to it
    with fs.open("greeting.txt", "a") as f {
        f.write("\nGoodbye!")
    }

    // read updated content
    console.println(fs.read("greeting.txt"))

    // clean up
    fs.remove("greeting.txt")
}
```

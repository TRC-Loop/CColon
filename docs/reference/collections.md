# Lists and Arrays

CColon has two collection types: **lists** (dynamic size) and **arrays** (fixed size).

## Lists

Lists are dynamic, ordered collections. They can grow and shrink at runtime, and they can hold values of mixed types.

### Creating lists

```
var list numbers = [1, 2, 3, 4, 5]
var list empty = []
var list mixed = ["hello", 42, true, 3.14]
```

### Accessing elements

Elements are accessed by index, starting at 0:

```
var list colors = ["red", "green", "blue"]
console.println(colors[0])    // "red"
console.println(colors[2])    // "blue"
```

Accessing an index out of bounds produces a runtime error.

### Modifying elements

```
var list items = [10, 20, 30]
items[1] = 99
// items is now [10, 99, 30]
```

### List methods

| Method       | Description                      | Returns |
|--------------|----------------------------------|---------|
| `.length()`  | Number of elements               | `int`   |
| `.append(v)` | Add an element to the end        | `nil`   |
| `.pop()`     | Remove and return the last element | value |
| `.tostring()` | String representation           | `string`|

```
var list nums = [1, 2, 3]
nums.append(4)                         // [1, 2, 3, 4]
console.println(nums.length().tostring())  // 4
var int last = nums.pop()              // last = 4, nums = [1, 2, 3]
```

### Iterating over lists

```
var list fruits = ["apple", "banana", "cherry"]
for fruit in fruits {
    console.println(fruit)
}
```

## Arrays (fixed size)

Arrays have a fixed length that is set at creation. You cannot append to or pop from an array.

### Creating arrays

Use the `fixed()` function:

```
var array coords = fixed([10, 20, 30])
var array flags = fixed([true, false, true])
```

### Accessing and modifying elements

Works the same as lists:

```
console.println(coords[0].tostring())    // 10
coords[1] = 99                           // [10, 99, 30]
```

### Array methods

| Method       | Description            | Returns  |
|--------------|------------------------|----------|
| `.length()`  | Number of elements     | `int`    |
| `.tostring()` | String representation | `string` |

### When to use arrays vs lists

Use **lists** when you need a collection that grows or shrinks during execution. Use **arrays** when the size is known and should not change. Arrays communicate intent: the reader knows the size is fixed.

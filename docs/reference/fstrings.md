# F-Strings

F-strings let you embed expressions directly inside string literals. Prefix a string with `f` and use `{expression}` to interpolate values.

## Basic usage

```
var string name = "World"
console.println(f"Hello, {name}!")  // Hello, World!
```

## Expressions

Any valid expression can go inside the braces:

```
var int x = 10
var int y = 20
console.println(f"{x} + {y} = {x + y}")  // 10 + 20 = 30
```

## Automatic string conversion

Non-string values are automatically converted to their string representation:

```
var int count = 42
console.println(f"count is {count}")    // count is 42

var float pi = 3.14
console.println(f"pi is {pi}")          // pi is 3.14

var bool done = true
console.println(f"done: {done}")        // done: true
```

## Escaping braces

Use `\{` and `\}` to include literal brace characters:

```
console.println(f"use \{braces\} like this")  // use {braces} like this
```

## String concatenation with +

You can also build strings with the `+` operator. When one side is a string, the other side is automatically converted:

```
var string msg = "count: " + 42       // "count: 42"
var string msg2 = 3.14 + " is pi"    // "3.14 is pi"
```

Both f-strings and `+` concatenation work well for building dynamic strings. F-strings are usually more readable for complex expressions.

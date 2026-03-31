# Hello World

## Your first program

Create a file called `hello.ccl`:

```
import console

function main() {
    console.println("Hello, World!")
}
```

Run it:

```sh
ccolon hello.ccl
```

Output:

```
Hello, World!
```

## Breaking it down

### `import console`

This brings in the `console` module, which provides functions for printing to the terminal and reading user input.

### `function main()`

Every CColon program needs a `main` function. It is the entry point and runs automatically when the file is executed.

### `console.println(...)`

Prints a value followed by a newline. You can pass strings, numbers, or any value.

## Something more interesting

```
import console

function greet(string name, int age) {
    console.println("Hello, " + name + "!")
    console.println("You are " + age.tostring() + " years old.")
}

function main() {
    var string name = console.scanp("What is your name? ")
    var string ageStr = console.scanp("How old are you? ")
    var int age = ageStr.toint()
    greet(name, age)
}
```

This program asks for your name and age, then greets you. It demonstrates variables, functions with parameters, user input, and type conversion.

## Next steps

- Try the [Interactive Shell](repl.md) for quick experiments
- Read the [Language Reference](../reference/variables.md) for the full syntax

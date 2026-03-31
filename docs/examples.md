# Examples

## FizzBuzz

```
import console

function fizzbuzz(int max) {
    for i in range(1, max + 1) {
        if (i % 15 == 0) {
            console.println("FizzBuzz")
        } else if (i % 3 == 0) {
            console.println("Fizz")
        } else if (i % 5 == 0) {
            console.println("Buzz")
        } else {
            console.println(i.tostring())
        }
    }
}

function main() {
    fizzbuzz(30)
}
```

## Fibonacci sequence

```
import console

function fib(int n) int {
    if (n <= 1) {
        return n
    }
    return fib(n - 1) + fib(n - 2)
}

function main() {
    for i in range(15) {
        console.print(fib(i).tostring() + " ")
    }
    console.println("")
}
```

Output: `0 1 1 2 3 5 8 13 21 34 55 89 144 233 377`

## Factorial

```
import console

function factorial(int n) int {
    if (n <= 1) {
        return 1
    }
    return n * factorial(n - 1)
}

function main() {
    for i in range(1, 11) {
        console.println(i.tostring() + "! = " + factorial(i).tostring())
    }
}
```

## Temperature converter

```
import console

function celsius_to_fahrenheit(float c) float {
    return c * 1.8 + 32.0
}

function fahrenheit_to_celsius(float f) float {
    return (f - 32.0) / 1.8
}

function main() {
    var string input = console.scanp("Enter temperature in Celsius: ")
    var float c = input.tofloat()
    var float f = celsius_to_fahrenheit(c)
    console.println(c.tostring() + " C = " + f.tostring() + " F")
}
```

## List operations

```
import console

function sum(list numbers) int {
    var int total = 0
    for n in numbers {
        total = total + n
    }
    return total
}

function main() {
    var list scores = [85, 92, 78, 95, 88]
    console.println("Scores: " + scores.tostring())
    console.println("Total: " + sum(scores).tostring())
    console.println("Count: " + scores.length().tostring())

    scores.append(100)
    console.println("After adding 100: " + scores.tostring())

    var int removed = scores.pop()
    console.println("Popped: " + removed.tostring())
}
```

## Number guessing game

```
import console

function main() {
    var int secret = 42
    var int attempts = 0

    console.println("I am thinking of a number between 1 and 100.")

    while (true) {
        var string input = console.scanp("Your guess: ")
        var int guess = input.toint()
        attempts = attempts + 1

        if (guess < secret) {
            console.println("Too low!")
        } else if (guess > secret) {
            console.println("Too high!")
        } else {
            console.println("Correct! You got it in " + attempts.tostring() + " attempts.")
            break
        }
    }
}
```

## Working with arrays

```
import console

function main() {
    var array rgb = fixed([255, 128, 0])
    console.println("R: " + rgb[0].tostring())
    console.println("G: " + rgb[1].tostring())
    console.println("B: " + rgb[2].tostring())

    rgb[1] = 200
    console.println("Updated color: " + rgb.tostring())
    console.println("Channels: " + rgb.length().tostring())
}
```

For a comprehensive example covering all language features, see [`demo.ccl`](https://github.com/TRC-Loop/CColon/blob/main/ccolon/examples/demo.ccl).

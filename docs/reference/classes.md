# Classes

Classes let you define custom types with fields and methods. CColon classes support public/private visibility, inheritance, and constructors.

## Defining a class

```
class Animal {
    var public string name = "unknown"
    var public int age = 0
    var private string id = "internal"

    public function init(string name, int age) {
        self.name = name
        self.age = age
    }

    public function speak() string {
        return self.name + " makes a sound"
    }
}
```

A class body contains **field declarations** and **method declarations**.

## Fields

Fields are declared with `var`, a visibility modifier, a type, a name, and a default value:

```
var public string name = "unknown"
var private int secret = 0
```

- `public` fields can be read and written from outside the class.
- `private` fields can only be accessed inside the class (through `self`).

Every field must have a default value.

## Methods

Methods are functions defined inside a class. Each method has a visibility modifier:

```
public function greet() string {
    return "Hello, " + self.name
}

private function internal() {
    // only callable from other methods in this class
}
```

Inside a method, use `self` to access the current instance's fields and methods:

```
public function info() string {
    return self.name + " (age " + self.age.tostring() + ")"
}
```

## Constructors

The `init` method is the constructor. It runs when you create a new instance:

```
class Dog {
    var public string name = ""
    var public string breed = ""

    public function init(string name, string breed) {
        self.name = name
        self.breed = breed
    }
}

var Dog d = Dog("Rex", "Labrador")
```

`init` does not return a value. The newly created instance is returned automatically.

## Creating instances

Call the class name like a function, passing arguments that match the `init` parameters:

```
var Animal cat = Animal("Whiskers", 3)
console.println(cat.name)        // "Whiskers"
console.println(cat.speak())     // "Whiskers makes a sound"
```

## Inheritance

Use `extends` to create a subclass:

```
class Dog extends Animal {
    var public string breed = "mixed"

    public function init(string name, int age, string breed) {
        super.init(name, age)
        self.breed = breed
    }

    public function speak() string {
        return self.name + " barks!"
    }
}
```

A subclass inherits all fields and methods from its parent. It can override methods by defining them with the same name.

### Calling parent methods

Use `super.methodName(args)` to call a method from the parent class:

```
public function init(string name, int age, string breed) {
    super.init(name, age)
    self.breed = breed
}
```

## String representation

When you call `.tostring()` on an instance or print it, CColon generates a string showing the class name, public fields, and public method names:

```
var Dog d = Dog("Rex", 5, "Labrador")
console.println(d.tostring())
// <Dog name="Rex", age=5, breed="Labrador" | speak(), info()>
```

## Full example

```
import console

class Shape {
    var public string name = "shape"

    public function init(string name) {
        self.name = name
    }

    public function area() float {
        return 0.0
    }
}

class Circle extends Shape {
    var public float radius = 0.0

    public function init(float radius) {
        super.init("circle")
        self.radius = radius
    }

    public function area() float {
        return 3.14159 * self.radius * self.radius
    }
}

function main() {
    var Circle c = Circle(5.0)
    console.println(c.name)                    // "circle"
    console.println(c.area().tostring())       // "78.53975"
}
```

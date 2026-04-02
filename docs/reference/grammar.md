# CColon EBNF Grammar

This document defines the complete grammar of the CColon programming language in Extended Backus-Naur Form (EBNF). It is derived directly from the lexer and parser source code and covers every construct the language accepts.

## Notation

| Symbol    | Meaning                          |
|-----------|----------------------------------|
| `=`       | Definition                       |
| `,`       | Concatenation (sequencing)       |
| `\|`      | Alternation                      |
| `[ ... ]` | Optional (zero or one)           |
| `{ ... }` | Repetition (zero or more)        |
| `( ... )` | Grouping                         |
| `" ... "` | Terminal string (literal token)   |
| `(* ... *)` | Comment                        |

---

## 1. Lexical Rules

### 1.1 Whitespace and Comments

Whitespace (spaces, tabs, carriage returns, newlines) is ignored between tokens. Two comment forms are supported:

```
line_comment  = "//" , { any_char - newline } ;
block_comment = "/*" , { any_char } , "*/" ;
```

### 1.2 Identifiers

```
letter     = "a" | "b" | ... | "z" | "A" | "B" | ... | "Z" | "_" ;
digit      = "0" | "1" | ... | "9" ;
identifier = ( letter ) , { letter | digit } ;
```

Identifiers that match a keyword are lexed as that keyword instead.

### 1.3 Keywords

```
keyword = "var" | "function" | "return" | "if" | "else"
        | "for" | "in" | "while" | "import"
        | "and" | "or" | "not"
        | "true" | "false"
        | "break" | "continue"
        | "class" | "extends" | "self" | "super"
        | "public" | "private"
        | "try" | "catch" | "throw"
        | "with" | "as"
        | "range" | "fixed" ;
```

### 1.4 Type Keywords

Type keywords are a subset of keywords that name built-in types. An identifier (class name) is also accepted wherever a type is expected.

```
type_keyword = "string" | "int" | "float" | "bool" | "list" | "array" | "dict" ;
type_name    = type_keyword | identifier ;
```

### 1.5 Literals

```
int_literal    = digit , { digit } ;
float_literal  = digit , { digit } , "." , digit , { digit } ;
string_literal = '"' , { string_char } , '"' ;
string_char    = any_char - ( '"' | "\\" )
               | "\\" , escape_char ;
escape_char    = "n" | "t" | '"' | "\\" ;
bool_literal   = "true" | "false" ;
```

### 1.6 Operators

```
operator = "+" | "-" | "*" | "/" | "%"
         | "=" | "==" | "!="
         | "<" | ">" | "<=" | ">="
         | "and" | "or" | "not" ;
```

### 1.7 Delimiters

```
delimiter = "(" | ")" | "{" | "}" | "[" | "]"
          | "," | "." | ":" | ";" ;
```

---

## 2. Grammar Productions

### 2.1 Program

A program is a sequence of statements. A `main()` function is the entry point.

```
program = { statement } ;
```

### 2.2 Statements

```
statement = import_stmt
          | var_decl
          | func_decl
          | if_stmt
          | while_stmt
          | for_in_stmt
          | return_stmt
          | break_stmt
          | continue_stmt
          | class_decl
          | try_catch_stmt
          | throw_stmt
          | with_stmt
          | expression_stmt
          | ";" ;
```

Stray semicolons between statements are silently consumed.

#### 2.2.1 Import

```
import_stmt = "import" , ( string_literal | identifier ) ;
```

When the argument is a string literal, it is treated as a file path (e.g. `import "utils.ccl"`). When it is an identifier, it refers to a standard library module (e.g. `import math`).

#### 2.2.2 Variable Declaration

```
var_decl = "var" , type_name , identifier , "=" , expression ;
```

#### 2.2.3 Function Declaration

```
func_decl = "function" , identifier , "(" , [ param_list ] , ")" , [ type_name ] , block ;
```

The optional `type_name` after the parameter list is the return type.

#### 2.2.4 Parameter List

```
param_list = param , { "," , param } ;
param      = type_name , identifier , [ "=" , expression ] ;
```

Parameters with default values (optional parameters) must come after all required parameters.

#### 2.2.5 Block

```
block = "{" , { statement } , "}" ;
```

#### 2.2.6 If / Else

```
if_stmt = "if" , "(" , expression , ")" , block , [ else_clause ] ;
else_clause = "else" , ( if_stmt | block ) ;
```

#### 2.2.7 While Loop

```
while_stmt = "while" , "(" , expression , ")" , block ;
```

#### 2.2.8 For-In Loop

```
for_in_stmt = "for" , identifier , "in" , expression , block ;
```

#### 2.2.9 Return

```
return_stmt = "return" , [ expression ] ;
```

A bare `return` (no value) is valid when followed by `}`, `;`, or end of file.

#### 2.2.10 Break and Continue

```
break_stmt    = "break" ;
continue_stmt = "continue" ;
```

#### 2.2.11 Throw

```
throw_stmt = "throw" , expression ;
```

#### 2.2.12 Try / Catch

```
try_catch_stmt = "try" , block , "catch" , "(" , identifier , identifier , ")" , block ;
```

The two identifiers inside `catch(...)` are the exception type and the binding name, in that order. For example: `catch(Error e)`.

#### 2.2.13 With / As

```
with_stmt = "with" , expression , "as" , identifier , block ;
```

#### 2.2.14 Expression Statement and Assignment

```
expression_stmt = expression , [ "=" , expression ] ;
```

If an `=` follows the left-hand expression, the statement becomes an assignment. The left-hand side can be an identifier, an index expression, or a field access.

#### 2.2.15 Class Declaration

```
class_decl = "class" , identifier , [ "extends" , identifier ] , "{" , { class_member } , "}" ;
class_member = field_decl | method_decl ;
```

#### 2.2.16 Field Declaration

```
field_decl = "var" , visibility , type_name , identifier , [ "=" , expression ] ;
visibility = "public" | "private" ;
```

#### 2.2.17 Method Declaration

```
method_decl = visibility , "function" , identifier , "(" , [ param_list ] , ")" , [ type_name ] , block ;
```

A method named `init` is always treated as private regardless of the declared visibility.

---

## 3. Expressions and Operator Precedence

The parser uses a Pratt (top-down operator precedence) approach. The following table lists precedence levels from lowest to highest:

| Precedence | Operators / Forms              | Associativity |
|------------|--------------------------------|---------------|
| 1          | `or`                           | Left          |
| 2          | `and`                          | Left          |
| 3          | `==` `!=`                      | Left          |
| 4          | `<` `>` `<=` `>=`             | Left          |
| 5          | `+` `-`                        | Left          |
| 6          | `*` `/` `%`                    | Left          |
| 7          | Unary: `-` `not`               | Prefix        |
| 8          | Call, index, dot access        | Left (postfix)|

### 3.1 Expression

```
expression = prefix , { infix_op , prefix } ;
```

The exact binding is governed by the precedence table above. In EBNF terms, each precedence level wraps the one above it, but the Pratt parser handles this with a numeric minimum-precedence parameter.

### 3.2 Prefix Expressions

```
prefix = int_literal
       | float_literal
       | string_literal
       | bool_literal
       | identifier
       | unary_expr
       | grouped_expr
       | self_expr
       | super_call
       | dict_literal
       | list_literal
       | range_expr
       | fixed_expr ;
```

#### Unary Expression

```
unary_expr = ( "-" | "not" ) , expression ;
```

#### Grouped Expression

```
grouped_expr = "(" , expression , ")" ;
```

#### Self

```
self_expr = "self" ;
```

#### Super Call

```
super_call = "super" , "." , identifier , "(" , [ arg_list ] , ")" ;
```

### 3.3 Infix / Postfix Expressions

#### Binary Expression

```
binary_expr = expression , binary_op , expression ;
binary_op   = "+" | "-" | "*" | "/" | "%"
            | "==" | "!=" | "<" | ">" | "<=" | ">="
            | "and" | "or" ;
```

#### Function Call

```
call_expr = expression , "(" , [ arg_list ] , ")" ;
arg_list  = expression , { "," , expression } ;
```

#### Method Call

```
method_call = expression , "." , identifier , "(" , [ arg_list ] , ")" ;
```

#### Field Access

```
field_access = expression , "." , identifier ;
```

When a dot-identifier is not followed by `(`, it is parsed as a field access rather than a method call.

#### Index Expression

```
index_expr = expression , "[" , expression , "]" ;
```

### 3.4 Collection Literals

#### List Literal

```
list_literal = "[" , [ expression , { "," , expression } ] , "]" ;
```

#### Dict Literal

```
dict_literal = "{" , [ dict_entry , { "," , dict_entry } , [ "," ] ] , "}" ;
dict_entry   = expression , ":" , expression ;
```

A trailing comma is permitted before the closing `}`.

#### Fixed Array

```
fixed_expr = "fixed" , "(" , "[" , [ expression , { "," , expression } ] , "]" , ")" ;
```

#### Range Expression

```
range_expr = "range" , "(" , expression , [ "," , expression ] , ")" ;
```

With one argument, `range(n)` produces the range `0..n`. With two arguments, `range(start, end)` produces `start..end`.

---

## 4. Notes

### Entry Point

Every CColon program must define a `function main()` as its entry point. The runtime calls `main()` after all top-level declarations are processed.

### Type System

Variable declarations, parameters, and return types use `type_name`, which accepts both built-in type keywords (`int`, `float`, `string`, `bool`, `list`, `array`, `dict`) and identifiers (for class types). Type annotations are present in the syntax but the language uses runtime type checking.

### Classes and Inheritance

Classes support single inheritance via `extends`. Fields use `var` with a visibility modifier. Methods use a visibility modifier before `function`. The `self` keyword refers to the current instance inside method bodies. The `super` keyword allows calling a parent class method via `super.method(args)`.

### Error Handling

`try`/`catch` blocks catch exceptions by type. The `throw` statement raises a value as an exception.

### Resource Management

The `with`/`as` statement binds the result of an expression to a name for the duration of a block, supporting resource-management patterns.

### Imports

File imports use a string literal path (`import "file.ccl"`). Standard library module imports use a bare identifier (`import math`).

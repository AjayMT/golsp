
# Syntax
## Syntax nodes
Golsp has six basic types of syntax nodes:
- comments
- identifiers
- literals
- expressions
- lists
- maps

### Comments
Comments in Golsp begin with `#` and end with a newline. Comments have no semantic meaning.
```python
# this is a comment
foo bar baz # this is a comment that begins in the middle of a line

```

### Identifiers
Identifiers are space-separated tokens that evaluate to values (literals, functions, lists or maps). Identifiers can contain all characters except syntactic delimiters and operators.
```python
a b c quux z00t $ _ # these are identifiers
```

### Literals
Literals can be of two types: **strings** and **numbers**.

String literals are delimited by `"`.
```python
"hello" "world" # these are strings
```

Numbers are contiguous groups of numeric characters surrounded by spaces or syntactic delimiters. Numbers can begin with a minus sign and can contain a single decimal point.
```python
1 2 3 4.5 -6 -7.8 # these are numbers
```

### Expressions
Expressions are delimited by square brackets and can contain other syntax nodes.
```python
[a b c 12 "hello"] # this is an expression
["doge" wow [much food]] # this is an expression that contains an expression
```

Expressions are also delimited by pairs of newlines (see README).
```python
# this:
printf "%v\n" [function
                argument quux
                fuzz cat
                foo
                bar baz]

# is automatically translated into this:
[printf "%v\n" [function
                 [argument quux]
                 [fuzz cat]
                 foo
                 bar baz]]

# note that the 'foo' was not wrapped in an expression -- Golsp
# does not automatically expression-ify single-token lines because you may
# not intend to call them as functions
# 'bar baz' was also not wrapped in an expression, because the surrounding
# expression ends on the same line
# Golsp will not wrap lines that are inside lists or maps in expression delimiters
```

Expressions consist of two parts: an **expression head** and a list of **arguments**. The expression head is the first node in the expression, and the arguments are all subsequent nodes.
```python
[a b c] # 'a' is the expression head, 'b' and 'c' are arguments
```

Expressions evaluate to different things depending on the type of the expression head.

If the expression head is a number, the expression simply evaluates to the number.
```python
[17 21 93 z b t] # => 17
```

If the expression head is a function, the function is applied to the arguments. If the function is defined for those arguments, it returns a result. Otherwise, the expression evaluates to [undefined](docs?id=undefined).
```python
[+ 1 2] # => 3
[+ "a" "b"] # => undefined
```

If the expression head is a string, list or map, the expression evaluates to different things depending on the arguments (see [Lists](docs?id=lists) and [Maps](docs?id=maps).)

### Lists
Lists are delimited by curly braces and can contain other syntax nodes.
```python
# this is a list containing numbers, a list of strings, and the
# result of the expression `[foo baz]`:
{ 1 2 3 { "a" "b" "c" } [foo baz] }
```

If a list is an [expression head](docs?id=expressions), it can be indexed and sliced as follows:
```python
# if no arguments are provided, the expression evaluates to the list
[{ 1 2 3 }] # => { 1 2 3 }

# list indices begin at 0
[{ 1 2 3 } 0] # => 1

# negative indices count backwards from the end of the list
[{ 1 2 3 } -1] # => 3

# two arguments (begin, inclusive and end, exclusive) slice the list
[{ 1 2 3 4 } 0 2] # => { 1 2 }

# `undefined` will slice until the end of the list
[{ 1 2 3 4 5 6 } 0 undefined] # => { 1 2 3 4 5 6 }

# a third argument produces a slice with a step
[{ 1 2 3 4 5 6 } 0 undefined 2] # => { 1 3 5 }

# steps can be negative, which will reverse the list
[{ 1 2 3 4 5 6 } -1 undefined -2] # => { 6 4 2 }
```

**Strings** can also be indexed and sliced like lists.

Golsp does not automatically wrap lines in expressions inside lists. This means that the following code:
```python
{
  a b
  c
}
```
is **not** converted into this:
```python
{
  [a b]
  [c]
}
```
This is because a line of space-separated syntax nodes (like `a b` in the example above) inside a list is usually meant to be multiple separate elements rather than being a single expression.

### Maps
Maps are delimited by parentheses and can contain pairs of syntax nodes joined by [zip operators](docs?id=zip).
```python
(
  "a": 1
  "b": 2
  "c": 3
  4: "d"
)
```

The syntax nodes to the left of the zip operators are the keys, and those to the right of the zip operators are the values. Keys can only be strings or numbers, but values can be any object.

```python
# this is a valid map
(
  "A": { 1 2 3 }
  "b": "c"
)

# this is not a valid map, since lists cannot be keys
(
  { 1 2 3 }: "A"
  "b": 2
)
```

If a map is an expression head, the provided arguments will be looked up in the map.
```python
# if no arguments are provided, the expression evaluates to the map itself
[( "A":1 "b":"z" )] # => ( "A":1 "b":"z" )

# providing a single argument produces a single value
[( "a":1 "b":2 ) "b"] # => 2

# providing multiple arguments produces a list of values
[( "Cat":"z00t" "beep":"boop" 4:"quux" ) "beep" 4 "foo"] # => { "boop" "quux" undefined }
```

Like with [lists](docs?id=lists), Golsp does not automatically wrap lines in expressions inside maps. So this:
```python
(
  "a": "b"
  "c": "d"
)
```
is **not** the same as this:
```python
(
  ["a": "b"]
  ["c": "d"]
)
```

## Operators
Golsp has three operators:
- spread: `...`
- zip: `:`
- dot: `.`

### Spread
`...`

The spread operator is a **postfix** operator that takes a list or string and distributes its contents into the surrounding expression. The operator has no effect on other types of values.
```python
# lists and strings are 'spread' into the surrounding expression
# strings of multiple characters spread into multiple strings of single characters
[{ 1 2 3 }...] # => [1 2 3]
{ "abc"... } # => { "a" "b" "c" }

# other types of values remain unchanged
{ 1... } # => { 1 }
```

The spread operator can be used in conjunction with the [zip](docs?id=zip) operator to 'zip' multiple keys and values together inside maps:
```python
(
  {"a" "b" "c" "d"}... : { 1 2 3 }...
) # => ( "a":1 "b":2 "c":3 )
```

### Zip
`:`

The zip operator is an **infix** operator that pairs keys and values inside maps.
```python
(
  "doge": "wow" # "doge" is paired with (i.e mapped to) "wow"
  "quux": [z00t] # "quux" is paired with the result of the expression [z00t]
  "asdf": { "z" "b" 2 } # "asdf" is paired with a list containing "z", "b" and 2
)
```

By default, the zip operator has no effect outside a map (except for the builtin [when](docs?id=when) function). A pair of syntax nodes zipped together will evaluate to the value of the first node.
```python
"A": 2 # => this simply evaluates to "A"
[foo]:bar # => this evaluates to the result of [foo]
```

The zip operator can be used in conjunction with the [spread](docs?id=spread) operator to 'zip' together multiple keys and values inside map:
```python
(
  {"a" "b" "c" "d"}... : { 1 2 3 }...
) # => ( "a":1 "b":2 "c":3 )
```

### Dot
`.`

The dot operator is an **infix** operator that looks up string keys in a map.
```python
def mymap ( "a":1 "b":2 "c":3 )

# this is effectively the same as [mymap "a"]
mymap.a # => 1

mymap.z # => undefined
```

The dot operator is not defined on non-map objects.
```python
def a 1
a.baz # => undefined
```

# Builtins
## Identifiers
### undefined
The `undefined` identifier behaves much like a literal. It evaluates to itself and is generally the result of a failed or undefined operation, such as:
- looking up a key that does not exist in a map
- indexing a list out of bounds
- calling a function with the wrong arguments

```python
# mathematical operations on non-numeric types are undefined
+ "a" "b" # => undefined
* 1 "hello" # => undefined

def mymap ( "foo": "bar" )
mymap "quux" # => undefined

def mylist { 1 2 "a" }
mylist 3 # => undefined
```

### \_\_filename\_\_
The `__filename__` identifier is string that contains the absolute path to the file being executed, or `"-"` if the Golsp interpreter is reading from stdin.

### \_\_dirname\_\_
The `__dirname__` identifier is string that contains the absolute path to the directory containing the file being executed, or `"."` if the Golsp interpreter is reading from stdin.

### \_\_args\_\_
The `__args__` identifier is a list of strings provided as command-line arguments to the file being executed.

```python
# program.golsp
printf "%v\n" __args__
```

```
$ golsp program.golsp hello --world
{hello --world }
```

## Functions
### def and const
`def` and `const` are the builtin assignment functions. They define symbols and functions within their local scope.

`def` and `const` take two arguments:
- an identifier or function pattern
- a value or function body

`const` defines constants and `def` defines variable symbols. This is true of function definitions as well -- `const` can only define functions for a single [pattern](docs?id=pattern-matching), unlike `def`.

`def` and `const` return the value defined.

```python
def a 1 # => 1
a # => 1
def a 2
a # => 2

const b "baz"
b # => "baz"
def b "asdf" # => undefined

def [f n] [+ 1 n] # => <function:f>
f 3 # => 4
```

#### Pattern matching
Functions can be defined differently for different sets of arguments, specified by patterns. For example:
```python
# factorial is defined for two patterns: 0 and n
def [factorial 0] 1
def [factorial n] [* n [factorial [- n 1]]]

factorial 5 # => 120
```

When a function is called, the arguments passed to it are compared against the patterns for which it is defined (in the order they are defined) until a match is found. If no matches are found, the expression evaluates to [undefined](docs?id=undefined).

Patterns can also match against and unpack lists and maps:
```python
def [add { a b }] [+ a b]
add { 1 2 } # => 3

def [pairHas3? { 3 a }] "yes"
def [pairHas3? { a 3 }] "yes"
def [pairHas3? { a b }] "no"
def [pairHas3? x] "not a pair"
pairHas3? { 1 3 } # => "yes"
pairHas3? "abc" # => "not a pair"
pairHas3? { 4 5 } # => "no"
pairHas3? { 3 2 } # => "yes"

# this function is only defined for maps with one key, where the key
# is equal to "name"
def [getName ( "name":n )] n
getName ( "name":"Ajay" ) # => "Ajay"
getName ( "name":"Zaphod Beeblebrox" "z00t":"qwer" ) # => undefined
getName 13 # => undefined
```

The spread operator functions differently inside patterns: instead of spreading, it gathers lists and maps into a single identifier. This is best illustrated with an example:
```python
# this function returns the head of a list
def [getHead { head tail... }] head

# this function returns the tail of a list
def [getTail { head tail... }] tail

getHead { 1 2 3 } # => 1
getTail { 1 2 3 } # => { 2 3 }

# this function extracts the value of the "name" key in an arbitrary map
def [greet ( "name":name rest... )] [sprintf "Hello, %v!" name]
def [greet ( keys... )] "Please introduce yourself!"
def [greet _] "You're not a map!"
greet ( "profession":"chef" "name":"Gordon Ramsay" ) # => "Hello, Gordon Ramsay!"
greet ( "foo":"bar" ) # => "Please introduce yourself!"
greet 12 # => "You're not a map!"
```

The spread operator allows patterns to match lists and maps of any size. It can also be used to define variadic functions:
```python
# this function returns all of the arguments passed to it as a list
def [asList args...] args
args 1 2 3 # => { 1 2 3 }

# this function takes at least two arguments
def [add x y] [+ x y]
def [add x ys...] [+ x [add ys...]]

add 1 # => undefined
add 1 2 # => 3
add 1 2 3 # => 6
```

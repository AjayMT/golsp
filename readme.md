
# Golsp
(pronounced "Go-lisp")
```python
printf "Hello, world!\n"
```

Golsp is a simple, interpreted lisp-like programming language. It is not intended to be fast, efficient or particularly sophisticated -- I wrote this as an exercise in programming language design, and for fun. As such, the language itself (`core/`) is feature complete (aside from a few trivial `TODO`s) and all that remains unfinished is the standard library.

Documentation coming soon!

**Table of contents**
- [Syntax and features](#syntax_and_features)
  - [Lists](#lists)
  - [Maps](#maps)
  - [Spread operator](#spread_operator)
  - [Pattern matching](#pattern_matching)
  - [Scopes, modules and concurrency](#scopes_modules_and_concurrency)
- [Installation](#installation)
- [Usage](#usage)
- [Contributing](#contributing)
- [Author(s)](#authors)

## <a name="syntax_and_features">❖</a> Syntax and features
On the surface, Golsp looks like an odd dialect of Lisp with some superficial syntactic changes:
```python
# comments begin with '#'
# expressions are enclosed by '[]'
[def [double x] [* x 2]]
[double 4] # => 8
```

My choice of brackets is not entirely arbitrary: `[` and `]` require very little effort to type, and it is an homage to Objective-C, which is one of my favourite languages. If you really don't like the look of dozens of nested square brackets (even I, a seasoned Golsper, sometimes tire of them), you can simply replace them with newlines.
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
# Golsp will not wrap lines that are inside lists or maps (more on those later)
# in expressions
```

Golsp is only similar to Lisp at a superficial level. At heart, it is much more like a stripped-down, simplified, functional version of Javascript. Golsp is a strong-but-dynamically typed language with two primitive types:
```python
1 1.1 -3.5 1200 # numbers
"hello" "foo" "bar" "baz" # strings
# there is no boolean type -- Golsp uses numbers instead

# the special 'undefined' identifier has no value and evaluates to itself
undefined
```

Variables and constants (including functions) are declared with `def` and `const`. `lambda` creates an anonymous function.
```python
def x 1 # x is a number
def [square n] [* n n] # square is a function
def square [lambda [n] [* n n]] # this is effectively the same as the previous statement
const a "hello" # 'a' is a string constant
def x 2 # this works
def a 3 # this does not
```

Golsp has two simple built-in data structures.

### <a name="lists">❖</a> Lists
```python
def mylist { 5 6 7 "a" + "c" }
```

Calling Golsp's lists 'lists' is somewhat misleading since they are actually immutable. Lists can be indexed and sliced in a variety of ways:
```python
# providing a single 'argument' indexes the list
mylist 0 # => 5
# negative indices count backwards from the end of the list
mylist -1 # => "c"

# two arguments slice the list -- begin (inclusive) and end (exclusive)
mylist 1 5 # => { 6 7 "a" + }
mylist 0 -1 # => { 5 6 7 "a" + }
# 'undefined' slices until the end of the list
mylist 2 undefined # => { 7 "a" + "c" }

# three arguments slice the list and skip elements -- begin, end and step
mylist 0 undefined 2 # => { 5 7 + }
# negative step reverses the list
mylist -1 0 -1 # => { "c" + "a" 7 6 }
mylist -2 undefined -2 # => { + 7 5 }
```

Strings can also be indexed and sliced like lists.

Golsp's parser will not automatically convert newlines to expression delimiters inside lists. This means that
```python
{
  a b c
}
```
is **NOT** evaluated as
```python
{
  [a b c]
}
```
since tokens inside lists are usually meant to be separate elements and not expressions.

### <a name="maps">❖</a> Maps
```python
def mymap (
  "a": 1
  "foo": "bar"
  12: "quux"
  "plus": +
)
```

Maps map literals (i.e strings and numbers) to arbitrary values. Like lists, maps are immutable. They are also ordered -- key-value pairs are inserted in the order they are specified.
```python
# single 'arguments' lookup a key
mymap "a" # => 1
mymap 12 # => "quux"

# multiple keys produce a list of values
mymap "foo" "plus" 12 # => { "bar" + "quux" }

# repeating a key overwrites its previous value
( "a":1 "a":2 ) # => ( "a":2 )
```

String keys can also be evaluated with the special 'dot' syntax:
```python
# this is equivalent to [mymap "a"]
mymap.a # => 1
mymap.plus # => +
```

As in lists, tokens bounded by newlines are not wrapped in expressions.
```python
(
  "a": "b"
)

# is NOT evaluated as

(
  ["a": "b"]
)
```

### <a name="spread_operator">❖</a> Spread operator (`...`)
The spread operator `...` takes a list, map or string and distributes its contents into the surrounding expression.
```python
def mylist { 1 2 3 }
printf "%v %v %v\n" mylist... # => prints "1 2 3"

def str "abc"
printf "%v %v %v\n" str... # => prints "a b c"
```

The spread operator makes typical list operations like appending and inserting very simple.
```python
def [append list value] { list... value }
def [insert list index value] { [list 0 index]... value [list index undefined]... }
def [join list1 list2] { list1... list2... }

append { 1 2 3 } 4 # => { 1 2 3 4 }
insert { 1 2 4 5 } 2 3 # => { 1 2 3 4 5 }
join { 1 2 3 } { 4 5 6 } # => { 1 2 3 4 5 6 }
```

Lists can also be evaluated as expressions.
```python
def expr { + 1 1 }
[expr...] # => 2
```

Unlike lists, spreading a map produces its keys. Maps are ordered, so they spread to their keys in the same order every time.
```python
def mymap ( "a":1 "b":2 "c":3 )
printf "%v %v %v\n" mymap... # => prints "a b c"
printf "%v %v %v\n" [mymap mymap...] # => prints "1 2 3"
```

The spread operator can 'zip' keys and values together when used inside a map.
```python
def keys { "foo" "bar" "baz" }
def values { 1 2 3 4 }
def map ( keys... : values... ) # => ( "foo":1 "bar":2 "baz":3 )

# since maps are ordered, constructing new maps from old ones is simple
def map2 (
  map... : [map map...]...
  "foo": 5
  "quux": "z00t"
) # => ( "foo":5 "bar":2 "baz":3 "quux":"z00t" )
```

### <a name="pattern_matching">❖</a> Pattern matching
Functions in Golsp are produced by the `def`, `const` and `lambda` builtins.
```python
def [square n] [* n n]
const [double n] [* 2 n]
def increment [lambda [x] [+ 1 x]]
```

As in many other functional languages, Golsp features pattern matching. Patterns function as implicit 'switch' statements -- arguments are compared against them in the order they are defined until a perfect match is found.
```python
# it is important to define [factorial 0] before [factorial n] -- otherwise
# this function never terminates
def [factorial 0] 1
[def [factorial n]
  * n [factorial [- n 1]]
]

factorial 6 # => 720
```

Patterns can also automatically de-structure data and 'gather' it (the opposite of spreading). This is best illustrated with an example:
```python
# 'len' finds the length of a list
def [len {}] 0
def [len { head tail... }] [+ 1 [len tail]]
len { 1 2 3 4 } # => 4

# functions can be variadic thanks to patterns
def [count args...] [len args]
count "a" "b" "c" 1 2 3 # => 6

# patterns can also match against and extract data from maps
def [greet ( "name":name rest... )] [sprintf "Hello, %v!" name]
def [greet ( keys... )] "Please introduce yourself!"
def [greet _] "You're not a map!"
greet ( "profession":"chef" "name":"Gordon Ramsay" ) # => Hello, Gordon Ramsay!
greet ( "foo":"bar" ) # => Please introduce yourself!
greet 12 # => You're not a map!
```

Pattern matching works well with the builtin `when` function and `types` module to provide simple and flexible polymorphism:
```python
const types [require "stdlib/types.golsp"] # basic type checking
const _ [require "stdlib/tools.golsp"] # map, filter and other higher-order functions

def [typeof x] [when
  [types.isNumber x]: "number"
  [types.isString x]: "string"
  [types.isFunction x]: "function"
  [types.isList x]: [sprintf "list(%v)" [typeof x...]]
  1: "map"
]
def [typeof xs...] [_.map typeof xs]

typeof 1 # => "number"
typeof 2 "a" typeof # => { "number" "string" "function" }
typeof { 1 2 3 "a" "b" "c" } # => "list({number number number string string string })"
```

### <a name="scopes_modules_and_concurrency">❖</a> Scopes, modules and concurrency
Functions in Golsp are evaluated in their own scopes -- they cannot re-bind symbols defined in outer scopes. Golsp also has a `do` builtin function that defines a scope of its own.
```python
[do
  def name "Ajay"
  printf "hello %v\n" name
]
```

`do` blocks evaluate to the result of the last statement in the block.
```python
def [f x] [do
  def doubled [* 2 x]
  def squared [* doubled doubled]
  def halved [/ squared 2]
  + 1 halved
]

f 2 # => 9
```

Since `do` blocks don't have side-effects, it is safe to execute them concurrently. This is what the builtin `go` function does.
```python
def x 1
[go
  def x 2
  sleep 1000
  printf "world %v\n" x
]
sleep 500
printf "hello %v " x
# prints "hello 1 world 2"
```
Golsp's `go` blocks are a thin layer atop Go's goroutines, which means they're lightweight and efficient.

Files are effectively the same as `do` blocks -- they define a scope, and they 'evaluate' to the result of the last statement. This is the basis of Golsp's module system (which is actually almost too simple to be a 'module system').
```python
##### a.golsp #####
def [double x] [* x 2]
def [square x] [* x x]

# this map gets exported since it is the last statement
(
  "double": double
  "square": square
)

##### b.golsp #####
# the 'require' function evaluates and imports a file
# the specified path is resolved relative to the current file ("b.golsp" in this case)
const a [require "a.golsp"]
a.double 3 # => 6
a.square 9 # => 81
```

`require` can also import standard library modules -- it will do so if the provided path begins with `stdlib/` (see Installation and `GOLSPPATH` below.)

## <a name="installation">❖</a> Installation
Unfortunately, Golsp only supports Linux and macOS at the moment. This installation process assumes that you have GNU make and Go installed, and that your `GOPATH` is set up correctly.

```sh
go get github.com/ajaymt/golsp
cd $GOPATH/src/github.com/ajaymt/golsp
make
go install
export GOLSPPATH="$GOPATH/src/github.com/ajaymt/golsp" # add this to your dotfile
```

## <a name="usage">❖</a> Usage
```sh
golsp [file] # execute 'file'
golsp -      # read from stdin
```

The CLI will eventually get better.

## <a name="contributing">❖</a> Contributing
Yes please! I will merge your code as long as it is:
- tested. A simple test will do -- I haven't written any comprehensive unit tests yet.
- readable. I'm not very picky about style -- I just like to follow a consistent naming convention. But please indent with tabs and be generous with whitespace.
- (reasonably) fast. Do not sacrifice a lot of generality and readability for speed, but don't write bubblesort either.

Here are some things I haven't done yet:
- written tests
- finished the CLI
- finished the builtin string formatter (see `formatStr` in `core/builtins.go`)
- finished the standard library (! high priority !)
- written documentation (! higher priority !)
- other miscellaneous `TODO`s in the codebase

Contributing is not limited to writing code. If you find a bug, want a feature or just want to discuss some ideas, please raise an issue!

## <a name="authors">❖</a> Author(s)
- Ajay Tatachar (ajaymt2@illinois.edu)

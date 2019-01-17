
# Syntax
Golsp has six basic types of syntax nodes:
- comments
- identifiers
- literals
- expressions
- lists
- maps

Golsp has three operators:
- spread
- zip
- dot

## Comments
Comments in Golsp begin with `#` and end with a newline. Comments have no semantic meaning.
```python
# this is a comment
foo bar baz # this is a comment that begins in the middle of a line

```

## Identifiers
Identifiers are space-separated tokens that evaluate to values (literals, functions, lists, maps or `undefined`). Identifiers can contain all characters except syntactic delimiters and operators.
```python
a b c quux z00t # these are identifiers
```

## Literals
Literals can be of two types: **strings** and **numbers**.

String literals are delimited by `"`.
```python
"hello" "world" # these are strings
```

Numbers are contiguous groups of numeric characters surrounded by spaces or syntactic delimiters. Numbers can begin with a minus sign and can contain a single decimal point.
```python
1 2 3 4.5 -6 -7.8 # these are numbers
```

## Expressions
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

If the expression head is a literal, the expression simply evaluates to the literal.
```python
["hello" 1 2] # => "hello"
[17 21 93 z b t] # => 17
```

If the expression head is a function, the function is applied to the arguments. If the function is defined for those arguments, it returns a result. Otherwise, the expression evaluates to `undefined`.
```python
[+ 1 2] # => 3
[+ "a" "b"] # => undefined
```

If the expression head is a list or map, the expression evaluates to different things depending on the arguments (see [Lists](docs?id=lists) and [Maps](docs?id=maps).)

## Lists
Lists are delimited by curly braces and can contain other syntax nodes.
```python
# this is a list containing numbers, a list of strings, and the
# result of the expression `[foo baz]`:
{ 1 2 3 { "a" "b" "c" } [foo baz] }
```

If a list is an expression head (see [Expressions](docs?id=expressions)) with one or more arguments, it will be indexed or sliced as follows:
```python
# list indices begin at 0
[{ 1 2 3 } 0] # => 1

# negative indices count backwards from the end of the list
[{ 1 2 3 } -1] # => 3

# two arguments (begin, inclusive and end, exclusive) slice the list
[{ 1 2 3 4 } 0 2] # => { 1 2 }

# `undefined` will slice until the end of the list
[{ 1 2 3 4 5 6 } 0 undefined] # => { 1 2 3 4 5 6 }

@ a third argument produces a slice with a step
[{ 1 2 3 4 5 6 } 0 undefined 2] # => { 1 3 5 }

# steps can be negative, which will reverse the list
[{ 1 2 3 4 5 6 } -1 undefined -2] # => { 6 4 2 }
```

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
This is because space-separated syntax nodes inside lists are usually intended to be separate list elements and not expressions.

## Maps
Maps are delimited by parentheses and can contain pairs of syntax nodes joined by zip operators (see [Zip](docs?id=zip)).
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

If a map is an expression head with one or more arguments, those arguments will be looked up in the map.
```python
# providing a single argument produces a single value
[( "a":1 "b":2 ) "b"] # => 2

# providing multiple arguments produces a list of values
[( "Cat":"z00t" "beep":"boop" 4:"quux" ) "beep" 4 "foo"] # => { "boop" "quux" undefined }
```

Like in [lists](docs?id=lists), Golsp does not automatically wrap lines in expressions inside maps. So this:
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


// Definitions

package main

// STNode: A single syntax tree node that has a 'head' (i.e value), type,
// list of child nodes and operator flags for the two postfix operators

type STNodeType int
const (
	STNodeTypeScope STNodeType = 0
	STNodeTypeExpression STNodeType = 1
	STNodeTypeStringLiteral STNodeType = 2
	STNodeTypeNumberLiteral STNodeType = 3
	STNodeTypeList STNodeType = 4
	STNodeTypeIdentifier STNodeType = 5
	STNodeTypeComment STNodeType = 6
)

type STNode struct {
	Head string
	Type STNodeType
	Children []STNode
	Spread bool
	Zip bool
}

// /STNode

// GolspScope: A 'scope' object that has a parent scope and a map of strings
// to Golsp objects

type GolspScope struct {
	Parent *GolspScope
	Identifiers map[string]GolspObject
}

// /GolspScope

// GolspFunction: A Golsp function object that cantains a list of patterns
// for which it is defined and an expression (i.e function body) for each
// pattern. If it is a builtin function (i.e implemented in Go), it contains a
// pointer to a Go function with a specific signature

type GolspBuiltinFunction func(GolspScope, []GolspObject) GolspObject
type GolspFunction struct {
	FunctionPatterns [][]STNode
	FunctionBodies []STNode
	BuiltinFunc GolspBuiltinFunction
}

// /GolspFunction

// GolspObject: A container for basic values in Golsp that wraps literals,
// functions and lists. Contains a type, value (for literals), function struct
// (for functions) and a list of elements (for lists). The scope property is the
// scope in which the object was created, primarily used to create closures

type GolspObjectType int
const (
	GolspObjectTypeBuiltinArgument GolspObjectType = -1
	GolspObjectTypeLiteral GolspObjectType = 0
	GolspObjectTypeFunction GolspObjectType = 1
	GolspObjectTypeList GolspObjectType = 2
)

type GolspObject struct {
	Scope GolspScope
	Type GolspObjectType
	Function GolspFunction
	Value STNode
	Elements []GolspObject
}

// /GolspObject

// UNDEFINED: The name and value of the special 'undefined' identifier
const UNDEFINED = "undefined"

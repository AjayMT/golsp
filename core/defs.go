
// Definitions

package golsp

import (
	"fmt"
	"strconv"
)

// STNode: A single syntax tree node that has a 'head' (i.e value), type,
// list of child nodes and flags/fields for the two operators

type STNodeType int
const (
	STNodeTypeScope STNodeType = 0
	STNodeTypeExpression STNodeType = 1
	STNodeTypeStringLiteral STNodeType = 2
	STNodeTypeNumberLiteral STNodeType = 3
	STNodeTypeList STNodeType = 4
	STNodeTypeMap STNodeType = 5
	STNodeTypeIdentifier STNodeType = 6
	STNodeTypeComment STNodeType = 7
)

type STNode struct {
	Head string
	Type STNodeType
	Children []STNode
	Spread bool
	Zip *STNode
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
// functions, lists and maps. Contains a type, value (for literals), function struct
// (for functions), list of elements (for lists) and map (for maps). The scope property
// is the scope in which the object was created, primarily used to create closures

type GolspObjectType int
const (
	GolspObjectTypeBuiltinArgument GolspObjectType = -1
	GolspObjectTypeLiteral GolspObjectType = 0
	GolspObjectTypeFunction GolspObjectType = 1
	GolspObjectTypeList GolspObjectType = 2
	GolspObjectTypeMap GolspObjectType = 3
)

type GolspObject struct {
	Scope GolspScope
	Type GolspObjectType
	Function GolspFunction
	Value STNode
	Elements []GolspObject
	Map map[string]GolspObject
	MapKeys []GolspObject
}

// /GolspObject

// GolspObject constructors

// GolspBuiltinFunctionObject: Produce a Golsp function object from a
// GolspBuiltinFunction-type function
// `fn`: the builtin function
// this function returns the GolspObject containing the builtin function
func GolspBuiltinFunctionObject(fn GolspBuiltinFunction) GolspObject {
	return GolspObject{
		Type: GolspObjectTypeFunction,
		Function: GolspFunction{BuiltinFunc: fn},
	}
}

// GolspUndefinedObject: Produce the UNDEFINED identifier object.
// Used because structs and other data structures are mutable by default and cannot
// be stored in consts
// this function returns the UNDEFINED GolspObject
func GolspUndefinedObject() GolspObject {
	return GolspObject{
		Type: GolspObjectTypeLiteral,
		Value: STNode{
			Head: UNDEFINED,
			Type: STNodeTypeIdentifier,
		},
	}
}

// GolspStringObject: Produce a Golsp string object from a string
// `str`: the string
// this function returns the produced GolspObject
func GolspStringObject(str string) GolspObject {
	return GolspObject{
		Type: GolspObjectTypeLiteral,
		Value: STNode{
			Head: fmt.Sprintf("\"%s\"", str),
			Type: STNodeTypeStringLiteral,
		},
	}
}

// GolspNumberObject: Produce a Golsp number object from a number
// `num`: the number
// this function returns the produced GolspObject
func GolspNumberObject(num float64) GolspObject {
	var head string
	if float64(int(num)) == num {
		head = strconv.Itoa(int(num))
	} else { head = fmt.Sprintf("%g", num) }

	return GolspObject{
		Type: GolspObjectTypeLiteral,
		Value: STNode{
			Head: head,
			Type: STNodeTypeNumberLiteral,
		},
	}
}

// GolspMapObject: Produce a Golsp map object from a map of strings
// to GolspObjects. This function cannot produce maps that bind numbers
// to objects
// `gomap`: the map
// this function returns the produced GolspObject
func GolspMapObject(gomap map[string]GolspObject) GolspObject {
	object := GolspObject{
		Type: GolspObjectTypeMap,
		Map: make(map[string]GolspObject),
		MapKeys: make([]GolspObject, 0, len(gomap)),
	}
	for k, v := range gomap {
		strobj := GolspStringObject(k)
		object.Map[strobj.Value.Head] = v
		object.MapKeys = append(object.MapKeys, strobj)
	}

	return object
}

// GolspListObject: Produce a Golsp list object from a slice of strings.
// This function cannot produce lists that contain non-string objects
// `slice`: the slice
// this function returns the produced GolspObject
func GolspListObject(slice []string) GolspObject {
	object := GolspObject{
		Type: GolspObjectTypeList,
		Elements: make([]GolspObject, len(slice)),
	}
	for i, str := range slice {
		object.Elements[i] = GolspStringObject(str)
	}

	return object
}

// /GolspObject constructors

// names of special builtin identifiers
const UNDEFINED = "undefined"
const DIRNAME = "__dirname__"
const FILENAME = "__filename__"
const ARGS = "__args__"

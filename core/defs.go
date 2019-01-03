
// Definitions

package golsp

import (
	"fmt"
	"strconv"
)

// STNode: A single syntax tree node that has a 'head' (i.e value), type,
// list of child nodes and flags/fields for operators

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
	Dot *STNode
}

// /STNode

// Scope: A 'scope' struct that has a parent scope and a map of strings
// to objects

type Scope struct {
	Parent *Scope
	Identifiers map[string]Object
	Constants map[string]bool
}

// /Scope

// Function: A function struct that contains a name, list of patterns
// for which it is defined and an expression (i.e function body) for each
// pattern. If it is a builtin function (i.e implemented in Go), it contains a
// function pointer with a specific signature

type BuiltinFunction func(Scope, []Object) Object
type Function struct {
	Name string
	FunctionPatterns [][]STNode
	FunctionBodies []STNode
	BuiltinFunc BuiltinFunction
}

// /Function

// Object: A container for basic values that wraps literals, functions,
// lists and maps. Contains a type, value (for literals), function struct
// (for functions), list of elements (for lists) and map (for maps). The scope property
// is the scope in which the object was created, primarily used to create closures

type ObjectType int
const (
	ObjectTypeBuiltinArgument ObjectType = -1
	ObjectTypeLiteral ObjectType = 0
	ObjectTypeFunction ObjectType = 1
	ObjectTypeList ObjectType = 2
	ObjectTypeMap ObjectType = 3
)

type Object struct {
	Scope Scope
	Type ObjectType
	Function Function
	Value STNode
	Elements []Object
	Map map[string]Object
	MapKeys []Object
}

// /Object

// Object constructors

// BuiltinFunctionObject: Produce a function object from a BuiltinFunction-type function
// `name`: the name of the function
// `fn`: the builtin function
// this function returns the Object containing the builtin function
func BuiltinFunctionObject(name string, fn BuiltinFunction) Object {
	return Object{
		Type: ObjectTypeFunction,
		Function: Function{Name: name, BuiltinFunc: fn},
	}
}

// UndefinedObject: Produce the UNDEFINED identifier object.
// Used because structs and other data structures are mutable by default and cannot
// be stored in consts
// this function returns the UNDEFINED Object
func UndefinedObject() Object {
	return Object{
		Type: ObjectTypeLiteral,
		Value: STNode{
			Head: UNDEFINED,
			Type: STNodeTypeIdentifier,
		},
	}
}

// StringObject: Produce a string object from a string
// `str`: the string
// this function returns the produced Object
func StringObject(str string) Object {
	return Object{
		Type: ObjectTypeLiteral,
		Value: STNode{
			Head: fmt.Sprintf("\"%s\"", str),
			Type: STNodeTypeStringLiteral,
		},
	}
}

// NumberObject: Produce a number object from a number
// `num`: the number
// this function returns the produced Object
func NumberObject(num float64) Object {
	var head string
	if float64(int(num)) == num {
		head = strconv.Itoa(int(num))
	} else { head = fmt.Sprintf("%g", num) }

	return Object{
		Type: ObjectTypeLiteral,
		Value: STNode{
			Head: head,
			Type: STNodeTypeNumberLiteral,
		},
	}
}

// MapObject: Produce a map object from a map of strings
// to Objects. This function cannot produce maps that bind numbers
// to objects
// `gomap`: the map
// this function returns the produced Object
func MapObject(gomap map[string]Object) Object {
	object := Object{
		Type: ObjectTypeMap,
		Map: make(map[string]Object),
		MapKeys: make([]Object, 0, len(gomap)),
	}
	for k, v := range gomap {
		strobj := StringObject(k)
		object.Map[strobj.Value.Head] = v
		object.MapKeys = append(object.MapKeys, strobj)
	}

	return object
}

// ListObject: Produce a list object from a slice of strings.
// This function cannot produce lists that contain non-string objects
// `slice`: the slice
// this function returns the produced Object
func ListObject(slice []string) Object {
	object := Object{
		Type: ObjectTypeList,
		Elements: make([]Object, len(slice)),
	}
	for i, str := range slice {
		object.Elements[i] = StringObject(str)
	}

	return object
}

// /Object constructors

// names of special builtin identifiers
const UNDEFINED = "undefined"
const DIRNAME = "__dirname__"
const FILENAME = "__filename__"
const ARGS = "__args__"

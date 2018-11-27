
// Definitions

package main

type STNodeType int

const (
	STNodeTypeScope STNodeType = 0
	STNodeTypeExpression STNodeType = 1
	STNodeTypeStringLiteral STNodeType = 2
	STNodeTypeNumberLiteral STNodeType = 3
	STNodeTypeList STNodeType = 4
	STNodeTypeIdentifier STNodeType = 5
	STNodeTypeComment = 6
)

type STNode struct {
	Head string
	Type STNodeType
	Children []STNode
	Spread bool
}

type GolspScope struct {
	Parent *GolspScope
	Identifiers map[string]GolspObject
}

type GolspBuiltinFunctionBody func(GolspScope, []GolspObject) GolspObject

type GolspFunction struct {
	FunctionPatterns [][]STNode
	FunctionBodies []STNode

	BuiltinPatterns [][]STNode
	BuiltinBodies []GolspBuiltinFunctionBody
}

type GolspObjectType int

const (
	GolspObjectTypeBuiltinArgument GolspObjectType = -1
	GolspObjectTypeLiteral GolspObjectType = 0
	GolspObjectTypeFunction = 1
	GolspObjectTypeList GolspObjectType = 2
)

type GolspObject struct {
	Scope GolspScope
	Type GolspObjectType
	Function GolspFunction
	Value STNode
	Elements []GolspObject
}

const UNDEFINED = "undefined"

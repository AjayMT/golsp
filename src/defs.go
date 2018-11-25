
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
}

type GolspScope map[string]GolspObject

type GolspBuiltinFunctionBody func(GolspScope, []STNode) GolspObject

type GolspFunction struct {
	FunctionPatterns [][]STNode
	FunctionBodies []STNode

	BuiltinPatterns [][]STNode
	BuiltinBodies []GolspBuiltinFunctionBody
}

type GolspObjectType int

const (
	GolspObjectTypeLiteral GolspObjectType = 0
	GolspObjectTypeFunction = 1
	GolspObjectTypeList GolspObjectType = 2
)

type GolspObject struct {
	Type GolspObjectType
	Function GolspFunction
	Value STNode
	Elements []GolspObject
}

const UNDEFINED = "undefined"

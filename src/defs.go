
// Definitions

package main

type STNodeType int

const (
	STNodeTypeScope STNodeType = 0
	STNodeTypeExpression STNodeType = 1
	STNodeTypeStringLiteral STNodeType = 2
	STNodeTypeNumberLiteral STNodeType = 3
	STNodeTypeIdentifier STNodeType = 4
	STNodeTypeComment = 5
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

type GolspObject struct {
	IsFunction bool
	Function GolspFunction
	Value STNode
}

const UNDEFINED = "undefined"


package main

import (
	"strconv"
	g "github.com/ajaymt/golsp/core"
)

func typeCheck(objectType g.GolspObjectType, nodeType g.STNodeType) g.GolspBuiltinFunction {
	return func (scope g.GolspScope, args []g.GolspObject) g.GolspObject {
		arguments := g.EvalArgs(scope, args)
		if len(arguments) < 1 {
			return g.Builtins.Identifiers[g.UNDEFINED]
		}
		if arguments[0].Type != objectType {
			return g.GolspNumberObject(0.0)
		}
		if objectType == g.GolspObjectTypeLiteral && arguments[0].Value.Type != nodeType {
			return g.GolspNumberObject(0.0)
		}

		return g.GolspNumberObject(1.0)
	}
}

func parseNumber(scope g.GolspScope, args []g.GolspObject) g.GolspObject {
	arguments := g.EvalArgs(scope, args)
	str := arguments[0].Value.Head
	num, err := strconv.ParseFloat(str[1:len(str) - 1], 64)
	if err != nil { return g.Builtins.Identifiers[g.UNDEFINED] }

	return g.GolspNumberObject(num)
}

var Exports = g.GolspMapObject(map[string]g.GolspObject{
	"isString": g.GolspBuiltinFunctionObject("isString",
		typeCheck(g.GolspObjectTypeLiteral, g.STNodeTypeStringLiteral)),
	"isNumber": g.GolspBuiltinFunctionObject("isNumber",
		typeCheck(g.GolspObjectTypeLiteral, g.STNodeTypeNumberLiteral)),
	"isFunction": g.GolspBuiltinFunctionObject("isFunction",
		typeCheck(g.GolspObjectTypeFunction, g.STNodeTypeIdentifier)),
	"isList": g.GolspBuiltinFunctionObject("isList",
		typeCheck(g.GolspObjectTypeList, g.STNodeTypeIdentifier)),
	"isMap": g.GolspBuiltinFunctionObject("isMap",
		typeCheck(g.GolspObjectTypeMap, g.STNodeTypeIdentifier)),

	"parseNumber": g.GolspBuiltinFunctionObject("parseNumber", parseNumber),
})

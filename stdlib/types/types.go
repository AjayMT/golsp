
package main

import (
	"strconv"
	g "github.com/ajaymt/golsp/core"
)

func typeCheck(objectType g.ObjectType, nodeType g.STNodeType) g.BuiltinFunction {
	return func (scope g.Scope, args []g.Object) g.Object {
		arguments := g.EvalArgs(scope, args)
		if len(arguments) < 1 {
			return g.Builtins.Identifiers[g.UNDEFINED]
		}
		if arguments[0].Type != objectType {
			return g.NumberObject(0.0)
		}
		if objectType == g.ObjectTypeLiteral && arguments[0].Value.Type != nodeType {
			return g.NumberObject(0.0)
		}

		return g.NumberObject(1.0)
	}
}

func parseNumber(scope g.Scope, args []g.Object) g.Object {
	arguments := g.EvalArgs(scope, args)
	str := arguments[0].Value.Head
	num, err := strconv.ParseFloat(str[1:len(str) - 1], 64)
	if err != nil { return g.Builtins.Identifiers[g.UNDEFINED] }

	return g.NumberObject(num)
}

var Exports = g.MapObject(map[string]g.Object{
	"isString": g.BuiltinFunctionObject("isString",
		typeCheck(g.ObjectTypeLiteral, g.STNodeTypeStringLiteral)),
	"isNumber": g.BuiltinFunctionObject("isNumber",
		typeCheck(g.ObjectTypeLiteral, g.STNodeTypeNumberLiteral)),
	"isFunction": g.BuiltinFunctionObject("isFunction",
		typeCheck(g.ObjectTypeFunction, g.STNodeTypeIdentifier)),
	"isList": g.BuiltinFunctionObject("isList",
		typeCheck(g.ObjectTypeList, g.STNodeTypeIdentifier)),
	"isMap": g.BuiltinFunctionObject("isMap",
		typeCheck(g.ObjectTypeMap, g.STNodeTypeIdentifier)),

	"parseNumber": g.BuiltinFunctionObject("parseNumber", parseNumber),
})

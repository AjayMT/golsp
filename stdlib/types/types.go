
package main

import g "github.com/ajaymt/golsp/core"

func typeCheck(objectType g.GolspObjectType, nodeType g.STNodeType) g.GolspObject {
	fn := func (scope g.GolspScope, args []g.GolspObject) g.GolspObject {
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

	return g.GolspBuiltinFunctionObject(fn)
}

var Exports = g.GolspMapObject(map[string]g.GolspObject{
	"isString": typeCheck(g.GolspObjectTypeLiteral, g.STNodeTypeStringLiteral),
	"isNumber": typeCheck(g.GolspObjectTypeLiteral, g.STNodeTypeNumberLiteral),
	"isFunction": typeCheck(g.GolspObjectTypeFunction, g.STNodeTypeIdentifier),
	"isList": typeCheck(g.GolspObjectTypeList, g.STNodeTypeIdentifier),
	"isMap": typeCheck(g.GolspObjectTypeMap, g.STNodeTypeIdentifier),
})

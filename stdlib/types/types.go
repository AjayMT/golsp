
package main

import golsp "github.com/ajaymt/golsp/core"

func isString(scope golsp.GolspScope, args []golsp.GolspObject) golsp.GolspObject {
	arguments := golsp.EvalArgs(scope, args)

	if len(arguments) < 0 {
		return golsp.Builtins.Identifiers[golsp.UNDEFINED]
	}

	result := 0
	if arguments[0].Type == golsp.GolspObjectTypeLiteral &&
		arguments[0].Value.Type == golsp.STNodeTypeStringLiteral {
		result = 1
	}

	return golsp.GolspNumberObject(float64(result))
}

var Exports = golsp.GolspObject{
	Type: golsp.GolspObjectTypeMap,
	Map: map[string]golsp.GolspObject{
		"\"isString\"": golsp.GolspBuiltinFunctionObject(isString),
	},
	MapKeys: []golsp.GolspObject{
		golsp.GolspStringObject("isString"),
	},
}

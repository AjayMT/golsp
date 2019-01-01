
package main

import (
	"os"
	"strconv"
	g "github.com/ajaymt/golsp/core"
)

func exit(scope g.GolspScope, args []g.GolspObject) g.GolspObject {
	arguments := g.EvalArgs(scope, args)
	n, _ := strconv.ParseFloat(arguments[0].Value.Head, 64)
	os.Exit(int(n))
	return g.Builtins.Identifiers[g.UNDEFINED]
}

var Exports = g.GolspMapObject(map[string]g.GolspObject{
	"exit": g.GolspBuiltinFunctionObject("exit", exit),
})


package main

import (
	"os"
	"strconv"
	g "github.com/ajaymt/golsp/core"
)

func exit(scope g.Scope, args []g.Object) g.Object {
	arguments := g.EvalArgs(scope, args)
	n, _ := strconv.ParseFloat(arguments[0].Value.Head, 64)
	os.Exit(int(n))
	return g.Builtins.Identifiers[g.UNDEFINED]
}

var Exports = g.MapObject(map[string]g.Object{
	"exit": g.BuiltinFunctionObject("exit", exit),
})

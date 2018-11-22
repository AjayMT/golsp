
// Evaluator

package main

import (
	"math"
	"fmt"
	"strings"
	"strconv"
)

type GolspScope map[string]GolspFunction

type GolspBuiltinFunctionBody func(GolspScope, []STNode) STNode

type GolspFunction struct {
	FunctionPatterns [][]STNode
	FunctionBodies []STNode

	BuiltinPatterns [][]STNode
	BuiltinBodies []GolspBuiltinFunctionBody
}

var Builtins = GolspScope{
	"undefined": GolspFunction{
		FunctionPatterns: make([][]STNode, 0),
		FunctionBodies: []STNode{GolspUndefinedIdentifier},
		BuiltinPatterns: make([][]STNode, 0),
		BuiltinBodies: make([]GolspBuiltinFunctionBody, 0),
	},

	"=": GolspFunction{
		FunctionPatterns: make([][]STNode, 0),
		FunctionBodies: make([]STNode, 0),
		BuiltinPatterns: [][]STNode{
			[]STNode{
				STNode{Head: "a", Type: STNodeTypeIdentifier, Children: make([]STNode, 0)},
				STNode{Head: "b", Type: STNodeTypeIdentifier, Children: make([]STNode, 0)},
			},
		},
		BuiltinBodies: []GolspBuiltinFunctionBody{GolspBuiltinEquals},
	},

	"+": GolspFunction{
		FunctionPatterns: make([][]STNode, 0),
		FunctionBodies: make([]STNode, 0),
		BuiltinPatterns: [][]STNode{
			[]STNode{
				STNode{Head: "n", Type: STNodeTypeIdentifier, Children: make([]STNode, 0)},
			},
		},
		BuiltinBodies: []GolspBuiltinFunctionBody{GolspBuiltinPlus},
	},

	"printf": GolspFunction{
		FunctionPatterns: make([][]STNode, 0),
		FunctionBodies: make([]STNode, 0),
		BuiltinPatterns: [][]STNode{
			[]STNode{
				STNode{Head: "s", Type: STNodeTypeIdentifier, Children: make([]STNode, 0)},
			},
		},
		BuiltinBodies: []GolspBuiltinFunctionBody{GolspBuiltinPrintf},
	},
}

var GolspUndefinedIdentifier = STNode{
	Head: "undefined",
	Type: STNodeTypeIdentifier,
	Children: make([]STNode, 0),
}

func GolspBuiltinEquals(scope GolspScope, arguments []STNode) STNode {
	if len(arguments) == 0 {
		panic("'=' function requires at least one argument")
	}

	symbol := arguments[0]
	value := GolspUndefinedIdentifier
	pattern := make([]STNode, 0)

	if len(arguments) > 1 {
		value = arguments[1]
	}

	if symbol.Type == STNodeTypeExpression {
		head := symbol.Children[0]
		if head.Type != STNodeTypeIdentifier {
			panic("symbol must be identifier")
		}

		pattern = symbol.Children[1:]
		for i, _ := range pattern {
			for pattern[i].Type == STNodeTypeExpression {
				pattern[i] = eval(scope, pattern[i])
			}
		}

		symbol = head
	}

	_, exists := scope[symbol.Head]

	if !exists {
		scope[symbol.Head] = GolspFunction{
			FunctionPatterns: make([][]STNode, 0),
			FunctionBodies: make([]STNode, 0),
			BuiltinPatterns: make([][]STNode, 0),
			BuiltinBodies: make([]GolspBuiltinFunctionBody, 0),
		}
	}

	patternexists := false
	patternindex := 0

	for index, p := range scope[symbol.Head].FunctionPatterns {
		i := 0
		for i, node := range p {
			if i >= len(pattern) { continue }
			if node.Type == pattern[i].Type { i++ }
		}

		if i == len(p) && i == len(pattern) {
			patternexists = true
			patternindex = index
		}
	}

	if patternexists {
		scope[symbol.Head].FunctionBodies[patternindex] = value
		return value
	}

	newfn := GolspFunction{
		FunctionPatterns: append(scope[symbol.Head].FunctionPatterns, pattern),
		FunctionBodies: append(scope[symbol.Head].FunctionBodies, value),
		BuiltinPatterns: scope[symbol.Head].BuiltinPatterns,
		BuiltinBodies: scope[symbol.Head].BuiltinBodies,
	}

	scope[symbol.Head] = newfn

	return value
}

func GolspBuiltinPlus(scope GolspScope, arguments []STNode) STNode {
	for i, _ := range arguments {
		for arguments[i].Head != "undefined" &&
			(arguments[i].Type == STNodeTypeIdentifier ||
			arguments[i].Type == STNodeTypeExpression) {
			arguments[i] = eval(scope, arguments[i])
		}
	}

	argtype := arguments[0].Type

	for _, a := range arguments {
		if a.Type != argtype {
			panic("cannot add arguments of different types")
		}
	}

	nsum := 0.0
	strsum := ""

	for _, v := range arguments {
		text := v.Head
		if argtype == STNodeTypeStringLiteral {
			text = text[1:len(text) - 1]
		}

		n, _ := strconv.ParseFloat(text, 64)
		nsum += n
		strsum += text
	}

	if argtype == STNodeTypeStringLiteral {
		return STNode{
			Head: "\"" + strsum + "\"",
			Type: STNodeTypeStringLiteral,
			Children: make([]STNode, 0),
		}
	}

	return STNode{
		Head: fmt.Sprintf("%v", nsum),
		Type: STNodeTypeNumberLiteral,
		Children: make([]STNode, 0),
	}
}

func GolspBuiltinPrintf(scope GolspScope, arguments []STNode) STNode {
	text := arguments[0].Head
	text = text[1:len(text) - 1]

	var args []interface{}

	for _, v := range arguments[1:] {
		for v.Type != STNodeTypeStringLiteral &&
			v.Type != STNodeTypeNumberLiteral &&
			v.Head != "undefined" {
			v = eval(scope, v)
		}

		if v.Type == STNodeTypeNumberLiteral {
			n, _ := strconv.ParseFloat(v.Head, 64)
			args = append(args, n)
		} else if v.Type == STNodeTypeStringLiteral {
			str := v.Head[1:len(v.Head) - 1]
			args = append(args, str)
		} else {
			args = append(args, v.Head)
		}
	}

	// TODO: replace all literal escape sequences with actual escape characters
	text = strings.Replace(text, "\\n", "\n", -1)
	text = strings.Replace(text, "\\\"", "\"", -1)

	fmt.Printf(text, args...)

	return arguments[0]
}

func compareNodes(a STNode, b STNode) bool {
	if a.Type == STNodeTypeIdentifier {
		return true
	}

	return a.Head == b.Head
}

func matchPatterns(fn GolspFunction, pattern []STNode) int {
	patterns := fn.FunctionPatterns

	if len(fn.BuiltinPatterns) > 0 {
		patterns = fn.BuiltinPatterns
	}

	bestmatchscore := 0
	bestmatchindex := 0

	for i, p := range patterns {
		score := 0
		minlen := int(math.Min(float64(len(p)), float64(len(pattern))))

		for j := 0; j < minlen; j++ {
			if compareNodes(p[j], pattern[j]) { score++ }
		}

		if score > bestmatchscore {
			bestmatchscore = score
			bestmatchindex = i
		}
	}

	return bestmatchindex
}

func Eval(root STNode) STNode {
	defaultProgramScope := make(GolspScope)

	for k, v := range Builtins {
		defaultProgramScope[k] = v
	}

	return eval(defaultProgramScope, root)
}

func eval(scope GolspScope, root STNode) STNode {
	if root.Type == STNodeTypeScope {
		var result STNode

		for _, child := range root.Children {
			result = eval(scope, child)
		}

		return result
	}

	if root.Type == STNodeTypeStringLiteral ||
		root.Type == STNodeTypeNumberLiteral {
		return root
	}

	// root has to be expression or identifier

	if len(root.Children) == 0 && root.Type == STNodeTypeExpression {
		panic("empty expression")
	}

	exprhead := root
	if root.Type == STNodeTypeExpression {
		exprhead = root.Children[0]
	}

	if exprhead.Type != STNodeTypeIdentifier {
		// expression head (i.e function name) is not identifier, so eval it
		exprhead = eval(scope, exprhead)
	}

	if exprhead.Type != STNodeTypeIdentifier {
		// expression head must be literal
		// 'calling' a literal like a function simply produces the literal
		return exprhead
	}

	var arguments []STNode
	if root.Type == STNodeTypeExpression {
		for _, child := range root.Children[1:] {
			arguments = append(arguments, child)
		}
	}

	fn, exists := scope[exprhead.Head]

	if !exists {
		return GolspUndefinedIdentifier
	}

	builtin := len(fn.BuiltinPatterns) > 0
	patternindex := matchPatterns(fn, arguments)

	if builtin {
		return fn.BuiltinBodies[patternindex](scope, arguments)
	}

	// eval function

	pattern := fn.FunctionPatterns[patternindex]
	newscope := make(GolspScope)
	for k, v := range scope {
		newscope[k] = v
	}

	if len(arguments) < len(pattern) {
		// curry function
	}

	for i, symbol := range pattern {
		newscope[symbol.Head] = GolspFunction{
			FunctionPatterns: [][]STNode{make([]STNode, 0)},
			FunctionBodies: []STNode{arguments[i]},
			BuiltinPatterns: make([][]STNode, 0),
			BuiltinBodies: make([]GolspBuiltinFunctionBody, 0),
		}
	}

	return eval(newscope, fn.FunctionBodies[patternindex])
}

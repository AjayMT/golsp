
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

func GolspBuiltinPrintf(scope GolspScope, arguments []STNode) STNode {
	text := arguments[0].Head
	text = text[1:len(text) - 1]

	var args []interface{}

	for _, v := range arguments[1:] {
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

	// root has to be expression

	if len(root.Children) == 0 {
		panic("empty expression")
	}

	exprhead := root.Children[0]

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
	for _, child := range root.Children[1:] {
		arguments = append(arguments, eval(scope, child))
	}

	fn, exists := scope[exprhead.Head]

	if !exists {
		// undefined expression head
	}

	builtin := len(fn.BuiltinPatterns) > 0
	patternindex := matchPatterns(fn, arguments)

	if builtin {
		return fn.BuiltinBodies[patternindex](scope, arguments)
	}

	// eval function

	return STNode{Head: "0", Type: STNodeTypeNumberLiteral, Children: []STNode{}}
}

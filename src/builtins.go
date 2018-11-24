
// Builtins

package main

import (
	"fmt"
	"strings"
	"strconv"
)

var Builtins = GolspScope{
	"undefined": GolspObject{
		IsFunction: false,
		Function: GolspEmptyFunction(),
		Value: GolspUndefinedIdentifier(),
	},

	"=": GolspObject{
		IsFunction: true,
		Function: GolspFunction{
			FunctionPatterns: make([][]STNode, 0),
			FunctionBodies: make([]STNode, 0),
			BuiltinPatterns: [][]STNode{
				[]STNode{
					STNode{
						Head: "a",
						Type: STNodeTypeIdentifier,
						Children: make([]STNode, 0),
					},
					STNode{
						Head: "b",
						Type: STNodeTypeIdentifier,
						Children: make([]STNode, 0),
					},
				},
			},
			BuiltinBodies: []GolspBuiltinFunctionBody{GolspBuiltinEquals},
		},
		Value: GolspUndefinedIdentifier(),
	},

	"+": GolspObject{
		IsFunction: true,
		Function: GolspFunction{
			FunctionPatterns: make([][]STNode, 0),
			FunctionBodies: make([]STNode, 0),
			BuiltinPatterns: [][]STNode{
				[]STNode{
					STNode{
						Head: "n",
						Type: STNodeTypeIdentifier,
						Children: make([]STNode, 0),
					},
				},
			},
			BuiltinBodies: []GolspBuiltinFunctionBody{GolspBuiltinPlus},
		},
		Value: GolspUndefinedIdentifier(),
	},

	"printf": GolspObject{
		IsFunction: true,
		Function: GolspFunction{
			FunctionPatterns: make([][]STNode, 0),
			FunctionBodies: make([]STNode, 0),
			BuiltinPatterns: [][]STNode{
				[]STNode{
					STNode{
						Head: "s",
						Type: STNodeTypeIdentifier,
						Children: make([]STNode, 0),
					},
				},
			},
			BuiltinBodies: []GolspBuiltinFunctionBody{GolspBuiltinPrintf},
		},
		Value: GolspUndefinedIdentifier(),
	},
}

func GolspBuiltinEquals(scope GolspScope, arguments []STNode) STNode {
	if len(arguments) < 2 {
		return GolspUndefinedIdentifier()
	}

	symbol := arguments[0]
	value := arguments[1]
	if symbol.Type != STNodeTypeIdentifier &&
		symbol.Type != STNodeTypeExpression {
		return GolspUndefinedIdentifier()
	}

	if symbol.Type == STNodeTypeIdentifier {
		for !isResolved(scope, value) {
			value = eval(scope, value)
		}

		scope[symbol.Head] = GolspObject{
			IsFunction: false,
			Function: GolspEmptyFunction(),
			Value: value,
		}

		return value
	}

	pattern := make([]STNode, 0)
	head := symbol.Children[0]
	if head.Type != STNodeTypeIdentifier {
		return GolspUndefinedIdentifier()
	}

	pattern = symbol.Children[1:]
	for i, _ := range pattern {
		for pattern[i].Type == STNodeTypeExpression {
			pattern[i] = eval(scope, pattern[i])
		}
	}

	symbol = head
	_, exists := scope[symbol.Head]
	if !exists {
		scope[symbol.Head] = GolspObject{
			IsFunction: true,
			Function: GolspEmptyFunction(),
			Value: GolspUndefinedIdentifier(),
		}
	}

	patternexists := false
	patternindex := 0
	for index, p := range scope[symbol.Head].Function.FunctionPatterns {
		i := 0
		for i, node := range p {
			if i >= len(pattern) { continue }
			if node.Type == pattern[i].Type { i++ }
		}

		if i == len(p) && i == len(pattern) {
			patternexists = true
			patternindex = index
			break
		}
	}

	if patternexists {
		scope[symbol.Head].Function.FunctionBodies[patternindex] = value
		return value
	}

	newfn := GolspFunction{
		FunctionPatterns: append(scope[symbol.Head].Function.FunctionPatterns, pattern),
		FunctionBodies: append(scope[symbol.Head].Function.FunctionBodies, value),
		BuiltinPatterns: scope[symbol.Head].Function.BuiltinPatterns,
		BuiltinBodies: scope[symbol.Head].Function.BuiltinBodies,
	}

	scope[symbol.Head] = GolspObject{
		IsFunction: true,
		Function: newfn,
		Value: GolspUndefinedIdentifier(),
	}

	return value
}

func GolspBuiltinPlus(scope GolspScope, arguments []STNode) STNode {
	for i, _ := range arguments {
		for !isResolved(scope, arguments[i]) {
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
		for !isResolved(scope, v) {
			v = eval(scope, v)
		}

		if v.Type == STNodeTypeNumberLiteral {
			n, _ := strconv.ParseFloat(v.Head, 64)
			args = append(args, n)
		} else if v.Type == STNodeTypeStringLiteral {
			str := v.Head[1:len(v.Head) - 1]
			args = append(args, str)
		} else if v.Head == "undefined" {
			args = append(args, "<undefined>")
		} else {
			args = append(args, fmt.Sprintf("<function:%v>", v.Head))
		}
	}

	// TODO: replace all literal escape sequences with actual escape characters
	text = strings.Replace(text, "\\n", "\n", -1)
	text = strings.Replace(text, "\\\"", "\"", -1)

	fmt.Printf(text, args...)

	return arguments[0]
}

func GolspUndefinedIdentifier() STNode {
	return STNode{
		Head: "undefined",
		Type: STNodeTypeIdentifier,
		Children: make([]STNode, 0),
	}
}

func GolspEmptyFunction() GolspFunction {
	return GolspFunction{
		FunctionPatterns: make([][]STNode, 0),
		FunctionBodies: make([]STNode, 0),
		BuiltinPatterns: make([][]STNode, 0),
		BuiltinBodies: make([]GolspBuiltinFunctionBody, 0),
	}
}

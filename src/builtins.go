
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
			BuiltinPatterns: [][]STNode{make([]STNode, 0)},
			BuiltinBodies: []GolspBuiltinFunctionBody{GolspBuiltinPlus},
		},
		Value: GolspUndefinedIdentifier(),
	},

	"-": GolspObject{
		IsFunction: true,
		Function: GolspFunction{
			FunctionPatterns: make([][]STNode, 0),
			FunctionBodies: make([]STNode, 0),
			BuiltinPatterns: [][]STNode{make([]STNode, 0)},
			BuiltinBodies: []GolspBuiltinFunctionBody{GolspBuiltinMinus},
		},
		Value: GolspUndefinedIdentifier(),
	},

	"*": GolspObject{
		IsFunction: true,
		Function: GolspFunction{
			FunctionPatterns: make([][]STNode, 0),
			FunctionBodies: make([]STNode, 0),
			BuiltinPatterns: [][]STNode{make([]STNode, 0)},
			BuiltinBodies: []GolspBuiltinFunctionBody{GolspBuiltinMultiply},
		},
		Value: GolspUndefinedIdentifier(),
	},

	"/": GolspObject{
		IsFunction: true,
		Function: GolspFunction{
			FunctionPatterns: make([][]STNode, 0),
			FunctionBodies: make([]STNode, 0),
			BuiltinPatterns: [][]STNode{make([]STNode, 0)},
			BuiltinBodies: []GolspBuiltinFunctionBody{GolspBuiltinDivide},
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
		valuescope := MakeScope(scope)
		for !isResolved(valuescope, value) {
			value = eval(valuescope, value)
		}

		scope[symbol.Head] = GolspObject{
			IsFunction: false,
			Function: GolspEmptyFunction(),
			Value: value,
		}

		return symbol
	}

	pattern := make([]STNode, 0)
	head := symbol.Children[0]
	if head.Type != STNodeTypeIdentifier {
		return GolspUndefinedIdentifier()
	}

	pattern = symbol.Children[1:]
	for i, _ := range pattern {
		patternscope := MakeScope(scope)
		for pattern[i].Type == STNodeTypeExpression {
			pattern[i] = eval(patternscope, pattern[i])
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
		return symbol
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

	return symbol
}

func GolspBuiltinPlus(scope GolspScope, arguments []STNode) STNode {
	argscope := MakeScope(scope)
	for i, _ := range arguments {
		for !isResolved(argscope, arguments[i]) {
			arguments[i] = eval(argscope, arguments[i])
		}
	}

	for _, a := range arguments {
		if a.Type != STNodeTypeNumberLiteral {
			return GolspUndefinedIdentifier()
		}
	}

	sum := 0.0
	for _, v := range arguments {
		n, _ := strconv.ParseFloat(v.Head, 64)
		sum += n
	}

	return STNode{
		Head: fmt.Sprintf("%v", sum),
		Type: STNodeTypeNumberLiteral,
		Children: make([]STNode, 0),
	}
}

func GolspBuiltinMinus(scope GolspScope, arguments []STNode) STNode {
	argscope := MakeScope(scope)
	for i, _ := range arguments {
		for !isResolved(argscope, arguments[i]) {
			arguments[i] = eval(argscope, arguments[i])
		}
	}

	for _, a := range arguments {
		if a.Type != STNodeTypeNumberLiteral {
			return GolspUndefinedIdentifier()
		}
	}

	sum := 0.0
	if len(arguments) > 0 {
		n, _ := strconv.ParseFloat(arguments[0].Head, 64)
		sum += n
	}

	for _, v := range arguments[1:] {
		n, _ := strconv.ParseFloat(v.Head, 64)
		sum -= n
	}

	return STNode{
		Head: fmt.Sprintf("%v", sum),
		Type: STNodeTypeNumberLiteral,
		Children: make([]STNode, 0),
	}
}

func GolspBuiltinMultiply(scope GolspScope, arguments []STNode) STNode {
	argscope := MakeScope(scope)
	for i, _ := range arguments {
		for !isResolved(argscope, arguments[i]) {
			arguments[i] = eval(argscope, arguments[i])
		}
	}

	for _, a := range arguments {
		if a.Type != STNodeTypeNumberLiteral {
			return GolspUndefinedIdentifier()
		}
	}

	product := 1.0
	for _, v := range arguments {
		n, _ := strconv.ParseFloat(v.Head, 64)
		product *= n
	}

	return STNode{
		Head: fmt.Sprintf("%v", product),
		Type: STNodeTypeNumberLiteral,
		Children: make([]STNode, 0),
	}
}

func GolspBuiltinDivide(scope GolspScope, arguments []STNode) STNode {
	argscope := MakeScope(scope)
	for i, _ := range arguments {
		for !isResolved(argscope, arguments[i]) {
			arguments[i] = eval(argscope, arguments[i])
		}
	}

	for _, a := range arguments {
		if a.Type != STNodeTypeNumberLiteral {
			return GolspUndefinedIdentifier()
		}
	}

	numerator := 1.0
	if len(arguments) > 0 {
		n, _ := strconv.ParseFloat(arguments[0].Head, 64)
		numerator *= n
	}

	denominator := 1.0
	for _, v := range arguments[1:] {
		n, _ := strconv.ParseFloat(v.Head, 64)
		denominator *= n
	}

	return STNode{
		Head: fmt.Sprintf("%v", numerator / denominator),
		Type: STNodeTypeNumberLiteral,
		Children: make([]STNode, 0),
	}
}

func GolspBuiltinPrintf(scope GolspScope, arguments []STNode) STNode {
	text := arguments[0].Head
	text = text[1:len(text) - 1]

	argscope := MakeScope(scope)
	var args []interface{}
	for _, v := range arguments[1:] {
		for !isResolved(argscope, v) {
			v = eval(argscope, v)
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

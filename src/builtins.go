
// Builtins

package main

import (
	"fmt"
	"strings"
	"strconv"
)

var Builtins = GolspScope{
	UNDEFINED: GolspObject{
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

func GolspBuiltinEquals(scope GolspScope, arguments []STNode) GolspObject {
	if len(arguments) < 2 {
		return scope[UNDEFINED]
	}

	symbol := arguments[0]
	value := arguments[1]

	if symbol.Type != STNodeTypeIdentifier &&
		symbol.Type != STNodeTypeExpression {
		return Eval(scope, symbol)
	}

	if symbol.Type == STNodeTypeIdentifier {
		valuescope := MakeScope(scope)
		scope[symbol.Head] = Eval(valuescope, value)
		return scope[symbol.Head]
	}

	pattern := make([]STNode, 0)
	head := symbol.Children[0]
	if head.Type != STNodeTypeIdentifier {
		return scope[UNDEFINED]
	}

	pattern = symbol.Children[1:]
	for i, _ := range pattern {
		patternscope := MakeScope(scope)
		for pattern[i].Type == STNodeTypeExpression {
			obj := Eval(patternscope, pattern[i])
			if !obj.IsFunction {
				pattern[i] = obj.Value
			}
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
			if i >= len(pattern) { break }
			if node.Type != pattern[i].Type { continue }
			if node.Type != STNodeTypeIdentifier &&
				node.Head != pattern[i].Head { continue }

			i++
		}

		if i == len(p) && i == len(pattern) {
			patternexists = true
			patternindex = index
			break
		}
	}

	if patternexists {
		scope[symbol.Head].Function.FunctionBodies[patternindex] = value
		return scope[symbol.Head]
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

	return scope[symbol.Head]
}

func GolspBuiltinPlus(scope GolspScope, args []STNode) GolspObject {
	var arguments []GolspObject
	argscope := MakeScope(scope)
	for _, a := range args {
		arguments = append(arguments, Eval(argscope, a))
	}

	for _, a := range arguments {
		if a.Value.Type != STNodeTypeNumberLiteral {
			return scope[UNDEFINED]
		}
	}

	sum := 0.0
	for _, v := range arguments {
		n, _ := strconv.ParseFloat(v.Value.Head, 64)
		sum += n
	}

	val := STNode{
		Head: fmt.Sprintf("%v", sum),
		Type: STNodeTypeNumberLiteral,
		Children: make([]STNode, 0),
	}

	return GolspObject{
		IsFunction: false,
		Function: GolspEmptyFunction(),
		Value: val,
	}
}

func GolspBuiltinMinus(scope GolspScope, args []STNode) GolspObject {
	var arguments []GolspObject
	argscope := MakeScope(scope)
	for _, a := range args {
		arguments = append(arguments, Eval(argscope, a))
	}

	for _, a := range arguments {
		if a.Value.Type != STNodeTypeNumberLiteral {
			return scope[UNDEFINED]
		}
	}

	sum := 0.0
	if len(arguments) > 0 {
		n, _ := strconv.ParseFloat(arguments[0].Value.Head, 64)
		sum += n
	}

	for _, v := range arguments[1:] {
		n, _ := strconv.ParseFloat(v.Value.Head, 64)
		sum -= n
	}

	val := STNode{
		Head: fmt.Sprintf("%v", sum),
		Type: STNodeTypeNumberLiteral,
		Children: make([]STNode, 0),
	}

	return GolspObject{
		IsFunction: false,
		Function: GolspEmptyFunction(),
		Value: val,
	}
}

func GolspBuiltinMultiply(scope GolspScope, args []STNode) GolspObject {
	var arguments []GolspObject
	argscope := MakeScope(scope)
	for _, a := range args {
		arguments = append(arguments, Eval(argscope, a))
	}

	for _, a := range arguments {
		if a.Value.Type != STNodeTypeNumberLiteral {
			return scope[UNDEFINED]
		}
	}

	product := 1.0
	for _, v := range arguments {
		n, _ := strconv.ParseFloat(v.Value.Head, 64)
		product *= n
	}

	value := STNode{
		Head: fmt.Sprintf("%v", product),
		Type: STNodeTypeNumberLiteral,
		Children: make([]STNode, 0),
	}

	return GolspObject{
		IsFunction: false,
		Function: GolspEmptyFunction(),
		Value: value,
	}
}

func GolspBuiltinDivide(scope GolspScope, args []STNode) GolspObject {
	var arguments []GolspObject
	argscope := MakeScope(scope)
	for _, a := range args {
		arguments = append(arguments, Eval(argscope, a))
	}

	for _, a := range arguments {
		if a.Value.Type != STNodeTypeNumberLiteral {
			return scope[UNDEFINED]
		}
	}

	numerator := 1.0
	if len(arguments) > 0 {
		n, _ := strconv.ParseFloat(arguments[0].Value.Head, 64)
		numerator *= n
	}

	denominator := 1.0
	for _, v := range arguments[1:] {
		n, _ := strconv.ParseFloat(v.Value.Head, 64)
		denominator *= n
	}

	val := STNode{
		Head: fmt.Sprintf("%v", numerator / denominator),
		Type: STNodeTypeNumberLiteral,
		Children: make([]STNode, 0),
	}

	return GolspObject{
		IsFunction: false,
		Function: GolspEmptyFunction(),
		Value: val,
	}
}

func GolspBuiltinPrintf(scope GolspScope, arguments []STNode) GolspObject {
	text := arguments[0].Head
	text = text[1:len(text) - 1]

	argscope := MakeScope(scope)
	var args []interface{}
	for _, a := range arguments[1:] {
		v := Eval(argscope, a)

		if v.IsFunction {
			args = append(args, fmt.Sprintf("<function:%v>", a.Head))
			continue
		}

		if v.Value.Type == STNodeTypeNumberLiteral {
			n, _ := strconv.ParseFloat(v.Value.Head, 64)
			args = append(args, n)
		} else if v.Value.Type == STNodeTypeStringLiteral {
			str := v.Value.Head[1:len(v.Value.Head) - 1]
			args = append(args, str)
		} else {
			args = append(args, "<undefined>")
		}
	}

	// TODO: replace all literal escape sequences with actual escape characters
	text = strings.Replace(text, "\\n", "\n", -1)
	text = strings.Replace(text, "\\\"", "\"", -1)

	fmt.Printf(text, args...)

	return GolspObject{
		IsFunction: false,
		Function: GolspEmptyFunction(),
		Value: STNode{
			Head: fmt.Sprintf("\"%v\"", text),
			Type: STNodeTypeStringLiteral,
			Children: make([]STNode, 0),
		},
	}
}

func GolspUndefinedIdentifier() STNode {
	return STNode{
		Head: UNDEFINED,
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

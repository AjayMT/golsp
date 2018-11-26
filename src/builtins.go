
// Builtins

package main

import (
	"fmt"
	"strings"
	"strconv"
)

var Builtins = GolspScope{
	Parent: nil,
	Identifiers: nil,
}

func InitializeBuiltins() {
	Builtins.Identifiers = map[string]GolspObject{
		UNDEFINED: GolspObject{
			Scope: GolspEmptyScope(),
			Type: GolspObjectTypeLiteral,
			Function: GolspEmptyFunction(),
			Value: GolspUndefinedIdentifier(),
		},

		"=": GolspObject{
			Scope: GolspEmptyScope(),
			Type: GolspObjectTypeFunction,
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
			Scope: GolspEmptyScope(),
			Type: GolspObjectTypeFunction,
			Function: GolspFunction{
				FunctionPatterns: make([][]STNode, 0),
				FunctionBodies: make([]STNode, 0),
				BuiltinPatterns: [][]STNode{make([]STNode, 0)},
				BuiltinBodies: []GolspBuiltinFunctionBody{GolspBuiltinPlus},
			},
			Value: GolspUndefinedIdentifier(),
		},

		"-": GolspObject{
			Scope: GolspEmptyScope(),
			Type: GolspObjectTypeFunction,
			Function: GolspFunction{
				FunctionPatterns: make([][]STNode, 0),
				FunctionBodies: make([]STNode, 0),
				BuiltinPatterns: [][]STNode{make([]STNode, 0)},
				BuiltinBodies: []GolspBuiltinFunctionBody{GolspBuiltinMinus},
			},
			Value: GolspUndefinedIdentifier(),
		},

		"*": GolspObject{
			Scope: GolspEmptyScope(),
			Type: GolspObjectTypeFunction,
			Function: GolspFunction{
				FunctionPatterns: make([][]STNode, 0),
				FunctionBodies: make([]STNode, 0),
				BuiltinPatterns: [][]STNode{make([]STNode, 0)},
				BuiltinBodies: []GolspBuiltinFunctionBody{GolspBuiltinMultiply},
			},
			Value: GolspUndefinedIdentifier(),
		},

		"/": GolspObject{
			Scope: GolspEmptyScope(),
			Type: GolspObjectTypeFunction,
			Function: GolspFunction{
				FunctionPatterns: make([][]STNode, 0),
				FunctionBodies: make([]STNode, 0),
				BuiltinPatterns: [][]STNode{make([]STNode, 0)},
				BuiltinBodies: []GolspBuiltinFunctionBody{GolspBuiltinDivide},
			},
			Value: GolspUndefinedIdentifier(),
		},

		"sprintf": GolspObject{
			Scope: GolspEmptyScope(),
			Type: GolspObjectTypeFunction,
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
				BuiltinBodies: []GolspBuiltinFunctionBody{GolspBuiltinSprintf},
			},
			Value: GolspUndefinedIdentifier(),
		},

		"printf": GolspObject{
			Scope: GolspEmptyScope(),
			Type: GolspObjectTypeFunction,
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
}

func GolspBuiltinEquals(scope GolspScope, arguments []STNode) GolspObject {
	if len(arguments) < 2 {
		return Builtins.Identifiers[UNDEFINED]
	}

	symbol := arguments[0]
	value := arguments[1]

	if symbol.Type != STNodeTypeIdentifier &&
		symbol.Type != STNodeTypeExpression {
		return Eval(scope, symbol)
	}

	if symbol.Type == STNodeTypeIdentifier {
		valuescope := MakeScope(&scope)
		scope.Identifiers[symbol.Head] = Eval(valuescope, value)
		return scope.Identifiers[symbol.Head]
	}

	pattern := make([]STNode, 0)
	head := symbol.Children[0]
	if head.Type != STNodeTypeIdentifier {
		return Builtins.Identifiers[UNDEFINED]
	}

	pattern = symbol.Children[1:]
	for i, _ := range pattern {
		patternscope := MakeScope(&scope)
		for pattern[i].Type == STNodeTypeExpression {
			obj := Eval(patternscope, pattern[i])
			if obj.Type == GolspObjectTypeLiteral {
				pattern[i] = obj.Value
			}
		}
	}

	symbol = head
	_, exists := scope.Identifiers[symbol.Head]
	if !exists {
		scope.Identifiers[symbol.Head] = GolspObject{
			Scope: MakeScope(&scope),
			Type: GolspObjectTypeFunction,
			Function: GolspEmptyFunction(),
			Value: GolspUndefinedIdentifier(),
		}
	}

	patternexists := false
	patternindex := 0
	for index, p := range scope.Identifiers[symbol.Head].Function.FunctionPatterns {
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
		scope.Identifiers[symbol.Head].Function.FunctionBodies[patternindex] = value
		return scope.Identifiers[symbol.Head]
	}

	newfn := GolspFunction{
		FunctionPatterns: append(scope.Identifiers[symbol.Head].Function.FunctionPatterns, pattern),
		FunctionBodies: append(scope.Identifiers[symbol.Head].Function.FunctionBodies, value),
		BuiltinPatterns: scope.Identifiers[symbol.Head].Function.BuiltinPatterns,
		BuiltinBodies: scope.Identifiers[symbol.Head].Function.BuiltinBodies,
	}

	scope.Identifiers[symbol.Head] = GolspObject{
		Scope: MakeScope(&scope),
		Type: GolspObjectTypeFunction,
		Function: newfn,
		Value: GolspUndefinedIdentifier(),
	}

	return scope.Identifiers[symbol.Head]
}

func GolspBuiltinPlus(scope GolspScope, args []STNode) GolspObject {
	arguments := make([]GolspObject, len(args))
	argscope := MakeScope(&scope)
	for i, a := range args {
		arguments[i] = Eval(argscope, a)
	}

	for _, a := range arguments {
		if a.Value.Type != STNodeTypeNumberLiteral {
			return Builtins.Identifiers[UNDEFINED]
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
		Scope: MakeScope(&scope),
		Type: GolspObjectTypeLiteral,
		Function: GolspEmptyFunction(),
		Value: val,
	}
}

func GolspBuiltinMinus(scope GolspScope, args []STNode) GolspObject {
	arguments := make([]GolspObject, len(args))
	argscope := MakeScope(&scope)
	for i, a := range args {
		arguments[i] = Eval(argscope, a)
	}

	for _, a := range arguments {
		if a.Value.Type != STNodeTypeNumberLiteral {
			return Builtins.Identifiers[UNDEFINED]
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
		Scope: MakeScope(&scope),
		Type: GolspObjectTypeLiteral,
		Function: GolspEmptyFunction(),
		Value: val,
	}
}

func GolspBuiltinMultiply(scope GolspScope, args []STNode) GolspObject {
	arguments := make([]GolspObject, len(args))
	argscope := MakeScope(&scope)
	for i, a := range args {
		arguments[i] = Eval(argscope, a)
	}

	for _, a := range arguments {
		if a.Value.Type != STNodeTypeNumberLiteral {
			return Builtins.Identifiers[UNDEFINED]
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
		Scope: MakeScope(&scope),
		Type: GolspObjectTypeLiteral,
		Function: GolspEmptyFunction(),
		Value: value,
	}
}

func GolspBuiltinDivide(scope GolspScope, args []STNode) GolspObject {
	arguments := make([]GolspObject, len(args))
	argscope := MakeScope(&scope)
	for i, a := range args {
		arguments[i] = Eval(argscope, a)
	}

	for _, a := range arguments {
		if a.Value.Type != STNodeTypeNumberLiteral {
			return Builtins.Identifiers[UNDEFINED]
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
		Scope: MakeScope(&scope),
		Type: GolspObjectTypeLiteral,
		Function: GolspEmptyFunction(),
		Value: val,
	}
}

func formatStr(text string, objects []GolspObject) string {
	args := make([]interface{}, len(objects))
	for i, v := range objects {
		if v.Type == GolspObjectTypeFunction {
			args[i] = "<function>"
			continue
		}

		if v.Type == GolspObjectTypeList {
			args[i] = fmt.Sprintf("{%v}",
				formatStr(strings.Repeat("%v ", len(v.Elements)), v.Elements))
			continue
		}

		if v.Value.Type == STNodeTypeNumberLiteral {
			n, _ := strconv.ParseFloat(v.Value.Head, 64)
			args[i] = n
		} else if v.Value.Type == STNodeTypeStringLiteral {
			str := v.Value.Head[1:len(v.Value.Head) - 1]
			args[i] = str
		} else {
			args[i] = "<undefined>"
		}
	}

	// TODO: replace all literal escape sequences with actual escape characters
	text = strings.Replace(text, "\\n", "\n", -1)
	text = strings.Replace(text, "\\\"", "\"", -1)

	return fmt.Sprintf(text, args...)
}

func GolspBuiltinSprintf(scope GolspScope, arguments []STNode) GolspObject {
	text := arguments[0].Head
	text = text[1:len(text) - 1]

	argscope := MakeScope(&scope)
	objects := make([]GolspObject, len(arguments) - 1)
	for i, a := range arguments[1:] {
		objects[i] = Eval(argscope, a)
	}

	return GolspObject{
		Scope: MakeScope(&scope),
		Type: GolspObjectTypeLiteral,
		Function: GolspEmptyFunction(),
		Value: STNode{
			Head: fmt.Sprintf("\"%v\"", formatStr(text, objects)),
			Type: STNodeTypeStringLiteral,
			Children: make([]STNode, 0),
		},
	}
}

func GolspBuiltinPrintf(scope GolspScope, arguments []STNode) GolspObject {
	obj := GolspBuiltinSprintf(scope, arguments)
	fmt.Printf(obj.Value.Head[1:len(obj.Value.Head) - 1])

	return obj
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

func GolspEmptyScope() GolspScope {
	return GolspScope{
		Parent: nil,
		Identifiers: nil,
	}
}

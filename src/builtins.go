
// Builtins

package main

import (
	"fmt"
	"strings"
	"strconv"
	"time"
)

var Builtins = GolspScope{}

func InitializeBuiltins() {
	Builtins.Identifiers = map[string]GolspObject{
		UNDEFINED: GolspObject{
			Type: GolspObjectTypeLiteral,
			Value: GolspUndefinedIdentifier(),
		},

		"=": GolspObject{
			Type: GolspObjectTypeFunction,
			Function: GolspFunction{
				BuiltinPatterns: [][]STNode{
					[]STNode{
						STNode{
							Head: "a",
							Type: STNodeTypeIdentifier,
						},
						STNode{
							Head: "b",
							Type: STNodeTypeIdentifier,
						},
					},
				},
				BuiltinBodies: []GolspBuiltinFunctionBody{GolspBuiltinEquals},
			},
		},

		"+": GolspObject{
			Type: GolspObjectTypeFunction,
			Function: GolspFunction{
				BuiltinPatterns: [][]STNode{make([]STNode, 0)},
				BuiltinBodies: []GolspBuiltinFunctionBody{GolspBuiltinPlus},
			},
		},

		"-": GolspObject{
			Type: GolspObjectTypeFunction,
			Function: GolspFunction{
				BuiltinPatterns: [][]STNode{make([]STNode, 0)},
				BuiltinBodies: []GolspBuiltinFunctionBody{GolspBuiltinMinus},
			},
		},

		"*": GolspObject{
			Type: GolspObjectTypeFunction,
			Function: GolspFunction{
				BuiltinPatterns: [][]STNode{make([]STNode, 0)},
				BuiltinBodies: []GolspBuiltinFunctionBody{GolspBuiltinMultiply},
			},
		},

		"/": GolspObject{
			Type: GolspObjectTypeFunction,
			Function: GolspFunction{
				BuiltinPatterns: [][]STNode{make([]STNode, 0)},
				BuiltinBodies: []GolspBuiltinFunctionBody{GolspBuiltinDivide},
			},
		},

		"==": GolspBuiltinComparisonFunction("=="),
		"!=": GolspBuiltinComparisonFunction("!="),
		">": GolspBuiltinComparisonFunction(">"),
		"<": GolspBuiltinComparisonFunction("<"),
		">=": GolspBuiltinComparisonFunction(">="),
		"<=": GolspBuiltinComparisonFunction("<="),

		"if": GolspObject{
			Type: GolspObjectTypeFunction,
			Function: GolspFunction{
				BuiltinPatterns: [][]STNode{
					[]STNode{
						STNode{
							Head: "cond",
							Type: STNodeTypeIdentifier,
						},
						STNode{
							Head: "expr",
							Type: STNodeTypeIdentifier,
						},
					},
				},
				BuiltinBodies: []GolspBuiltinFunctionBody{GolspBuiltinIf},
			},
		},

		"do": GolspObject{
			Type: GolspObjectTypeFunction,
			Function: GolspFunction{
				BuiltinPatterns: [][]STNode{make([]STNode, 0)},
				BuiltinBodies: []GolspBuiltinFunctionBody{GolspBuiltinDo},
			},
		},

		"go": GolspObject{
			Type: GolspObjectTypeFunction,
			Function: GolspFunction{
				BuiltinPatterns: [][]STNode{make([]STNode, 0)},
				BuiltinBodies: []GolspBuiltinFunctionBody{GolspBuiltinGo},
			},
		},

		"sleep": GolspObject{
			Type: GolspObjectTypeFunction,
			Function: GolspFunction{
				BuiltinPatterns: [][]STNode{make([]STNode, 0)},
				BuiltinBodies: []GolspBuiltinFunctionBody{GolspBuiltinSleep},
			},
		},

		"sprintf": GolspObject{
			Type: GolspObjectTypeFunction,
			Function: GolspFunction{
				BuiltinPatterns: [][]STNode{
					[]STNode{
						STNode{
							Head: "s",
							Type: STNodeTypeIdentifier,
						},
					},
				},
				BuiltinBodies: []GolspBuiltinFunctionBody{GolspBuiltinSprintf},
			},
		},

		"printf": GolspObject{
			Type: GolspObjectTypeFunction,
			Function: GolspFunction{
				BuiltinPatterns: [][]STNode{
					[]STNode{
						STNode{
							Head: "s",
							Type: STNodeTypeIdentifier,
						},
					},
				},
				BuiltinBodies: []GolspBuiltinFunctionBody{GolspBuiltinPrintf},
			},
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
		newscope := MakeScope(&scope)
		scope.Identifiers[symbol.Head] = GolspObject{
			Scope: newscope,
			Type: GolspObjectTypeFunction,
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
	}

	return GolspObject{
		Scope: MakeScope(&scope),
		Type: GolspObjectTypeLiteral,
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
	}

	return GolspObject{
		Scope: MakeScope(&scope),
		Type: GolspObjectTypeLiteral,
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
	}

	return GolspObject{
		Scope: MakeScope(&scope),
		Type: GolspObjectTypeLiteral,
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
	}

	return GolspObject{
		Scope: MakeScope(&scope),
		Type: GolspObjectTypeLiteral,
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
		Value: STNode{
			Head: fmt.Sprintf("\"%v\"", formatStr(text, objects)),
			Type: STNodeTypeStringLiteral,
		},
	}
}

func GolspBuiltinPrintf(scope GolspScope, arguments []STNode) GolspObject {
	obj := GolspBuiltinSprintf(scope, arguments)
	fmt.Printf(obj.Value.Head[1:len(obj.Value.Head) - 1])

	return obj
}

func GolspBuiltinDo(scope GolspScope, arguments []STNode) GolspObject {
	scopenode := STNode{
		Type: STNodeTypeScope,
		Children: arguments,
	}

	return Eval(scope, scopenode)
}

func GolspBuiltinGo(scope GolspScope, arguments []STNode) GolspObject {
	go GolspBuiltinDo(scope, arguments)

	return Builtins.Identifiers[UNDEFINED]
}

func GolspBuiltinSleep(scope GolspScope, arguments []STNode) GolspObject {
	argscope := MakeScope(&scope)
	argobjects := make([]GolspObject, len(arguments))
	for i, a := range arguments {
		argobjects[i] = Eval(argscope, a)
	}

	if argobjects[0].Type != GolspObjectTypeLiteral ||
		argobjects[0].Value.Type != STNodeTypeNumberLiteral {
		return Builtins.Identifiers[UNDEFINED]
	}

	duration, _ := strconv.ParseFloat(argobjects[0].Value.Head, 64)
	time.Sleep(time.Duration(duration) * time.Millisecond)

	return Builtins.Identifiers[UNDEFINED]
}

func GolspBuiltinIf(scope GolspScope, arguments []STNode) GolspObject {
	if len(arguments) < 2 {
		return Builtins.Identifiers[UNDEFINED]
	}

	argscope := MakeScope(&scope)
	condObj := Eval(argscope, arguments[0])
	cond := false

	if condObj.Type == GolspObjectTypeFunction { cond = true }
	if condObj.Type == GolspObjectTypeList { cond = len(condObj.Elements) > 0 }
	if condObj.Type == GolspObjectTypeLiteral {
		if condObj.Value.Type == STNodeTypeStringLiteral {
			cond = len(condObj.Value.Head) > 2
		}

		if condObj.Value.Type == STNodeTypeNumberLiteral {
			n, _ := strconv.ParseFloat(condObj.Value.Head, 64)
			cond = n != 0
		}

		if condObj.Value.Head == UNDEFINED { cond = false }
	}

	if cond { return Eval(argscope, arguments[1]) }

	if len(arguments) > 2 {
		return Eval(argscope, arguments[2])
	}

	return Builtins.Identifiers[UNDEFINED]
}

func GolspBuiltinComparisonFunction(op string) GolspObject {
	fn := func (scope GolspScope, args []STNode) GolspObject {
		argscope := MakeScope(&scope)
		arguments := make([]GolspObject, len(args))
		for i, c := range args {
			arguments[i] = Eval(argscope, c)
		}

		if len(arguments) != 2 {
			return Builtins.Identifiers[UNDEFINED]
		}

		// TODO handle lists?

		if arguments[0].Type != GolspObjectTypeLiteral ||
			arguments[1].Type != GolspObjectTypeLiteral {
			return GolspObject{
				Scope: MakeScope(&scope),
				Type: GolspObjectTypeLiteral,
				Value: STNode{
					Head: "0",
					Type: STNodeTypeNumberLiteral,
				},
			}
		}

		if arguments[0].Value.Head == UNDEFINED ||
			arguments[1].Value.Head == UNDEFINED {
			result := arguments[0].Value.Head == UNDEFINED &&
				arguments[1].Value.Head == UNDEFINED &&
				strings.Contains(op, "=")

			resultint := 0
			if result { resultint = 1 }

			return GolspObject{
				Scope: MakeScope(&scope),
				Type: GolspObjectTypeLiteral,
				Value: STNode{
					Head: strconv.Itoa(resultint),
					Type: STNodeTypeNumberLiteral,
				},
			}
		}

		argtype := arguments[0].Value.Type
		if arguments[1].Value.Type != argtype {
			return Builtins.Identifiers[UNDEFINED]
		}

		var value1 interface{}
		var value2 interface{}

		if argtype == STNodeTypeStringLiteral {
			value1 = arguments[0].Value.Head[1:len(arguments[0].Value.Head) - 1]
			value2 = arguments[1].Value.Head[1:len(arguments[1].Value.Head) - 1]
		} else {
			value1, _ = strconv.ParseFloat(arguments[0].Value.Head, 64)
			value2, _ = strconv.ParseFloat(arguments[1].Value.Head, 64)
		}

		resultint := 0
		result := false
		switch op {
		case "==":
			result = value1 == value2
		case "!=":
			result = value1 != value2
		case ">":
			if argtype == STNodeTypeStringLiteral {
				result = value1.(string) > value2.(string)
			} else { result = value1.(float64) > value2.(float64) }
		case "<":
			if argtype == STNodeTypeStringLiteral {
				result = value1.(string) < value2.(string)
			} else { result = value1.(float64) < value2.(float64) }
		case ">=":
			if argtype == STNodeTypeStringLiteral {
				result = value1.(string) >= value2.(string)
			} else { result = value1.(float64) >= value2.(float64) }
		case "<=":
			if argtype == STNodeTypeStringLiteral {
				result = value1.(string) <= value2.(string)
			} else { result = value1.(float64) <= value2.(float64) }
		}

		if result { resultint = 1 }

		return GolspObject{
			Scope: MakeScope(&scope),
			Type: GolspObjectTypeLiteral,
			Value: STNode{
				Head: strconv.Itoa(resultint),
				Type: STNodeTypeNumberLiteral,
			},
		}
	}

	return GolspObject{
		Type: GolspObjectTypeFunction,
		Function: GolspFunction{
			BuiltinPatterns: [][]STNode{
				[]STNode{
					STNode{
						Head: "a",
						Type: STNodeTypeIdentifier,
					},
					STNode{
						Head: "b",
						Type: STNodeTypeIdentifier,
					},
				},
			},
			BuiltinBodies: []GolspBuiltinFunctionBody{fn},
		},
	}
}

func GolspUndefinedIdentifier() STNode {
	return STNode{
		Head: UNDEFINED,
		Type: STNodeTypeIdentifier,
	}
}

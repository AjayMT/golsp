
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

		"=": GolspBuiltinFunctionObject(GolspBuiltinEquals),
		"lambda": GolspBuiltinFunctionObject(GolspBuiltinLambda),

		"+": GolspBuiltinMathFunction("+"),
		"-": GolspBuiltinMathFunction("-"),
		"*": GolspBuiltinMathFunction("*"),
		"/": GolspBuiltinMathFunction("/"),
		"%": GolspBuiltinMathFunction("%"),

		"==": GolspBuiltinComparisonFunction("=="),
		"!=": GolspBuiltinComparisonFunction("!="),
		">": GolspBuiltinComparisonFunction(">"),
		"<": GolspBuiltinComparisonFunction("<"),
		">=": GolspBuiltinComparisonFunction(">="),
		"<=": GolspBuiltinComparisonFunction("<="),

		"if": GolspBuiltinFunctionObject(GolspBuiltinIf),
		"do": GolspBuiltinFunctionObject(GolspBuiltinDo),
		"go": GolspBuiltinFunctionObject(GolspBuiltinGo),
		"sleep": GolspBuiltinFunctionObject(GolspBuiltinSleep),
		"sprintf": GolspBuiltinFunctionObject(GolspBuiltinSprintf),
		"printf": GolspBuiltinFunctionObject(GolspBuiltinPrintf),
	}
}

func comparePatterns(pattern1 []STNode, pattern2 []STNode) bool {
	for i, node1 := range pattern1 {
		if i >= len(pattern2) { return false }

		node2 := pattern2[i]

		if node1.Type != node2.Type { return false }

		if node1.Type == STNodeTypeStringLiteral ||
			node1.Type == STNodeTypeNumberLiteral {
			if node1.Head != node2.Head { return false }
		}

		if node1.Spread != node2.Spread { return false }

		if node1.Type == STNodeTypeList {
			if !comparePatterns(node1.Children, node2.Children) {
				return false
			}
		}
	}

	return true
}

func GolspBuiltinEquals(scope GolspScope, arguments []GolspObject) GolspObject {
	if len(arguments) < 2 {
		return Builtins.Identifiers[UNDEFINED]
	}

	if arguments[0].Type != GolspObjectTypeBuiltinArgument ||
		arguments[1].Type != GolspObjectTypeBuiltinArgument {
		return Builtins.Identifiers[UNDEFINED]
	}

	symbol := arguments[0].Value
	value := arguments[1].Value

	if symbol.Type != STNodeTypeIdentifier &&
		symbol.Type != STNodeTypeExpression {
		return Eval(scope, symbol)
	}

	if symbol.Type == STNodeTypeIdentifier {
		_, builtin := Builtins.Identifiers[symbol.Head]
		if builtin {
			return Builtins.Identifiers[UNDEFINED]
		}

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
	_, builtin := Builtins.Identifiers[symbol.Head]
	if builtin {
		return Builtins.Identifiers[UNDEFINED]
	}

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
		if comparePatterns(pattern, p) {
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
	}

	scope.Identifiers[symbol.Head] = GolspObject{
		Scope: MakeScope(&scope),
		Type: GolspObjectTypeFunction,
		Function: newfn,
	}

	return scope.Identifiers[symbol.Head]
}

func GolspBuiltinLambda(scope GolspScope, arguments []GolspObject) GolspObject {
	if len(arguments) < 2 {
		return Builtins.Identifiers[UNDEFINED]
	}

	for _, arg := range arguments {
		if arg.Type != GolspObjectTypeBuiltinArgument {
			return Builtins.Identifiers[UNDEFINED]
		}
	}

	if arguments[0].Value.Type != STNodeTypeExpression {
		return Builtins.Identifiers[UNDEFINED]
	}

	pattern := arguments[0].Value.Children
	body := arguments[1].Value

	for i, _ := range pattern {
		patternscope := MakeScope(&scope)
		for pattern[i].Type == STNodeTypeExpression {
			obj := Eval(patternscope, pattern[i])
			if obj.Type == GolspObjectTypeLiteral {
				pattern[i] = obj.Value
			}
		}
	}

	fn := GolspFunction{
		FunctionPatterns: [][]STNode{pattern},
		FunctionBodies: []STNode{body},
	}

	return GolspObject{
		Scope: MakeScope(&scope),
		Type: GolspObjectTypeFunction,
		Function: fn,
	}
}

func GolspBuiltinMathFunction(op string) GolspObject {
	fn := func (scope GolspScope, args []GolspObject) GolspObject {
		arguments := evalArgs(scope, args)
		for _, a := range arguments {
			if a.Value.Type != STNodeTypeNumberLiteral {
				return Builtins.Identifiers[UNDEFINED]
			}
		}

		result := 0.0

		switch op {
		case "+":
			for _, v := range arguments {
				n, _ := strconv.ParseFloat(v.Value.Head, 64)
				result += n
			}
		case "-":
			if len(arguments) > 0 {
				n, _ := strconv.ParseFloat(arguments[0].Value.Head, 64)
				result += n
			}

			for _, v := range arguments[1:] {
				n, _ := strconv.ParseFloat(v.Value.Head, 64)
				result -= n
			}
		case "*":
			result = 1.0
			for _, v := range arguments {
				n, _ := strconv.ParseFloat(v.Value.Head, 64)
				result *= n
			}
		case "/":
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

			result = numerator / denominator
		case "%":
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

			result = float64(int(numerator) % int(denominator))
		}

		return GolspObject{
			Type: GolspObjectTypeLiteral,
			Value: STNode{
				Head: fmt.Sprintf("%v", result),
				Type: STNodeTypeNumberLiteral,
			},
		}
	}

	return GolspBuiltinFunctionObject(fn)
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

func GolspBuiltinSprintf(scope GolspScope, args []GolspObject) GolspObject {
	arguments := evalArgs(scope, args)

	if arguments[0].Value.Type != STNodeTypeStringLiteral {
		return Builtins.Identifiers[UNDEFINED]
	}

	text := arguments[0].Value.Head
	text = text[1:len(text) - 1]

	return GolspObject{
		Type: GolspObjectTypeLiteral,
		Value: STNode{
			Head: fmt.Sprintf("\"%v\"", formatStr(text, arguments[1:])),
			Type: STNodeTypeStringLiteral,
		},
	}
}

func GolspBuiltinPrintf(scope GolspScope, arguments []GolspObject) GolspObject {
	obj := GolspBuiltinSprintf(scope, arguments)
	if obj.Value.Head != UNDEFINED {
		fmt.Printf(obj.Value.Head[1:len(obj.Value.Head) - 1])
	}

	return obj
}

func GolspBuiltinDo(scope GolspScope, arguments []GolspObject) GolspObject {
	for _, a := range arguments {
		if a.Type != GolspObjectTypeBuiltinArgument {
			return Builtins.Identifiers[UNDEFINED]
		}
	}

	args := make([]STNode, len(arguments))

	for i, c := range arguments {
		args[i] = c.Value
	}

	scopenode := STNode{
		Type: STNodeTypeScope,
		Children: args,
	}

	return Eval(scope, scopenode)
}

func GolspBuiltinGo(scope GolspScope, arguments []GolspObject) GolspObject {
	go GolspBuiltinDo(scope, arguments)

	return Builtins.Identifiers[UNDEFINED]
}

func GolspBuiltinSleep(scope GolspScope, arguments []GolspObject) GolspObject {
	argobjects := evalArgs(scope, arguments)

	if argobjects[0].Type != GolspObjectTypeLiteral ||
		argobjects[0].Value.Type != STNodeTypeNumberLiteral {
		return Builtins.Identifiers[UNDEFINED]
	}

	duration, _ := strconv.ParseFloat(argobjects[0].Value.Head, 64)
	time.Sleep(time.Duration(duration) * time.Millisecond)

	return Builtins.Identifiers[UNDEFINED]
}

func GolspBuiltinIf(scope GolspScope, args []GolspObject) GolspObject {
	arguments := []GolspObject{Builtins.Identifiers[UNDEFINED]}

	if len(args) == 0 { return Builtins.Identifiers[UNDEFINED] }

	if args[0].Type == GolspObjectTypeBuiltinArgument {
		argscope := MakeScope(&scope)
		if args[0].Value.Spread {
			spread := SpreadNode(argscope, args[0].Value)
			arguments = spread
		} else {
			arguments[0] = Eval(argscope, args[0].Value)
		}
	} else { arguments[0] = args[0] }

	condObj := arguments[0]
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

	if cond {
		if len(arguments) > 1 { return arguments[1] }
		if len(args) > 1 { return evalArgs(scope, args[1:2])[0] }
	}

	if len(arguments) > 2 { return arguments[2] }
	if len(args) > 2 { return evalArgs(scope, args[2:3])[0] }

	return Builtins.Identifiers[UNDEFINED]
}

func GolspBuiltinComparisonFunction(op string) GolspObject {
	fn := func (scope GolspScope, args []GolspObject) GolspObject {
		arguments := evalArgs(scope, args)

		if len(arguments) != 2 {
			return Builtins.Identifiers[UNDEFINED]
		}

		// TODO handle lists?

		if arguments[0].Type != GolspObjectTypeLiteral ||
			arguments[1].Type != GolspObjectTypeLiteral {
			return GolspObject{
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
			Type: GolspObjectTypeLiteral,
			Value: STNode{
				Head: strconv.Itoa(resultint),
				Type: STNodeTypeNumberLiteral,
			},
		}
	}

	return GolspBuiltinFunctionObject(fn)
}

func GolspBuiltinFunctionObject(fn GolspBuiltinFunction) GolspObject {
	return GolspObject{
		Type: GolspObjectTypeFunction,
		Function: GolspFunction{BuiltinFunc: fn},
	}
}

func evalArgs(scp GolspScope, args []GolspObject) []GolspObject {
	scope := MakeScope(&scp)
	var arguments []GolspObject
	for _, child := range args {
		if child.Type == GolspObjectTypeBuiltinArgument {
			node := child.Value
			if node.Spread {
				spread := SpreadNode(scope, node)
				arguments = append(arguments, spread...)
			} else {
				arguments = append(arguments, Eval(scope, node))
			}
		} else {
			arguments = append(arguments, child)
		}
	}

	return arguments
}

func GolspUndefinedIdentifier() STNode {
	return STNode{
		Head: UNDEFINED,
		Type: STNodeTypeIdentifier,
	}
}

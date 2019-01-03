
// Builtins

package golsp

import (
	"fmt"
	"strings"
	"strconv"
	"time"
	"sync"
	"path/filepath"
	"io/ioutil"
	"os"
	"plugin"
)

var Builtins = Scope{}
var WaitGroup sync.WaitGroup

// InitializeBuiltins: Initialize the default builtin scope ('Builtins')
// with builtin identifiers
func InitializeBuiltins(dirname string, filename string, args []string) {
	identifiers := map[string]Object{
		UNDEFINED: UndefinedObject(),
		DIRNAME: StringObject(dirname),
		FILENAME: StringObject(filename),
		ARGS: ListObject(args),

		"def": BuiltinFunctionObject("def", BuiltinDef),
		"const": BuiltinFunctionObject("const", BuiltinConst),
		"lambda": BuiltinFunctionObject("lambda", BuiltinLambda),
		"require": BuiltinFunctionObject("require", BuiltinRequire),
		"if": BuiltinFunctionObject("if", BuiltinIf),
		"when": BuiltinFunctionObject("when", BuiltinWhen),
		"do": BuiltinFunctionObject("do", BuiltinDo),
		"go": BuiltinFunctionObject("go", BuiltinGo),
		"sleep": BuiltinFunctionObject("sleep", BuiltinSleep),
		"sprintf": BuiltinFunctionObject("sprintf", BuiltinSprintf),
		"printf": BuiltinFunctionObject("printf", BuiltinPrintf),

		"+": BuiltinMathFunction("+"),
		"-": BuiltinMathFunction("-"),
		"*": BuiltinMathFunction("*"),
		"/": BuiltinMathFunction("/"),
		"%": BuiltinMathFunction("%"),

		"==": BuiltinComparisonFunction("=="),
		"!=": BuiltinComparisonFunction("!="),
		">": BuiltinComparisonFunction(">"),
		"<": BuiltinComparisonFunction("<"),
		">=": BuiltinComparisonFunction(">="),
		"<=": BuiltinComparisonFunction("<="),
	}

	Builtins.Identifiers = identifiers
	Builtins.Constants = make(map[string]bool)
	for k, _ := range identifiers { Builtins.Constants[k] = true }
}

// comparePatterns: Compare two function potterns (as passed to the '=' function)
// to check whether they are identical. This function is used to check for and
// redefine existing function patterns
// `pattern1`: the first pattern
// `pattern2`: the second pattern
// this function returns whether the two patterns are identical
func comparePatterns(pattern1 []STNode, pattern2 []STNode) bool {
	if len(pattern1) != len(pattern2) { return false }

	for i, node1 := range pattern1 {
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
		if node1.Type == STNodeTypeMap {
			if !comparePatterns(node1.Children, node2.Children) {
				return false
			}

			zip1 := make([]STNode, len(node1.Children))
			zip2 := make([]STNode, len(node2.Children))
			for j, z := range node1.Children { zip1[j] = *z.Zip }
			for j, z := range node2.Children { zip2[j] = *z.Zip }

			if !comparePatterns(zip1, zip2) { return false }
		}
	}

	return true
}

// isConstant: Check whether an identifier is constant within a given scope
// and its parents
// `scope`: the scope
// `identifier`: the identifier
// this function returns whether the identifer is a constant
func isConstant(scope Scope, identifier string) bool {
	constant, exists := scope.Constants[identifier]
	if exists { return constant }
	if scope.Parent != nil { return isConstant(*scope.Parent, identifier) }

	return false
}

// see 'assign'
func BuiltinDef(scope Scope, arguments []Object) Object {
	return assign(scope, arguments, false)
}
func BuiltinConst(scope Scope, arguments []Object) Object {
	return assign(scope, arguments, true)
}

// assign: The builtin 'def' and 'const' functions. These functions (re)bind identifiers
// to objects and function patterns to expressions. They only act within their immediate
// scope and do not cause side-effects elsewhere
// `scope`: the scope within which the function is evaluated
// `arguments`: the arguments passed to the function
// `constant`: whether the identifier is being set as a constant
// this function returns a result object -- for '=', this is the value that the
// identifier or pattern was bound to
func assign(scope Scope, arguments []Object, constant bool) Object {
	if len(arguments) < 2 {
		return Builtins.Identifiers[UNDEFINED]
	}

	// as of now, '=' does not take spread expressions as arguments
	if arguments[0].Type != ObjectTypeBuiltinArgument ||
		arguments[1].Type != ObjectTypeBuiltinArgument {
		return Builtins.Identifiers[UNDEFINED]
	}

	symbol := arguments[0].Value
	value := arguments[1].Value

	// attempting to assign to a literal or list fails
	if symbol.Type != STNodeTypeIdentifier &&
		symbol.Type != STNodeTypeExpression {
		return Builtins.Identifiers[UNDEFINED]
	}

	if symbol.Type == STNodeTypeIdentifier {
		// attempting to assign to a constant identifier fails
		if isConstant(scope, symbol.Head) {
			return Builtins.Identifiers[UNDEFINED]
		}

		// if the symbol is an identifier, the value is evaluated immediately
		// and symbol is bound to it
		obj := Eval(MakeScope(&scope), value)
		if obj.Type == ObjectTypeFunction { obj.Function.Name = symbol.Head }
		scope.Identifiers[symbol.Head] = obj
		if constant { scope.Constants[symbol.Head] = true }
		return scope.Identifiers[symbol.Head]
	}

	// at this point the symbol must be an expression, i.e '[functionName pattern...]'

	head := symbol.Children[0]
	if head.Type != STNodeTypeIdentifier {
		return Builtins.Identifiers[UNDEFINED]
	}

	pattern := symbol.Children[1:]
	for i, _ := range pattern {
		patternscope := MakeScope(&scope)
		for pattern[i].Type == STNodeTypeExpression {
			obj := Eval(patternscope, pattern[i])
			if obj.Type == ObjectTypeLiteral {
				pattern[i] = obj.Value
			}
		}
	}

	symbol = head
	if isConstant(scope, symbol.Head) { return Builtins.Identifiers[UNDEFINED] }

	_, exists := scope.Identifiers[symbol.Head]
	if !exists {
		newscope := MakeScope(&scope)
		scope.Identifiers[symbol.Head] = Object{
			Scope: newscope,
			Type: ObjectTypeFunction,
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
		if constant { scope.Constants[symbol.Head] = true }
		return scope.Identifiers[symbol.Head]
	}

	newfn := Function{
		Name: symbol.Head,
		FunctionPatterns: append(scope.Identifiers[symbol.Head].Function.FunctionPatterns, pattern),
		FunctionBodies: append(scope.Identifiers[symbol.Head].Function.FunctionBodies, value),
	}

	scope.Identifiers[symbol.Head] = Object{
		Scope: MakeScope(&scope),
		Type: ObjectTypeFunction,
		Function: newfn,
	}

	if constant { scope.Constants[symbol.Head] = true }
	return scope.Identifiers[symbol.Head]
}

// BuiltinLambda: The builtin 'lambda' function. This produces a function-type
// object with one pattern and one expression
// this function returns the function object that is produced
func BuiltinLambda(scope Scope, arguments []Object) Object {
	if len(arguments) < 2 {
		return Builtins.Identifiers[UNDEFINED]
	}

	// as of now, 'lambda' does not take spread expressions as arguments
	if arguments[0].Type != ObjectTypeBuiltinArgument ||
		arguments[1].Type != ObjectTypeBuiltinArgument {
		return Builtins.Identifiers[UNDEFINED]
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
			if obj.Type == ObjectTypeLiteral {
				pattern[i] = obj.Value
			}
		}
	}

	fn := Function{
		FunctionPatterns: [][]STNode{pattern},
		FunctionBodies: []STNode{body},
	}

	return Object{
		Scope: MakeScope(&scope),
		Type: ObjectTypeFunction,
		Function: fn,
	}
}

// BuiltinRequire: The builtin 'require' function. This function evaluates a
// file and returns the Object that the file exports.
func BuiltinRequire(scope Scope, args []Object) Object {
	arguments := EvalArgs(scope, args)

	if len(arguments) < 1 {
		return Builtins.Identifiers[UNDEFINED]
	}
	if arguments[0].Value.Type != STNodeTypeStringLiteral {
		return Builtins.Identifiers[UNDEFINED]
	}

	dirnode := LookupIdentifier(scope, DIRNAME)
	dirname := dirnode.Value.Head[1:len(dirnode.Value.Head) - 1]
	rawpath := arguments[0].Value.Head[1:len(arguments[0].Value.Head) - 1]
	if strings.HasPrefix(rawpath, "stdlib/") {
		// TODO find a better way to do this
		dirname = os.Getenv("GOLSPPATH")
	}

	resolvedpath := filepath.Join(dirname, rawpath)

	if strings.HasSuffix(resolvedpath, ".so") {
		plug, err := plugin.Open(resolvedpath)
		if err != nil { return Builtins.Identifiers[UNDEFINED] }
		exportssym, err := plug.Lookup("Exports")
		if err != nil { return Builtins.Identifiers[UNDEFINED] }
		return *exportssym.(*Object)
	}

	file, err := os.Open(resolvedpath)
	if err != nil { return Builtins.Identifiers[UNDEFINED] }
	data, err := ioutil.ReadAll(file)
	if err != nil { return Builtins.Identifiers[UNDEFINED] }

	return Run(filepath.Dir(resolvedpath), resolvedpath, []string{}, string(data))
}

// BuiltinMathFunction: Produce a builtin function for a given math operator
// `op`: the math operator, one of + - * / %
// this function returns a Object containing the builtin function for the math operator
func BuiltinMathFunction(op string) Object {
	// fn: the produced math function that performs an operation specified by `op`
	// this function returns the result of the math operation
	fn := func (scope Scope, args []Object) Object {
		// math operations are undefined on non-numbers
		arguments := EvalArgs(scope, args)
		for _, a := range arguments {
			if a.Value.Type != STNodeTypeNumberLiteral {
				return Builtins.Identifiers[UNDEFINED]
			}
		}

		result := 0.0
		numbers := make([]float64, len(arguments))
		for i, arg := range arguments {
			numbers[i], _ = strconv.ParseFloat(arg.Value.Head, 64)
		}

		switch op {
		case "+":
			for _, n := range numbers { result += n }
		case "-":
			if len(numbers) > 0 { result += numbers[0] }
			for _, n := range numbers[1:] { result -= n }
		case "*":
			result = 1.0
			for _, n := range numbers { result *= n }
		case "/":
			numerator := 1.0
			if len(numbers) > 0 { numerator *= numbers[0] }
			denominator := 1.0
			for _, n := range numbers[1:] { denominator *= n }
			result = numerator / denominator
		case "%":
			numerator := 1.0
			if len(numbers) > 0 { numerator *= numbers[0] }
			denominator := 1.0
			for _, n := range numbers[1:] { denominator *= n }
			result = float64(int(numerator) % int(denominator))
		}

		return NumberObject(result)
	}

	return BuiltinFunctionObject(op, fn)
}

// formatStr: Format a Go-style format string with a set of Object arguments
// `text`: the format string
// `objects`: the objects to serialize into the string
// this function returns the formatted string
func formatStr(text string, objects []Object) string {
	args := make([]interface{}, len(objects))
	for i, v := range objects {
		if v.Type == ObjectTypeFunction {
			args[i] = fmt.Sprintf("<function:%v>", v.Function.Name)
			continue
		}

		if v.Type == ObjectTypeList {
			args[i] = fmt.Sprintf("{%v}",
				formatStr(strings.Repeat("%v ", len(v.Elements)), v.Elements))
			continue
		}

		if v.Type == ObjectTypeMap {
			strs := make([]string, 0, len(v.MapKeys))
			for _, key := range v.MapKeys {
				str := fmt.Sprintf("%v: %v", key.Value.Head,
					formatStr("%v", []Object{v.Map[key.Value.Head]}))
				strs = append(strs, str)
			}
			args[i] = fmt.Sprintf("map(%v)", strings.Join(strs, ", "))
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

// BuiltinSprintf: The builtin 'sprintf' function. This function formats a
// Go-style format string with a set of arguments
// this function returns the formatted string
func BuiltinSprintf(scope Scope, args []Object) Object {
	arguments := EvalArgs(scope, args)

	if arguments[0].Value.Type != STNodeTypeStringLiteral {
		return Builtins.Identifiers[UNDEFINED]
	}

	text := arguments[0].Value.Head
	text = text[1:len(text) - 1]

	return StringObject(formatStr(text, arguments[1:]))
}

// BuiltinPrintf: The builtin 'printf' function. This function formats
// a Go-style format string with a set of arguments and writes the result to
// stdout
// this function returns the formatted string
func BuiltinPrintf(scope Scope, arguments []Object) Object {
	obj := BuiltinSprintf(scope, arguments)
	if obj.Value.Head != UNDEFINED {
		fmt.Printf(obj.Value.Head[1:len(obj.Value.Head) - 1])
	}

	return obj
}

// BuiltinDo: The builtin 'do' function. This function evaluates a series of
// statements within an enclosed, isolated scope
// this function returns the result of evaluating the final statement
// in the scope
func BuiltinDo(scope Scope, arguments []Object) Object {
	// no support for spread arguments yet
	for _, a := range arguments {
		if a.Type != ObjectTypeBuiltinArgument {
			return Builtins.Identifiers[UNDEFINED]
		}
	}

	args := make([]STNode, len(arguments))
	for i, c := range arguments { args[i] = c.Value }

	scopenode := STNode{
		Type: STNodeTypeScope,
		Children: args,
	}

	return Eval(scope, scopenode)
}

// BuiltinGo: The builtin 'go' function. This function concurrently evaluates
// a series of statements within an enclosed, isolated scope
// this function returns UNDEFINED
func BuiltinGo(scope Scope, arguments []Object) Object {
	WaitGroup.Add(1)
	go func () {
		defer WaitGroup.Done()
		BuiltinDo(scope, arguments)
	}()

	return Builtins.Identifiers[UNDEFINED]
}

// BuiltinSleep: the builtin 'sleep' function. This function waits for a
// specified number of milliseconds
// this function returns UNDEFINED
func BuiltinSleep(scope Scope, arguments []Object) Object {
	argobjects := EvalArgs(scope, arguments)

	if argobjects[0].Type != ObjectTypeLiteral ||
		argobjects[0].Value.Type != STNodeTypeNumberLiteral {
		return Builtins.Identifiers[UNDEFINED]
	}

	duration, _ := strconv.ParseFloat(argobjects[0].Value.Head, 64)
	time.Sleep(time.Duration(duration) * time.Millisecond)

	return Builtins.Identifiers[UNDEFINED]
}

// objectToBoolean: Convert a Object to a boolean. This function defines the
// conditions of the builtin 'if' and 'when' functions
// `obj`: the object
// this function returns true or false depending on the type and contents of obj
func objectToBoolean(obj Object) bool {
	if obj.Type == ObjectTypeLiteral {
		if obj.Value.Type == STNodeTypeNumberLiteral {
			return obj.Value.Head != "0"
		}
		if obj.Value.Type == STNodeTypeStringLiteral {
			return len(obj.Value.Head) > 2
		}
	}
	if obj.Type == ObjectTypeList { return len(obj.Elements) > 0 }
	if obj.Type == ObjectTypeMap { return len(obj.MapKeys) > 0 }
	if obj.Type == ObjectTypeFunction { return true }

	return false
}

// BuiltinIf: the builtin 'if' function. This function evaluates a predicate
// and evaluates one of two expressions depending on the result of the predicate
// the function returns the result of the expression that is evaluated, or
// UNDEFINED
func BuiltinIf(scope Scope, args []Object) Object {
	arguments := []Object{Builtins.Identifiers[UNDEFINED]}

	if len(args) == 0 { return Builtins.Identifiers[UNDEFINED] }

	if args[0].Type == ObjectTypeBuiltinArgument {
		argscope := MakeScope(&scope)
		if args[0].Value.Spread {
			spread := SpreadNode(argscope, args[0].Value)
			arguments = spread
		} else {
			arguments[0] = Eval(argscope, args[0].Value)
		}
	} else { arguments[0] = args[0] }

	if objectToBoolean(arguments[0]) {
		if len(arguments) > 1 { return arguments[1] }
		if len(args) > 1 { return EvalArgs(scope, args[1:2])[0] }
	}

	if len(arguments) > 2 { return arguments[2] }
	if len(args) > 2 { return EvalArgs(scope, args[2:3])[0] }

	return Builtins.Identifiers[UNDEFINED]
}

// BuiltinWhen: the builtin 'when' function. This function takes a set of predicate-body
// pairs (expressed as zipped expressions), evaluates the predicates one by one and evaluates
// a 'body' when it reaches a predicate that is true.
// this function returns the result of the body expression that is evaluated, or UNDEFINED
func BuiltinWhen(scope Scope, args []Object) Object {
	for _, arg := range args {
		if arg.Type != ObjectTypeBuiltinArgument {
			return Builtins.Identifiers[UNDEFINED]
		}
	}

	scp := MakeScope(&scope)
	for _, arg := range args {
		obj := Eval(scp, arg.Value)
		if objectToBoolean(obj) {
			if arg.Value.Zip == nil { return Builtins.Identifiers[UNDEFINED] }
			return Eval(scp, *arg.Value.Zip)
		}
	}

	return Builtins.Identifiers[UNDEFINED]
}

// BuiltinComparisonFunction: This function produces a builtin comparison function
// for the specified operator
// `op`: the comparison operator, one of == != > < >= <=
// this function retuns the produced builtin function
func BuiltinComparisonFunction(op string) Object {
	// fn: the builtin comparison function. This function compares numbers and strings as of now
	// this function returns the result of the comparison operator
	fn := func (scope Scope, args []Object) Object {
		arguments := EvalArgs(scope, args)
		if len(arguments) != 2 {
			return Builtins.Identifiers[UNDEFINED]
		}

		// TODO handle lists?

		if arguments[0].Type != ObjectTypeLiteral ||
			arguments[1].Type != ObjectTypeLiteral {
			return NumberObject(0)
		}

		if arguments[0].Value.Head == UNDEFINED ||
			arguments[1].Value.Head == UNDEFINED {
			result := arguments[0].Value.Head == UNDEFINED &&
				arguments[1].Value.Head == UNDEFINED &&
				strings.Contains(op, "=")

			resultint := 0
			if result { resultint = 1 }

			return NumberObject(float64(resultint))
		}

		argtype := arguments[0].Value.Type
		if arguments[1].Value.Type != argtype {
			return Builtins.Identifiers[UNDEFINED]
		}

		str1, str2 := "", ""
		num1, num2 := 0.0, 0.0

		if argtype == STNodeTypeStringLiteral {
			str1 = arguments[0].Value.Head[1:len(arguments[0].Value.Head) - 1]
			str2 = arguments[1].Value.Head[1:len(arguments[1].Value.Head) - 1]
		} else {
			num1, _ = strconv.ParseFloat(arguments[0].Value.Head, 64)
			num2, _ = strconv.ParseFloat(arguments[1].Value.Head, 64)
		}

		resultnum := 0.0
		result := false
		switch op {
		case "==":
			result = num1 == num2
			if argtype == STNodeTypeStringLiteral { result = str1 == str2 }
		case "!=":
			result = num1 != num2
			if argtype == STNodeTypeStringLiteral { result = str1 != str2 }
		case ">":
			result = num1 > num2
			if argtype == STNodeTypeStringLiteral { result = str1 > str2 }
		case "<":
			result = num1 < num2
			if argtype == STNodeTypeStringLiteral { result = str1 < str2 }
		case ">=":
			result = num1 >= num2
			if argtype == STNodeTypeStringLiteral { result = str1 >= str2 }
		case "<=":
			result = num1 <= num2
			if argtype == STNodeTypeStringLiteral { result = str1 <= str2 }
		}

		if result { resultnum = 1 }

		return NumberObject(resultnum)
	}

	return BuiltinFunctionObject(op, fn)
}

// EvalArgs: evaluate a list of arguments passed to builtin functions,
// primarily used to handle spreading
// `scp`: the scope within which to evaluate the arguments
// `args`: the arguments to evaluate
// this function returns the evaluated arguments as a list of Objects
func EvalArgs(scp Scope, args []Object) []Object {
	scope := MakeScope(&scp)
	arguments := make([]Object, 0, len(args))
	for _, child := range args {
		if child.Type == ObjectTypeBuiltinArgument {
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


// Evaluator

package main

import "math"

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

func MakeScope(scope GolspScope) GolspScope {
	newscope := make(GolspScope)
	for k, v := range scope {
		newscope[k] = v
	}

	return newscope
}

func Eval(scope GolspScope, root STNode) GolspObject {
	if root.Type == STNodeTypeScope {
		newscope := MakeScope(scope)

		var result GolspObject
		for _, child := range root.Children {
			result = Eval(newscope, child)
		}

		return result
	}

	if root.Type == STNodeTypeStringLiteral ||
		root.Type == STNodeTypeNumberLiteral {
		return GolspObject{
			IsFunction: false,
			Function: GolspEmptyFunction(),
			Value: root,
		}
	}

	if root.Type == STNodeTypeIdentifier {
		obj, exists := scope[root.Head]
		if !exists {
			return scope[UNDEFINED]
		}

		return obj
	}

	if len(root.Children) == 0 {
		return scope[UNDEFINED]
	}

	exprscope := MakeScope(scope)
	exprhead := Eval(exprscope, root.Children[0])
	if !exprhead.IsFunction {
		return exprhead
	}

	var arguments []STNode
	for _, child := range root.Children[1:] {
		arguments = append(arguments, child)
	}

	fn := exprhead.Function
	builtin := len(fn.BuiltinPatterns) > 0
	patternindex := matchPatterns(fn, arguments)

	if builtin {
		return fn.BuiltinBodies[patternindex](scope, arguments)
	}

	// Eval function

	var argobjects []GolspObject
	for i, arg := range arguments {
		obj := Eval(exprscope, arg)
		argobjects = append(argobjects, obj)

		if !obj.IsFunction {
			arguments[i] = obj.Value
		}
	}

	patternindex = matchPatterns(fn, arguments)
	pattern := fn.FunctionPatterns[patternindex]
	newscope := MakeScope(scope)

	if len(arguments) < len(pattern) {
		return scope[UNDEFINED]
	}

	for i, symbol := range pattern {
		if symbol.Type != STNodeTypeIdentifier {
			continue
		}

		newscope[symbol.Head] = argobjects[i]
	}

	return Eval(newscope, fn.FunctionBodies[patternindex])
}

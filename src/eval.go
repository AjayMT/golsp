
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

func isResolved(scope GolspScope, symbol STNode) bool {
	if symbol.Type == STNodeTypeStringLiteral ||
		symbol.Type == STNodeTypeNumberLiteral {
		return true
	}

	if symbol.Type == STNodeTypeIdentifier &&
		(scope[symbol.Head].IsFunction || symbol.Head == "undefined") {
		return true
	}

	return false
}

func Eval(root STNode) STNode {
	return eval(Builtins, root)
}

func eval(scope GolspScope, root STNode) STNode {
	if root.Type == STNodeTypeScope {
		newscope := make(GolspScope)
		for k, v := range scope {
			newscope[k] = v
		}

		var result STNode
		for _, child := range root.Children {
			result = eval(newscope, child)
		}

		return result
	}

	if root.Type == STNodeTypeStringLiteral ||
		root.Type == STNodeTypeNumberLiteral {
		return root
	}

	if root.Type == STNodeTypeIdentifier {
		obj, exists := scope[root.Head]
		if !exists {
			return GolspUndefinedIdentifier()
		}

		if obj.IsFunction {
			return root
		}

		return obj.Value
	}

	// root has to be expression or function identifier

	if len(root.Children) == 0 {
		return GolspUndefinedIdentifier()
	}

	exprhead := root.Children[0]
	for !isResolved(scope, exprhead) {
		exprhead = eval(scope, exprhead)
	}

	if exprhead.Type != STNodeTypeIdentifier {
		return exprhead
	}

	var arguments []STNode
	for _, child := range root.Children[1:] {
		arguments = append(arguments, child)
	}

	obj, _ := scope[exprhead.Head]
	fn := obj.Function
	builtin := len(fn.BuiltinPatterns) > 0
	patternindex := matchPatterns(fn, arguments)

	if builtin {
		return fn.BuiltinBodies[patternindex](scope, arguments)
	}

	// eval function

	for i, _ := range arguments {
		for !isResolved(scope, arguments[i]) {
			arguments[i] = eval(scope, arguments[i])
		}
	}

	patternindex = matchPatterns(fn, arguments)
	pattern := fn.FunctionPatterns[patternindex]
	newscope := make(GolspScope)
	for k, v := range scope {
		newscope[k] = v
	}

	if len(arguments) < len(pattern) {
		return GolspUndefinedIdentifier()
	}

	for i, symbol := range pattern {
		if symbol.Type != STNodeTypeIdentifier {
			continue
		}

		obj, exists := scope[arguments[i].Head]
		if exists {
			newscope[symbol.Head] = obj
		} else {
			newscope[symbol.Head] = GolspObject{
				IsFunction: false,
				Function: GolspEmptyFunction(),
				Value: arguments[i],
			}
		}
	}

	return eval(newscope, fn.FunctionBodies[patternindex])
}


// Evaluator

package main

import (
	"math"
	"strconv"
)

func compareNodes(a STNode, b STNode) bool {
	if a.Type == STNodeTypeIdentifier {
		return true
	}

	if a.Type != b.Type { return false }

	if a.Type == STNodeTypeList {
		for i, c := range a.Children {
			if !compareNodes(c, b.Children[i]) {
				return false
			}
		}
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

func LookupIdentifier(scope GolspScope, identifier string) GolspObject {
	obj, exists := scope.Identifiers[identifier]
	if exists { return obj }

	if scope.Parent != nil {
		return LookupIdentifier(*(scope.Parent), identifier)
	}

	return Builtins.Identifiers[UNDEFINED]
}

func MakeScope(parent *GolspScope) GolspScope {
	newscope := GolspScope{
		Parent: parent,
		Identifiers: make(map[string]GolspObject),
	}

	return newscope
}

func Eval(scope GolspScope, root STNode) GolspObject {
	if root.Type == STNodeTypeScope {
		newscope := MakeScope(&scope)

		var result GolspObject
		for _, child := range root.Children {
			result = Eval(newscope, child)
		}

		return result
	}

	if root.Type == STNodeTypeStringLiteral ||
		root.Type == STNodeTypeNumberLiteral {
		return GolspObject{
			Type: GolspObjectTypeLiteral,
			Value: root,
		}
	}

	if root.Type == STNodeTypeList {
		elements := make([]GolspObject, len(root.Children))
		for i, c := range root.Children {
			elements[i] = Eval(scope, c)
		}

		return GolspObject{
			Type: GolspObjectTypeList,
			Value: GolspUndefinedIdentifier(),
			Elements: elements,
		}
	}

	if root.Type == STNodeTypeIdentifier {
		obj := LookupIdentifier(scope, root.Head)
		return obj
	}

	if len(root.Children) == 0 {
		return Builtins.Identifiers[UNDEFINED]
	}

	exprhead := Eval(MakeScope(&scope), root.Children[0])

	if exprhead.Type == GolspObjectTypeLiteral {
		return exprhead
	}

	if exprhead.Type == GolspObjectTypeList {
		if len(root.Children) == 1 {
			return exprhead
		}

		indexobj := Eval(scope, root.Children[1])
		if indexobj.Value.Type != STNodeTypeNumberLiteral {
			return Builtins.Identifiers[UNDEFINED]
		}

		index, _ := strconv.Atoi(indexobj.Value.Head)
		if index < 0 { index += len(exprhead.Elements) }
		if index < 0 || index >= len(exprhead.Elements) {
			return Builtins.Identifiers[UNDEFINED]
		}

		return exprhead.Elements[index]
	}

	arguments := make([]STNode, len(root.Children) - 1)
	for i, child := range root.Children[1:] {
		arguments[i] = child
	}

	fn := exprhead.Function
	builtin := len(fn.BuiltinPatterns) > 0
	patternindex := matchPatterns(fn, arguments)

	if builtin {
		return fn.BuiltinBodies[patternindex](scope, arguments)
	}

	// Eval function

	argscope := MakeScope(&scope)
	argobjects := make([]GolspObject, len(arguments))
	for i, arg := range arguments {
		obj := Eval(argscope, arg)
		argobjects[i] = obj

		if obj.Type != GolspObjectTypeFunction {
			arguments[i] = obj.Value
		}
	}

	patternindex = matchPatterns(fn, arguments)
	pattern := fn.FunctionPatterns[patternindex]

	if len(arguments) < len(pattern) {
		return Builtins.Identifiers[UNDEFINED]
	}

	for i, symbol := range pattern {
		if symbol.Type != STNodeTypeIdentifier {
			continue
		}

		exprhead.Scope.Identifiers[symbol.Head] = argobjects[i]
	}

	return Eval(exprhead.Scope, fn.FunctionBodies[patternindex])
}

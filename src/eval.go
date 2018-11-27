
// Evaluator

package main

import (
	"math"
	"strconv"
)

func comparePatternNode(pattern STNode, arg GolspObject) bool {
	if pattern.Type == STNodeTypeIdentifier { return true }

	if pattern.Type == STNodeTypeStringLiteral ||
		pattern.Type == STNodeTypeNumberLiteral {
		return arg.Value.Head == pattern.Head
	}

	if pattern.Type == STNodeTypeList {
		if arg.Type != GolspObjectTypeList { return false }
		if len(pattern.Children) != len(arg.Elements) { return false }

		for i, c := range pattern.Children {
			if !comparePatternNode(c, arg.Elements[i]) {
				return false
			}
		}
	}

	return true
}

func matchPatterns(fn GolspFunction, pattern []GolspObject) int {
	patterns := fn.FunctionPatterns
	bestmatchscore := 0
	bestmatchindex := 0

	for i, p := range patterns {
		score := 0
		minlen := int(math.Min(float64(len(p)), float64(len(pattern))))

		for j := 0; j < minlen; j++ {
			if comparePatternNode(p[j], pattern[j]) { score++ }
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

func copyObjectScope(object GolspObject) GolspObject {
	newobject := GolspObject{
		Type: object.Type,
		Value: object.Value,
		Function: object.Function,
		Elements: object.Elements,
		Scope: GolspScope{
			Parent: object.Scope.Parent,
			Identifiers: make(map[string]GolspObject),
		},
	}

	for k, o := range object.Scope.Identifiers {
		newobject.Scope.Identifiers[k] = copyObjectScope(o)
	}

	return newobject
}

func isolateScope(scope GolspScope) map[string]GolspObject {
	identifiers := make(map[string]GolspObject)

	if scope.Parent != nil {
		identifiers = isolateScope(*(scope.Parent))
	}

	for k, o := range scope.Identifiers {
		identifiers[k] = copyObjectScope(o)
	}

	return identifiers
}

func evalSlice(list GolspObject, arguments []GolspObject) GolspObject {
	for _, arg := range arguments {
		if arg.Value.Type != STNodeTypeNumberLiteral {
			return Builtins.Identifiers[UNDEFINED]
		}
	}

	if len(arguments) == 0 { return list }

	if len(arguments) == 1 {
		indexf, _ := strconv.ParseFloat(arguments[0].Value.Head, 64)
		index := int(indexf)
		if index < 0 { index += len(list.Elements) }
		if index < 0 || index >= len(list.Elements) {
			return Builtins.Identifiers[UNDEFINED]
		}
		return list.Elements[index]
	}

	startf, _ := strconv.ParseFloat(arguments[0].Value.Head, 64)
	start := int(startf)
	endf, _ := strconv.ParseFloat(arguments[1].Value.Head, 64)
	end := int(endf)
	step := 1
	if len(arguments) > 2 {
		stepf, _ := strconv.ParseFloat(arguments[2].Value.Head, 64)
		step = int(stepf)
	}

	if start < 0 { start += len(list.Elements) }
	if end < 0 { end += len(list.Elements) }

	if start < 0 || start >= len(list.Elements) {
		return Builtins.Identifiers[UNDEFINED]
	}
	if end < 0 || end > len(list.Elements) {
		return Builtins.Identifiers[UNDEFINED]
	}

	slice := GolspObject{
		Type: GolspObjectTypeList,
		Elements: make([]GolspObject, 0),
	}

	for i := start; i != end && i < len(list.Elements) && i >= 0; i += step {
		if i > end && step > 0 { break }
		if i < end && step < 0 { break }

		slice.Elements = append(slice.Elements, list.Elements[i])
	}

	return slice
}

func Eval(scope GolspScope, root STNode) GolspObject {
	if root.Type == STNodeTypeScope {
		newscope := GolspScope{
			Identifiers: isolateScope(scope),
		}

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
			Elements: elements,
		}
	}

	if root.Type == STNodeTypeIdentifier {
		return LookupIdentifier(scope, root.Head)
	}

	if len(root.Children) == 0 {
		return Builtins.Identifiers[UNDEFINED]
	}

	exprhead := Eval(MakeScope(&scope), root.Children[0])

	if exprhead.Type == GolspObjectTypeLiteral {
		return exprhead
	}

	if exprhead.Type == GolspObjectTypeList {
		argscope := MakeScope(&scope)
		argobjects := make([]GolspObject, len(root.Children) - 1)
		for i, c := range root.Children[1:] {
			argobjects[i] = Eval(argscope, c)
		}

		return evalSlice(exprhead, argobjects)
	}

	arguments := make([]STNode, len(root.Children) - 1)
	for i, child := range root.Children[1:] {
		arguments[i] = child
	}

	fn := exprhead.Function
	builtin := len(fn.BuiltinPatterns) > 0

	if builtin {
		return fn.BuiltinBodies[0](scope, arguments)
	}

	// Eval function

	argscope := MakeScope(&scope)
	argobjects := make([]GolspObject, len(arguments))
	for i, arg := range arguments {
		obj := Eval(argscope, arg)
		argobjects[i] = obj
	}

	patternindex := matchPatterns(fn, argobjects)
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


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

func copyFunction(fn GolspFunction) GolspFunction {
	copy := GolspFunction{
		FunctionPatterns: make([][]STNode, len(fn.FunctionPatterns)),
		FunctionBodies: make([]STNode, len(fn.FunctionBodies)),
	}

	for i, p := range fn.FunctionPatterns { copy.FunctionPatterns[i] = p }
	for i, p := range fn.FunctionBodies { copy.FunctionBodies[i] = p }
	copy.BuiltinFunc = fn.BuiltinFunc

	return copy
}

func copyObjectScope(object GolspObject) GolspObject {
	newobject := GolspObject{
		Type: object.Type,
		Value: object.Value,
		Function: copyFunction(object.Function),
		Elements: make([]GolspObject, len(object.Elements)),
		Scope: GolspScope{
			Parent: object.Scope.Parent,
			Identifiers: make(map[string]GolspObject),
		},
	}

	for k, o := range object.Scope.Identifiers {
		newobject.Scope.Identifiers[k] = copyObjectScope(o)
	}

	for i, e := range object.Elements {
		newobject.Elements[i] = copyObjectScope(e)
	}

	return newobject
}

func IsolateScope(scope GolspScope) map[string]GolspObject {
	identifiers := make(map[string]GolspObject)

	if scope.Parent != nil {
		identifiers = IsolateScope(*(scope.Parent))
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

func SpreadNode(scope GolspScope, node STNode) []GolspObject {
	nodescope := MakeScope(&scope)
	obj := Eval(nodescope, node)

	if obj.Type != GolspObjectTypeList {
		return []GolspObject{obj}
	}

	return obj.Elements
}

func Eval(scope GolspScope, root STNode) GolspObject {
	if root.Type == STNodeTypeScope {
		newscope := GolspScope{
			Identifiers: IsolateScope(scope),
		}

		var result GolspObject
		for _, child := range root.Children {
			if child.Spread {
				spread := SpreadNode(newscope, child)
				result = spread[len(spread) - 1]
			} else {
				result = Eval(newscope, child)
			}
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

	if root.Type == STNodeTypeIdentifier {
		return LookupIdentifier(scope, root.Head)
	}

	if root.Type == STNodeTypeList {
		var elements []GolspObject
		for _, c := range root.Children {
			if c.Spread {
				spread := SpreadNode(scope, c)
				for _, elem := range spread { elements = append(elements, elem) }
			} else {
				elements = append(elements, Eval(MakeScope(&scope), c))
			}
		}

		return GolspObject{
			Type: GolspObjectTypeList,
			Elements: elements,
		}
	}

	if len(root.Children) == 0 {
		return Builtins.Identifiers[UNDEFINED]
	}

	var exprhead GolspObject
	var argobjects []GolspObject
	argscope := MakeScope(&scope)

	if root.Children[0].Spread {
		spread := SpreadNode(scope, root.Children[0])

		if len(spread) == 0 {
			return Builtins.Identifiers[UNDEFINED]
		}

		exprhead = spread[0]
		argobjects = spread[1:]
	} else {
		exprhead = Eval(MakeScope(&scope), root.Children[0])
	}

	if exprhead.Type == GolspObjectTypeFunction {
		exprhead.Scope.Identifiers = make(map[string]GolspObject)
	}

	if exprhead.Type == GolspObjectTypeLiteral {
		return exprhead
	}

	if exprhead.Type == GolspObjectTypeList {
		for _, c := range root.Children[1:] {
			if c.Spread {
				spread := SpreadNode(argscope, c)
				for _, obj := range spread { argobjects = append(argobjects, obj) }
			} else {
				argobjects = append(argobjects, Eval(argscope, c))
			}
		}

		return evalSlice(exprhead, argobjects)
	}

	fn := exprhead.Function
	builtin := fn.BuiltinFunc != nil

	if builtin {
		for _, c := range root.Children[1:] {
			obj := GolspObject{
				Type: GolspObjectTypeBuiltinArgument,
				Value: c,
			}
			argobjects = append(argobjects, obj)
		}

		return fn.BuiltinFunc(scope, argobjects)
	}

	// Eval function

	for _, c := range root.Children[1:] {
		if c.Spread {
			spread := SpreadNode(argscope, c)
			for _, obj := range spread { argobjects = append(argobjects, obj) }
		} else {
			argobjects = append(argobjects, Eval(argscope, c))
		}
	}

	patternindex := matchPatterns(fn, argobjects)
	pattern := fn.FunctionPatterns[patternindex]

	if len(argobjects) < len(pattern) {
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


// Evaluator

package golsp

import (
	"math"
	"strconv"
	"fmt"
	"sync"
)

// runtimeWaitGroup is the wait group used to wait for 'go' blocks to complete
var runtimeWaitGroup sync.WaitGroup
func RuntimeWaitGroup() *sync.WaitGroup {
	return &runtimeWaitGroup
}

// comparePatternNode: Compare a node in a function pattern with an argument object
// `pattern`: the pattern node
// `arg`: the argument to compare with the pattern
// this function returns whether the argument matches the pattern node
func comparePatternNode(pattern STNode, arg Object) bool {
	// identifiers i.e non-literal patterns match everything
	if pattern.Type == STNodeTypeIdentifier { return true }

	// literal patterns match arguments that have the same value
	if pattern.Type == STNodeTypeStringLiteral ||
		pattern.Type == STNodeTypeNumberLiteral {
		return arg.Value.Head == pattern.Head
	}

	// map patterns match if all the specified keys and values match
	// value-only matching i.e `[foo ( quux: "hello" )]` does not work yet
	if pattern.Type == STNodeTypeMap {
		if arg.Type != ObjectTypeMap { return false }

		for i, c := range pattern.Children {
			if c.Spread && c.Type == STNodeTypeIdentifier {
				return len(arg.MapKeys) >= i
			}
			if len(arg.MapKeys) <= i { return false }
			if c.Type == STNodeTypeStringLiteral || c.Type == STNodeTypeNumberLiteral {
				value, exists := arg.Map[c.Head]
				if !exists { return false }
				if c.Zip != nil {
					if !comparePatternNode(*c.Zip, value) { return false }
				}
			}
		}

		if len(arg.MapKeys) > len(pattern.Children) { return false }
	}

	// list patterns match if each of their elements match and the lists
	// are of the same length, after accounting for spreading
	if pattern.Type == STNodeTypeList {
		if arg.Type != ObjectTypeList { return false }

		for i, c := range pattern.Children {
			if c.Spread && c.Type == STNodeTypeIdentifier {
				return len(arg.Elements) >= i
			}
			if len(arg.Elements) <= i { return false }
			if !comparePatternNode(c, arg.Elements[i]) { return false }
		}

		if len(arg.Elements) > len(pattern.Children) { return false }
	}

	return true
}

// matchPatterns: Match a list of arguments to a particular function pattern
// `fn`: the function whose patterns to check
// `arguments`: the list of arguments to match to a pattern
// this function returns the index of the best-matching pattern in function's
// list of patterns, and whether a matching pattern was found
func matchPatterns(fn Function, arguments []Object) (int, bool) {
	patterns := fn.FunctionPatterns
	bestscore := 0
	bestdiff := -1
	matchindex := 0
	found := false

	for i, p := range patterns {
		score := 0
		diff := 0
		minlen := len(p)
		if len(p) > len(arguments) {
			diff = len(p) - len(arguments)
			minlen = len(arguments)
		}

		if len(p) == 0 { found = true }

		for j := 0; j < minlen; j++ {
			if comparePatternNode(p[j], arguments[j]) {
				found = true
				score++
			}
			if p[j].Spread {
				score += len(arguments) - 1 - j
				break
			}
		}

		if bestdiff == -1 { bestdiff = diff }
		if score > bestscore || (score == bestscore && diff < bestdiff) {
			matchindex = i
		}
		if diff < bestdiff { bestdiff = diff }
		if score > bestscore { bestscore = score }
	}

	return matchindex, found
}

// LookupIdentifier: lookup an identifier within a particular scope
// `scope`: the scope in which to search for the identifier
// `identifier`: the name of the identifier
// this function returns the object corresponding to the identifier
// or UNDEFINED
func LookupIdentifier(scope Scope, identifier string) Object {
	obj, exists := scope.Identifiers[identifier]
	if exists { return obj }

	if scope.Parent != nil {
		return LookupIdentifier(*(scope.Parent), identifier)
	}

	return Builtins.Identifiers[UNDEFINED]
}

// MakeScope: construct a new child scope that descends from a parent scope
// `parent`: the parent scope
// this function returns a new Scope struct whose Parent points to
// parent
func MakeScope(parent *Scope) Scope {
	newscope := Scope{
		Parent: parent,
		Identifiers: make(map[string]Object),
		Constants: make(map[string]bool, len(parent.Constants)),
	}
	for k, v := range parent.Constants { newscope.Constants[k] = v }

	return newscope
}

// CopyFunction: Copy a Function struct
// `fn`: the function to copy
// this function returns a copy of fn
func CopyFunction(fn Function) Function {
	fncopy := Function{
		Name: fn.Name,
		FunctionPatterns: make([][]STNode, len(fn.FunctionPatterns)),
		FunctionBodies: make([]STNode, len(fn.FunctionBodies)),
		BuiltinFunc: fn.BuiltinFunc,
	}
	copy(fncopy.FunctionPatterns, fn.FunctionPatterns)
	copy(fncopy.FunctionBodies, fn.FunctionBodies)

	return fncopy
}

// CopyObject: Copy an Object
// `object`: the object to copy
// this function returns a copy of object. Note that it does not copy
// object.Value since that property is never modified
func CopyObject(object Object) Object {
	newobject := Object{
		Type: object.Type,
		Value: object.Value,
		Function: CopyFunction(object.Function),
		Elements: make([]Object, len(object.Elements)),
		MapKeys: make([]Object, len(object.MapKeys)),
		Map: make(map[string]Object, len(object.Map)),
		Scope: Scope{
			Parent: object.Scope.Parent,
			Identifiers: make(map[string]Object, len(object.Scope.Identifiers)),
			Constants: make(map[string]bool, len(object.Scope.Constants)),
		},
	}

	for k, o := range object.Scope.Identifiers { newobject.Scope.Identifiers[k] = CopyObject(o) }
	for k, v := range object.Scope.Constants { newobject.Scope.Constants[k] = v }
	for i, e := range object.Elements { newobject.Elements[i] = CopyObject(e) }
	for i, k := range object.MapKeys { newobject.MapKeys[i] = CopyObject(k) }
	for k, v := range object.Map { newobject.Map[k] = CopyObject(v) }

	return newobject
}

// IsolateScope: 'Isolate' a scope object by copying all values from its parent
// scopes into the scope struct, effectively orphaning it and flattening its
// inheritance tree
// `scope`: the scope to isolate
// this function returns the isolated scope
func IsolateScope(scope Scope) Scope {
	newscope := Scope{
		Identifiers: make(map[string]Object, len(scope.Identifiers)),
		Constants: make(map[string]bool, len(scope.Constants)),
	}
	if scope.Parent != nil {
		parent := IsolateScope(*(scope.Parent))
		for k, obj := range parent.Identifiers {
			obj.Scope.Parent = &newscope
			newscope.Identifiers[k] = obj
		}
		for k, v := range parent.Constants { newscope.Constants[k] = v }
	}
	for k, o := range scope.Identifiers {
		obj := CopyObject(o)
		obj.Scope.Parent = &newscope
		newscope.Identifiers[k] = obj
	}
	for k, v := range scope.Constants { newscope.Constants[k] = v }

	return newscope
}

// EvalSlice: Evaluate a slice expression, i.e `[list begin end step]`
// `list`: the list or string that is sliced
// `arguments`: the arguments passed in the expression
// this function returns a slice of the list/string or UNDEFINED
func EvalSlice(list Object, arguments []Object) Object {
	if len(arguments) == 0 { return list }

	listlen := len(list.Elements)
	if list.Type == ObjectTypeLiteral { listlen = len(list.Value.Head) - 2 }

	if len(arguments) == 1 {
		indexf, _ := strconv.ParseFloat(arguments[0].Value.Head, 64)
		index := int(indexf)
		if index < 0 { index += listlen }
		if index < 0 || index >= listlen {
			return Builtins.Identifiers[UNDEFINED]
		}

		if list.Type == ObjectTypeList { return list.Elements[index] }

		liststr := []rune(list.Value.Head[1:listlen + 1])
		str := fmt.Sprintf("\"%s\"", string(liststr[index:index + 1]))

		return Object{
			Type: ObjectTypeLiteral,
			Value: STNode{
				Type: STNodeTypeStringLiteral,
				Head: str,
			},
		}
	}

	startf, _ := strconv.ParseFloat(arguments[0].Value.Head, 64)
	start := int(startf)
	end := listlen
	step := 1

	if len(arguments) > 2 && arguments[2].Value.Type == STNodeTypeNumberLiteral {
		stepf, _ := strconv.ParseFloat(arguments[2].Value.Head, 64)
		step = int(stepf)
		if step == 0 { return Builtins.Identifiers[UNDEFINED] }
		if step < 0 { end = -listlen - 1 }
	}

	if arguments[1].Value.Type == STNodeTypeNumberLiteral {
		endf, _ := strconv.ParseFloat(arguments[1].Value.Head, 64)
		end = int(endf)
	}

	if start < 0 { start += listlen }
	if end < 0 { end += listlen }

	slice := Object{
		Type: list.Type,
		Elements: make([]Object, 0, listlen),
	}
	slicestr := make([]rune, 0, listlen)
	var liststr []rune
	if list.Type == ObjectTypeLiteral {
		liststr = []rune(list.Value.Head[1:listlen + 1])
	}

	if start < 0 || start >= listlen {
		if list.Type == ObjectTypeLiteral {
			slice.Value = STNode{
				Type: STNodeTypeStringLiteral,
				Head: fmt.Sprintf("\"%s\"", string(slicestr)),
			}
		}

		return slice
	}

	for i := start; i != end; i += step {
		if i >= listlen { break }
		if i < 0 { break }

		if slice.Type == ObjectTypeList {
			slice.Elements = append(slice.Elements, list.Elements[i])
		} else {
			slicestr = append(slicestr, liststr[i])
		}
	}

	if list.Type == ObjectTypeLiteral {
		slice.Value = STNode{
			Type: STNodeTypeStringLiteral,
			Head: fmt.Sprintf("\"%s\"", string(slicestr)),
		}
	}

	return slice
}

// EvalMap: Lookup key(s) in a map
// `glmap`: the map object
// `arguments`: the key or keys to look up
// this function returns the object or list of objects that the key(s) map to
func EvalMap(glmap Object, arguments []Object) Object {
	if len(arguments) == 0 { return glmap }
	if len (arguments) == 1 {
		value, exists := glmap.Map[arguments[0].Value.Head]
		if arguments[0].Type != ObjectTypeLiteral || !exists {
			return Builtins.Identifiers[UNDEFINED]
		}
		return value
	}

	values := make([]Object, len(arguments))
	for i, arg := range arguments {
		value, exists := glmap.Map[arg.Value.Head]
		if arg.Type != ObjectTypeLiteral || !exists {
			values[i] = Builtins.Identifiers[UNDEFINED]
		} else {
			values[i] = value
		}
	}

	return Object{
		Type: ObjectTypeList,
		Elements: values,
	}
}

// SpreadNode: Apply the spread operator to a syntax tree node
// `scope`: the scope within which the node is being spread
// `node`: the node to spread
// this function returns the list of Objects that the node spreads to
func SpreadNode(scope Scope, node STNode) []Object {
	nodescope := MakeScope(&scope)
	obj := Eval(nodescope, node)
	if obj.Value.Head == UNDEFINED { return make([]Object, 0) }

	if obj.Type != ObjectTypeList &&
		obj.Type != ObjectTypeMap &&
		obj.Value.Type != STNodeTypeStringLiteral {
		return []Object{obj}
	}

	if obj.Type == ObjectTypeList { return obj.Elements }
	if obj.Type == ObjectTypeMap { return obj.MapKeys }

	str := obj.Value.Head[1:len(obj.Value.Head) - 1]
	objects := make([]Object, len(str))

	for i, r := range str {
		objects[i] = Object{
			Type: ObjectTypeLiteral,
			Value: STNode{
				Type: STNodeTypeStringLiteral,
				Head: fmt.Sprintf("\"%s\"", string(r)),
			},
		}
	}

	return objects
}

// bindArguments: Bind the arguments passed to a function to the function
// object's Scope property
// `exprhead`: the 'expression head' i.e function object
// `pattern`: the matched pattern, based on which arguments will be bound
// to identifiers
// `argobjects`: the arguments passed to the function that will be bound to
// identifiers
func bindArguments(exprhead Object, pattern []STNode, argobjects []Object) {
	for i, symbol := range pattern {
		if symbol.Type == STNodeTypeStringLiteral || symbol.Type == STNodeTypeNumberLiteral {
			continue
		}

		if symbol.Type == STNodeTypeIdentifier {
			if symbol.Spread {
				exprhead.Scope.Identifiers[symbol.Head] = Object{
					Type: ObjectTypeList,
					Elements: argobjects[i:],
				}
				break
			}
			exprhead.Scope.Identifiers[symbol.Head] = argobjects[i]
			continue
		}

		if argobjects[i].Type == ObjectTypeList && symbol.Type == STNodeTypeList {
			bindArguments(exprhead, symbol.Children, argobjects[i].Elements)
		}

		if argobjects[i].Type == ObjectTypeMap && symbol.Type == STNodeTypeMap {
			// this is a giant mess. clean it up

			mapped := make(map[string]bool)
			mappatternindex := 0
			for iterindex, child := range symbol.Children {
				mappatternindex = iterindex
				if !(child.Type == STNodeTypeNumberLiteral ||
					child.Type == STNodeTypeStringLiteral) {
					break
				}

				if child.Zip == nil { continue }

				value, exists := argobjects[i].Map[child.Head]
				if !exists { continue }

				bindArguments(exprhead, []STNode{*child.Zip}, []Object{value})
				mapped[child.Head] = true
			}

			keys := make([]Object, 0, len(argobjects[i].MapKeys))
			values := make([]Object, 0, len(argobjects[i].MapKeys))
			for _, key := range argobjects[i].MapKeys {
				if !mapped[key.Value.Head] {
					keys = append(keys, key)
					values = append(values, argobjects[i].Map[key.Value.Head])
				}
			}

			patternkeys := symbol.Children[mappatternindex:]
			patternvalues := make([]STNode, 0, len(patternkeys))
			for _, c := range patternkeys {
				if c.Zip == nil { continue }
				patternvalues = append(patternvalues, *c.Zip)
			}

			bindArguments(exprhead, patternkeys, keys)
			bindArguments(exprhead, patternvalues, values)
		}
	}
}

// evalDot: Evaluate 'dot' property access operator on map objects
// `obj`: the (map) object
// `root`: the syntax tree node associated with the dot operator
// this function returns the value from the map object
func evalDot(obj Object, root STNode) Object {
	if root.Dot == nil { return obj }
	if obj.Type != ObjectTypeMap { return Builtins.Identifiers[UNDEFINED] }
	if root.Dot.Type != STNodeTypeIdentifier { return Builtins.Identifiers[UNDEFINED] }

	key := fmt.Sprintf("\"%s\"", root.Dot.Head)
	value, exists := obj.Map[key]
	if !exists { return Builtins.Identifiers[UNDEFINED] }

	return evalDot(value, *root.Dot)
}

// Eval: Evaluate a syntax tree node within a scope
// `scope`: the scope within which to evaluate the node
// `root`: the root node to evaluate
// this function returns the result of evaluating the node as an Object
func Eval(scope Scope, root STNode) Object {
	// root node is a scope -- it evaluates to the result of the last expression
	// in the scope
	// scope nodes are isolated from their parents to ensure that they do not
	// cause side-effects, especially important for 'go' blocks
	if root.Type == STNodeTypeScope {
		newscope := IsolateScope(scope)
		var result Object
		for _, child := range root.Children {
			if child.Spread {
				spread := SpreadNode(newscope, child)
				result = spread[len(spread) - 1]
			} else {
				result = Eval(newscope, child)
			}
		}

		return evalDot(CopyObject(result), root)
	}

	// string and number literals simply evaluate to themselves
	if root.Type == STNodeTypeNumberLiteral || root.Type == STNodeTypeStringLiteral {
		result := Object{
			Type: ObjectTypeLiteral,
			Value: root,
		}
		return evalDot(result, root)
	}

	// identifers evaluate to their corresponding values within the scope or UNDEFINED
	if root.Type == STNodeTypeIdentifier {
		return evalDot(LookupIdentifier(scope, root.Head), root)
	}

	// 'list' type syntax tree nodes evaluate to 'list' type Objects
	// note that list elements are evaluated immediately, unlike quote expressions
	// in Lisp
	if root.Type == STNodeTypeList {
		elements := make([]Object, 0, len(root.Children))
		for _, c := range root.Children {
			if c.Spread {
				elements = append(elements, SpreadNode(scope, c)...)
			} else {
				elements = append(elements, Eval(MakeScope(&scope), c))
			}
		}

		result := Object{
			Type: ObjectTypeList,
			Elements: elements,
		}
		return evalDot(result, root)
	}

	// 'map' type syntax tree nodes evaluate to maps
	if root.Type == STNodeTypeMap {
		obj := Object{
			Type: ObjectTypeMap,
			Map: make(map[string]Object, len(root.Children)),
			MapKeys: make([]Object, 0, len(root.Children)),
		}

		for _, c := range root.Children {
			if c.Zip == nil { continue }
			var left []Object
			var right []Object

			if c.Spread {
				left = SpreadNode(scope, c)
			} else {
				left = []Object{Eval(MakeScope(&scope), c)}
			}
			if c.Zip.Spread {
				right = SpreadNode(scope, *c.Zip)
			} else {
				right = []Object{Eval(MakeScope(&scope), *c.Zip)}
			}

			minlen := int(math.Min(float64(len(left)), float64(len(right))))
			for index := 0; index < minlen; index++ {
				if left[index].Type != ObjectTypeLiteral {
					continue
				}

				_, exists := obj.Map[left[index].Value.Head]
				obj.Map[left[index].Value.Head] = right[index]
				if !exists {
					obj.MapKeys = append(obj.MapKeys, left[index])
				}
			}
		}

		return evalDot(obj, root)
	}

	// at this point the root node must be an expression

	// empty expressions evaluate to UNDEFINED
	if len(root.Children) == 0 {
		return evalDot(Builtins.Identifiers[UNDEFINED], root)
	}

	// exprhead is the head of the expression, aka the function
	// that is being called, list that is being sliced, etc...
	// argobjects is the rest of the expression, the arguments passed
	// to exprhead
	// arguments are evaluated in their own scope (argscope) to prevent side effects
	var exprhead Object
	argobjects := make([]Object, 0, len(root.Children))
	argscope := MakeScope(&scope)

	if root.Children[0].Spread {
		spread := SpreadNode(scope, root.Children[0])
		if len(spread) == 0 { return Builtins.Identifiers[UNDEFINED] }
		exprhead = spread[0]
		argobjects = spread[1:]
	} else {
		exprhead = Eval(MakeScope(&scope), root.Children[0])
	}

	// the function's argument scope is cleared every time it is called
	// since the arguments will be bound again
	if exprhead.Type == ObjectTypeFunction {
		exprhead.Scope.Identifiers = make(map[string]Object, len(exprhead.Scope.Identifiers))
	}

	// evaluating an expression with a number literal or UNDEFINED head
	// produces the literal or UNDEFINED
	// i.e [1 2 3] evals to 1, [undefined a b c] evals to undefined
	if exprhead.Type == ObjectTypeLiteral &&
		(exprhead.Value.Type == STNodeTypeNumberLiteral ||
		exprhead.Value.Head == UNDEFINED) {
		return evalDot(exprhead, root)
	}

	// if exprhead is a list or string literal, slice it
	// if it is a map, lookup key
	if exprhead.Type == ObjectTypeList ||
		exprhead.Type == ObjectTypeMap ||
		exprhead.Value.Type == STNodeTypeStringLiteral {
		for _, c := range root.Children[1:] {
			if c.Spread {
				argobjects = append(argobjects, SpreadNode(argscope, c)...)
			} else {
				argobjects = append(argobjects, Eval(argscope, c))
			}
		}

		if exprhead.Type == ObjectTypeMap {
			return evalDot(EvalMap(exprhead, argobjects), root)
		}

		return evalDot(EvalSlice(exprhead, argobjects), root)
	}

	// at this point the expression must be a function call

	fn := exprhead.Function
	builtin := fn.BuiltinFunc != nil

	// builtin functions are called without evaluating the
	// argument syntax tree nodes, these functions can decide how to eval
	// arguments on their own
	if builtin {
		for _, c := range root.Children[1:] {
			obj := Object{
				Type: ObjectTypeBuiltinArgument,
				Value: c,
			}
			argobjects = append(argobjects, obj)
		}

		return evalDot(fn.BuiltinFunc(scope, argobjects), root)
	}

	// at this point the expression must be a calling a user-defined function
	// all arguments are evaluated immediately
	for _, c := range root.Children[1:] {
		if c.Spread {
			argobjects = append(argobjects, SpreadNode(argscope, c)...)
		} else {
			argobjects = append(argobjects, Eval(argscope, c))
		}
	}

	patternindex, patternfound := matchPatterns(fn, argobjects)
	if !patternfound { return Builtins.Identifiers[UNDEFINED] }
	pattern := fn.FunctionPatterns[patternindex]

	// calling a function with fewer arguments than required evaluates to UNDEFINED
	// might possibly implement automatic partial evaluation in the future
	if len(argobjects) < len(pattern) {
		return Builtins.Identifiers[UNDEFINED]
	}

	bindArguments(exprhead, pattern, argobjects)

	return evalDot(Eval(exprhead.Scope, fn.FunctionBodies[patternindex]), root)
}

// Run: Run a Golsp program
// `program`: the program to run
// `dirname`: the directory of the program file
// `filename`: the name of the program file
// `args`: command line arguments passed to the program
// this function returns the result of running the program
func Run(dirname string, filename string, args []string, program string) Object {
	InitializeBuiltins(dirname, filename, args)
	result := Eval(Builtins, MakeST(Tokenize(program)))
	defer RuntimeWaitGroup().Wait()

	return result
}

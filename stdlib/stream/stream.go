
package main

import (
	g "github.com/ajaymt/golsp/core"
	"sync"
)

type Stream struct {
	Queue g.List
	mux sync.Mutex
}

func stream(_ g.Scope, _ []g.Object) g.Object {
	s := Stream{}

	read := func(_ g.Scope, _ []g.Object) g.Object {
		for s.Queue.Length == 0 {}

		s.mux.Lock()
		obj := s.Queue.First.Object
		s.Queue = s.Queue.Slice(1, s.Queue.Length).Elements
		s.mux.Unlock()

		return obj
	}

	flush := func(_ g.Scope, _ []g.Object) g.Object {
		s.mux.Lock()
		q := s.Queue
		s.Queue = g.List{}
		s.mux.Unlock()

		return g.Object{
			Type: g.ObjectTypeList,
			Elements: q,
		}
	}

	write := func(scope g.Scope, args []g.Object) g.Object {
		arguments := g.EvalArgs(scope, args)
		fn := arguments[0]
		data := arguments[1:]
		for _, obj := range data {
			arglist := g.List{}
			arglist.Append(obj)
			result := g.CallFunction(fn, arglist)
			s.mux.Lock()
			s.Queue.Append(result)
			s.mux.Unlock()
		}

		return g.UndefinedObject()
	}

	return g.MapObject(map[string]g.Object{
		"read": g.BuiltinFunctionObject("read", read),
		"flush": g.BuiltinFunctionObject("flush", flush),
		"write": g.BuiltinFunctionObject("write", write),
	})
}

var Exports = g.BuiltinFunctionObject("stream", stream)


package main

import (
	"os"
	"io"
	"bufio"
	g "github.com/ajaymt/golsp/core"
)

var openFiles = []*bufio.ReadWriter{
	bufio.NewReadWriter(bufio.NewReader(os.Stdin), nil),
	bufio.NewReadWriter(nil, bufio.NewWriter(os.Stdout)),
	bufio.NewReadWriter(nil, bufio.NewWriter(os.Stderr)),
}

func open(scope g.Scope, args []g.Object) g.Object {
	arguments := g.EvalArgs(scope, args)
	filename, _ := g.ToString(arguments[0])

	f, err := os.OpenFile(filename, os.O_RDWR | os.O_CREATE, 0755)
	if err != nil { return g.UndefinedObject() }

	reader := bufio.NewReader(f)
	writer := bufio.NewWriter(f)
	openFiles = append(openFiles, bufio.NewReadWriter(reader, writer))

	return g.NumberObject(float64(len(openFiles) - 1))
}

func read(scope g.Scope, args []g.Object) g.Object {
	arguments := g.EvalArgs(scope, args)
	indexf, _ := g.ToNumber(arguments[0])
	nf, _ := g.ToNumber(arguments[1])
	index, n := int(indexf), int(nf)
	if index < 0 || index >= len(openFiles) { return g.UndefinedObject() }
	if n < 0 { return g.UndefinedObject() }

	readwriter := openFiles[index]
	if readwriter.Reader == nil { return g.UndefinedObject() }

	bytes := make([]byte, n)
	_, err := readwriter.Reader.Read(bytes)
	if err != nil && err != io.EOF { return g.UndefinedObject() }

	return g.StringObject(string(bytes))
}

func readAll(scope g.Scope, args []g.Object) g.Object {
	arguments := g.EvalArgs(scope, args)
	indexf, _ := g.ToNumber(arguments[0])
	index := int(indexf)
	if index < 0 || index >= len(openFiles) { return g.UndefinedObject() }

	readwriter := openFiles[index]
	if readwriter.Reader == nil { return g.UndefinedObject() }

	bytes := make([]byte, 0)
	b, err := readwriter.Reader.ReadByte()
	for ; err == nil; b, err = readwriter.Reader.ReadByte() {
		bytes = append(bytes, b)
	}

	if err != io.EOF { return g.UndefinedObject() }

	return g.StringObject(string(bytes))
}

func readUntil(scope g.Scope, args []g.Object) g.Object {
	arguments := g.EvalArgs(scope, args)
	indexf, _ := g.ToNumber(arguments[0])
	index := int(indexf)
	delim, _ := g.ToString(arguments[1])
	if index < 0 || index >= len(openFiles) || len(delim) == 0 {
		return g.UndefinedObject()
	}

	readwriter := openFiles[index]
	if readwriter.Reader == nil { return g.UndefinedObject() }

	bytes, err := readwriter.Reader.ReadBytes(delim[0])
	if err != nil && err != io.EOF { return g.UndefinedObject() }

	return g.StringObject(string(bytes[:len(bytes) - 1]))
}

func write(scope g.Scope, args []g.Object) g.Object {
	arguments := g.EvalArgs(scope, args)
	str, _ := g.ToString(arguments[1])
	indexf, _ := g.ToNumber(arguments[0])
	index := int(indexf)
	if index < 0 || index >= len(openFiles) { return g.UndefinedObject() }

	readwriter := openFiles[index]
	if readwriter.Writer == nil { return g.UndefinedObject() }

	nwritten, err := readwriter.Writer.WriteString(str)
	if err != nil { return g.UndefinedObject() }

	err = readwriter.Writer.Flush()
	if err != nil { return g.UndefinedObject() }

	return g.NumberObject(float64(nwritten))
}

func seek(_ g.Scope, args []g.Object) g.Object {
	// TODO
	return g.UndefinedObject()
}

func stat(_ g.Scope, args []g.Object) g.Object {
	// TODO
	return g.UndefinedObject()
}

func exit(scope g.Scope, args []g.Object) g.Object {
	arguments := g.EvalArgs(scope, args)
	n, _ := g.ToNumber(arguments[0])
	os.Exit(int(n))
	return g.UndefinedObject()
}

var Exports = g.MapObject(map[string]g.Object{
	"stdin": g.NumberObject(0.0),
	"stdout": g.NumberObject(1.0),
	"stderr": g.NumberObject(2.0),
	"open": g.BuiltinFunctionObject("open", open),
	"read": g.BuiltinFunctionObject("read", read),
	"readAll": g.BuiltinFunctionObject("readAll", readAll),
	"readUntil": g.BuiltinFunctionObject("readUntil", readUntil),
	"write": g.BuiltinFunctionObject("write", write),
	"seek": g.BuiltinFunctionObject("seek", seek),
	"stat": g.BuiltinFunctionObject("stat", stat),
	"exit": g.BuiltinFunctionObject("exit", exit),
})

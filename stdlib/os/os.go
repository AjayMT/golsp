
package main

import (
	"os"
	"io"
	"io/ioutil"
	"bufio"
	g "github.com/ajaymt/golsp/core"
)

type file struct {
	file *os.File
	reader *bufio.Reader
	writer *bufio.Writer
}

var openFiles = []file{
	file{file: os.Stdin, reader: bufio.NewReader(os.Stdin), writer: nil},
	file{file: os.Stdout, reader: nil, writer: bufio.NewWriter(os.Stdout)},
	file{file: os.Stderr, reader: nil, writer: bufio.NewWriter(os.Stderr)},
}

func cropen(scope g.Scope, args []g.Object, create bool) g.Object {
	arguments := g.EvalArgs(scope, args)
	filename, _ := g.ToString(arguments[0])

	mode := os.O_RDWR
	if create { mode = os.O_RDWR | os.O_CREATE }
	f, err := os.OpenFile(filename, mode, 0644)
	if err != nil { return g.UndefinedObject() }

	reader := bufio.NewReader(f)
	writer := bufio.NewWriter(f)
	openFiles = append(openFiles, file{file: f, reader: reader, writer: writer})

	return g.NumberObject(float64(len(openFiles) - 1))
}

func open(s g.Scope, a []g.Object) g.Object { return cropen(s, a, false) }
func create(s g.Scope, a []g.Object) g.Object { return cropen(s, a, true) }

func rm(scope g.Scope, args []g.Object, all bool) g.Object {
	arguments := g.EvalArgs(scope, args)
	path, _ := g.ToString(arguments[0])

	rmf := os.Remove
	if all { rmf = os.RemoveAll }
	err := rmf(path)
	if err != nil { return g.NumberObject(0.0) }

	return g.NumberObject(1.0)
}

func remove(s g.Scope, a []g.Object) g.Object { return rm(s, a, false) }
func removeAll(s g.Scope, a []g.Object) g.Object { return rm(s, a, true) }

func mkdir(scope g.Scope, args []g.Object) g.Object {
	arguments := g.EvalArgs(scope, args)
	path, _ := g.ToString(arguments[0])

	err := os.MkdirAll(path, 0755)
	if err != nil { return g.NumberObject(0.0) }

	return g.NumberObject(1.0)
}

func read(scope g.Scope, args []g.Object) g.Object {
	arguments := g.EvalArgs(scope, args)
	indexf, _ := g.ToNumber(arguments[0])
	nf, _ := g.ToNumber(arguments[1])
	index, n := int(indexf), int(nf)
	if index < 0 || index >= len(openFiles) { return g.UndefinedObject() }
	if n < 0 { return g.UndefinedObject() }

	readwriter := openFiles[index]
	if readwriter.reader == nil { return g.UndefinedObject() }

	bytes := make([]byte, n)
	_, err := readwriter.reader.Read(bytes)
	if err != nil && err != io.EOF { return g.UndefinedObject() }

	return g.StringObject(string(bytes))
}

func readAll(scope g.Scope, args []g.Object) g.Object {
	arguments := g.EvalArgs(scope, args)
	indexf, _ := g.ToNumber(arguments[0])
	index := int(indexf)
	if index < 0 || index >= len(openFiles) { return g.UndefinedObject() }

	readwriter := openFiles[index]
	if readwriter.reader == nil { return g.UndefinedObject() }

	bytes := make([]byte, 0)
	b, err := readwriter.reader.ReadByte()
	for ; err == nil; b, err = readwriter.reader.ReadByte() {
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
	if readwriter.reader == nil { return g.UndefinedObject() }

	bytes, err := readwriter.reader.ReadBytes(delim[0])
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
	if readwriter.writer == nil { return g.UndefinedObject() }

	nwritten, err := readwriter.writer.WriteString(str)
	if err != nil { return g.UndefinedObject() }

	err = readwriter.writer.Flush()
	if err != nil { return g.UndefinedObject() }

	return g.NumberObject(float64(nwritten))
}

func seek(scope g.Scope, args []g.Object) g.Object {
	arguments := g.EvalArgs(scope, args)
	indexf, _ := g.ToNumber(arguments[0])
	index := int(indexf)
	if index < 0 || index >= len(openFiles) { return g.UndefinedObject() }
	posf, _ := g.ToNumber(arguments[1])
	pos := int64(posf)
	if pos < 0 { return g.UndefinedObject() }
	whencef, _ := g.ToNumber(arguments[2])
	whence := int(whencef)
	if whence < 0 || whence > 2 { return g.UndefinedObject() }

	file := openFiles[index]
	newpos, err := file.file.Seek(pos, whence)
	if err != nil { return g.UndefinedObject() }

	return g.NumberObject(float64(newpos))
}

func fileInfoToObject(fi os.FileInfo) g.Object {
	isDir := 0.0
	if fi.IsDir() { isDir = 1.0 }

	return g.MapObject(map[string]g.Object{
		"name": g.StringObject(fi.Name()),
		"size": g.NumberObject(float64(fi.Size())),
		"isDir": g.NumberObject(isDir),
	})
}

func stat(scope g.Scope, args []g.Object) g.Object {
	arguments := g.EvalArgs(scope, args)
	filename, _ := g.ToString(arguments[0])

	fileinfo, err := os.Stat(filename)
	if err != nil { return g.UndefinedObject() }

	return fileInfoToObject(fileinfo)
}

func readDir(scope g.Scope, args []g.Object) g.Object {
	arguments := g.EvalArgs(scope, args)
	dirname, _ := g.ToString(arguments[0])

	dirinfo, err := ioutil.ReadDir(dirname);
	if err != nil { return g.UndefinedObject() }

	contents := g.List{}
	for _, file := range dirinfo {
		contents.Append(fileInfoToObject(file))
	}

	// TODO write a list-object contructor?
	return g.Object{
		Type: g.ObjectTypeList,
		Elements: contents,
	}
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
	"create": g.BuiltinFunctionObject("create", create),
	"remove": g.BuiltinFunctionObject("remove", remove),
	"removeAll": g.BuiltinFunctionObject("removeAll", removeAll),
	"mkdir": g.BuiltinFunctionObject("mkdir", mkdir),
	"read": g.BuiltinFunctionObject("read", read),
	"readAll": g.BuiltinFunctionObject("readAll", readAll),
	"readUntil": g.BuiltinFunctionObject("readUntil", readUntil),
	"write": g.BuiltinFunctionObject("write", write),
	"seek": g.BuiltinFunctionObject("seek", seek),
	"stat": g.BuiltinFunctionObject("stat", stat),
	"readDir": g.BuiltinFunctionObject("readDir", readDir),
	"exit": g.BuiltinFunctionObject("exit", exit),
})

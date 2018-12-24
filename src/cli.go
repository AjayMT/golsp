
// CLI

package main

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"github.com/ajaymt/golsp/src/golsp"
)

func main() {
	filename := "-"
	dirname := "."
	file := os.Stdin
	var args []string

	if len(os.Args) > 1 {
		filename = os.Args[1]
		args = os.Args[2:]
	}

	if filename != "-" {
		filename, _ = filepath.Abs(filename)
		dirname = filepath.Dir(filename)
		file, _ = os.Open(filename)
	}

	input, _ := ioutil.ReadAll(file)
	golsp.Run(dirname, filename, args, string(input))
}


// CLI

package main

import (
	"io/ioutil"
	"os"
	// "fmt"
)

func main() {
	input, _ := ioutil.ReadAll(os.Stdin)

	InitializeBuiltins()
	tokens := Tokenize(string(input))
	tree := MakeST(tokens)
	Eval(Builtins, tree)
	// fmt.Println(PrintST(tree))
}

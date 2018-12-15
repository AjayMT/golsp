
// CLI

package main

import (
	"io/ioutil"
	"os"
	"github.com/ajaymt/golsp/src/golsp"
)

func main() {
	input, _ := ioutil.ReadAll(os.Stdin)

	golsp.InitializeBuiltins()
	tokens := golsp.Tokenize(string(input))
	tree := golsp.MakeST(tokens)
	golsp.Eval(golsp.Builtins, tree)
}

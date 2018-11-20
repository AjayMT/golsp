
// CLI

package main

import (
	"fmt"
	"io/ioutil"
	"os"
)

func main() {
	input, _ := ioutil.ReadAll(os.Stdin)

	tokens := Tokenize(string(input))
	tree := MakeST(tokens)
	fmt.Printf("Syntax tree:\n%v\n", PrintST(tree))
}

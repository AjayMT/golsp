
// CLI

package main

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"github.com/ajaymt/golsp/src/golsp"
)

func main() {
	input, _ := ioutil.ReadAll(os.Stdin)
	dirname, _ := filepath.Abs(".")
	golsp.Run(dirname, string(input))
}

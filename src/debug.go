
// Debugging functions

package main

import (
	"strconv"
	"strings"
)

func PrintST(root STNode) string {
	str := ""

	str += "\nHead: \"" + root.Head +
		"\"\nType: " + strconv.Itoa(int(root.Type)) +
		"\nChildren: ("

	for _, child := range root.Children {
		childstr := PrintST(child)
		lines := strings.Split(childstr, "\n")
		for i := 0; i < len(lines); i++ {
			lines[i] = "  " + lines[i]
		}

		str += strings.Join(lines, "\n")
	}

	str += "\n),"

	return str
}

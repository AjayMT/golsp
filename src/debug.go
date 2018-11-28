
// Debugging functions

package main

import (
	"strconv"
	"strings"
)

func PrintElements(list GolspObject) string {
	str := ""

	for _, elem := range list.Elements {
		if elem.Type == GolspObjectTypeList {
			str += "{ " + PrintElements(elem) + " }"
		}

		str += " " + elem.Value.Head
	}

	return str
}

func PrintST(root STNode) string {
	str := ""

	str += "\nHead: \"" + root.Head +
		"\"\nType: " + strconv.Itoa(int(root.Type)) +
		"\nSpread: " + strconv.FormatBool(root.Spread) +
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

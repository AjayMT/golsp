
package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"strconv"
	"unicode"
)

type STNodeType int

const (
	Program STNodeType = 0
	Expr STNodeType = 1
	String STNodeType = 2
	Number STNodeType = 3
	Identifier STNodeType = 4
)

type STNode struct {
	Head string
	Type STNodeType
	Children []STNode
}

var Delimiters = map[string]string{
	"": "",
	"[": "]",
	"]": "[",
	"\"": "\"",
}

var DelimiterTypes = map[string]STNodeType{
	"": Program,
	"[": Expr,
	"\"": String,
}

func MakeST(tokens []string) STNode {
	root, _ := makeST(tokens[0], tokens[1:])
	return root
}

func makeST(head string, tokens []string) (STNode, []string) {
	nodetype, delimiter := DelimiterTypes[head]

	current := STNode{
		Head: head,
		Type: nodetype,
		Children: make([]STNode, 0),
	}

	if !delimiter {
		current.Type = Identifier
		_, err := strconv.ParseFloat(head, 64)

		if err == nil {
			current.Type = Number
		}

		return current, tokens
	}

	for len(tokens) > 0 {
		token := tokens[0]
		tokens = tokens[1:]

		if token == Delimiters[current.Head] {
			return current, tokens
		}

		newchildren, newtokens := makeST(token, tokens)
		current.Children = append(current.Children, newchildren)
		tokens = newtokens
	}

	return current, tokens
}

func Tokenize(input string) []string {
	input = strings.TrimSpace(input)
	runes := []rune(input)

	token := ""
	tokens := []string{token}

	for _, r := range runes {
		_, delimiter := Delimiters[string(r)]
		if !delimiter && !unicode.IsSpace(r) {
			token += string(r)
			continue
		}

		if len(token) > 0 {
			tokens = append(tokens, token)
		}

		token = ""
		if delimiter {
			tokens = append(tokens, string(r))
		}
	}

	tokens = append(tokens, "")

	return tokens
}

func printST(root STNode) string {
	str := ""

	str += "\nHead: \"" + root.Head +
		"\"\nType: " + strconv.Itoa(int(root.Type)) +
		"\nChildren: ("

	for _, child := range root.Children {
		childstr := printST(child)
		lines := strings.Split(childstr, "\n")
		for i := 0; i < len(lines); i++ {
			lines[i] = "  " + lines[i]
		}

		str += strings.Join(lines, "\n")
	}

	str += "\n),"

	return str
}

func main() {
	input, _ := ioutil.ReadAll(os.Stdin)

	tokens := Tokenize(string(input))
	fmt.Printf("tokens: %v\n", tokens)

	tree := MakeST(tokens)
	fmt.Printf("Syntax tree:\n%v\n", printST(tree))
}

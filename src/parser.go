
// Parser

package main

import (
	"strings"
	"strconv"
	"unicode"
)

type OperatorType int
const (
	OperatorTypeSpread OperatorType = 0
	OperatorTypeZip OperatorType = 1
)

var Operators = []string{"...", ":"}
var OperatorTypes = map[string]OperatorType{
	"...": OperatorTypeSpread,
	":": OperatorTypeZip,
}

var LiteralDelimiters = map[string]string{"\"": "\"", "#": "\n"}
var LiteralDelimiterTypes = map[string]STNodeType{
	"\"": STNodeTypeStringLiteral,
	"#": STNodeTypeComment,
}

var TokenDelimiters = map[string]string{
	"": "",
	"[": "]",
	"]": "",
	"{": "}",
	"}": "",
}
var TokenDelimiterTypes = map[string]STNodeType{
	"": STNodeTypeScope,
	"[": STNodeTypeExpression,
	"{": STNodeTypeList,
}

// MakeST: construct a syntax tree from a list of tokens
// `tokens`: list of tokens to parse
func MakeST(tokens []string) STNode {
	root := STNode{Type: STNodeTypeScope}
	root.Children, _ = makeST(tokens[0], tokens[1:])

	return pruneComments(root)
}

// makeST: recursively construct a syntax tree from a list of tokens
// `delim`: the leading delimeter of the current expression
// `tokens`: remaining tokens to parse
// this function returns a list of nodes within the current expression
// and a list of remaining unparsed tokens
func makeST(delim string, tokens []string) ([]STNode, []string) {
	nodes := make([]STNode, 0, len(tokens))
	i := 0

	for ; i < len(tokens); i++ {
		if tokens[i] == TokenDelimiters[delim] { return nodes, tokens[i + 1:] }

		current := STNode{
			Head: tokens[i],
			Type: STNodeTypeIdentifier,
			Children: make([]STNode, 0),
		}

		// check if current token is a delimiter '[]' or '{}'
		// parse recursively if so
		delimtype, isDelimiter := TokenDelimiterTypes[current.Head]
		if isDelimiter {
			var newtokens []string
			current.Type = delimtype
			current.Children, newtokens = makeST(current.Head, tokens[i + 1:])
			i = -1
			tokens = newtokens
			nodes = append(nodes, current)
			continue
		}

		// check if current token is an extended literal i.e a string or comment
		literaltype, isLiteral := LiteralDelimiterTypes[string(current.Head[0])]
		if isLiteral {
			current.Type = literaltype
			nodes = append(nodes, current)
			continue
		}

		// check if current token is a number literal
		_, err := strconv.ParseFloat(current.Head, 64)
		if err == nil {
			current.Type = STNodeTypeNumberLiteral
			nodes = append(nodes, current)
			continue
		}

		// check if current token is an operator
		optype, isOperator := OperatorTypes[current.Head]
		if isOperator && len(nodes) > 0 {
			switch optype {
			case OperatorTypeSpread:
				nodes[len(nodes) - 1].Spread = true
			case OperatorTypeZip:
				nodes[len(nodes) - 1].Zip = true
			}
			continue
		}

		// current token must be an identifier
		nodes = append(nodes, current)
	}

	return nodes, tokens[i:]
}

// pruneComments: remove all comment nodes from a syntax tree
// `root`: root node of the syntax tree
func pruneComments(root STNode) STNode {
	newchildren := make([]STNode, 0, len(root.Children))
	for _, child := range root.Children {
		if child.Type == STNodeTypeComment { continue }
		newchildren = append(newchildren, pruneComments(child))
	}

	root.Children = newchildren

	return root
}

// parseLiteral: parse an extended literal, i.e a string or comment
// `delimiter`: leading delimiter of literal, either '"' or '#'
// `input`: list of unparsed characters following delimiter
// this function returns the number of characters it has parsed
// and a literal token
func parseLiteral(delimiter string, input []rune) (int, string) {
	escape := '\\'
	str := ""
	i := 0

	for ; i < len(input); i++ {
		if input[i] == escape {
			str += string(input[i])
			i++
			str += string(input[i])
			continue
		}

		if string(input[i]) == LiteralDelimiters[delimiter] {
			str += LiteralDelimiters[delimiter]
			i++
			break
		}

		str += string(input[i])
	}

	return i, str
}

// matchOperator: check if a list of characters contains an operator
// and find the correct operator if so
// `runes`: list of characters
// `index`: index at which to begin searching runes for an operator
// this function returns the index of the found operator in `Operators`
// (defined above) or -1 if runes does not contain an operator immediately
// after index
func matchOperator(runes []rune, index int) int {
	matchindex := -1
	matchscore := 0
	r := runes[index]

	for i, op := range Operators {
		score := 0
		if r != rune(op[0]) { continue }
		if index + len(op) > len(runes) { continue }

		opstr := string(runes[index:index + len(op)])
		if op == opstr { score = len(op) }

		if score > matchscore {
			matchscore = score
			matchindex = i
		}
	}

	return matchindex
}

// Tokenize: tokenize a string
// `input`: the string to tokenize
// this function returns a list of tokens
func Tokenize(input string) []string {
	input = strings.TrimSpace(input)
	runes := []rune(input)
	token := ""
	tokens := []string{token}

	for i := 0; i < len(runes); i++ {
		r := runes[i]
		_, literal := LiteralDelimiters[string(r)]
		if literal {
			if len(token) > 0 {
				tokens = append(tokens, token)
				token = ""
			}

			len, str := parseLiteral(string(r), runes[i + 1:])
			i += len
			tokens = append(tokens, string(r) + str)
			continue
		}

		opindex := matchOperator(runes, i)
		if opindex != -1 {
			if len(token) > 0 {
				tokens = append(tokens, token)
				token = ""
			}

			op := Operators[opindex]
			i += len(op) - 1
			tokens = append(tokens, op)
			continue
		}

		_, delimiter := TokenDelimiters[string(r)]
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

	if len(token) > 0 { tokens = append(tokens, token) }
	tokens = append(tokens, "")

	return tokens
}

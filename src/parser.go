
// Parser

package main

import (
	"strings"
	"strconv"
	"unicode"
)

var LiteralDelimiters = map[string]string{
	"\"": "\"",
	"#": "\n",
}

type OperatorType int

const (
	OperatorTypeSpread OperatorType = 0
	OperatorTypeZip OperatorType = 1
)

var Operators = []string{
	"...",
	":",
}

var OperatorTypes = map[string]OperatorType{
	"...": OperatorTypeSpread,
	":": OperatorTypeZip,
}

const LiteralEscape = '\\'

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

func MakeST(tokens []string) STNode {
	root, _ := makeST(tokens[0], tokens[1:])

	return pruneComments(root)
}

func pruneComments(root STNode) STNode {
	var newchildren []STNode
	for _, child := range root.Children {
		if child.Type == STNodeTypeComment {
			continue
		}

		newchildren = append(newchildren, pruneComments(child))
	}

	root.Children = newchildren

	return root
}

func makeST(head string, tokens []string) (STNode, []string) {
	nodetype, delimiter := TokenDelimiterTypes[head]
	_, operator := OperatorTypes[head]

	current := STNode{
		Head: head,
		Type: nodetype,
		Children: make([]STNode, 0),
	}

	if !delimiter && !operator {
		current.Type = STNodeTypeIdentifier
		_, err := strconv.ParseFloat(head, 64)

		if err == nil {
			current.Type = STNodeTypeNumberLiteral
		}

		literaltype, literal := LiteralDelimiterTypes[string(head[0])]

		if literal {
			current.Type = literaltype
		}

		nextoptype, nextop := OperatorTypes[tokens[0]]
		if nextop && nextoptype == OperatorTypeSpread {
			current.Spread = true
			tokens = tokens[1:]
		}

		return current, tokens
	}

	for len(tokens) > 0 {
		token := tokens[0]
		tokens = tokens[1:]

		if token == TokenDelimiters[current.Head] {
			if len(tokens) > 0 {
				nextoptype, nextop := OperatorTypes[tokens[0]]
				if nextop && nextoptype == OperatorTypeSpread {
					current.Spread = true
					tokens = tokens[1:]
				}
			}

			return current, tokens
		}

		newchildren, newtokens := makeST(token, tokens)
		current.Children = append(current.Children, newchildren)
		tokens = newtokens
	}

	return current, tokens
}

func parseLiteral(delimiter string, input []rune) (int, string) {
	str := ""
	i := 0

	for ; i < len(input); i++ {
		if input[i] == LiteralEscape {
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

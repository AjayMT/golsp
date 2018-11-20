
// Parser

package main

import (
	"strings"
	"strconv"
	"unicode"
)

type STNodeType int

const (
	STNodeTypeProgram STNodeType = 0
	STNodeTypeExpression STNodeType = 1
	STNodeTypeStringLiteral STNodeType = 2
	STNodeTypeNumberLiteral STNodeType = 3
	STNodeTypeIdentifier STNodeType = 4
	STNodeTypeComment = 5
)

type STNode struct {
	Head string
	Type STNodeType
	Children []STNode
}

var LiteralDelimiters = map[string]string{
	"\"": "\"",
	"#": "\n",
}

var LiteralEscape = '\\'

var LiteralDelimiterTypes = map[string]STNodeType{
	"\"": STNodeTypeStringLiteral,
	"#": STNodeTypeComment,
}

var TokenDelimiters = map[string]string{
	"": "",
	"[": "]",
	"]": "",
}

var TokenDelimiterTypes = map[string]STNodeType{
	"": STNodeTypeProgram,
	"[": STNodeTypeExpression,
}

func MakeST(tokens []string) STNode {
	root, _ := makeST(tokens[0], tokens[1:])
	return root
}

func makeST(head string, tokens []string) (STNode, []string) {
	nodetype, delimiter := TokenDelimiterTypes[head]

	current := STNode{
		Head: head,
		Type: nodetype,
		Children: make([]STNode, 0),
	}

	if !delimiter {
		current.Type = STNodeTypeIdentifier
		_, err := strconv.ParseFloat(head, 64)

		if err == nil {
			current.Type = STNodeTypeNumberLiteral
		}

		literaltype, literal := LiteralDelimiterTypes[string(head[0])]

		if literal {
			current.Type = literaltype
		}

		return current, tokens
	}

	for len(tokens) > 0 {
		token := tokens[0]
		tokens = tokens[1:]

		if token == TokenDelimiters[current.Head] {
			return current, tokens
		}

		newchildren, newtokens := makeST(token, tokens)
		current.Children = append(current.Children, newchildren)
		tokens = newtokens
	}

	return current, tokens
}

func parseLiteral(escape rune, delimiter string, input []rune) (int, string) {
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

func Tokenize(input string) []string {
	input = strings.TrimSpace(input)
	runes := []rune(input)

	token := ""
	tokens := []string{token}

	for i := 0; i < len(runes); i++ {
		r := runes[i]
		_, literal := LiteralDelimiters[string(r)]

		if literal {
			len, str := parseLiteral(LiteralEscape, string(r), runes[i + 1:])
			i += len
			tokens = append(tokens, string(r) + str)
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

	tokens = append(tokens, "")

	return tokens
}

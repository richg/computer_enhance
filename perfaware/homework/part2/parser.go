package main

import (
	"fmt"
)

type NodeType = uint8

const (
	ObjectNode NodeType = iota
	ArrayNode
	NumberNode
	StringLiteralNode
	TrueNode
	FalseNode
	NullNode
)

type Object struct {
	NodeType NodeType
	Members  []Member
}

type Member struct {
	Name  string
	Value any
}

type Array struct {
	NodeType NodeType
	items    []any
}

type Number struct {
	NodeType NodeType
	items    []any
}

type String struct {
	NodeType NodeType
	value    string
}

type Bool struct {
	NodeType NodeType
	value    bool
}

type Null struct {
	NodeType NodeType
	value    byte
}

var posPowers = [20]uint64{
	1,
	10,
	10 * 10,
	10 * 10 * 10,
	10 * 10 * 10 * 10,
	10 * 10 * 10 * 10 * 10,
	10 * 10 * 10 * 10 * 10 * 10,
	10 * 10 * 10 * 10 * 10 * 10 * 10,
	10 * 10 * 10 * 10 * 10 * 10 * 10 * 10,
	10 * 10 * 10 * 10 * 10 * 10 * 10 * 10 * 10,
	10 * 10 * 10 * 10 * 10 * 10 * 10 * 10 * 10 * 10,
	10 * 10 * 10 * 10 * 10 * 10 * 10 * 10 * 10 * 10 * 10,
	10 * 10 * 10 * 10 * 10 * 10 * 10 * 10 * 10 * 10 * 10 * 10,
	10 * 10 * 10 * 10 * 10 * 10 * 10 * 10 * 10 * 10 * 10 * 10 * 10,
	10 * 10 * 10 * 10 * 10 * 10 * 10 * 10 * 10 * 10 * 10 * 10 * 10 * 10,
	10 * 10 * 10 * 10 * 10 * 10 * 10 * 10 * 10 * 10 * 10 * 10 * 10 * 10 * 10,
	10 * 10 * 10 * 10 * 10 * 10 * 10 * 10 * 10 * 10 * 10 * 10 * 10 * 10 * 10 * 10,
	10 * 10 * 10 * 10 * 10 * 10 * 10 * 10 * 10 * 10 * 10 * 10 * 10 * 10 * 10 * 10 * 10,
	10 * 10 * 10 * 10 * 10 * 10 * 10 * 10 * 10 * 10 * 10 * 10 * 10 * 10 * 10 * 10 * 10 * 10,
	10 * 10 * 10 * 10 * 10 * 10 * 10 * 10 * 10 * 10 * 10 * 10 * 10 * 10 * 10 * 10 * 10 * 10 * 10,
}

var negPowers = [20]float64{
	1.0 / 10,
	1.0 / 10 / 10,
	1.0 / 10 / 10 / 10,
	1.0 / 10 / 10 / 10 / 10,
	1.0 / 10 / 10 / 10 / 10 / 10,
	1.0 / 10 / 10 / 10 / 10 / 10 / 10,
	1.0 / 10 / 10 / 10 / 10 / 10 / 10 / 10,
	1.0 / 10 / 10 / 10 / 10 / 10 / 10 / 10 / 10,
	1.0 / 10 / 10 / 10 / 10 / 10 / 10 / 10 / 10 / 10,
	1.0 / 10 / 10 / 10 / 10 / 10 / 10 / 10 / 10 / 10 / 10,
	1.0 / 10 / 10 / 10 / 10 / 10 / 10 / 10 / 10 / 10 / 10 / 10,
	1.0 / 10 / 10 / 10 / 10 / 10 / 10 / 10 / 10 / 10 / 10 / 10 / 10,
	1.0 / 10 / 10 / 10 / 10 / 10 / 10 / 10 / 10 / 10 / 10 / 10 / 10 / 10,
	1.0 / 10 / 10 / 10 / 10 / 10 / 10 / 10 / 10 / 10 / 10 / 10 / 10 / 10 / 10,
	1.0 / 10 / 10 / 10 / 10 / 10 / 10 / 10 / 10 / 10 / 10 / 10 / 10 / 10 / 10 / 10,
	1.0 / 10 / 10 / 10 / 10 / 10 / 10 / 10 / 10 / 10 / 10 / 10 / 10 / 10 / 10 / 10 / 10,
	1.0 / 10 / 10 / 10 / 10 / 10 / 10 / 10 / 10 / 10 / 10 / 10 / 10 / 10 / 10 / 10 / 10 / 10,
	1.0 / 10 / 10 / 10 / 10 / 10 / 10 / 10 / 10 / 10 / 10 / 10 / 10 / 10 / 10 / 10 / 10 / 10 / 10,
	1.0 / 10 / 10 / 10 / 10 / 10 / 10 / 10 / 10 / 10 / 10 / 10 / 10 / 10 / 10 / 10 / 10 / 10 / 10 / 10,
	1.0 / 10 / 10 / 10 / 10 / 10 / 10 / 10 / 10 / 10 / 10 / 10 / 10 / 10 / 10 / 10 / 10 / 10 / 10 / 10 / 10,
}

type TokenScanner struct {
	current   int
	tokens    []JsonToken
	numTokens int
}

func (s *TokenScanner) Peek() JsonToken {
	if s.current >= s.numTokens {
		return JsonToken{}
	}
	return s.tokens[s.current]
}

func getExponentVal(n uint64) uint64 {
	res := uint64(1)
	for i := uint64(0); i < n; {
		res *= 10
		i++
	}
	return res
}

func parseNumber(val []byte) any {
	var sign = 1
	intPartStart := 0
	intPartEnd := len(val)
	if val[0] == '-' {
		sign = -1
		intPartStart = 1
	}
	fractionPartStart := -1
	fractionPartEnd := len(val)
	exponentPartStart := -1
	for i, c := range val {
		switch c {
		case '.':
			intPartEnd = i
			fractionPartStart = i + 1
		case 'e', 'E':
			fractionPartEnd = i
			exponentPartStart = i + 1
		}
	}
	i := intPartEnd - 1
	pow := 0
	var intPart uint64 = 0
	for i >= intPartStart {
		intPart += (uint64(val[i]) - 48) * posPowers[pow]
		pow++
		i--
	}
	var fractionPart float64 = 0.0
	if fractionPartStart > -1 {
		i := fractionPartStart
		pow := 0
		for i < fractionPartEnd {
			fractionPart += float64(val[i]-48) * negPowers[pow]
			pow++
			i++
		}
	}
	var expPart uint64 = 0
	if exponentPartStart > -1 {
		i := len(val) - 1
		pow := 0
		for i >= exponentPartStart {
			expPart += (uint64(val[i]) - 48) * posPowers[pow]
			pow++
			i--
		}
	}
	if fractionPartStart == -1 {
		// TODO: handle negative exponents
		return uint64(sign) * intPart * getExponentVal(expPart)
	} else {
		return float64(sign) * (float64(intPart) + fractionPart) * float64(getExponentVal(expPart))
	}

}

func parseObject(s *TokenScanner) map[string]any {
	o := make(map[string]any)
	if s.Peek().Type != OpenCurlyToken {
		panic(fmt.Sprintf("Unexpceted token type when parsing object %s", string(s.Peek().Type)))
	}
	s.current++
	t := s.Peek()
	i := 0
	var memberName string
	for t.Type != CloseCurlyToken && t.Value != nil {
		if i%4 == 0 { // key, colon, value, comma
			if t.Type != StringLiteralToken {
				panic(fmt.Sprintf("Failed to parse object - expected member attr name got token %s", t.String()))
			}
			memberName = string(t.Value)
		} else if i%2 == 0 {
			o[memberName] = parseValue(s, t)
		}
		s.current++
		t = s.Peek()
		i++
	}
	if t.Value == nil {
		panic("Parse error: no closing brace for object")
	}
	return o
}

func parseArray(s *TokenScanner) []any {
	a := make([]any, 0)
	if s.Peek().Type != OpenBraceToken {
		panic(fmt.Sprintf("Unexpceted token type when parsing object %s", string(s.Peek().Type)))
	}
	s.current++
	t := s.Peek()
	i := 0
	for t.Type != CloseBraceToken && t.Value != nil {
		if (i+1)%2 == 0 {
			if t.Type != CommaToken {
				panic(fmt.Sprintf("Failed to parse array - expected separator, got %s", t.String()))
			}
		} else {
			a = append(a, parseValue(s, t))
		}
		s.current++
		t = s.Peek()
		i++
	}
	if t.Value == nil {
		panic("Parse error: no closing brace for array")
	}
	return a
}

func parseValue(s *TokenScanner, t JsonToken) any {
	switch t.Type {
	case OpenCurlyToken:
		return parseObject(s)
	case OpenBraceToken:
		return parseArray(s)
	case TrueToken:
		return true
	case FalseToken:
		return false
	case NullToken:
		return nil
	case StringLiteralToken:
		return string(t.Value)
	case NumberToken:
		return parseNumber(t.Value)
	default:
		panic(fmt.Sprintf("why tho? %s", t.String()))
	}
}

func Parse(tokens []JsonToken) any {
	s := TokenScanner{0, tokens, len(tokens)}
	t := s.Peek()
	return parseValue(&s, t)
}

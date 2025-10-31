package main

import (
	"fmt"
	"os"
)

type TokenType = uint8

const (
	OpenBraceToken TokenType = iota
	CloseBraceToken
	OpenCurlyToken
	CloseCurlyToken
	ColonToken
	CommaToken
	NullToken
	TrueToken
	FalseToken
	NumberToken
	StringLiteralToken
)

var TokenNameLookup = [11]string{
	"OpenBrace",
	"CloseBrace",
	"OpenCurly",
	"CloseCurly",
	"Colon",
	"Comma",
	"Null",
	"True",
	"False",
	"Number",
	"StringLiteral",
}

const (
	EOF byte = 0x04
)

type JsonToken struct {
	Type  TokenType
	Value []byte
}

func (t *JsonToken) String() string {
	return fmt.Sprintf("%s %s", TokenNameLookup[t.Type], string(t.Value))
}

type Scanner struct {
	current int
	data    []byte
	dataLen int
}

func (s *Scanner) Continue() {
	s.current++
}

func (s *Scanner) Peek() byte {
	if s.current >= s.dataLen {
		return EOF
	}
	return s.data[s.current]
}

func (s *Scanner) PeekNext() byte {
	if s.current+1 >= s.dataLen {
		return EOF
	}
	return s.data[s.current+1]
}

func (s *Scanner) ConsumeByte() []byte {
	prev := s.current
	s.current++
	return s.data[prev:s.current]
}

func (s *Scanner) ConsumeStringLiteral() []byte {
	start := s.current
	s.current++
	var c = s.Peek()
	for c != EOF {
		if c == '"' && s.data[s.current-1] != '/' {
			s.current++
			// Exclude the quotes
			return s.data[start+1 : s.current-1]
		}
		s.current++
		c = s.Peek()
	}
	panic(fmt.Sprintf("EOF: Failed to find closing quote for string starting at %d", start))
}

func (s *Scanner) ConsumeDigits() {
	var c = s.Peek()
	if !(c >= '0' && c <= '9') {
		panic(fmt.Sprintf("Invalid number at loc %d, expected digit got %s", s.current, string(c)))
	}
	for c != EOF {
		if !(c >= '0' && c <= '9') {
			break
		}
		s.current++
		c = s.Peek()
	}
}

func (s *Scanner) ConsumeNumber() []byte {
	start := s.current
	var c = s.Peek()
	if c == '-' {
		n := s.PeekNext()
		if !(n >= '0' && n <= '9') {
			panic(fmt.Sprintf("Invalid number at %d - digits must follow a '-'", start))
		}
		s.current++
		c = s.Peek()
	} else if c == '0' {
		n := s.PeekNext()
		if n != '.' {
			panic(fmt.Sprintf("Invalid number at %d - leading zero must be followed by '.'", start))
		}
		s.current++
		c = s.Peek()
	}
	var EON = false
	for c != EOF && !EON {
		switch c {
		case '.':
			n := s.PeekNext()
			if !(n >= '0' && n <= '9') {
				panic(fmt.Sprintf("Invalid number at %d - digits must follow a '.'", start))
			}
			s.current++
		case 'e', 'E':
			n := s.PeekNext()
			if n == '+' || n == '-' {
				s.current++
			}
			n = s.PeekNext()
			if !(n >= '0' && n <= '9') {
				panic(fmt.Sprintf("Invalid number at %d - exponent notation requires =/- or digits after E", start))
			}
			s.current++
		case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
			s.ConsumeDigits()
		default:
			EON = true
		}
		c = s.Peek()
	}
	return s.data[start:s.current]
}

func (s *Scanner) ConsumeKeyword(match []byte) []byte {
	start := s.current
	for i := 0; i < len(match); i++ {
		if match[i] != s.Peek() {
			panic(fmt.Sprintf("Invalid token at %d", s.current))
		}
		s.current++
	}
	return s.data[start:s.current]
}

func Tokenize(fn string) []JsonToken {
	data, err := os.ReadFile(fn)
	if err != nil {
		panic("Invalid file")
	}
	tokens := make([]JsonToken, 0, 1000)
	scanner := Scanner{0, data, len(data)}
	for scanner.current < scanner.dataLen {
		currentChar := scanner.Peek()
		var tokenType TokenType
		var tokenVal []byte = nil
		switch currentChar {
		case '\t', ' ', '\n':
			// Should implement ConsumeWhitespace
			scanner.Continue()
		case '[':
			tokenType = OpenBraceToken
			tokenVal = scanner.ConsumeByte()
		case ']':
			tokenType = CloseBraceToken
			tokenVal = scanner.ConsumeByte()
		case '{':
			tokenType = OpenCurlyToken
			tokenVal = scanner.ConsumeByte()
		case '}':
			tokenType = CloseCurlyToken
			tokenVal = scanner.ConsumeByte()
		case ':':
			tokenType = ColonToken
			tokenVal = scanner.ConsumeByte()
		case ',':
			tokenType = CommaToken
			tokenVal = scanner.ConsumeByte()
		case '"':
			tokenType = StringLiteralToken
			tokenVal = scanner.ConsumeStringLiteral()
		case 'n':
			tokenType = NullToken
			tokenVal = scanner.ConsumeKeyword([]byte("null"))
		case 't':
			tokenType = TrueToken
			tokenVal = scanner.ConsumeKeyword([]byte("true"))
		case 'f':
			tokenType = FalseToken
			tokenVal = scanner.ConsumeKeyword([]byte("false"))
		default:
			c := scanner.Peek()
			if c == '-' || (c >= '0' && c <= '9') {
				tokenType = NumberToken
				tokenVal = scanner.ConsumeNumber()
			} else {
				panic(fmt.Sprintf("Unparsable token at %d", scanner.current))

			}
		}
		if tokenVal != nil {
			token := JsonToken{
				tokenType,
				tokenVal,
			}
			tokens = append(tokens, token)
		}
	}
	return tokens
}

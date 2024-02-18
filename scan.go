package tplexpr

import (
	"errors"
	"fmt"
	"unicode"
)

//go:generate go run golang.org/x/tools/cmd/stringer -type TokenType -trimprefix Token
type TokenType int

const (
	TokenValue TokenType = iota
	TokenIdent
	TokenNumber
	TokenLeftParen
	TokenRightParen
	TokenDot
	TokenComma
	TokenEOF
	TokenString
	TokenArrow
	TokenDeclare
	TokenGT
	TokenGE
	TokenEQ
	TokenNE
	TokenLE
	TokenLT
	TokenAND
	TokenOR
	TokenADD
	TokenSUB
	TokenMUL
	TokenDIV
	TokenBlock
	TokenEndBlock
	TokenIf
	TokenThen
	TokenElse
	TokenElseIf
	TokenEndIf
	TokenFor
	TokenIn
	TokenDo
	TokenBreak
	TokenContinue
	TokenEndFor
	TokenInclude
	TokenDiscard
	TokenEndDiscard
	TokenError
)

type Token struct {
	Type  TokenType
	Start int
	End   int
	Value []byte
}

const (
	scanValue = iota
	scanExpr
	scanVar
)

var keywordMap = map[string]TokenType{
	"block":      TokenBlock,
	"endblock":   TokenEndBlock,
	"if":         TokenIf,
	"then":       TokenThen,
	"else":       TokenElse,
	"elseif":     TokenElseIf,
	"endif":      TokenEndIf,
	"for":        TokenFor,
	"in":         TokenIn,
	"do":         TokenDo,
	"break":      TokenBreak,
	"continue":   TokenContinue,
	"endfor":     TokenEndFor,
	"declare":    TokenDeclare,
	"include":    TokenInclude,
	"discard":    TokenDiscard,
	"enddiscard": TokenEndDiscard,
}

var (
	ErrSyntax = errors.New("syntax error")
)

type Scanner struct {
	Err   error
	mode  int
	pos   int
	input []byte
}

func NewScanner(input []byte) Scanner {
	return Scanner{
		input: input,
	}
}

func (s *Scanner) Scan() (t Token) {
beginScan:
	switch s.mode {
	case scanValue:
		t.Start = s.pos

		if s.pos >= len(s.input) {
			t.End = s.pos
			t.Type = TokenEOF
			return
		}

		t.Type = TokenValue
		value := []byte{}
		wasDollar := false

		for {
			if s.pos >= len(s.input) {
				t.End = s.pos
				t.Value = value

				if wasDollar {
					t.Type = TokenError
					s.Err = fmt.Errorf("%w: Endingh with $", ErrSyntax)
				}
				return
			}

			c := s.input[s.pos]

			if wasDollar {
				wasDollar = false
				switch {
				case c == '$':
					value = append(value, '$')
					s.pos += 1
				case c == '{':
					s.mode = scanExpr
					s.pos += 1
					if len(value) == 0 {
						goto beginScan
					}
					return
				case isIdentByte(c):
					s.mode = scanVar
					if len(value) == 0 {
						goto beginScan
					}
					return
				default:
					t.Type = TokenError
					t.End = s.pos
					s.Err = fmt.Errorf("%w: unexpected char after $", ErrSyntax)
					return
				}
			} else {
				switch c {
				case '$':
					s.pos += 1
					wasDollar = true
					t.End = s.pos
					t.Value = value
				default:
					s.pos += 1
					value = append(value, c)
				}
			}

		}
	case scanExpr:
		// skip whitespace
		for s.pos < len(s.input) && unicode.IsSpace(rune(s.input[s.pos])) {
			s.pos++
		}

		t.Start = s.pos

		if s.pos >= len(s.input) {
			t.Type = TokenError
			t.End = s.pos
			s.Err = fmt.Errorf("%w: EOF in expression", ErrSyntax)
			return
		}

		c := s.input[s.pos]
		switch c {
		case '}':
			s.pos += 1
			s.mode = scanValue
			goto beginScan

		case '%':
			if s.pos+1 < len(s.input) && s.input[s.pos+1] == '}' {
				s.pos += 2
				s.mode = scanValue

				for s.pos < len(s.input) && unicode.IsSpace(rune(s.input[s.pos])) {
					s.pos += 1
				}
				goto beginScan
			}
			t.End = s.pos
			t.Type = TokenError
			s.Err = s.errUnexpectedInput()
			return
		case '(':
			s.pos += 1
			t.End = s.pos
			t.Type = TokenLeftParen
			return
		case ')':
			s.pos += 1
			t.End = s.pos
			t.Type = TokenRightParen
			return
		case '.':
			s.pos += 1
			t.End = s.pos
			t.Type = TokenDot
			return
		case ',':
			s.pos += 1
			t.End = s.pos
			t.Type = TokenComma
			return
		case '=':
			switch {
			case s.pos+1 >= len(s.input):
				// nop
			case s.input[s.pos+1] == '>':
				s.pos += 2
				t.End = s.pos
				t.Type = TokenArrow
				return
			case s.input[s.pos+1] == '=':
				s.pos += 2
				t.End = s.pos
				t.Type = TokenEQ
				return
			}

			t.End = s.pos
			t.Type = TokenError
			s.Err = s.errUnexpectedInput()
			return
		case '>':
			if s.pos+1 < len(s.input) && s.input[s.pos+1] == '=' {
				s.pos += 2
				t.Type = TokenGT
			} else {
				s.pos += 1
				t.Type = TokenGE
			}
			t.End = s.pos
			return
		case '<':
			if s.pos+1 < len(s.input) && s.input[s.pos+1] == '=' {
				s.pos += 2
				t.Type = TokenLE
			} else {
				s.pos += 1
				t.Type = TokenLT
			}
			t.End = s.pos
			return
		case '!':
			if s.pos+1 < len(s.input) && s.input[s.pos+1] == '=' {
				s.pos += 2
				t.End = s.pos
				t.Type = TokenNE
				return
			}
			t.End = s.pos
			t.Type = TokenError
			s.Err = s.errUnexpectedInput()
			return
		case '&':
			if s.pos+1 < len(s.input) && s.input[s.pos+1] == '&' {
				s.pos += 2
				t.End = s.pos
				t.Type = TokenAND
				return
			}
			t.End = s.pos
			t.Type = TokenError
			s.Err = s.errUnexpectedInput()
			return
		case '|':
			if s.pos+1 < len(s.input) && s.input[s.pos+1] == '|' {
				s.pos += 2
				t.End = s.pos
				t.Type = TokenOR
				return
			}
			t.End = s.pos
			t.Type = TokenError
			s.Err = s.errUnexpectedInput()
			return
		case '+':
			if s.pos+1 < len(s.input) && unicode.IsDigit(rune(s.input[s.pos+1])) {
				t = s.scanNumber(c)
				return
			}
			s.pos += 1
			t.End = s.pos
			t.Type = TokenADD
			return
		case '-':
			if s.pos+1 < len(s.input) && unicode.IsDigit(rune(s.input[s.pos+1])) {
				t = s.scanNumber(c)
				return
			}
			s.pos += 1
			t.End = s.pos
			t.Type = TokenSUB
			return
		case '*':
			s.pos += 1
			t.End = s.pos
			t.Type = TokenMUL
			return
		case '/':
			s.pos += 1
			t.End = s.pos
			t.Type = TokenDIV
			return
		case '"', '\'':
			s.pos += 1
			quote := c
			esc := false
			value := []byte{}

			for {
				if s.pos >= len(s.input) {
					t.End = s.pos
					t.Type = TokenError
					s.Err = fmt.Errorf("%w: EOF in string literal", ErrSyntax)
					return
				}

				c = s.input[s.pos]

				if esc {
					esc = false
					switch c {
					case '\\', '"', '\'':
						s.pos += 1
						value = append(value, c)
					case 'r':
						s.pos += 1
						value = append(value, '\r')
					case 'n':
						s.pos += 1
						value = append(value, '\n')
					case 'b':
						s.pos += 1
						value = append(value, '\b')
					default:
						t.End = s.pos
						t.Type = TokenError
						s.Err = fmt.Errorf("%w: bad escape sequence (%c)", ErrSyntax, c)
						return
					}
				} else {
					switch c {
					case '\\':
						s.pos += 1
						esc = true
					case quote:
						s.pos += 1
						t.End = s.pos
						t.Type = TokenString
						t.Value = value
						return
					default:
						s.pos += 1
						value = append(value, c)
					}
				}
			}

		default:
			if isIdentStartByte(c) {
				value := []byte{c}
				s.pos += 1

				for s.pos < len(s.input) {
					c = s.input[s.pos]

					if isIdentByte(c) {
						s.pos += 1
						value = append(value, c)
					} else {
						break
					}
				}

				// check if we have a keyword
				if tt, ok := keywordMap[string(value)]; ok {
					t.Type = tt
				} else {
					t.Type = TokenIdent
				}

				t.End = s.pos
				t.Value = value
				return
			}
			if isNumberStartByte(c) {
				t = s.scanNumber(c)
				return
			}

			t.End = s.pos
			t.Type = TokenError
			s.Err = s.errUnexpectedInput()
			return
		}
	case scanVar:
		t.Start = s.pos
		value := []byte{}
		for s.pos < len(s.input) && isIdentByte(s.input[s.pos]) {
			value = append(value, s.input[s.pos])
			s.pos += 1
		}
		t.End = s.pos
		t.Type = TokenIdent
		t.Value = value
		s.mode = scanValue
		return
	default:
		panic("tplexpr: invalid scan mode")
	}
}

func (s *Scanner) scanNumber(startByte byte) (t Token) {
	c := startByte

	value := []byte{c}
	s.pos += 1

	for s.pos < len(s.input) {
		c = s.input[s.pos]

		if isNumberByte(c) {
			s.pos += 1
			value = append(value, c)
		} else {
			break
		}
	}

	t.End = s.pos
	t.Type = TokenNumber
	t.Value = value
	return
}

func (s *Scanner) errUnexpectedInput() error {
	return fmt.Errorf("%w: unexpected input: %.5s", ErrSyntax, s.input[s.pos:])
}

func isIdentByte(c byte) bool {
	switch {
	case c >= 'a' && c <= 'z', c >= 'A' && c <= 'Z', c >= '0' && c <= '9', c == '_':
		return true
	default:
		return false
	}
}

func isIdentStartByte(c byte) bool {
	switch {
	case c >= 'a' && c <= 'z', c >= 'A' && c <= 'Z', c == '_':
		return true
	default:
		return false
	}
}

func isNumberByte(c byte) bool {
	switch {
	case c >= '0' && c <= '9', c == '.', c == 'e', c == 'E', c == '+', c == '-':
		return true
	default:
		return false
	}
}

func isNumberStartByte(c byte) bool {
	switch {
	case c >= '0' && c <= '9', c == '+', c == '-':
		return true
	default:
		return false
	}
}

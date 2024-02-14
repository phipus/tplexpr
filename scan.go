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
	TokenLeftParen
	TokenRightParen
	TokenDot
	TokenComma
	TokenEOF
	TokenString
	TokenArrow
	TokenDeclare
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
		switch {
		case c == '}':
			s.pos += 1
			s.mode = scanValue
			goto beginScan
		case c == '(':
			s.pos += 1
			t.End = s.pos
			t.Type = TokenLeftParen
			return
		case c == ')':
			s.pos += 1
			t.End = s.pos
			t.Type = TokenRightParen
			return
		case c == '.':
			s.pos += 1
			t.End = s.pos
			t.Type = TokenDot
			return
		case c == ',':
			s.pos += 1
			t.End = s.pos
			t.Type = TokenComma
			return
		case c == '=':
			if s.pos+1 < len(s.input) && s.input[s.pos+1] == '>' {
				s.pos += 2
				t.End = s.pos
				t.Type = TokenArrow
				return
			}
			t.End = s.pos
			t.Type = TokenError
			s.Err = s.errUnexpectedInput()
			return
		case c == ':':
			if s.pos+1 < len(s.input) && s.input[s.pos+1] == '=' {
				s.pos += 2
				t.End = s.pos
				t.Type = TokenDeclare
				return
			}
			t.End = s.pos
			t.Type = TokenError
			s.Err = s.errUnexpectedInput()
			return
		case c == '"', c == '\'':
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
		case isIdentByte(c):
			value := []byte{c}
			s.pos += 1

		readIdent:
			for s.pos < len(s.input) {
				c = s.input[s.pos]
				switch {
				case isIdentByte(c):
					s.pos += 1
					value = append(value, c)
				default:
					break readIdent
				}
			}

			t.End = s.pos
			t.Type = TokenIdent
			t.Value = value
			return
		default:
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

package tplexpr

import (
	"testing"
)

func TestScanType(t *testing.T) {
	type testCase struct {
		input  string
		tokens []TokenType
	}

	testCases := []testCase{
		{"Hello World", []TokenType{TokenValue, TokenEOF}},
		{"Hello $World", []TokenType{TokenValue, TokenIdent, TokenEOF}},
		{"Hello $World!", []TokenType{TokenValue, TokenIdent, TokenValue, TokenEOF}},
		{`Hello ${"WORLD"}`, []TokenType{TokenValue, TokenString, TokenEOF}},
		{`$Hello World`, []TokenType{TokenIdent, TokenValue}},
		{`${Hello "World"}`, []TokenType{TokenIdent, TokenString, TokenEOF}},
		{`$v."$it"`, []TokenType{TokenIdent, TokenValue, TokenIdent, TokenValue, TokenEOF}},
		{`${v."$it"}`, []TokenType{TokenIdent, TokenDot, TokenString, TokenEOF}},
	}

	for _, testCase := range testCases {
		t.Logf("Scanning %s", testCase.input)

		s := NewScanner([]byte(testCase.input))

		for _, expected := range testCase.tokens {
			actual := s.Scan()
			if expected != actual.Type {
				t.Errorf("    Expected %s got %s", expected, actual.Type)
			}
		}
	}
}

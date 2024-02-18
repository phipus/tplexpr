package html

import (
	"io"

	"golang.org/x/net/html"
)

type Scanner struct {
	t     *html.Tokenizer
	l0    html.Token
	hasL0 bool
}

func NewScanner(r io.Reader) Scanner {
	return Scanner{t: html.NewTokenizer(r)}
}

func (s *Scanner) Token() html.Token {
	if !s.hasL0 {
		s.t.Next()
		s.l0 = s.t.Token()
		s.hasL0 = true
	}
	return s.l0
}

func (s *Scanner) Consume() {
	s.hasL0 = false
}

func (s *Scanner) Err() error {
	return s.t.Err()
}

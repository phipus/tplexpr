package tplexpr

import "fmt"

type Parser struct {
	s         Scanner
	lookahead []Token
}

func NewParser(input []byte) Parser {
	return Parser{
		s: NewScanner(input),
	}
}

func (p *Parser) lookAhead(n int) Token {
	for {
		if n < len(p.lookahead) {
			return p.lookahead[n]
		}
		if len(p.lookahead) > 0 {
			last := p.lookahead[len(p.lookahead)-1]
			switch last.Type {
			case TokenEOF, TokenError:
				return last
			}
		}

		t := p.s.Scan()
		p.lookahead = append(p.lookahead, t)
	}
}

func (p *Parser) getToken() (t Token) {
	return p.lookAhead(0)
}

func (p *Parser) consume() {
	if len(p.lookahead) == 1 {
		p.lookahead = p.lookahead[:0]
	} else {
		p.lookahead = p.lookahead[1:]
	}
}

func (p *Parser) errUnexpected(expected string) error {
	t := p.getToken()
	if t.Type == TokenError {
		return p.s.Err
	}
	if len(expected) > 0 {
		return fmt.Errorf("%w: unexpected token %s. Expected %s", ErrSyntax, t.Type, expected)
	}
	return fmt.Errorf("%w: unexpected token %s", ErrSyntax, t.Type)
}

func (p *Parser) parseAtom() (n Node, err error) {
	t := p.getToken()

	switch t.Type {
	case TokenValue:
		p.consume()
		n = &ValueNode{string(t.Value)}
		return
	case TokenIdent:
		p.consume()
		n = &VarNode{string(t.Value)}
		return
	case TokenString:
		p.consume()
		subp := NewParser(t.Value)
		n, err = subp.Parse()
		return
	case TokenLeftParen:
		// find the matching closing paren
		open := 1
		afterParen := Token{}
	lookAfterParen:
		for i := 1; ; i++ {
			t := p.lookAhead(i)
			switch t.Type {
			case TokenEOF, TokenError:
				err = p.errUnexpected(")")
				return
			case TokenLeftParen:
				open += 1
			case TokenRightParen:
				open -= 1
				if open <= 0 {
					afterParen = p.lookAhead(i + 1)
					break lookAfterParen
				}
			}
		}

		if afterParen.Type == TokenArrow {
			args := []string{}
			p.consume()
			for {
				t = p.getToken()
				if t.Type == TokenRightParen {
					break
				}
				if t.Type != TokenIdent {
					err = p.errUnexpected("identifier")
					return
				}
				args = append(args, string(t.Value))
				p.consume()

				t = p.getToken()
				if t.Type != TokenComma {
					break
				}
				p.consume()
			}

			// consume the closing brace
			if t.Type != TokenRightParen {
				err = p.errUnexpected(")")
				return
			}
			p.consume()

			// consume the arrow
			t = p.getToken()
			if t.Type != TokenArrow {
				err = p.errUnexpected("=>")
				return
			}
			p.consume()

			// consume the string
			t = p.getToken()
			if t.Type != TokenString {
				err = p.errUnexpected("string")
				return
			}
			p.consume()

			subp := NewParser(t.Value)
			n, err = subp.Parse()
			if err != nil {
				return
			}
			n = &SubprogNode{args, n}
		} else {
			p.consume()
			n, err = p.Parse()
			if err != nil {
				return
			}
			t = p.getToken()
			if t.Type != TokenRightParen {
				err = p.errUnexpected(")")
				return
			}
			p.consume()
		}
		return
	default:
		err = p.errUnexpected("")
		return

	}
}

func (p *Parser) parsePostfix() (n Node, err error) {
	n, err = p.parseAtom()
	if err != nil {
		return
	}

	for {
		t := p.getToken()

		switch t.Type {
		case TokenLeftParen:
			var args []Node
			args, err = p.parseArgList()
			if err != nil {
				return
			}

			if v, ok := n.(*VarNode); ok {
				n = &CallNode{v.Name, args}
			} else {
				err = fmt.Errorf("%w: Value is not callable", ErrSyntax)
				return
			}
		case TokenDot:
			p.consume()
			t = p.getToken()

			switch t.Type {
			case TokenIdent:
				name := string(t.Value)
				p.consume()

				t = p.getToken()
				if t.Type == TokenLeftParen {
					var args []Node
					args, err = p.parseArgList()
					if err != nil {
						return
					}
					n = &CallNode{name, append([]Node{n}, args...)}
				} else {
					n = &AttrNode{n, name}
				}
			default:
				err = p.errUnexpected("")
				return
			}
		default:
			return

		}
	}
}

func (p *Parser) parseArgList() (args []Node, err error) {
	t := p.getToken()
	if t.Type != TokenLeftParen {
		err = p.errUnexpected("(")
		return
	}
	p.consume()

	for {
		t = p.getToken()
		if t.Type == TokenRightParen {
			break
		}

		var arg Node
		arg, err = p.ParseExpr()
		if err != nil {
			return
		}
		args = append(args, arg)

		t = p.getToken()

		if t.Type != TokenComma {
			break
		}
		p.consume()
	}

	t = p.getToken()
	if t.Type != TokenRightParen {
		err = p.errUnexpected(")")
		return
	}
	p.consume()
	return
}

func (p *Parser) ParseExpr() (n Node, err error) {
	return p.parsePostfix()
}

func (p *Parser) Parse() (n Node, err error) {
	nodes := []Node{}

	for {
		t := p.getToken()
		if t.Type == TokenEOF {
			break
		}

		n, err = p.ParseExpr()
		if err != nil {
			return
		}

		nodes = append(nodes, n)
	}

	switch len(nodes) {
	case 0:
		n = &ValueNode{""}
	case 1:
		n = nodes[0]
	default:
		n = &EmitNode{nodes}
	}
	return
}

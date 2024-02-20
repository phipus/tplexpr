package tplexpr

import (
	"fmt"
	"strconv"
	"strings"
)

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
	case TokenIdent:
		p.consume()
		n = &VarNode{string(t.Value)}
		return
	case TokenString:
		p.consume()
		subp := NewParser(t.Value)
		n, err = subp.Parse()
		return
	case TokenNumber:
		p.consume()
		value := string(t.Value)
		_, err = strconv.ParseFloat(value, 64)
		n = &NumberNode{value}
		return
	case TokenObject:
		p.consume()

		t = p.getToken()
		if t.Type != TokenLeftParen {
			err = p.errUnexpected("(")
			return
		}
		p.consume()

		t = p.getToken()
		if t.Type == TokenRightParen {
			p.consume()
			n = &ObjectNode{}
			return
		}
		var extend Node
		if p.getToken().Type != TokenIdent || p.lookAhead(1).Type != TokenArrow {
			extend, err = p.ParseExpr()
			if err != nil {
				return
			}
			t = p.getToken()
			switch t.Type {
			case TokenRightParen:
				p.consume()
				n = &ObjectNode{extend, nil}
				return
			case TokenComma:
				p.consume()
			default:
				err = p.errUnexpected(", | )")
				return
			}
		}

		keys := []ObjectKey{}
		for {
			t = p.getToken()
			if t.Type == TokenRightParen {
				break
			}
			if t.Type != TokenIdent {
				err = p.errUnexpected("identifier")
				return
			}
			key := string(t.Value)
			p.consume()

			t = p.getToken()
			if t.Type != TokenArrow {
				err = p.errUnexpected("=>")
				return
			}
			p.consume()

			var value Node
			value, err = p.ParseExpr()
			if err != nil {
				return
			}

			keys = append(keys, ObjectKey{key, value})

			t = p.getToken()
			if t.Type != TokenComma {
				break
			}
			p.consume()
		}
		if t.Type != TokenRightParen {
			err = p.errUnexpected(")")
			return
		}
		p.consume()
		n = &ObjectNode{extend, keys}
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
			n, err = p.ParseExpr()
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
			case TokenThen:
				p.consume()
				var args []Node
				args, err = p.parseArgList()
				if err != nil {
					return
				}

				switch len(args) {
				case 0:
					n = &ValueNode{""}
				case 1:
					n = &OrNode{[]Node{
						&AndNode{[]Node{n, args[0]}},
						&ValueNode{""},
					}}
				case 2:
					n = &OrNode{[]Node{
						&AndNode{[]Node{n, args[0]}},
						args[1],
					}}
				default:
					err = fmt.Errorf("%w: .then reiquires 0 to 2 arguments", ErrSyntax)
				}
				return
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

func (p *Parser) parseBinaryOP(defs map[TokenType]int, next func() (Node, error)) (n Node, err error) {
	n, err = next()
	if err != nil {
		return
	}

	ops := []BinaryOP{}

	for {
		t := p.getToken()
		op, ok := defs[t.Type]
		if !ok {
			break
		}

		p.consume()

		e, err := next()
		if err != nil {
			return nil, err
		}
		ops = append(ops, BinaryOP{op, e})
	}

	if len(ops) > 0 {
		n = &BinaryOPNode{n, ops}
	}
	return
}

var factorDefs = map[TokenType]int{
	TokenMUL: MUL,
	TokenDIV: DIV,
}

func (p *Parser) parseFactor() (n Node, err error) {
	return p.parseBinaryOP(factorDefs, p.parsePostfix)
}

var termDefs = map[TokenType]int{
	TokenADD: ADD,
	TokenDIV: DIV,
}

func (p *Parser) parseTerm() (n Node, err error) {
	return p.parseBinaryOP(termDefs, p.parseFactor)
}

func (p *Parser) parseCompare() (n Node, err error) {
	n, err = p.parseTerm()
	if err != nil {
		return
	}

	t := p.getToken()
	cmp := 0
	switch t.Type {
	case TokenGT:
		cmp = GT
	case TokenGE:
		cmp = GE
	case TokenEQ:
		cmp = EQ
	case TokenNE:
		cmp = NE
	case TokenLE:
		cmp = LE
	case TokenLT:
		cmp = LT
	default:
		return
	}

	p.consume()
	r, err := p.parseTerm()
	if err != nil {
		return
	}

	n = &CompareNode{cmp, n, r}
	return
}

func (p *Parser) parseAND() (n Node, err error) {
	n, err = p.parseCompare()
	if err != nil {
		return
	}

	nodes := []Node{n}

	for {
		t := p.getToken()
		if t.Type != TokenAND {
			break
		}
		p.consume()

		n, err = p.parseCompare()
		if err != nil {
			return
		}

		nodes = append(nodes, n)
	}

	if len(nodes) > 1 {
		n = &AndNode{nodes}
	} else {
		n = nodes[0]
	}
	return
}

func (p *Parser) parseOR() (n Node, err error) {
	n, err = p.parseAND()
	if err != nil {
		return
	}

	nodes := []Node{n}

	for {
		t := p.getToken()
		if t.Type != TokenOR {
			break
		}
		p.consume()

		n, err = p.parseAND()
		if err != nil {
			return
		}

		nodes = append(nodes, n)
	}

	if len(nodes) > 1 {
		n = &OrNode{nodes}
	} else {
		n = nodes[0]
	}
	return
}

func (p *Parser) ParseExpr() (n Node, err error) {
	return p.parseOR()
}

var templateEndTokens = map[TokenType]bool{
	TokenEndBlock: true,
}

func (p *Parser) parseTemplate() (n Node, err error) {
	t := p.getToken()
	if t.Type != TokenBlock {
		err = p.errUnexpected("template")
		return
	}
	p.consume()

	t = p.getToken()
	if t.Type != TokenLeftParen {
		err = p.errUnexpected("(")
		return
	}
	p.consume()

	t = p.getToken()
	if t.Type != TokenIdent {
		err = p.errUnexpected("identifier")
		return
	}
	name := string(t.Value)
	p.consume()

	args := []string{}

	t = p.getToken()
	if t.Type == TokenComma {
		p.consume()

		for {
			t = p.getToken()
			if t.Type == TokenRightParen {
				break
			}
			if t.Type != TokenIdent {
				err = p.errUnexpected("identifier | )")
				return
			}

			arg := string(t.Value)
			args = append(args, arg)
			p.consume()

			t = p.getToken()
			if t.Type == TokenComma {
				p.consume()
			} else {
				break
			}
		}
	}

	if t.Type != TokenRightParen {
		err = p.errUnexpected(")")
		return
	}
	p.consume()

	stmts, err := p.parseStmtList(templateEndTokens)
	if err != nil {
		return
	}
	p.consume()

	n = &BlockNode{name, args, stmts}
	return
}

var ifEndTokens = map[TokenType]bool{
	TokenElse:   true,
	TokenElseIf: true,
	TokenEndIf:  true,
}

func (p *Parser) parseIf() (n Node, err error) {
	t := p.getToken()
	if t.Type != TokenIf {
		err = p.errUnexpected("if")
		return
	}
	p.consume()

	expr, err := p.ParseExpr()
	if err != nil {
		return
	}

	t = p.getToken()
	if t.Type != TokenThen {
		err = p.errUnexpected("then")
		return
	}
	p.consume()

	body, err := p.parseStmtList(ifEndTokens)

	branches := []IfBranch{{expr, body}}

	for {
		t = p.getToken()
		if t.Type != TokenElseIf {
			break
		}
		p.consume()

		expr, err = p.ParseExpr()
		if err != nil {
			return
		}
		t = p.getToken()
		if t.Type != TokenThen {
			err = p.errUnexpected("then")
			return
		}
		p.consume()
		body, err = p.parseStmtList(ifEndTokens)
		if err != nil {
			return
		}
		branches = append(branches, IfBranch{expr, body})
	}

	if t.Type == TokenElse {
		p.consume()
		body, err = p.parseStmtList(ifEndTokens)
		if err != nil {
			return
		}
		t = p.getToken()
	} else {
		body = nil
	}

	if t.Type != TokenEndIf {
		err = p.errUnexpected("endif")
		return
	}
	p.consume()

	n = &IfNode{branches, body}
	return
}

var forEndTokenMap = map[TokenType]bool{
	TokenEndFor: true,
}

func (p *Parser) parseFor() (n Node, err error) {
	t := p.getToken()
	if t.Type != TokenFor {
		err = p.errUnexpected("for")
		return
	}
	p.consume()

	t = p.getToken()
	if t.Type != TokenIdent {
		err = p.errUnexpected("identifier")
		return
	}
	varName := string(t.Value)
	p.consume()

	t = p.getToken()
	if t.Type != TokenIn {
		err = p.errUnexpected("in")
		return
	}
	p.consume()

	expr, err := p.ParseExpr()
	if err != nil {
		return
	}

	t = p.getToken()
	if t.Type != TokenDo {
		err = p.errUnexpected("do")
		return
	}
	p.consume()

	body, err := p.parseStmtList(forEndTokenMap)
	if err != nil {
		return
	}
	p.consume()

	n = &ForNode{varName, expr, body}
	return
}

func (p *Parser) parseDeclare() (n Node, err error) {
	t := p.getToken()
	if t.Type != TokenDeclare {
		err = p.errUnexpected("declare")
		return
	}
	p.consume()

	t = p.getToken()
	if t.Type != TokenLeftParen {
		err = p.errUnexpected("(")
		return
	}
	p.consume()

	t = p.getToken()
	if t.Type != TokenIdent {
		err = p.errUnexpected("identifier")
		return
	}
	name := string(t.Value)
	p.consume()

	t = p.getToken()
	if t.Type != TokenComma {
		err = p.errUnexpected(",")
		return
	}
	p.consume()

	n, err = p.ParseExpr()
	if err != nil {
		return
	}

	t = p.getToken()
	if t.Type != TokenRightParen {
		err = p.errUnexpected(")")
		return
	}
	p.consume()

	n = &DeclareNode{name, n}
	return
}

func (p *Parser) parseInclude() (n Node, err error) {
	t := p.getToken()
	if t.Type != TokenInclude {
		err = p.errUnexpected("include")
		return
	}
	p.consume()

	t = p.getToken()
	if t.Type != TokenLeftParen {
		err = p.errUnexpected("(")
		return
	}
	p.consume()

	name, err := p.ParseExpr()
	if err != nil {
		return
	}

	t = p.getToken()
	if t.Type != TokenRightParen {
		err = p.errUnexpected(")")
		return
	}
	p.consume()

	n = &IncludeNode{name}
	return
}

var discardEndTokenMap = map[TokenType]bool{
	TokenEndDiscard: true,
}

func (p *Parser) parseDiscard() (n Node, err error) {
	t := p.getToken()
	if t.Type != TokenDiscard {
		err = p.errUnexpected("discard")
		return
	}
	p.consume()

	body, err := p.parseStmtList(discardEndTokenMap)
	if err != nil {
		return
	}
	p.consume()

	n = &DiscardNode{body}
	return
}

func (p *Parser) ParseStmt() (n Node, err error) {
	t := p.getToken()
	switch t.Type {
	case TokenValue:
		p.consume()
		n = &ValueNode{string(t.Value)}
	case TokenBlock:
		n, err = p.parseTemplate()
	case TokenIf:
		n, err = p.parseIf()
	case TokenFor:
		n, err = p.parseFor()
	case TokenDeclare:
		n, err = p.parseDeclare()
	case TokenInclude:
		n, err = p.parseInclude()
	case TokenDiscard:
		n, err = p.parseDiscard()
	default:
		n, err = p.ParseExpr()
	}
	return
}

func (p *Parser) parseStmtList(endTokenMap map[TokenType]bool) (stmts []Node, err error) {
	for {
		t := p.getToken()

		if endTokenMap[t.Type] {
			return
		}

		switch t.Type {
		case TokenEOF:
			expected := make([]string, 0, len(endTokenMap))
			for tt, ok := range endTokenMap {
				if ok {
					expected = append(expected, strings.ToLower(tt.String()))
				}
			}
			err = p.errUnexpected(strings.Join(expected, " | "))
			return
		case TokenError:
			err = p.s.Err
			return
		}

		stmt, err := p.ParseStmt()
		if err != nil {
			return stmts, err
		}
		stmts = append(stmts, stmt)
	}
}

func (p *Parser) Parse() (n Node, err error) {
	nodes := []Node{}

	for {
		t := p.getToken()
		if t.Type == TokenEOF {
			break
		}

		n, err = p.ParseStmt()
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

type loopJump struct {
	idx  int
	kind int
}

const (
	loopJumpNext = iota
	loopJumpEnd
)

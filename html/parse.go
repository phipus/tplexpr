package html

import (
	"fmt"
	"io"
	"regexp"
	"strings"

	"github.com/phipus/tplexpr"
	"golang.org/x/net/html"
)

func ParseReader(r io.Reader) (tplexpr.Node, error) {
	s := NewScanner(r)

	body := []tplexpr.Node{}
	err := parse(&body, &s)
	if err != nil && err != io.EOF {
		return nil, err
	}
	return &tplexpr.EmitNode{Nodes: body}, nil
}

func ParseString(s string) (n tplexpr.Node, err error) {
	return ParseReader(strings.NewReader(s))
}

var identRegex = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)

func ParseTemplateReader(ctx *tplexpr.CompileContext, name string, r io.Reader) error {
	n, err := ParseReader(r)
	if err != nil {
		return err
	}
	return ctx.CompileTemplate(name, n)
}

func ParseTemplateString(ctx *tplexpr.CompileContext, name string, tpl string) error {
	n, err := ParseString(tpl)
	if err != nil {
		return err
	}
	return ctx.CompileTemplate(name, n)
}

func parse(to *[]tplexpr.Node, s *Scanner) (err error) {
	for {
		t := s.Token()
		switch t.Type {
		case html.ErrorToken:
			err = s.Err()
			return
		case html.TextToken:
			n, err := parseString(t.Data)
			if err != nil {
				return err
			}
			if value, ok := n.(*tplexpr.ValueNode); ok {
				value.Value = html.EscapeString(value.Value)
				*to = append(*to, value)
			} else {
				*to = append(*to, &TextNode{n})
			}

			s.Consume()

		case html.StartTagToken:
			switch t.Data {
			case "tx-switch":
				err = parseSwitch(to, s)
				if err != nil {
					return err
				}
			case "tx-block":
				err = parseBlock(to, s)
				if err != nil {
					return err
				}
			case "tx-for":
				err = parseFor(to, s)
				if err != nil {
					return err
				}
			case "tx-declare":
				err = parseDeclare(to, s)
				if err != nil {
					return err
				}
			case "tx-discard":
				err = parseDiscard(to, s)
				if err != nil {
					return err
				}
			case "tx-slot":
				err = parseSlot(to, s)
				if err != nil {
					return err
				}
			default:
				if len(t.Attr) <= 0 {
					*to = append(*to, &tplexpr.ValueNode{Value: fmt.Sprintf("<%s>", t.Data)})
				} else {
					*to = append(*to, &tplexpr.ValueNode{Value: fmt.Sprintf("<%s", t.Data)})
					err = parseAttrs(to, t.Attr)
					if err != nil {
						return
					}
					*to = append(*to, &tplexpr.ValueNode{Value: ">"})
				}
				s.Consume()
			}
		case html.EndTagToken:
			switch t.Data {
			case "tx-switch", "tx-case", "tx-default",
				"tx-block", "tx-for", "tx-declare", "tx-discard", "tx-slot":
				return nil
			default:
				*to = append(*to, &tplexpr.ValueNode{Value: fmt.Sprintf("</%s>", t.Data)})
				s.Consume()
			}

		case html.SelfClosingTagToken:
			switch t.Data {
			case "tx-switch":
				err = parseSwitch(to, s)
				if err != nil {
					return err
				}
			case "tx-block":
				err = parseBlock(to, s)
				if err != nil {
					return err
				}
			case "tx-for":
				err = parseFor(to, s)
				if err != nil {
					return err
				}
			case "tx-declare":
				err = parseDeclare(to, s)
				if err != nil {
					return err
				}
			case "tx-discard":
				err = parseDiscard(to, s)
				if err != nil {
					return err
				}
			case "tx-slot":
				err = parseSlot(to, s)
				if err != nil {
					return err
				}
			default:
				*to = append(*to, &tplexpr.ValueNode{Value: fmt.Sprintf("<%s", t.Data)})
				err = parseAttrs(to, t.Attr)
				if err != nil {
					return
				}
				*to = append(*to, &tplexpr.ValueNode{Value: "/>"})
				s.Consume()
			}

		case html.CommentToken:
			n, err := parseString(t.Data)
			if err != nil {
				return err
			}
			*to = append(*to, &CommentNode{n})
			s.Consume()
		case html.DoctypeToken:
			*to = append(*to, parseDoctype(t.Data, t.Attr))
			s.Consume()
		default:
			return fmt.Errorf("html: invalid token type %s", t.Type)
		}
	}
}

func parseAttrs(to *[]tplexpr.Node, attrs []html.Attribute) error {
	for _, a := range attrs {
		key := ""
		if a.Namespace != "" {
			key = fmt.Sprintf(" %s:%s=\"", a.Namespace, a.Key)
		} else {
			key = fmt.Sprintf(" %s=\"", a.Key)
		}
		n, err := parseString(a.Val)
		if err != nil {
			return err
		}
		if value, ok := n.(*tplexpr.ValueNode); ok {
			*to = append(*to, &tplexpr.ValueNode{Value: fmt.Sprintf(`%s%s"`, key, html.EscapeString(value.Value))})
		} else {
			*to = append(*to, &tplexpr.ValueNode{Value: key})
			*to = append(*to, &TextNode{n})
			*to = append(*to, &tplexpr.ValueNode{Value: "\""})
		}

	}
	return nil
}

func parseString(s string) (tplexpr.Node, error) {
	p := tplexpr.NewParser([]byte(s))
	return p.Parse()
}

func parseDoctype(data string, attrs []html.Attribute) (n tplexpr.Node) {
	bytes := []byte{}
	bytes = append(bytes, "<!DOCTYPE "...)
	bytes = append(bytes, html.EscapeString(data)...)
	if attrs != nil {
		var p, s string
		for _, a := range attrs {
			switch a.Key {
			case "public":
				p = a.Val
			case "system":
				s = a.Val
			}
		}
		if p != "" {
			bytes = append(bytes, " PUBLIC "...)
			bytes = append(bytes, Quote(p)...)
			if s != "" {
				bytes = append(bytes, ' ')
				bytes = append(bytes, Quote(s)...)
			}
		} else if s != "" {
			bytes = append(bytes, " SYSTEM "...)
			bytes = append(bytes, Quote(s)...)
		}
	}
	bytes = append(bytes, '>')
	n = &tplexpr.ValueNode{Value: string(bytes)}
	return
}

func getAttrsUnique(t *html.Token) (values map[string]string, err error) {
	values = map[string]string{}
	for _, a := range t.Attr {
		key := ""
		if a.Namespace != "" {
			key = fmt.Sprintf("%s:%s", a.Namespace, a.Key)
		} else {
			key = a.Key
		}

		if _, ok := values[key]; ok {
			err = fmt.Errorf("%w: dublicate attribute %s", tplexpr.ErrSyntax, key)
			return
		}
		values[key] = a.Val
	}
	return
}

func isEndTag(t *html.Token, tag string) bool {
	return t.Type == html.EndTagToken && t.Data == tag
}

func isOpenTag(t *html.Token, tag string) bool {
	return (t.Type == html.StartTagToken || t.Type == html.SelfClosingTagToken) && t.Data == tag
}

func parseSwitch(to *[]tplexpr.Node, s *Scanner) error {
	t := s.Token()
	if !isOpenTag(&t, "tx-switch") {
		return errUnexpected(s, &t, "<tx-switch ...")
	}

	attrs, err := getAttrsUnique(&t)
	if err != nil {
		return err
	}

	var expr tplexpr.Node
	if exprValue, ok := attrs["expr"]; ok {
		expr, err = parseString(exprValue)
		if err != nil {
			return err
		}
	}

	if t.Type == html.SelfClosingTagToken {
		if expr != nil {
			*to = append(*to, &SwitchNode{Expr: expr})
		}
		s.Consume()
		return nil
	}
	s.Consume()

	branches := []tplexpr.IfBranch{}
	alt := []tplexpr.Node{}
	hasAlt := false

	for {
		t = s.Token()
		switch {
		case t.Type == html.CommentToken:
			n, err := parseString(t.Data)
			if err != nil {
				return err
			}
			*to = append(*to, &CommentNode{n})
			s.Consume()
		case isOpenTag(&t, "tx-case"):
			attrs, err := getAttrsUnique(&t)
			if err != nil {
				return err
			}
			value, ok := attrs["value"]
			if !ok {
				return errAttrRequired("tx-case", "value")
			}
			expr, err := parseString(value)
			if err != nil {
				return err
			}

			if t.Type == html.SelfClosingTagToken {
				branches = append(branches, tplexpr.IfBranch{Expr: expr})
				s.Consume()
				continue
			}

			s.Consume()
			body := []tplexpr.Node{}
			err = parse(&body, s)
			if err != nil {
				return err
			}
			t = s.Token()
			if !isEndTag(&t, "tx-case") {
				return errUnexpected(s, &t, "</tx-case>")
			}
			branches = append(branches, tplexpr.IfBranch{Expr: expr, Body: body})
			s.Consume()

		case isOpenTag(&t, "tx-default"):
			if hasAlt {
				return fmt.Errorf("%w: in a <tx-switch> only one <tx-default> is allowed", tplexpr.ErrSyntax)
			}
			hasAlt = true
			if t.Type == html.SelfClosingTagToken {
				s.Consume()
				continue
			}
			s.Consume()
			err = parse(&alt, s)
			if err != nil {
				return err
			}
			t = s.Token()
			if !isEndTag(&t, "tx-default") {
				return errUnexpected(s, &t, "</tx-default>")
			}
			s.Consume()
		case isEndTag(&t, "tx-switch"):
			s.Consume()
			var n tplexpr.Node
			if expr != nil {
				n = &SwitchNode{Expr: expr, If: tplexpr.IfNode{Branches: branches, Alt: alt}}
			} else {
				n = &tplexpr.IfNode{Branches: branches, Alt: alt}
			}
			*to = append(*to, n)
			return nil
		default:
			return errUnexpected(s, &t, "</tx-switch>")
		}
	}
}

func parseBlock(to *[]tplexpr.Node, s *Scanner) error {
	t := s.Token()
	if !isOpenTag(&t, "tx-block") {
		return errUnexpected(s, &t, "<tx-block ...")
	}
	attrs, err := getAttrsUnique(&t)
	if err != nil {
		return err
	}
	name, ok := attrs["name"]
	if !ok {
		return errAttrRequired("tx-block", "name")
	}
	var args []string
	if argsValue := strings.TrimSpace(attrs["args"]); argsValue != "" {
		args = strings.Split(argsValue, ",")
		for i := range args {
			args[i] = strings.TrimSpace(args[i])
			arg := args[i]

			if arg == "" && i == len(args)-1 {
				break // allow trailing comma
			}
			if !identRegex.MatchString(arg) {
				return fmt.Errorf("%w: invalid argument name '%s'", tplexpr.ErrSyntax, arg)
			}
		}
	}

	if t.Type == html.SelfClosingTagToken {
		*to = append(*to, &tplexpr.BlockNode{Name: name, Args: args})
		s.Consume()
		return nil
	}
	s.Consume()
	body := []tplexpr.Node{}
	err = parse(&body, s)
	if err != nil {
		return err
	}
	t = s.Token()
	if !isEndTag(&t, "tx-block") {
		return errUnexpected(s, &t, "</tx-block>")
	}
	s.Consume()
	*to = append(*to, &tplexpr.BlockNode{Name: name, Args: args, Body: body})
	return nil
}

func parseFor(to *[]tplexpr.Node, s *Scanner) error {
	t := s.Token()
	if !isOpenTag(&t, "tx-for") {
		return errUnexpected(s, &t, "<tx-for ...")
	}
	attrs, err := getAttrsUnique(&t)
	if err != nil {
		return err
	}
	varName, ok := attrs["var"]
	if !ok {
		return errAttrRequired("tx-for", "var")
	}
	if !identRegex.MatchString(varName) {
		return fmt.Errorf("%w: invalid variable name '%s'", tplexpr.ErrSyntax, varName)
	}

	exprValue, ok := attrs["expr"]
	if !ok {
		return errAttrRequired("tx-for", "expr")
	}
	expr, err := parseString(exprValue)
	if err != nil {
		return err
	}

	if t.Type == html.SelfClosingTagToken {
		*to = append(*to, &tplexpr.ForNode{Var: varName, Expr: expr})
		s.Consume()
		return nil
	}
	s.Consume()
	body := []tplexpr.Node{}
	err = parse(&body, s)
	if err != nil {
		return err
	}
	t = s.Token()
	if !isEndTag(&t, "tx-for") {
		return errUnexpected(s, &t, "</tx-for>")
	}
	s.Consume()
	*to = append(*to, &tplexpr.ForNode{Var: varName, Expr: expr, Body: body})
	return nil
}

func parseDeclare(to *[]tplexpr.Node, s *Scanner) error {
	t := s.Token()
	if !isOpenTag(&t, "tx-declare") {
		return errUnexpected(s, &t, "<tx-declare ...")
	}
	attrs, err := getAttrsUnique(&t)
	if err != nil {
		return err
	}
	name, ok := attrs["name"]
	if !ok {
		return errAttrRequired("tx-declare", name)
	}

	if t.Type == html.SelfClosingTagToken {
		*to = append(*to, &tplexpr.DeclareNode{Name: name, Value: &tplexpr.ValueNode{Value: ""}})
		s.Consume()
		return nil
	}
	s.Consume()
	n := &tplexpr.EmitNode{}
	err = parse(&n.Nodes, s)
	if err != nil {
		return err
	}
	t = s.Token()
	if !isEndTag(&t, "tx-declare") {
		return errUnexpected(s, &t, "</tx-declare>")
	}
	s.Consume()
	*to = append(*to, &tplexpr.DeclareNode{Name: name, Value: n})
	return nil
}

func parseDiscard(to *[]tplexpr.Node, s *Scanner) error {
	t := s.Token()
	if !isOpenTag(&t, "tx-discard") {
		return errUnexpected(s, &t, "<tx-discard ...")
	}

	if t.Type == html.SelfClosingTagToken {
		s.Consume()
		return nil
	}
	s.Consume()
	body := []tplexpr.Node{}
	err := parse(&body, s)
	if err != nil {
		return err
	}
	t = s.Token()
	if !isEndTag(&t, "tx-discard") {
		return errUnexpected(s, &t, "</tx-discard>")
	}
	s.Consume()
	*to = append(*to, &tplexpr.DiscardNode{Body: body})
	return nil
}

func parseSlot(to *[]tplexpr.Node, s *Scanner) error {
	t := s.Token()
	if !isOpenTag(&t, "tx-slot") {
		return errUnexpected(s, &t, "<tx-slot ...")
	}
	attrs, err := getAttrsUnique(&t)
	if err != nil {
		return err
	}
	exprValue, ok := attrs["expr"]
	if !ok {
		return errAttrRequired("tx-slot", "expr")
	}
	expr, err := parseString(exprValue)
	if err != nil {
		return err
	}

	if t.Type == html.SelfClosingTagToken {
		*to = append(*to, expr)
		s.Consume()
		return nil
	}
	s.Consume()
	alt := []tplexpr.Node{}
	err = parse(&alt, s)
	if err != nil {
		return err
	}
	t = s.Token()
	if !isEndTag(&t, "tx-slot") {
		return errUnexpected(s, &t, "</tx-slot>")
	}
	s.Consume()

	*to = append(*to, &tplexpr.OrNode{
		Exprs: []tplexpr.Node{
			&tplexpr.DynCallNode{
				Value: &tplexpr.SubprogNode{
					Prog: expr,
				},
			},
			&tplexpr.DynCallNode{
				Value: &tplexpr.SubprogNode{
					Prog: &tplexpr.EmitNode{Nodes: alt},
				},
			},
		},
	})
	return nil
}

func errUnexpected(s *Scanner, t *html.Token, expected string) error {
	unexpectedValue := ""

	switch t.Type {
	case html.ErrorToken:
		if s.Err() != io.EOF {
			return s.Err()
		}
		unexpectedValue = "EOF"
	case html.TextToken:
		unexpectedValue = fmt.Sprintf("Text: %s", t.Data)
	case html.StartTagToken:
		unexpectedValue = fmt.Sprintf("<%s...", t.Data)
	case html.EndTagToken:
		unexpectedValue = fmt.Sprintf("</%s>", t.Data)
	case html.CommentToken:
		unexpectedValue = fmt.Sprintf("<!--%s-->", t.Data)
	case html.DoctypeToken:
		unexpectedValue = fmt.Sprintf("<!DOCTYPE %s ...", t.Data)
	default:
		return fmt.Errorf("html: unexpected token type")
	}

	if expected != "" {
		return fmt.Errorf("%w: expected %s but got %s", tplexpr.ErrSyntax, expected, unexpectedValue)
	}
	return fmt.Errorf("%w: unexpected %s", tplexpr.ErrSyntax, unexpectedValue)
}

func errAttrRequired(tag string, attr string) error {
	return fmt.Errorf("%w: <%s ... requires %s attribute", tplexpr.ErrSyntax, tag, attr)
}

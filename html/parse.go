package html

import (
	"errors"
	"fmt"
	"io"
	"regexp"
	"strings"

	"github.com/phipus/tplexpr"
	"golang.org/x/net/html"
)

func ParseReader(r io.Reader) (n tplexpr.Node, err error) {
	h, err := html.Parse(r)
	if err != nil {
		return
	}
	return ParseNode(h)
}

func ParseNode(h *html.Node) (tplexpr.Node, error) {
	n, err := parseHtmlNode(h)
	if err == errPlaintextAbort {
		err = nil
	}
	return n, err
}

var identRegex = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)

func parseHtmlNode(h *html.Node) (n tplexpr.Node, err error) {
	switch h.Type {
	case html.ErrorNode:
		err = fmt.Errorf("html: can not use an ErrorNode")
		return
	case html.TextNode:
		p := tplexpr.NewParser([]byte(h.Data))
		n, err = p.Parse()
		if err != nil {
			return
		}
		n = &TextNode{n}
		return
	case html.DocumentNode:
		nodes := []tplexpr.Node{}
		for c := h.FirstChild; c != nil; c = c.NextSibling {
			n, err = parseHtmlNode(c)
			if err != nil {
				return
			}
			nodes = append(nodes, n)
		}
		n = &tplexpr.EmitNode{Nodes: nodes}
		return
	case html.ElementNode:
		// No-op
	case html.CommentNode:
		p := tplexpr.NewParser([]byte(h.Data))
		n, err = p.Parse()
		if err != nil {
			return
		}
		n = &CommentNode{n}
		return
	case html.DoctypeNode:
		bytes := []byte{}
		bytes = append(bytes, "<!DOCTYPE "...)
		bytes = append(bytes, html.EscapeString(h.Data)...)
		if h.Attr != nil {
			var p, s string
			for _, a := range h.Attr {
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
	case html.RawNode:
		p := tplexpr.NewParser([]byte(h.Data))
		n, err = p.Parse()
		return
	default:
		err = fmt.Errorf("html: unknown node type")
		return
	}

	// test for special tx-* nodes and compile them differently
	switch h.Data {
	case "tx-switch":
		var expr tplexpr.Node
		for _, a := range h.Attr {
			if a.Namespace == "" && a.Key == "expr" {
				p := tplexpr.NewParser([]byte(a.Val))
				expr, err = p.Parse()
				if err != nil {
					return
				}
				break
			}
		}

		ifNode := tplexpr.IfNode{}
		hasAlt := false
		for c := h.FirstChild; c != nil; c = c.NextSibling {
			switch c.Type {
			case html.CommentNode:
				return parseHtmlNode(c)

			case html.ElementNode:
				switch c.Data {
				case "tx-case":
					br := tplexpr.IfBranch{}
					for _, a := range c.Attr {
						if a.Namespace == "" && a.Key == "expr" {
							p := tplexpr.NewParser([]byte(a.Val))
							br.Expr, err = p.Parse()
							if err != nil {
								return
							}
							break
						}
					}
					if br.Expr == nil {
						err = fmt.Errorf("%w: tx-case requires expr attribute", tplexpr.ErrSyntax)
						return
					}

					for cc := c.FirstChild; cc != nil; cc = cc.NextSibling {
						n, err = parseHtmlNode(cc)
						if err != nil {
							return
						}
						br.Body = append(br.Body, n)
					}

					ifNode.Branches = append(ifNode.Branches, br)
					continue

				case "tx-default":
					if hasAlt {
						err = fmt.Errorf("%w: tx-switch can only contain one tx-default", tplexpr.ErrSyntax)
						return
					}
					hasAlt = true
					for cc := c.FirstChild; cc != nil; cc = cc.NextSibling {
						n, err = parseHtmlNode(cc)
						if err != nil {
							return
						}
						ifNode.Alt = append(ifNode.Alt, n)
					}
					continue
				}
			}

			err = fmt.Errorf("%w: tx-switch can only contain tx-case and tx-default", tplexpr.ErrSyntax)
			return
		}

		if expr != nil {
			n = &SwitchNode{expr, ifNode}
		} else {
			p := &tplexpr.IfNode{}
			*p = ifNode
			n = p
		}
		return
	case "tx-template":
		name := ""
		args := []string{}

		for _, a := range h.Attr {
			if a.Namespace == "" && a.Key == "name" {
				name = a.Val
			}
			if a.Namespace == "" && a.Key == "args" {
				args = strings.Split(a.Val, ",")
				for i, arg := range args {
					args[i] = strings.TrimSpace(arg)
					if !identRegex.MatchString(args[i]) {
						err = fmt.Errorf("%w: invalid argument name '%s'", tplexpr.ErrSyntax, args[i])
						return
					}
				}
			}
		}

		body := []tplexpr.Node{}

		for c := h.FirstChild; c != nil; c = c.NextSibling {
			n, err = parseHtmlNode(c)
			if err != nil {
				return
			}
			body = append(body, n)
		}
		n = &tplexpr.TemplateNode{Name: name, Args: args, Body: body}
		return
	case "tx-for":
		var expr tplexpr.Node
		var varName string
		for _, a := range h.Attr {
			if a.Namespace == "" && a.Key == "expr" {
				p := tplexpr.NewParser([]byte(a.Val))
				expr, err = p.Parse()
				if err != nil {
					return
				}
			} else if a.Namespace == "" && a.Key == "var" {
				varName = a.Val
			}
		}
		if expr == nil {
			err = fmt.Errorf("%w tx-for: expr required", tplexpr.ErrSyntax)
			return
		}
		if !identRegex.MatchString(varName) {
			err = fmt.Errorf("%w: invalid variable name '%s'", tplexpr.ErrSyntax, varName)
			return
		}

		body := []tplexpr.Node{}
		for c := h.FirstChild; c != nil; c = c.NextSibling {
			n, err = parseHtmlNode(c)
			if err != nil {
				return
			}
			body = append(body, n)
		}
		n = &tplexpr.ForNode{Var: varName, Expr: expr, Body: body}
		return
	}

	nodes := []tplexpr.Node{}
	nodes = append(nodes, &tplexpr.ValueNode{Value: fmt.Sprintf("<%s", h.Data)})

	for _, a := range h.Attr {
		key := ""
		if a.Namespace != "" {
			key = fmt.Sprintf(" %s:%s=\"", a.Namespace, a.Key)
		} else {
			key = fmt.Sprintf(" %s=\"", a.Key)
		}
		nodes = append(nodes, &tplexpr.ValueNode{Value: key})
		p := tplexpr.NewParser([]byte(a.Val))
		n, err = p.Parse()
		if err != nil {
			return
		}
		nodes = append(nodes, &TextNode{n})
		nodes = append(nodes, &tplexpr.ValueNode{Value: "\""})

	}

	if voidElements[h.Data] {
		if h.FirstChild != nil {
			err = fmt.Errorf("html: void element <%s> has child nodes", h.Data)
			return
		}
		nodes = append(nodes, &tplexpr.ValueNode{Value: "/>"})
		n = &tplexpr.EmitNode{Nodes: nodes}
		return
	}

	nodes = append(nodes, &tplexpr.ValueNode{Value: ">"})

	// Add initial newline where there is danger of a newline beging ignored.
	if c := h.FirstChild; c != nil && c.Type == html.TextNode && strings.HasPrefix(c.Data, "\n") {
		switch h.Data {
		case "pre", "listing", "textarea":
			nodes = append(nodes, &tplexpr.ValueNode{Value: "\n"})
		}
	}

	// Render any child nodes
	if childTextNodesAreLiteral(h) {
		for c := h.FirstChild; c != nil; c = c.NextSibling {
			if c.Type == html.TextNode {
				p := tplexpr.NewParser([]byte(c.Data))
				n, err = p.Parse()
				if err != nil {
					return
				}
				nodes = append(nodes, n)
			} else {
				n, err = parseHtmlNode(c)
				if err != nil {
					return
				}
				nodes = append(nodes, n)
			}
		}
		if h.Data == "plaintext" {
			err = errPlaintextAbort
			n = &tplexpr.EmitNode{Nodes: nodes}
			return
		}
	} else {
		for c := h.FirstChild; c != nil; c = c.NextSibling {
			n, err = parseHtmlNode(c)
			if err != nil {
				return
			}
			nodes = append(nodes, n)
		}
	}

	// Render the closing tag
	nodes = append(nodes, &tplexpr.ValueNode{Value: fmt.Sprintf("</%s>", h.Data)})
	n = &tplexpr.EmitNode{Nodes: nodes}
	return
}

// Section 12.1.2, "Elements", gives this list of void elements. Void elements
// are those that can't have any contents.
var voidElements = map[string]bool{
	"area":   true,
	"base":   true,
	"br":     true,
	"col":    true,
	"embed":  true,
	"hr":     true,
	"img":    true,
	"input":  true,
	"keygen": true, // "keygen" has been removed from the spec, but are kept here for backwards compatibility.
	"link":   true,
	"meta":   true,
	"param":  true,
	"source": true,
	"track":  true,
	"wbr":    true,
}

func childTextNodesAreLiteral(n *html.Node) bool {
	// Per WHATWG HTML 13.3, if the parent of the current node is a style,
	// script, xmp, iframe, noembed, noframes, or plaintext element, and the
	// current node is a text node, append the value of the node's data
	// literally. The specification is not explicit about it, but we only
	// enforce this if we are in the HTML namespace (i.e. when the namespace is
	// "").
	// NOTE: we also always include noscript elements, although the
	// specification states that they should only be rendered as such if
	// scripting is enabled for the node (which is not something we track).
	if n.Namespace != "" {
		return false
	}
	switch n.Data {
	case "iframe", "noembed", "noframes", "noscript", "plaintext", "script", "style", "xmp":
		return true
	default:
		return false
	}
}

// errPlaintextAbort is returned from parseHtmlNode when a <plaintext> element
// has been rendered. No more end tags should be rendered after that.
var errPlaintextAbort = errors.New("html: internal error (plaintext abort)")

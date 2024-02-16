package html

import (
	"errors"
	"fmt"
	"strings"

	"github.com/phipus/tplexpr"
	"golang.org/x/net/html"
)

func compileHtmlNode1(ctx *tplexpr.CompileContext, n *html.Node) (err error) {
	switch n.Type {
	case html.ErrorNode:
		return fmt.Errorf("html: can not render an ErrorNode")
	case html.TextNode:
		p := tplexpr.NewParser([]byte(n.Data))
		n, err := p.Parse()
		if err != nil {
			return err
		}
		ctx.PushOutputFilter(HtmlEscapeFilter)
		err = n.Compile(ctx, tplexpr.CompileEmit)
		if err != nil {
			return err
		}
		ctx.PopOutputFilter()
		return nil

	case html.DocumentNode:
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			err = compileHtmlNode1(ctx, c)
			if err != nil {
				return
			}
		}
		return nil
	case html.ElementNode:
		// No-op
	case html.CommentNode:
		p := tplexpr.NewParser([]byte(n.Data))
		n, err := p.Parse()
		if err != nil {
			return err
		}
		ctx.EmitValue("<!--")
		ctx.PushOutputFilter(CommentEscapeFilter)
		err = n.Compile(ctx, tplexpr.CompileEmit)
		if err != nil {
			return err
		}
		ctx.PopOutputFilter()
		ctx.EmitValue("-->")
		return nil

	case html.DoctypeNode:
		ctx.EmitValue("<!DOCTYPE ")
		ctx.EmitValue(html.EscapeString(string(n.Data)))
		if n.Attr != nil {
			var p, s string
			for _, a := range n.Attr {
				switch a.Key {
				case "public":
					p = a.Val
				case "system":
					s = a.Val
				}
			}
			if p != "" {
				ctx.EmitValue(" PUBLIC ")
				ctx.EmitValue(Quote(p))
				if s != "" {
					ctx.EmitValue(" ")
					ctx.EmitValue(Quote(s))
				}
			} else if s != "" {
				ctx.EmitValue(" SYSTEM ")
				ctx.EmitValue(Quote(s))
			}
		}
		ctx.EmitValue(">")
		return nil
	case html.RawNode:
		p := tplexpr.NewParser([]byte(n.Data))
		n, err := p.Parse()
		if err != nil {
			return err
		}
		return n.Compile(ctx, tplexpr.CompileEmit)
	default:
		return fmt.Errorf("html: unknown node type")

	}

	ctx.EmitValue("<")
	ctx.EmitValue(n.Data)

	for _, a := range n.Attr {
		key := ""
		if a.Namespace != "" {
			key = fmt.Sprintf(" %s:%s=\"", a.Namespace, a.Key)
		} else {
			key = fmt.Sprintf(" %s=\"", a.Key)
		}
		ctx.EmitValue(key)
		p := tplexpr.NewParser([]byte(a.Val))
		n, err := p.Parse()
		if err != nil {
			return err
		}
		ctx.PushOutputFilter(HtmlEscapeFilter)
		err = n.Compile(ctx, tplexpr.CompileEmit)
		if err != nil {
			return err
		}
		ctx.PopOutputFilter()
		ctx.EmitValue("\"")
	}

	if voidElements[n.Data] {
		if n.FirstChild != nil {
			return fmt.Errorf("html: void element <%s> has child nodes", n.Data)
		}
		ctx.EmitValue("/>")
		return
	}

	// Add initial newline where there is danger of a newline beging ignored.
	if c := n.FirstChild; c != nil && c.Type == html.TextNode && strings.HasPrefix(c.Data, "\n") {
		switch n.Data {
		case "pre", "listing", "textarea":
			ctx.EmitValue("\n")
		}
	}

	// Render any child nodes
	if childTextNodesAreLiteral(n) {
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			if c.Type == html.TextNode {
				p := tplexpr.NewParser([]byte(n.Data))
				n, err := p.Parse()
				if err != nil {
					return err
				}
				err = n.Compile(ctx, tplexpr.CompileEmit)
				if err != nil {
					return err
				}
			} else {
				err = compileHtmlNode1(ctx, c)
				if err != nil {
					return err
				}
			}
		}
		if n.Data == "plaintext" {
			return errPlaintextAbort
		}
	} else {
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			err = compileHtmlNode1(ctx, c)
			if err != nil {
				return err
			}
		}
	}

	// Render the closing tag
	ctx.EmitValue(fmt.Sprintf("</%s>", n.Data))
	return nil
}

func compileHtmlNode(ctx *tplexpr.CompileContext, n *html.Node) (err error) {
	err = compileHtmlNode1(ctx, n)
	if err == errPlaintextAbort {
		err = nil
	}
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

// errPlaintextAbort is returned from compileHtmlNode1 when a <plaintext> element
// has been rendered. No more end tags should be rendered after that.
var errPlaintextAbort = errors.New("html: internal error (plaintext abort)")

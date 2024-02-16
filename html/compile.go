package html

import (
	"io"
	"strings"

	"github.com/phipus/tplexpr"
	"golang.org/x/net/html"
)

func Compile(r io.Reader, ctx *tplexpr.CompileContext, mode int) (err error) {
	n, err := ParseReader(r)
	if err != nil {
		return
	}

	err = n.Compile(ctx, mode)
	return
}

func CompileString(s string, ctx *tplexpr.CompileContext, mode int) (err error) {
	r := strings.NewReader(s)
	return Compile(r, ctx, mode)
}

func CompileNode(h *html.Node, ctx *tplexpr.CompileContext, mode int) error {
	n, err := ParseNode(h)
	if err != nil {
		return err
	}
	return n.Compile(ctx, mode)
}

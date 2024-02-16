package html

import (
	"io"
	"strings"

	"github.com/phipus/tplexpr"
	"golang.org/x/net/html"
)

func Compile(r io.Reader) (code []tplexpr.Instr, c tplexpr.Context, err error) {
	node, err := html.Parse(r)
	if err != nil {
		return
	}

	ctx := tplexpr.NewCompileContext()

	err = compileHtmlNode1(&ctx, node)
	if err == nil {
		code, c = ctx.Compile()
	}
	return
}

func CompileString(s string) (code []tplexpr.Instr, c tplexpr.Context, err error) {
	r := strings.NewReader(s)
	return Compile(r)
}

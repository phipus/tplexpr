package html

import (
	"bytes"
	"io"
	"io/fs"
	"os"
	"strings"

	"github.com/phipus/tplexpr"
	"github.com/phipus/tplexpr/web"
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

func CompileTemplate(name string, r io.Reader, ctx *tplexpr.CompileContext) error {
	n, err := ParseReader(r)
	if err != nil {
		return err
	}
	return ctx.CompileTemplate(name, n)
}

func CompileTemplateString(name, s string, ctx *tplexpr.CompileContext) error {
	r := strings.NewReader(s)
	return CompileTemplate(name, r, ctx)
}

func CompileGlobFS(fsys fs.FS, pattern string, ctx *tplexpr.CompileContext) error {
	files, err := fs.Glob(fsys, pattern)
	if err != nil {
		return err
	}

	for _, file := range files {
		err = CompileFSFile(fsys, file, ctx)
		if err != nil {
			return err
		}
	}
	return nil
}

func CompileFSFile(fsys fs.FS, fileName string, ctx *tplexpr.CompileContext) error {
	file, err := fsys.Open(fileName)
	if err != nil {
		return err
	}
	defer file.Close()
	return CompileTemplate(fileName, file, ctx)
}

func CompileFile(fileName string, ctx *tplexpr.CompileContext) error {
	file, err := os.Open(fileName)
	if err != nil {
		return err
	}
	defer file.Close()
	return CompileTemplate(fileName, file, ctx)
}

type Plugin struct {
	NoBuiltins bool
}

var _ tplexpr.Plugin = &Plugin{}

func (p *Plugin) ParseTemplate(name string, data []byte, ctx *tplexpr.CompileContext) (bool, error) {
	switch web.FileNameExtension(name) {
	case ".html", ".htm":
		err := ParseTemplateReader(ctx, name, bytes.NewReader(data))
		return true, err
	default:
		return false, nil
	}
}

func (p *Plugin) InitContext(ctx *tplexpr.Context) {
	if !p.NoBuiltins {
		ctx.Declare("buildQueryParams", tplexpr.FuncValue(BuiltinBuildQueryParams))
		ctx.Declare("escapeQuery", tplexpr.FuncValue(BuiltinQueryEscape))
		ctx.Declare("escapePath", tplexpr.FuncValue(BuiltinPathEscape))
	}
}

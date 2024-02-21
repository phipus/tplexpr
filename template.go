package tplexpr

import (
	"io"
	"strings"
)

type Template struct {
	Code []Instr
}

type Store interface {
	Render(w io.Writer, name string, vars VarScope) error
}

func FileNameExtension(name string) string {
	ext := ""
	switch dotIdx := strings.LastIndexByte(name, '.'); dotIdx {
	case -1, 0:
		// nop
	default:
		ext = name[dotIdx:]
	}
	return ext
}

func Render(s Store, w io.Writer, tpl string, vb *ScopeBuilder) error {
	vars, err := vb.Build()
	if err != nil {
		return err
	}
	return s.Render(w, tpl, vars)
}

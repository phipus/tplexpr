package tplexpr

import (
	"io"
)

type Template struct {
	Code []Instr
}

type Store interface {
	Render(w io.Writer, name string, vars VarScope) error
}

func Render(s Store, w io.Writer, tpl string, vb *ScopeBuilder) error {
	vars, err := vb.Build()
	if err != nil {
		return err
	}
	return s.Render(w, tpl, vars)
}

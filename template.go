package tplexpr

import (
	"io"
)

type Template struct {
	Code []Instr
}

type Store interface {
	Render(w io.Writer, name string, vars Vars) error
}

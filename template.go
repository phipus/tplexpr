package tplexpr

import (
	"io"
	"net/http"
)

type Vars = map[string]Value

type Template struct {
	Code []Instr
}

type Store interface {
	Render(w io.Writer, name string, vars Vars) error
}

type WebStore struct {
	Store        Store
	ContentTypes ContentTypeResolver
}

type ContentTypeResolver interface {
	ResolveContentType(name string) (string, bool)
}

func (s *WebStore) Render(w http.ResponseWriter, status int, name string, vars Vars) error {
	contentType, ok := s.ContentTypes.ResolveContentType(name)
	if ok {
		w.Header().Set("Content-Type", contentType)
	}
	w.WriteHeader(status)
	return s.Store.Render(w, name, vars)
}

func ResolveContentType(name string)

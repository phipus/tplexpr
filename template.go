package tplexpr

import (
	"io"
	"net/http"
	"strings"
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
	contentType := ""
	ok := false
	if s.ContentTypes != nil {
		contentType, ok = s.ContentTypes.ResolveContentType(name)
	} else {
		contentType, ok = ResolveWebContentType(name)
	}
	if ok {
		w.Header().Set("Content-Type", contentType)
	}
	w.WriteHeader(status)
	return s.Store.Render(w, name, vars)
}

func ResolveWebContentType(name string) (string, bool) {
	ext := ""
	switch dotIdx := strings.LastIndexByte(name, '.'); dotIdx {
	case -1, 0:
		// nop
	default:
		ext = name[dotIdx:]
	}

	contentType := ""
	switch ext {
	case ".htm", ".html":
		contentType = "text/html"
	case ".css":
		contentType = "text/css"
	case ".js":
		contentType = "text/javascript"
	case ".json":
		contentType = "application/json"
	case ".txt":
		contentType = "text/plain"
	default:
		return "", false
	}

	return contentType, true
}

type webContentTypeResolver struct{}

var WebContentTypeResolver ContentTypeResolver = webContentTypeResolver{}

func (webContentTypeResolver) ResolveContentType(name string) (string, bool) {
	return ResolveWebContentType(name)
}

package web

import (
	"net/http"
	"strings"

	"github.com/phipus/tplexpr"
)

type Store struct {
	s tplexpr.Store
	r ContentTypeResolver
}

func NewStore(s tplexpr.Store, r ContentTypeResolver) *Store {
	return &Store{s, r}
}

func (s *Store) Render(w http.ResponseWriter, status int, name string, vars tplexpr.Vars) error {
	contentType := ""
	ok := false
	if s.r != nil {
		contentType, ok = s.r.ResolveContentType(name)
	} else {
		contentType, ok = ResolveWebContentType(name)
	}
	if ok {
		w.Header().Set("Content-Type", contentType)
	}
	w.WriteHeader(status)
	return s.s.Render(w, name, vars)
}

type ContentTypeResolver interface {
	ResolveContentType(name string) (string, bool)
}

type webContentTypeResolver struct{}

var WebContentTypeResolver ContentTypeResolver = webContentTypeResolver{}

func (webContentTypeResolver) ResolveContentType(name string) (string, bool) {
	return ResolveWebContentType(name)
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

func ResolveWebContentType(name string) (string, bool) {
	ext := FileNameExtension(name)

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

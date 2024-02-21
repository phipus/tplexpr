package html

import (
	"net/http"

	"github.com/phipus/tplexpr"
)

type WebStore struct {
	Store    tplexpr.Store
	Resolver ContentTypeResolver
}

func (s *WebStore) Render(w http.ResponseWriter, status int, name string, vars tplexpr.VarScope) error {
	contentType := ""
	ok := false
	if s.Resolver != nil {
		contentType, ok = s.Resolver.ResolveContentType(name)
	} else {
		contentType, ok = ResolveWebContentType(name)
	}
	if ok {
		w.Header().Set("Content-Type", contentType)
	}
	w.WriteHeader(status)
	return s.Store.Render(w, name, vars)
}

type ContentTypeResolver interface {
	ResolveContentType(name string) (string, bool)
}

type webContentTypeResolver struct{}

var WebContentTypeResolver ContentTypeResolver = webContentTypeResolver{}

func (webContentTypeResolver) ResolveContentType(name string) (string, bool) {
	return ResolveWebContentType(name)
}

func ResolveWebContentType(name string) (string, bool) {
	ext := tplexpr.FileNameExtension(name)

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

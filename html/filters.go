package html

import (
	"github.com/phipus/tplexpr"
	"golang.org/x/net/html"
)

type htmlEscapeFilter struct{}

var HtmlEscapeFilter tplexpr.ValueFilter = htmlEscapeFilter{}

func (htmlEscapeFilter) Filter(s string) (string, error) {
	return html.EscapeString(s), nil
}

type commentEscapeFilter struct{}

var CommentEscapeFilter tplexpr.ValueFilter = commentEscapeFilter{}

func (commentEscapeFilter) Filter(s string) (string, error) {
	return EscapeComment(s), nil
}

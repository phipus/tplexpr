package html

import (
	"io"

	"golang.org/x/net/html"
)

func Parse(r io.Reader) {
	node, err := html.Parse(r)
}

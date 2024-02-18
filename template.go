package tplexpr

import "net/http"

type Vars = map[string]Value

type Template struct {
	Code []Instr
}

func (c *Context) RenderTemplateHTML(status int, w http.ResponseWriter, name string, vars Vars) error {
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(status)
	return c.EvalTemplateWriter(name, vars, w)
}

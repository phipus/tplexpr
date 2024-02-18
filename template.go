package tplexpr

type Vars = map[string]Value

type Template struct {
	Code []Instr
}

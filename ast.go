package tplexpr

type Node interface {
	Compile(ctx *CompileContext, mode int) error
}

type ValueNode struct {
	Value string
}

type VarNode struct {
	Name string
}

type CallNode struct {
	Name string
	Args []Node
}

type DynCallNode struct {
	Value Node
	Args  []Node
}

type EmitNode struct {
	nodes []Node
}

type AttrNode struct {
	Expr Node
	Name string
}

type SubprogNode struct {
	Args []string
	Prog Node
}

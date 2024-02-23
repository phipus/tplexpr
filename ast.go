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

type CompoundNode struct {
	Nodes []Node
}

type AttrNode struct {
	Expr Node
	Name string
}

type SubprogNode struct {
	Args []string
	Prog Node
}

type CompareNode struct {
	Compare int
	Left    Node
	Right   Node
}

type AndNode struct {
	Exprs []Node
}

type OrNode struct {
	Exprs []Node
}

type BinaryOP struct {
	Op   int
	Expr Node
}

type BinaryOPNode struct {
	Expr Node
	Ops  []BinaryOP
}

type NumberNode struct {
	Value string
}

type BlockNode struct {
	Name string
	Args []string
	Body []Node
}

type IfBranch struct {
	Expr Node
	Body []Node
}

type IfNode struct {
	Branches []IfBranch
	Alt      []Node
}

type DeclareNode struct {
	Name  string
	Value Node
}

type ForNode struct {
	Var  string
	Expr Node
	Body []Node
}

type IncludeNode struct {
	Name Node
}

type DiscardNode struct {
	Body []Node
}

type ObjectKey struct {
	Key   string
	Value Node
}

type ObjectNode struct {
	Extend Node
	Keys   []ObjectKey
}

type SingleValueNode struct {
	Node Node
}

package tplexpr

type CompileContext struct {
	code     []Instr
	subprogs []Subprog
}

func NewCompileContext() CompileContext {
	return CompileContext{}
}

const (
	CompileEmit = iota
	CompilePush
)

func (c *CompileContext) setCode(code []Instr) {
	c.code = code
}

func (c *CompileContext) PushInstr(op, iarg int, sarg string) {
	c.code = append(c.code, Instr{op, iarg, sarg})
}

func (c *CompileContext) Value(mode int, value string) {
	switch mode {
	case CompileEmit:
		c.PushInstr(emit, 0, value)
	case CompilePush:
		c.PushInstr(push, 0, value)
	}
}

func (c *CompileContext) Code() []Instr {
	return c.code
}

func (c *CompileContext) pushSubprog(args []string, code []Instr) int {
	idx := len(c.subprogs)
	c.subprogs = append(c.subprogs, Subprog{args, code})
	return idx
}

func (c *CompileContext) WithSubprog(args []string, f func() error) (int, error) {
	defer c.setCode(c.code)
	c.code = nil
	err := f()
	if err != nil {
		return 0, err
	}

	return c.pushSubprog(args, c.code), nil
}

func (c *CompileContext) Var(mode int, name string) {
	switch mode {
	case CompileEmit:
		c.PushInstr(emitFetch, 0, name)
	case CompilePush:
		c.PushInstr(pushFetch, 0, name)
	}
}

func (c *CompileContext) Call(mode int, name string, argc int) {
	switch mode {
	case CompileEmit:
		c.PushInstr(emitCall, argc, name)
	case CompilePush:
		c.PushInstr(pushCall, argc, name)
	}
}

func (c *CompileContext) Subprog(mode int, subprogIdx int) {
	switch mode {
	case CompileEmit:
		c.PushInstr(emitSubprog, subprogIdx, "")
	case CompilePush:
		c.PushInstr(pushSubprog, subprogIdx, "")
	}
}

func (c *CompileContext) Attr(mode int, name string) {
	switch mode {
	case CompileEmit:
		c.PushInstr(emitAttr, 0, name)
	case CompilePush:
		c.PushInstr(pushAttr, 0, name)
	}
}

func (c *CompileContext) Compile() (code []Instr, ctx Context) {
	code = c.code
	ctx = NewContext()
	ctx.subprogs = c.subprogs
	return
}

func (n *ValueNode) Compile(ctx *CompileContext, mode int) error {
	ctx.Value(mode, n.Value)
	return nil
}

func (n *VarNode) Compile(ctx *CompileContext, mode int) error {
	ctx.Var(mode, n.Name)
	return nil
}

func (n *CallNode) Compile(ctx *CompileContext, mode int) error {
	for _, arg := range n.Args {
		err := arg.Compile(ctx, CompilePush)
		if err != nil {
			return err
		}
	}

	ctx.Call(mode, n.Name, len(n.Args))
	return nil
}

func (n *EmitNode) Compile(ctx *CompileContext, mode int) error {
	for _, node := range n.nodes {
		err := node.Compile(ctx, mode)
		if err != nil {
			return err
		}
	}
	return nil
}

func (n *AttrNode) Compile(ctx *CompileContext, mode int) error {
	err := n.Expr.Compile(ctx, CompilePush)
	if err != nil {
		return err
	}
	ctx.Attr(mode, n.Name)
	return nil
}

func (n *SubprogNode) Compile(ctx *CompileContext, mode int) error {
	index, err := ctx.WithSubprog(n.Args, func() error {
		return n.Prog.Compile(ctx, CompileEmit)
	})
	if err != nil {
		return err
	}

	ctx.Subprog(mode, index)
	return nil
}

package tplexpr

import (
	"errors"
	"fmt"
)

type CompileContext struct {
	code           []Instr
	subprogs       []Subprog
	loopJumps      *[]loopJump
	valueFilters   []ValueFilter
	valueFilterMap map[ValueFilter]int
	templates      map[string]Template
}

func NewCompileContext() CompileContext {
	return CompileContext{
		valueFilterMap: map[ValueFilter]int{},
		templates:      map[string]Template{},
	}
}

const (
	CompileEmit = iota
	CompilePush
)

type ValueFilter interface {
	Filter(s string) (string, error)
}

var ErrTemplateExists = errors.New("template exists already")

func (c *CompileContext) CompileTemplate(name string, node Node) error {
	_, ok := c.templates[name]
	if ok {
		return fmt.Errorf("create template '%s': %w", name, ErrTemplateExists)
	}
	defer c.setCode(c.code)
	c.code = nil

	err := node.Compile(c, CompileEmit)
	if err != nil {
		return err
	}
	c.templates[name] = Template{c.code}
	return nil
}

func (c *CompileContext) ParseTemplate(name string, data []byte) error {
	p := NewParser(data)
	n, err := p.Parse()
	if err != nil {
		return err
	}
	return c.CompileTemplate(name, n)
}

func (c *CompileContext) setCode(code []Instr) {
	c.code = code
}

func (c *CompileContext) setLoopJumps(loopJumps *[]loopJump) {
	c.loopJumps = loopJumps
}

func (c *CompileContext) pushInstr(op, iarg int, sarg string) {
	c.code = append(c.code, Instr{op, iarg, sarg})
}

func (c *CompileContext) Value(mode int, value string) {
	switch mode {
	case CompileEmit:
		c.pushInstr(emit, 0, value)
	case CompilePush:
		c.pushInstr(push, 0, value)
	}
}

func (c *CompileContext) EmitValue(value string) {
	c.pushInstr(emit, 0, value)
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

func (c *CompileContext) WithLoopJumps(loopJumps *[]loopJump, f func() error) error {
	defer c.setLoopJumps(c.loopJumps)
	c.loopJumps = loopJumps
	return f()
}

func (c *CompileContext) Var(mode int, name string) {
	switch mode {
	case CompileEmit:
		c.pushInstr(emitFetch, 0, name)
	case CompilePush:
		c.pushInstr(pushFetch, 0, name)
	}
}

func (c *CompileContext) Call(mode int, name string, argc int) {
	switch mode {
	case CompileEmit:
		c.pushInstr(emitCall, argc, name)
	case CompilePush:
		c.pushInstr(pushCall, argc, name)
	}
}

func (c *CompileContext) DynCall(mode int, argc int) {
	switch mode {
	case CompileEmit:
		c.pushInstr(emitCallDyn, argc, "")
	case CompilePush:
		c.pushInstr(pushCallDyn, argc, "")
	}
}

func (c *CompileContext) Subprog(mode int, subprogIdx int) {
	switch mode {
	case CompileEmit:
		c.pushInstr(emitSubprog, subprogIdx, "")
	case CompilePush:
		c.pushInstr(pushSubprog, subprogIdx, "")
	}
}

func (c *CompileContext) Attr(mode int, name string) {
	switch mode {
	case CompileEmit:
		c.pushInstr(emitAttr, 0, name)
	case CompilePush:
		c.pushInstr(pushAttr, 0, name)
	}
}

func (c *CompileContext) Compare(mode int, cmp int) {
	switch mode {
	case CompileEmit:
		c.pushInstr(emitCompare, cmp, "")
	case CompilePush:
		c.pushInstr(pushCompare, cmp, "")
	}
}

func (c *CompileContext) PushPeek() {
	c.pushInstr(pushPeek, 0, "")
}

func (c *CompileContext) DiscardPop() {
	c.pushInstr(discardPop, 0, "")
}

func (c *CompileContext) BinaryOP(mode int, op int) {
	switch mode {
	case CompileEmit:
		c.pushInstr(emitBinaryOP, op, "")
	case CompilePush:
		c.pushInstr(pushBinaryOP, op, "")
	}
}

func (c *CompileContext) Number(mode int, value string) {
	switch mode {
	case CompileEmit:
		c.pushInstr(emitNumber, 0, value)
	case CompilePush:
		c.pushInstr(pushNumber, 0, value)
	}
}

func (c *CompileContext) PushOutputFilter(f ValueFilter) {
	idx, ok := c.valueFilterMap[f]
	if !ok {
		idx = len(c.valueFilters)
		c.valueFilters = append(c.valueFilters, f)
		c.valueFilterMap[f] = idx
	}
	c.pushInstr(pushOutputFilter, idx, "")
}

func (c *CompileContext) PopOutputFilter() {
	c.pushInstr(popOutputFilter, 0, "")
}

func (c *CompileContext) Compile() (code []Instr, ctx Context) {
	code = c.code
	ctx = NewContext()
	ctx.subprogs = c.subprogs
	ctx.valueFilters = c.valueFilters
	ctx.templates = map[string]Template{}
	for name, tpl := range c.templates {
		ctx.templates[name] = tpl
	}
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

func (n *DynCallNode) Compile(ctx *CompileContext, mode int) error {
	err := n.Value.Compile(ctx, CompilePush)
	if err != nil {
		return err
	}
	for _, arg := range n.Args {
		err = arg.Compile(ctx, CompilePush)
		if err != nil {
			return err
		}
	}
	ctx.DynCall(mode, len(n.Args))
	return nil
}

func (n *EmitNode) Compile(ctx *CompileContext, mode int) error {
	for _, node := range n.Nodes {
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

func (n *CompareNode) Compile(ctx *CompileContext, mode int) error {
	err := n.Left.Compile(ctx, CompilePush)
	if err != nil {
		return err
	}
	err = n.Right.Compile(ctx, CompilePush)
	if err != nil {
		return err
	}

	ctx.Compare(mode, n.Compare)
	return nil
}

func (n *AndNode) Compile(ctx *CompileContext, mode int) error {
	jumpLabels := []int{}
	lastIdx := len(n.Exprs) - 1

	for i, n := range n.Exprs {
		err := n.Compile(ctx, CompilePush)
		if err != nil {
			return err
		}

		if i != lastIdx {
			jumpLabels = append(jumpLabels, len(ctx.code))
			ctx.pushInstr(jumpFalse, 0, "")
			ctx.pushInstr(discardPop, 0, "")
		}
	}

	for _, lb := range jumpLabels {
		ctx.code[lb].iarg = len(ctx.code) - lb - 1
	}

	switch mode {
	case CompileEmit:
		ctx.pushInstr(emitPop, 0, "")
	}
	return nil
}

func (n *OrNode) Compile(ctx *CompileContext, mode int) error {
	jumpLabels := []int{}
	lastIdx := len(n.Exprs) - 1

	for i, n := range n.Exprs {
		err := n.Compile(ctx, CompilePush)
		if err != nil {
			return err
		}

		if i != lastIdx {
			jumpLabels = append(jumpLabels, len(ctx.code))
			ctx.pushInstr(jumpTrue, 0, "")
			ctx.pushInstr(discardPop, 0, "")
		}
	}

	for _, lb := range jumpLabels {
		ctx.code[lb].iarg = len(ctx.code) - lb - 1
	}

	switch mode {
	case CompileEmit:
		ctx.pushInstr(emitPop, 0, "")
	}
	return nil
}

func (n *BinaryOPNode) Compile(ctx *CompileContext, mode int) error {
	if len(n.Ops) == 0 {
		return n.Expr.Compile(ctx, mode)
	}

	err := n.Expr.Compile(ctx, CompilePush)
	if err != nil {
		return err
	}
	lastIndex := len(n.Ops) - 1
	for i, op := range n.Ops {
		err := op.Expr.Compile(ctx, CompilePush)
		if err != nil {
			return err
		}
		if i == lastIndex {
			ctx.BinaryOP(mode, op.Op)
		} else {
			ctx.BinaryOP(CompilePush, op.Op)
		}
	}
	return nil
}

func (n *NumberNode) Compile(ctx *CompileContext, mode int) error {
	ctx.Number(mode, n.Value)
	return nil
}

func (n *BlockNode) Compile(ctx *CompileContext, mode int) error {
	subprog, err := ctx.WithSubprog(n.Args, func() error {
		for _, n := range n.Body {
			err := n.Compile(ctx, CompileEmit)
			if err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return err
	}

	ctx.Subprog(CompilePush, subprog)
	ctx.pushInstr(declarePop, 0, n.Name)
	return nil
}

func (n *IfNode) Compile(ctx *CompileContext, mode int) (err error) {
	const (
		labelNextBranchStart = iota
		labelEnd
	)

	type jumpLabel struct {
		idx    int
		branch int
		label  int
	}

	jumpLabels := []jumpLabel{}
	branchStarts := make([]int, len(n.Branches)+1) // +1 for the alternative

	for i := range n.Branches {
		if i != 0 {
			// discard expr from previous branch
			ctx.pushInstr(discardPop, 0, "")
		}

		b := &n.Branches[i]
		branchStarts[i] = len(ctx.code)

		err = b.Expr.Compile(ctx, CompilePush)
		if err != nil {
			return err
		}

		// jump to the start of the next branch
		jumpLabels = append(jumpLabels, jumpLabel{len(ctx.code), i, labelNextBranchStart})
		ctx.pushInstr(jumpFalse, 0, "")

		ctx.pushInstr(discardPop, 0, "")
		for _, n := range b.Body {
			err = n.Compile(ctx, mode)
			if err != nil {
				return err
			}
		}

		jumpLabels = append(jumpLabels, jumpLabel{len(ctx.code), i, labelEnd})
		ctx.pushInstr(jump, 0, "")
	}

	// discard expr from the last branch if it was skipped
	if len(n.Branches) > 0 {
		ctx.pushInstr(discardPop, 0, "")
	}

	// compile the alternative (else) branch
	branchStarts[len(n.Branches)] = len(ctx.code)
	for _, n := range n.Alt {
		err = n.Compile(ctx, mode)
		if err != nil {
			return
		}
	}

	end := len(ctx.code)

	// update the labels
	for _, lb := range jumpLabels {
		switch lb.label {
		case labelNextBranchStart:
			ctx.code[lb.idx].iarg = branchStarts[lb.branch+1] - lb.idx - 1
		case labelEnd:
			ctx.code[lb.idx].iarg = end - lb.idx - 1
		}

	}

	return
}

func (n *DeclareNode) Compile(ctx *CompileContext, mode int) error {
	err := n.Value.Compile(ctx, CompilePush)
	if err != nil {
		return err
	}
	ctx.pushInstr(declarePop, 0, n.Name)
	return nil
}

func (n *ForNode) Compile(ctx *CompileContext, mode int) error {
	var loopJumps []loopJump
	err := n.Expr.Compile(ctx, CompilePush)
	if err != nil {
		return err
	}

	ctx.pushInstr(pushIter, 0, "")
	ctx.pushInstr(beginScope, 0, "")
	ctx.pushInstr(push, 0, "")
	ctx.pushInstr(declarePop, 0, n.Var)

	nextIndex := len(ctx.code)
	ctx.pushInstr(iterNextOrJump, 0, n.Var)
	err = ctx.WithLoopJumps(&loopJumps, func() error {
		for _, n := range n.Body {
			err := n.Compile(ctx, mode)
			if err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return err
	}

	ctx.pushInstr(jump, nextIndex-len(ctx.code)-1, "")
	endIndex := len(ctx.code)
	ctx.code[nextIndex].iarg = endIndex - nextIndex - 1

	// update the loop labels
	for _, jmp := range loopJumps {
		switch jmp.kind {
		case loopJumpNext:
			ctx.code[jmp.idx].iarg = nextIndex - jmp.idx - 1
		case loopJumpEnd:
			ctx.code[jmp.idx].iarg = endIndex - jmp.idx - 1
		}
	}
	return nil
}

func (n *IncludeNode) Compile(ctx *CompileContext, mode int) error {
	if name, ok := n.Name.(*ValueNode); ok {
		ctx.pushInstr(includeTemplate, 0, name.Value)
	} else {
		err := n.Name.Compile(ctx, CompilePush)
		if err != nil {
			return err
		}
		ctx.pushInstr(includeTemplateDyn, 0, "")
	}
	return nil
}

type discardFilter struct{}

var DiscardFilter ValueFilter = discardFilter{}

func (discardFilter) Filter(s string) (string, error) {
	return "", nil
}

func (n *DiscardNode) Compile(ctx *CompileContext, mode int) error {
	ctx.PushOutputFilter(DiscardFilter)
	for _, n := range n.Body {
		err := n.Compile(ctx, mode)
		if err != nil {
			return err
		}
	}
	ctx.PopOutputFilter()
	return nil
}

func (n *ObjectNode) Compile(ctx *CompileContext, mode int) error {
	if n.Extend == nil {
		ctx.pushInstr(pushObject, 0, "")
	} else {
		err := n.Extend.Compile(ctx, CompilePush)
		if err != nil {
			return err
		}
		ctx.pushInstr(extendObject, 0, "")
	}
	for _, key := range n.Keys {
		err := key.Value.Compile(ctx, CompilePush)
		if err != nil {
			return err
		}
		ctx.pushInstr(assignKey, 0, key.Key)
	}

	if mode == CompileEmit {
		ctx.pushInstr(emitPop, 0, "")
	}
	return nil
}

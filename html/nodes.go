package html

import "github.com/phipus/tplexpr"

type TextNode struct {
	Value tplexpr.Node
}

func (n *TextNode) Compile(ctx *tplexpr.CompileContext, mode int) error {
	ctx.PushOutputFilter(HtmlEscapeFilter)
	err := n.Value.Compile(ctx, mode)
	if err == nil {
		ctx.PopOutputFilter()
	}
	return err
}

type CommentNode struct {
	Body tplexpr.Node
}

func (n *CommentNode) Compile(ctx *tplexpr.CompileContext, mode int) error {
	ctx.EmitValue("<!--")
	ctx.PushOutputFilter(CommentEscapeFilter)
	err := n.Body.Compile(ctx, mode)
	if err != nil {
		return err
	}
	ctx.PopOutputFilter()
	ctx.EmitValue("-->")
	return nil
}

type SwitchNode struct {
	Expr tplexpr.Node
	If   tplexpr.IfNode
}

type PushPeek struct{}

func (PushPeek) Compile(ctx *tplexpr.CompileContext, mode int) error {
	ctx.PushPeek()
	return nil
}

func (n *SwitchNode) Compile(ctx *tplexpr.CompileContext, mode int) error {
	err := n.Expr.Compile(ctx, tplexpr.CompilePush)
	if err != nil {
		return err
	}

	branches := make([]tplexpr.IfBranch, len(n.If.Branches))
	for i := range n.If.Branches {
		branches[i] = tplexpr.IfBranch{
			Expr: &tplexpr.CompareNode{
				Compare: tplexpr.EQ,
				Left:    &PushPeek{},
				Right:   n.If.Branches[i].Expr,
			},
			Body: n.If.Branches[i].Body,
		}
	}

	ifNode := tplexpr.IfNode{Branches: branches, Alt: n.If.Alt}
	err = ifNode.Compile(ctx, mode)
	if err != nil {
		return err
	}
	ctx.DiscardPop()
	return nil
}

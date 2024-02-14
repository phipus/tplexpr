package tplexpr

type Instr struct {
	op   int
	iarg int
	sarg string
}

const (
	emit = iota
	push
	emitFetch
	pushFetch
	emitCall
	pushCall
	emitCallDyn
	pushCallDyn
	emitAttr
	pushAttr
	emitSubprog
	pushSubprog
)

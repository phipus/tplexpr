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
	emitCompare
	pushCompare
	jump
	jumpTrue
	jumpFalse
	emitPop
	discardPop
	pushNot
	emitNot
)

const (
	EQ = iota
	NE
	GT
	GE
	LT
	LE
)

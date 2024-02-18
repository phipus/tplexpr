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
	storePop
	declarePop
	pushPeek
	emitNot
	pushNot
	emitBinaryOP
	pushBinaryOP
	emitNumber
	pushNumber
	pushIter
	iterNextOrJump
	discardIter
	beginScope
	endScope
	pushOutputFilter
	popOutputFilter
	includeTemplate
	includeTemplateDyn
)

// Compare constants
const (
	EQ = iota
	NE
	GT
	GE
	LT
	LE
)

// Binary OP Constatns
const (
	ADD = iota
	SUB
	MUL
	DIV
)

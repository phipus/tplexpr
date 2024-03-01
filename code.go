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
	emitCallSubprogNA
	pushCallSubprogNA
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
	emitNil
	pushNil
	pushIter
	iterNextOrJump
	discardIter
	beginScope
	endScope
	pushOutputFilter
	popOutputFilter
	emitTemplate
	pushTemplate
	emitTemplateDyn
	pushTemplateDyn
	assignKey
	assignKeyDyn
	pushObject
	extendObject
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

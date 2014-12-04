package backend

type BinOp int
type UnaryOp int
type RelOp int

const (
	Sub BinOp = iota
	Add
	Mul
	Div
	Mod
	And
	Or
)

const (
	Not UnaryOp = iota
	Ord
	Chr
	Neg
	Len
)

const (
	LT RelOp = iota
	LE
	GT
	GE
	EQ
	NE
)

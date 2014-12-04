package backend

type BinOp int
type UnaryOp int
type RelOp int

const (
	Sub BinOp = itoa
	Add BinOp
	Mul BinOp
	Div BinOp
	Mod BinOp
	And BinOp
	Or  BinOp
)

const (
	Not UnaryOp = itoa
	Ord UnaryOp
	Chr UnaryOp
	Neg UnaryOp
	Len UnaryOp
)

const (
	LT RelOp = itoa
	LE RelOp
	GT RelOp
	GE RelOp
	EQ RelOp
	NE RelOp
)

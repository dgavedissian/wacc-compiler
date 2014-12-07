package backend

var regWidth = 4

const (
	Sub string = "-"
	Add string = "+"
	Mul string = "*"
	Div string = "/"
	Mod string = "%"
	And string = "&&"
	Or  string = "||"
)

const (
	Not string = "!"
	Ord string = "ord"
	Chr string = "chr"
	Neg string = "-"
	Len string = "len"
)

const (
	LT string = "<"
	LE string = "<="
	GT string = ">"
	GE string = ">="
	EQ string = "=="
	NE string = "!="
)

const (
	RuntimeOverflowLabel         string = "_wacc_throw_overflow_error"
	RuntimeCheckDivZeroLabel     string = "_wacc_check_divide_by_zero"
	RuntimeCheckArrayBoundsLabel string = "_wacc_check_array_bounds"
)

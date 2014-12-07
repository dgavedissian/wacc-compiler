package backend

import (
	"fmt"
	"strings"
	"unicode"

	"../frontend"
)

type Expr interface {
	expr()
	Repr() string
	Weight() int

	allocateRegisters(*RegisterAllocatorContext, int)
}

// Type annotation pseudo-expression
type TypeExpr struct {
	Type frontend.Type
}

type IntConstExpr struct {
	Value int
}

type BoolConstExpr struct {
	Value bool
}

type CharConstExpr struct {
	Value rune
	Size  int
}

type StringConstExpr struct {
	Value string
}

type PointerConstExpr struct {
	Value int
}

type ArrayConstExpr struct {
	Type  frontend.Type
	Elems []Expr
}

type LocationExpr struct {
	Label string
}

type VarExpr struct {
	Name string
}

type MemExpr struct {
	Address *RegisterExpr
	Offset  int
}

type RegisterExpr struct {
	Id int
}

type StackLocationExpr struct {
	Id int
}

type ArrayElemExpr struct {
	Array Expr
	Index Expr
}

type PairElemExpr struct {
	Fst     bool
	Operand *VarExpr
}

type UnaryExpr struct {
	Operator string
	Operand  Expr
	Type     frontend.Type
}

type BinaryExpr struct {
	Operator string
	Left     Expr
	Right    Expr
	Type     frontend.Type
}

type NewPairExpr struct {
	Left  Expr
	Right Expr
}

type CallExpr struct {
	Label *LocationExpr
	Args  []Expr
}

func (TypeExpr) expr()          {}
func (e TypeExpr) Repr() string { return "TYPE" }
func (TypeExpr) Weight() int    { return 0 }

func (IntConstExpr) expr()          {}
func (e IntConstExpr) Repr() string { return fmt.Sprintf("INT %v", e.Value) }
func (IntConstExpr) Weight() int    { return 1 }

func (BoolConstExpr) expr()          {}
func (e BoolConstExpr) Repr() string { return fmt.Sprintf("BOOL %v", e.Value) }
func (BoolConstExpr) Weight() int    { return 1 }

func (CharConstExpr) expr() {}
func (e CharConstExpr) Repr() string {
	if unicode.IsPrint(e.Value) {
		return fmt.Sprintf("CHAR \"%v\"", string(e.Value))
	} else {
		return fmt.Sprintf("CHAR %v", e.Value)
	}
}
func (CharConstExpr) Weight() int { return 1 }

func (StringConstExpr) expr()          {}
func (e StringConstExpr) Repr() string { return fmt.Sprintf("STRING %#v", e.Value) }
func (StringConstExpr) Weight() int    { return 1 }

func (PointerConstExpr) expr()          {}
func (e PointerConstExpr) Repr() string { return fmt.Sprintf("PTR 0x%x", e.Value) }
func (PointerConstExpr) Weight() int    { return 1 }

func (ArrayConstExpr) expr() {}
func (e ArrayConstExpr) Repr() string {
	rs := make([]string, len(e.Elems))
	for i, v := range e.Elems {
		rs[i] = v.Repr()
	}
	return "ARRAY [" + strings.Join(rs, ", ") + "]"
}

func (e ArrayConstExpr) Weight() int { return len(e.Elems) }
func (LocationExpr) expr()           {}
func (e LocationExpr) Repr() string  { return "LABEL " + e.Label }
func (LocationExpr) Weight() int     { return 1 }

func (VarExpr) expr()          {}
func (e VarExpr) Repr() string { return "VAR " + e.Name }
func (VarExpr) Weight() int    { return 1 }

func (MemExpr) expr()          {}
func (e MemExpr) Repr() string { return fmt.Sprintf("MEM %v +%v", e.Address.Repr(), e.Offset) }
func (MemExpr) Weight() int    { return 1 }

func (RegisterExpr) expr()          {}
func (e RegisterExpr) Repr() string { return fmt.Sprintf("r%d", e.Id) }
func (RegisterExpr) Weight() int    { return 1 }

func (StackLocationExpr) expr()          {}
func (e StackLocationExpr) Repr() string { return fmt.Sprintf("STACK_%d", e.Id) }
func (StackLocationExpr) Weight() int    { return 1 }

func (ArrayElemExpr) expr() {}
func (e ArrayElemExpr) Repr() string {
	return fmt.Sprintf("ARRAY ELEM %v IN %v", e.Index.Repr(), e.Array.Repr())
}
func (ArrayElemExpr) Weight() int { return 1 }

func (PairElemExpr) expr() {}
func (e PairElemExpr) Repr() string {
	if e.Fst {
		return fmt.Sprintf("FST %v", e.Operand.Repr())
	} else {
		return fmt.Sprintf("SND %v", e.Operand.Repr())
	}
}
func (PairElemExpr) Weight() int { return 1 }

func (UnaryExpr) expr() {}
func (e UnaryExpr) Repr() string {
	return fmt.Sprintf("UNARY %v %v (%v)", e.Type.Repr(), e.Operator, e.Operand.Repr())
}
func (e UnaryExpr) Weight() int { return e.Operand.Weight() + 1 }

func (BinaryExpr) expr() {}
func (e BinaryExpr) Repr() string {
	return fmt.Sprintf("BINARY %v %v (%v) (%v)", e.Type.Repr(), e.Operator, e.Left.Repr(), e.Right.Repr())
}
func (e BinaryExpr) Weight() int { return e.Left.Weight() + e.Right.Weight() + 1 }

func (NewPairExpr) expr() {}
func (e NewPairExpr) Repr() string {
	return fmt.Sprintf("NEWPAIR %v %v", e.Left.Repr(), e.Right.Repr())
}
func (e NewPairExpr) Weight() int { return e.Left.Weight() + e.Right.Weight() + 1 }

func (CallExpr) expr() {}
func (e CallExpr) Repr() string {
	args := make([]string, len(e.Args))
	for i, arg := range e.Args {
		args[i] = arg.Repr()
	}
	return fmt.Sprintf("CALL %v (%s)", e.Label.Label, strings.Join(args, ", "))
}
func (e CallExpr) Weight() int {
	x := 1
	for _, arg := range e.Args {
		x += arg.Weight()
	}
	return x
}

type InstrNode struct {
	Instr      Instr
	stackSpace int
	Next       *InstrNode
}

type Instr interface {
	instr()
	Repr() string

	allocateRegisters(*RegisterAllocatorContext)
	generateCode(*GeneratorContext)
}

type NoOpInstr struct {
}

type LabelInstr struct {
	Label string
}

type ReadInstr struct {
	Dst Expr // LValueExpr
}

type FreeInstr struct {
	Object Expr // LValueExpr
}

type ReturnInstr struct {
	Expr Expr
}

type ExitInstr struct {
	Expr Expr
}

type PrintInstr struct {
	Expr Expr
	Type frontend.Type
}

type MoveInstr struct {
	Dst Expr // LValueExpr
	Src Expr
}

type NotInstr struct {
	Dst Expr // LValueExpr
	Src Expr
}

type NegInstr struct {
	Expr Expr
}

type CmpInstr struct {
	Left     Expr
	Right    Expr
	Dst      Expr
	Operator string
}

type JmpInstr struct {
	Dst *InstrNode
}

type JmpCondInstr struct {
	Dst  *InstrNode
	Cond Expr
}

//
// Meta data
//
type DeclareInstr struct {
	Var  *VarExpr
	Type frontend.Type
}

type PushScopeInstr struct {
}

type PopScopeInstr struct {
}

type DeclareTypeInstr struct {
	Dst  Expr
	Type *TypeExpr
}

func (DeclareInstr) instr() {}
func (e DeclareInstr) Repr() string {
	return fmt.Sprintf("DECLARE %v OF TYPE %v", e.Var.Name, e.Type.Repr())
}

func (PushScopeInstr) instr() {}
func (e PushScopeInstr) Repr() string {
	return fmt.Sprintf("PUSH SCOPE")
}

func (PopScopeInstr) instr() {}
func (e PopScopeInstr) Repr() string {
	return fmt.Sprintf("POP SCOPE")
}

func (DeclareTypeInstr) instr() {}
func (e DeclareTypeInstr) Repr() string {
	return fmt.Sprintf("TYPE OF %v IS %v", e.Dst.Repr(), e.Type.Type.Repr())
}

// Second stage instructions

// Shift
type Shift interface {
	shift()
	Repr() string
}

type LSL struct {
	Value int
}

func (*LSL) expr()          {}
func (*LSL) shift()         {}
func (s *LSL) Repr() string { return fmt.Sprintf("lsl #%v", s.Value) }

// Binary operations
type AddInstr struct {
	Dst      *RegisterExpr
	Op1      *RegisterExpr
	Op2      Expr
	Op2Shift Shift
}

type SubInstr struct {
	Dst      *RegisterExpr
	Op1      *RegisterExpr
	Op2      Expr
	Op2Shift Shift
}

type MulInstr struct {
	Dst *RegisterExpr
	Op1 *RegisterExpr
	Op2 *RegisterExpr
}

type DivInstr struct {
	Dst *RegisterExpr
	Op1 *RegisterExpr
	Op2 *RegisterExpr
}

type AndInstr struct {
	Dst *RegisterExpr
	Op1 *RegisterExpr
	Op2 *RegisterExpr
}

type OrInstr struct {
	Dst *RegisterExpr
	Op1 *RegisterExpr
	Op2 *RegisterExpr
}

type CallInstr struct {
	Label *LocationExpr
}

type HeapAllocInstr struct {
	Dst  *RegisterExpr
	Size int
}

func (NoOpInstr) instr()       {}
func (NoOpInstr) Repr() string { return "NOOP" }

func (LabelInstr) instr() {}
func (i LabelInstr) Repr() string {
	return fmt.Sprintf("LABEL %s", i.Label)
}

func (ReadInstr) instr() {}
func (i ReadInstr) Repr() string {
	return fmt.Sprintf("READ %s", i.Dst.Repr())
}

func (FreeInstr) instr() {}
func (i FreeInstr) Repr() string {
	return fmt.Sprintf("FREE %s", i.Object.Repr())
}

func (ReturnInstr) instr() {}
func (i ReturnInstr) Repr() string {
	return fmt.Sprintf("RETURN %s", i.Expr.Repr())
}

func (ExitInstr) instr() {}
func (i ExitInstr) Repr() string {
	return fmt.Sprintf("EXIT %s", i.Expr.Repr())
}

func (PrintInstr) instr() {}
func (i PrintInstr) Repr() string {
	if i.Type == nil {
		return fmt.Sprintf("PRINT <nil> %s", i.Expr.Repr())
	} else {
		return fmt.Sprintf("PRINT %v %s", i.Type.Repr(), i.Expr.Repr())
	}
}

func (MoveInstr) instr() {}
func (i MoveInstr) Repr() string {
	return fmt.Sprintf("MOVE (%s) (%s)", i.Dst.Repr(), i.Src.Repr())
}

func (NotInstr) instr() {}
func (i NotInstr) Repr() string {
	return fmt.Sprintf("NOT (%s) (%s)", i.Dst.Repr(), i.Src.Repr())
}

func (NegInstr) instr() {}
func (i NegInstr) Repr() string {
	return fmt.Sprintf("NEG (%v)", i.Expr.Repr())
}

func (CmpInstr) instr() {}
func (i CmpInstr) Repr() string {
	return fmt.Sprintf("CMP %v (%v) (%v) (%v)", i.Operator, i.Dst.Repr(), i.Left.Repr(), i.Right.Repr())
}

func (JmpInstr) instr() {}
func (i JmpInstr) Repr() string {
	return fmt.Sprintf("JMP (%s)", i.Dst.Instr.(*LabelInstr).Repr())
}

func (JmpCondInstr) instr() {}
func (i JmpCondInstr) Repr() string {
	return fmt.Sprintf("J%s (%s)", i.Cond, i.Dst.Instr.(*LabelInstr).Repr())
}

func (*AddInstr) instr() {}
func (i *AddInstr) Repr() string {
	if i.Op2Shift != nil {
		return fmt.Sprintf("ADD %v %v %v %v", i.Dst.Repr(), i.Op1.Repr(), i.Op2.Repr(), i.Op2Shift.Repr())
	} else {
		return fmt.Sprintf("ADD %v %v %v", i.Dst.Repr(), i.Op1.Repr(), i.Op2.Repr())
	}
}

func (*SubInstr) instr() {}
func (i *SubInstr) Repr() string {
	if i.Op2Shift != nil {
		return fmt.Sprintf("SUB %v %v %v %v", i.Dst.Repr(), i.Op1.Repr(), i.Op2.Repr(), i.Op2Shift.Repr())
	} else {
		return fmt.Sprintf("SUB %v %v %v", i.Dst.Repr(), i.Op1.Repr(), i.Op2.Repr())
	}
}

func (*MulInstr) instr() {}
func (i *MulInstr) Repr() string {
	return fmt.Sprintf("MUL %v %v %v", i.Dst.Repr(), i.Op1.Repr(), i.Op2.Repr())
}

func (*DivInstr) instr() {}
func (i *DivInstr) Repr() string {
	return fmt.Sprintf("DIV %v %v %v", i.Dst.Repr(), i.Op1.Repr(), i.Op2.Repr())
}

func (*AndInstr) instr() {}
func (i *AndInstr) Repr() string {
	return fmt.Sprintf("AND %v %v %v", i.Dst.Repr(), i.Op1.Repr(), i.Op2.Repr())
}

func (*OrInstr) instr() {}
func (i *OrInstr) Repr() string {
	return fmt.Sprintf("OR %v %v %v", i.Dst.Repr(), i.Op1.Repr(), i.Op2.Repr())
}

func (*CallInstr) instr() {}
func (i *CallInstr) Repr() string {
	return fmt.Sprintf("CALL %v", i.Label.Label)
}

func (*HeapAllocInstr) instr() {}
func (i *HeapAllocInstr) Repr() string {
	return fmt.Sprintf("ALLOC %v SIZE %v", i.Dst.Repr(), i.Size)
}

/*
		Toothless defends this code

	                         ^\    ^
	                        / \\  / \
	                       /.  \\/   \      |\___/|
	    *----*           / / |  \\    \  __/  O  O\
	    |   /          /  /  |   \\    \_\/  \     \
	   / /\/         /   /   |    \\   _\/    '@___@
	  /  /         /    /    |     \\ _\/       |U
	  |  |       /     /     |      \\\/        |
	  \  |     /_     /      |       \\  )   \ _|_
	  \   \       ~-./_ _    |    .- ; (  \_ _ _,\'
	  ~    ~.           .-~-.|.-*      _        {-,
	   \      ~-. _ .-~                 \      /\'
	    \                   }            {   .*
	     ~.                 '-/        /.-~----.
	       ~- _             /        >..----.\\\
	           ~ - - - - ^}_ _ _ _ _ _ _.-\\\

		To whoever reads from here onwards, I'm sorry...
*/
func DrawIFGraph(iform *IFContext) {
	// Transform into a list
	var list []Instr

	// Functions
	for _, f := range iform.functions {
		node := f
		for node != nil {
			list = append(list, node.Instr)
			node = node.Next
		}
	}

	// Main
	node := iform.main
	for node != nil {
		list = append(list, node.Instr)
		node = node.Next
	}
	instrCount := len(list)

	// Referrals
	var referredBy Instr
	referStack := 0

	// Iterate
	for i, instr := range list {
		fmt.Printf("%d  ", i)

		// Are we a label?
		if _, ok := instr.(*LabelInstr); ok {
			// Does anyone else refer to me?
			for j := 0; j < instrCount; j++ {
				switch jmp := list[j].(type) {
				case *JmpInstr:
					if jmp.Dst.Instr == instr {
						referredBy = jmp
						referStack++
						break
					}
				}
			}

			fmt.Printf("|")
			for l := 0; l < referStack; l++ {
				fmt.Printf("<-")
			}
		} else {
			fmt.Printf("|")

			// Have we reached the referred by?
			if instr == referredBy {
				referStack--
				referredBy = nil
				fmt.Printf("-'")
			}

			for l := 0; l < referStack; l++ {
				fmt.Printf(" |")
			}
		}

		// Instruction
		fmt.Printf("  %s\n", instr.Repr())
	}
}

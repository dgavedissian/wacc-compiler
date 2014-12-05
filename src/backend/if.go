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

	allocateRegisters(*RegisterAllocatorContext, int)
}

type IntConstExpr struct {
	Value int
}

type CharConstExpr struct {
	Value rune
	Size  int
}

type ArrayExpr struct {
	Type  frontend.BasicType
	Elems []Expr
}

type LocationExpr struct {
	Label string
}

type VarExpr struct {
	Name string
}

type MemExpr struct {
	Address Expr
}

type RegisterExpr struct {
	Id int
}

type BinOpExpr struct {
	Left  Expr
	Right Expr
}

type NotExpr struct {
	Operand Expr
}

type OrdExpr struct {
	Operand Expr
}

type ChrExpr struct {
	Operand Expr
}

type NegExpr struct {
	Operand Expr
}

type LenExpr struct {
	Operand Expr
}

type NewPairExpr struct {
	Left  Expr
	Right Expr
}

type CallExpr struct {
	Ident string
	Args  []Expr
}

func (IntConstExpr) expr()          {}
func (e IntConstExpr) Repr() string { return fmt.Sprintf("INT %v", e.Value) }

func (CharConstExpr) expr() {}
func (e CharConstExpr) Repr() string {
	if unicode.IsPrint(e.Value) {
		return fmt.Sprintf("CHAR \"%v\"", string(e.Value))
	} else {
		return fmt.Sprintf("CHAR %v", e.Value)
	}
}

func (LocationExpr) expr()          {}
func (e LocationExpr) Repr() string { return "LABEL " + e.Label }

func (VarExpr) expr()          {}
func (e VarExpr) Repr() string { return "VAR " + e.Name }

func (RegisterExpr) expr()          {}
func (e RegisterExpr) Repr() string { return fmt.Sprintf("r%d", e.Id) }

func (ArrayExpr) expr() {}
func (e ArrayExpr) Repr() string {
	rs := make([]string, len(e.Elems))
	for i, v := range e.Elems {
		rs[i] = v.Repr()
	}
	return "ARRAYCONST [" + strings.Join(rs, ", ") + "]"
}

func (BinOpExpr) expr() {}
func (e BinOpExpr) Repr() string {
	return fmt.Sprintf("BINOP %s %s", e.Left.Repr(), e.Right.Repr())
}

func (NotExpr) expr() {}
func (e NotExpr) Repr() string {
	return fmt.Sprintf("NOT %v", e.Operand)
}
func (OrdExpr) expr() {}
func (e OrdExpr) Repr() string {
	return fmt.Sprintf("Ord %v", e.Operand)
}
func (ChrExpr) expr() {}
func (e ChrExpr) Repr() string {
	return fmt.Sprintf("Chr %v", e.Operand)
}
func (NegExpr) expr() {}
func (e NegExpr) Repr() string {
	return fmt.Sprintf("Neg ^v", e.Operand)
}
func (LenExpr) expr() {}
func (e LenExpr) Repr() string {
	return fmt.Sprintf("Len %v", e.Operand)
}
func (NewPairExpr) expr() {}
func (e NewPairExpr) Repr() string {
	return fmt.Sprintf("NEWPAIR %v %v", e.Left.Repr(), e.Right.Repr())
}

func (CallExpr) expr() {}
func (e CallExpr) Repr() string {
	args := make([]string, len(e.Args))
	for i, arg := range e.Args {
		args[i] = arg.Repr()
	}
	return fmt.Sprintf("CALL %v (%s)", e.Ident, strings.Join(args, ", "))
}

type InstrNode struct {
	Instr Instr
	Next  *InstrNode
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
}

type MoveInstr struct {
	Dst Expr // LValueExpr
	Src Expr
}

type TestInstr struct {
	Cond Expr
}

type JmpInstr struct {
	Dst *InstrNode
}

type JmpEqualInstr struct {
	Dst *InstrNode
}

type PushScopeInstr struct {
}

type PopScopeInstr struct {
}

func (PushScopeInstr) instr() {}
func (e PushScopeInstr) Repr() string {
	return fmt.Sprintf("PUSH SCOPE")
}

func (PopScopeInstr) instr() {}
func (e PopScopeInstr) Repr() string {
	return fmt.Sprintf("POP SCOPE")
}

// Second stage instructions
type AddInstr struct {
	Dst *RegisterExpr
	Op1 *RegisterExpr
	Op2 *RegisterExpr
}

type CallInstr struct {
	Ident string
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
	return fmt.Sprintf("PRINT %s", i.Expr.Repr())
}

func (MoveInstr) instr() {}
func (i MoveInstr) Repr() string {
	return fmt.Sprintf("MOVE (%s) (%s)", i.Dst.Repr(), i.Src.Repr())
}

func (TestInstr) instr() {}
func (i TestInstr) Repr() string {
	return fmt.Sprintf("TEST (%s)", i.Cond.Repr())
}

func (JmpInstr) instr() {}
func (i JmpInstr) Repr() string {
	return fmt.Sprintf("JMP (%s)", i.Dst.Instr.(*LabelInstr).Repr())
}

func (JmpEqualInstr) instr() {}
func (i JmpEqualInstr) Repr() string {
	return fmt.Sprintf("JEQ (%s)", i.Dst.Instr.(*LabelInstr).Repr())
}

func (*AddInstr) instr() {}
func (i *AddInstr) Repr() string {
	return fmt.Sprintf("ADD %v %v %v", i.Dst.Repr(), i.Op1.Repr(), i.Op2.Repr())
}

func (*CallInstr) instr() {}
func (i *CallInstr) Repr() string {
	return fmt.Sprintf("CALL %v", i.Ident)
}

type IFContext struct {
	labels    map[string]Instr
	main      *InstrNode
	functions map[string]*InstrNode
	current   *InstrNode
	nextTemp  int
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

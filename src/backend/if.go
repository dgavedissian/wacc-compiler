package backend

import (
	"fmt"
	"strings"
	"unicode"

	"../frontend"
)

//
// Expressions
//
type Expr interface {
	Copy() Expr

	expr()
	Repr() string
	Weight() int

	allocateRegisters(*RegisterAllocatorContext, *RegisterExpr)
}

// Type annotation pseudo-expression
type TypeExpr struct {
	Type frontend.Type
}

type IntConstExpr struct {
	Value int
}

type FloatConstExpr struct {
	Value float32
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

type StackArgumentExpr struct {
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

type StructElemExpr struct {
	StructIdent *VarExpr
	ElemIdent   *VarExpr
	ElemOffset  int
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

type NewStructExpr struct {
	Label *LocationExpr
	Args  []Expr
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
func (e TypeExpr) Copy() Expr   { return &TypeExpr{e.Type} }

func (IntConstExpr) expr()          {}
func (e IntConstExpr) Repr() string { return fmt.Sprintf("INT %v", e.Value) }
func (IntConstExpr) Weight() int    { return 1 }
func (e IntConstExpr) Copy() Expr   { return &IntConstExpr{e.Value} }

func (FloatConstExpr) expr()          {}
func (e FloatConstExpr) Repr() string { return fmt.Sprintf("FLOAT %v", e.Value) }
func (FloatConstExpr) Weight() int    { return 1 }
func (e FloatConstExpr) Copy() Expr   { return &FloatConstExpr{e.Value} }

func (BoolConstExpr) expr()          {}
func (e BoolConstExpr) Repr() string { return fmt.Sprintf("BOOL %v", e.Value) }
func (BoolConstExpr) Weight() int    { return 1 }
func (e BoolConstExpr) Copy() Expr   { return &BoolConstExpr{e.Value} }

func (CharConstExpr) expr() {}
func (e CharConstExpr) Repr() string {
	if unicode.IsPrint(e.Value) {
		return fmt.Sprintf("CHAR \"%v\"", string(e.Value))
	} else {
		return fmt.Sprintf("CHAR %v", e.Value)
	}
}
func (CharConstExpr) Weight() int  { return 1 }
func (e CharConstExpr) Copy() Expr { return &CharConstExpr{e.Value, e.Size} }

func (StringConstExpr) expr()          {}
func (e StringConstExpr) Repr() string { return fmt.Sprintf("STRING %#v", e.Value) }
func (StringConstExpr) Weight() int    { return 1 }
func (e StringConstExpr) Copy() Expr   { return &StringConstExpr{e.Value} }

func (PointerConstExpr) expr()          {}
func (e PointerConstExpr) Repr() string { return fmt.Sprintf("PTR 0x%x", e.Value) }
func (PointerConstExpr) Weight() int    { return 1 }
func (e PointerConstExpr) Copy() Expr   { return &PointerConstExpr{e.Value} }

func (ArrayConstExpr) expr() {}
func (e ArrayConstExpr) Repr() string {
	rs := make([]string, len(e.Elems))
	for i, v := range e.Elems {
		rs[i] = v.Repr()
	}
	return "ARRAY [" + strings.Join(rs, ", ") + "]"
}
func (e ArrayConstExpr) Weight() int { return len(e.Elems) }
func (e ArrayConstExpr) Copy() Expr {
	copyElems := make([]Expr, len(e.Elems))
	for i, v := range e.Elems {
		copyElems[i] = v.Copy()
	}
	return &ArrayConstExpr{e.Type, copyElems}
}

func (LocationExpr) expr()          {}
func (e LocationExpr) Repr() string { return "LABEL " + e.Label }
func (LocationExpr) Weight() int    { return 1 }
func (e LocationExpr) Copy() Expr   { return &LocationExpr{e.Label} }

func (VarExpr) expr()          {}
func (e VarExpr) Repr() string { return "VAR " + e.Name }
func (VarExpr) Weight() int    { return 1 }
func (e VarExpr) Copy() Expr   { return &VarExpr{e.Name} }

func (MemExpr) expr()          {}
func (e MemExpr) Repr() string { return fmt.Sprintf("MEM %v +%v", e.Address.Repr(), e.Offset) }
func (MemExpr) Weight() int    { return 1 }
func (e MemExpr) Copy() Expr   { return &MemExpr{e.Address, e.Offset} }

func (RegisterExpr) expr()          {}
func (e RegisterExpr) Repr() string { return fmt.Sprintf("r%d", e.Id) }
func (RegisterExpr) Weight() int    { return 1 }
func (e RegisterExpr) Copy() Expr   { return &RegisterExpr{e.Id} }

func (StackLocationExpr) expr()          {}
func (e StackLocationExpr) Repr() string { return fmt.Sprintf("STACK_%d", e.Id) }
func (StackLocationExpr) Weight() int    { return 1 }
func (e StackLocationExpr) Copy() Expr   { return &StackLocationExpr{e.Id} }

func (StackArgumentExpr) expr()          {}
func (e StackArgumentExpr) Repr() string { return fmt.Sprintf("STARG_%d", e.Id) }
func (StackArgumentExpr) Weight() int    { return 1 }
func (e StackArgumentExpr) Copy() Expr   { return &StackArgumentExpr{e.Id} }

func (ArrayElemExpr) expr() {}
func (e ArrayElemExpr) Repr() string {
	return fmt.Sprintf("ARRAY ELEM %v IN %v", e.Index.Repr(), e.Array.Repr())
}
func (ArrayElemExpr) Weight() int  { return 1 }
func (e ArrayElemExpr) Copy() Expr { return &ArrayElemExpr{e.Array.Copy(), e.Index.Copy()} }

func (PairElemExpr) expr() {}
func (e PairElemExpr) Repr() string {
	if e.Fst {
		return fmt.Sprintf("FST %v", e.Operand.Repr())
	} else {
		return fmt.Sprintf("SND %v", e.Operand.Repr())
	}
}
func (PairElemExpr) Weight() int  { return 1 }
func (e PairElemExpr) Copy() Expr { return &PairElemExpr{e.Fst, e.Operand.Copy().(*VarExpr)} }

func (StructElemExpr) expr() {}
func (e StructElemExpr) Repr() string {
	return fmt.Sprintf("%v %v", e.StructIdent.Repr(), e.ElemIdent.Repr())
}
func (StructElemExpr) Weight() int { return 1 }
func (e StructElemExpr) Copy() Expr {
	return &StructElemExpr{
		e.StructIdent.Copy().(*VarExpr),
		e.ElemIdent.Copy().(*VarExpr),
		e.ElemOffset,
	}
}

func (UnaryExpr) expr() {}
func (e UnaryExpr) Repr() string {
	return fmt.Sprintf("UNARY %v %v (%v)", e.Type.Repr(), e.Operator, e.Operand.Repr())
}
func (e UnaryExpr) Weight() int { return e.Operand.Weight() + 1 }
func (e UnaryExpr) Copy() Expr  { return &UnaryExpr{e.Operator, e.Operand.Copy(), e.Type} }

func (BinaryExpr) expr() {}
func (e BinaryExpr) Repr() string {
	return fmt.Sprintf("BINARY %v %v (%v) (%v)", e.Type.Repr(), e.Operator, e.Left.Repr(), e.Right.Repr())
}
func (e BinaryExpr) Weight() int { return e.Left.Weight() + e.Right.Weight() + 1 }
func (e BinaryExpr) Copy() Expr  { return &BinaryExpr{e.Operator, e.Left.Copy(), e.Right.Copy(), e.Type} }

func (NewStructExpr) expr() {}
func (e NewStructExpr) Repr() string {
	args := make([]string, len(e.Args))
	for i, arg := range e.Args {
		args[i] = arg.Repr()
	}
	return fmt.Sprintf("NEWSTRUCT %v %v",
		e.Label.Label, strings.Join(args, ", "))
}
func (e NewStructExpr) Weight() int {
	x := 1
	for _, arg := range e.Args {
		x += arg.Weight()
	}
	return x
}
func (e NewStructExpr) Copy() Expr {
	newArgs := make([]Expr, len(e.Args))
	for i, v := range e.Args {
		newArgs[i] = v.Copy()
	}
	return &NewStructExpr{e.Label.Copy().(*LocationExpr), newArgs}
}

func (NewPairExpr) expr() {}
func (e NewPairExpr) Repr() string {
	return fmt.Sprintf("NEWPAIR %v %v", e.Left.Repr(), e.Right.Repr())
}
func (e NewPairExpr) Weight() int { return e.Left.Weight() + e.Right.Weight() + 1 }
func (e NewPairExpr) Copy() Expr  { return &NewPairExpr{e.Left.Copy(), e.Right.Copy()} }

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
func (e CallExpr) Copy() Expr {
	newArgs := make([]Expr, len(e.Args))
	for i, v := range e.Args {
		newArgs[i] = v.Copy()
	}
	return &CallExpr{e.Label.Copy().(*LocationExpr), newArgs}
}

//
// Instructions
//
type InstrNode struct {
	Instr      Instr
	stackSpace int
	Next       *InstrNode
	Prev       *InstrNode
}

func (i *InstrNode) Copy() *InstrNode {
	return &InstrNode{
		Instr:      i.Instr.Copy(),
		stackSpace: i.stackSpace,
		Next:       i.Next,
		Prev:       i.Prev,
	}
}

type Instr interface {
	instr()
	Repr() string
	Copy() Instr

	allocateRegisters(*RegisterAllocatorContext)
	generateCode(*GeneratorContext)
}

type NoOpInstr struct {
}

type LabelInstr struct {
	Label string
}

type EvalInstr struct {
	Expr Expr
}

type ReadInstr struct {
	Dst  Expr // LValueExpr
	Type frontend.Type
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
	StackSize int
}

type PopScopeInstr struct {
	StackSize int
}

type LocaleInstr struct {
}

func (DeclareInstr) instr() {}
func (e DeclareInstr) Repr() string {
	return fmt.Sprintf("DECLARE %v OF TYPE %v", e.Var.Name, e.Type.Repr())
}
func (e DeclareInstr) Copy() Instr {
	return &DeclareInstr{e.Var.Copy().(*VarExpr), e.Type}
}

func (PushScopeInstr) instr() {}
func (e PushScopeInstr) Repr() string {
	return fmt.Sprintf("PUSH SCOPE %v", e.StackSize)
}
func (e PushScopeInstr) Copy() Instr {
	return &PushScopeInstr{}
}

func (PopScopeInstr) instr() {}
func (e PopScopeInstr) Repr() string {
	return fmt.Sprintf("POP SCOPE %v", e.StackSize)
}
func (e PopScopeInstr) Copy() Instr {
	return &PopScopeInstr{}
}

func (LocaleInstr) instr() {}
func (e LocaleInstr) Repr() string {
	return fmt.Sprintf("SET LOCALE")
}
func (e LocaleInstr) Copy() Instr {
	return &LocaleInstr{}
}

//
// Second stage instructions
//

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
	Type     frontend.Type
}

type SubInstr struct {
	Dst      *RegisterExpr
	Op1      *RegisterExpr
	Op2      Expr
	Op2Shift Shift
	Type     frontend.Type
}

type MulInstr struct {
	Dst  *RegisterExpr
	Op1  *RegisterExpr
	Op2  *RegisterExpr
	Type frontend.Type
}

type DivInstr struct {
	Dst  *RegisterExpr
	Op1  *RegisterExpr
	Op2  *RegisterExpr
	Type frontend.Type
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

// Unary operations
type NotInstr struct {
	Dst Expr // LValueExpr
	Src Expr
}

type NegInstr struct {
	Expr Expr
	Type frontend.Type
}

type CmpInstr struct {
	Left     Expr
	Right    Expr
	Dst      Expr
	Operator string
}

// Function call
type CallInstr struct {
	Label *LocationExpr
}

// Heap allocation
type HeapAllocInstr struct {
	Dst  *RegisterExpr
	Size int
}

// Push/pop register
type PushInstr struct {
	Op *RegisterExpr
}

type PopInstr struct {
	Op *RegisterExpr
}

// Runtime errors
type CheckNullDereferenceInstr struct {
	Ptr Expr
}

func (NoOpInstr) instr()       {}
func (NoOpInstr) Repr() string { return "NOOP" }
func (NoOpInstr) Copy() Instr  { return &NoOpInstr{} }

func (LabelInstr) instr() {}
func (i LabelInstr) Repr() string {
	return fmt.Sprintf("LABEL %s", i.Label)
}
func (i LabelInstr) Copy() Instr { return &LabelInstr{i.Label} }

func (EvalInstr) instr() {}
func (i EvalInstr) Repr() string {
	return fmt.Sprintf("EVAL %v", i.Expr.Repr())
}
func (i EvalInstr) Copy() Instr { return &EvalInstr{i.Expr.Copy()} }

func (ReadInstr) instr() {}
func (i ReadInstr) Repr() string {
	return fmt.Sprintf("READ %v %s", i.Type.Repr(), i.Dst.Repr())
}
func (i ReadInstr) Copy() Instr { return &ReadInstr{i.Dst.Copy(), i.Type} }

func (FreeInstr) instr() {}
func (i FreeInstr) Repr() string {
	return fmt.Sprintf("FREE %s", i.Object.Repr())
}
func (i FreeInstr) Copy() Instr { return &FreeInstr{i.Object.Copy()} }

func (ReturnInstr) instr() {}
func (i ReturnInstr) Repr() string {
	return fmt.Sprintf("RETURN %s", i.Expr.Repr())
}
func (i ReturnInstr) Copy() Instr { return &ReturnInstr{i.Expr.Copy()} }

func (ExitInstr) instr() {}
func (i ExitInstr) Repr() string {
	return fmt.Sprintf("EXIT %s", i.Expr.Repr())
}
func (i ExitInstr) Copy() Instr { return &ExitInstr{i.Expr.Copy()} }

func (PrintInstr) instr() {}
func (i PrintInstr) Repr() string {
	return fmt.Sprintf("PRINT %v %s", i.Type.Repr(), i.Expr.Repr())
}
func (i PrintInstr) Copy() Instr { return &PrintInstr{i.Expr.Copy(), i.Type} }

func (MoveInstr) instr() {}
func (i MoveInstr) Repr() string {
	return fmt.Sprintf("MOVE (%s) (%s)", i.Dst.Repr(), i.Src.Repr())
}
func (i MoveInstr) Copy() Instr { return &MoveInstr{i.Dst.Copy(), i.Src.Copy()} }

func (JmpInstr) instr() {}
func (i JmpInstr) Repr() string {
	return fmt.Sprintf("JMP (%s)", i.Dst.Instr.(*LabelInstr).Repr())
}
func (i JmpInstr) Copy() Instr { return &JmpInstr{i.Dst.Copy()} }

func (JmpCondInstr) instr() {}
func (i JmpCondInstr) Repr() string {
	return fmt.Sprintf("J%s (%s)", i.Cond.Repr(), i.Dst.Instr.(*LabelInstr).Repr())
}
func (i JmpCondInstr) Copy() Instr { return &JmpCondInstr{i.Dst.Copy(), i.Cond.Copy()} }

func (*AddInstr) instr() {}
func (i *AddInstr) Repr() string {
	// Is this a floating point instruction?
	prefix := ""
	if i.Type.Equals(frontend.BasicType{frontend.FLOAT}) {
		prefix = "F"
	}

	if i.Op2Shift != nil {
		return fmt.Sprintf("%vADD %v %v %v %v", prefix, i.Dst.Repr(), i.Op1.Repr(), i.Op2.Repr(), i.Op2Shift.Repr())
	} else {
		return fmt.Sprintf("%vADD %v %v %v", prefix, i.Dst.Repr(), i.Op1.Repr(), i.Op2.Repr())
	}
}
func (i *AddInstr) Copy() Instr {
	return &AddInstr{i.Dst.Copy().(*RegisterExpr), i.Op1.Copy().(*RegisterExpr), i.Op2.Copy().(*RegisterExpr), i.Op2Shift, i.Type}
}

func (*SubInstr) instr() {}
func (i *SubInstr) Repr() string {
	// Is this a floating point instruction?
	prefix := ""
	if i.Type.Equals(frontend.BasicType{frontend.FLOAT}) {
		prefix = "F"
	}

	if i.Op2Shift != nil {
		return fmt.Sprintf("%vSUB %v %v %v %v", prefix, i.Dst.Repr(), i.Op1.Repr(), i.Op2.Repr(), i.Op2Shift.Repr())
	} else {
		return fmt.Sprintf("%vSUB %v %v %v", prefix, i.Dst.Repr(), i.Op1.Repr(), i.Op2.Repr())
	}
}
func (i *SubInstr) Copy() Instr {
	return &SubInstr{i.Dst.Copy().(*RegisterExpr), i.Op1.Copy().(*RegisterExpr), i.Op2.Copy().(*RegisterExpr), i.Op2Shift, i.Type}
}

func (*MulInstr) instr() {}
func (i *MulInstr) Repr() string {
	// Is this a floating point instruction?
	prefix := ""
	if i.Type.Equals(frontend.BasicType{frontend.FLOAT}) {
		prefix = "F"
	}

	return fmt.Sprintf("%vMUL %v %v %v", prefix, i.Dst.Repr(), i.Op1.Repr(), i.Op2.Repr())
}
func (i *MulInstr) Copy() Instr {
	return &MulInstr{i.Dst.Copy().(*RegisterExpr), i.Op1.Copy().(*RegisterExpr), i.Op2.Copy().(*RegisterExpr), i.Type}
}

func (*DivInstr) instr() {}
func (i *DivInstr) Repr() string {
	// Is this a floating point instruction?
	prefix := ""
	if i.Type.Equals(frontend.BasicType{frontend.FLOAT}) {
		prefix = "F"
	}

	return fmt.Sprintf("%vDIV %v %v %v", prefix, i.Dst.Repr(), i.Op1.Repr(), i.Op2.Repr())
}
func (i *DivInstr) Copy() Instr {
	return &DivInstr{i.Dst.Copy().(*RegisterExpr), i.Op1.Copy().(*RegisterExpr), i.Op2.Copy().(*RegisterExpr), i.Type}
}

func (*AndInstr) instr() {}
func (i *AndInstr) Repr() string {
	return fmt.Sprintf("AND %v %v %v", i.Dst.Repr(), i.Op1.Repr(), i.Op2.Repr())
}
func (i *AndInstr) Copy() Instr {
	return &AndInstr{i.Dst.Copy().(*RegisterExpr), i.Op1.Copy().(*RegisterExpr), i.Op2.Copy().(*RegisterExpr)}
}

func (*OrInstr) instr() {}
func (i *OrInstr) Repr() string {
	return fmt.Sprintf("OR %v %v %v", i.Dst.Repr(), i.Op1.Repr(), i.Op2.Repr())
}
func (i *OrInstr) Copy() Instr {
	return &OrInstr{i.Dst.Copy().(*RegisterExpr), i.Op1.Copy().(*RegisterExpr), i.Op2.Copy().(*RegisterExpr)}
}

func (NotInstr) instr() {}
func (i NotInstr) Repr() string {
	return fmt.Sprintf("NOT (%s) (%s)", i.Dst.Repr(), i.Src.Repr())
}
func (i NotInstr) Copy() Instr {
	return &NotInstr{i.Dst.Copy(), i.Src.Copy()}
}

func (NegInstr) instr() {}
func (i NegInstr) Repr() string {
	// Is this a floating point instruction?
	prefix := ""
	if i.Type.Equals(frontend.BasicType{frontend.FLOAT}) {
		prefix = "F"
	}

	return fmt.Sprintf("%vNEG %v", prefix, i.Expr.Repr())
}
func (i NegInstr) Copy() Instr {
	return &NegInstr{i.Expr.Copy(), i.Type}
}

func (CmpInstr) instr() {}
func (i CmpInstr) Repr() string {
	return fmt.Sprintf("CMP %v (%v) (%v) (%v)", i.Operator, i.Dst.Repr(), i.Left.Repr(), i.Right.Repr())
}
func (i CmpInstr) Copy() Instr {
	return &CmpInstr{i.Left.Copy(), i.Right.Copy(), i.Dst.Copy(), i.Operator}
}

func (*CallInstr) instr() {}
func (i *CallInstr) Repr() string {
	return fmt.Sprintf("CALL %v", i.Label.Label)
}
func (i *CallInstr) Copy() Instr {
	return &CallInstr{i.Label.Copy().(*LocationExpr)}
}

func (*HeapAllocInstr) instr() {}
func (i *HeapAllocInstr) Repr() string {
	return fmt.Sprintf("ALLOC %v SIZE %v", i.Dst.Repr(), i.Size)
}
func (i *HeapAllocInstr) Copy() Instr {
	return &HeapAllocInstr{i.Dst.Copy().(*RegisterExpr), i.Size}
}

func (*PushInstr) instr() {}
func (i *PushInstr) Repr() string {
	return fmt.Sprintf("PUSH %v", i.Op)
}
func (i *PushInstr) Copy() Instr {
	return &PushInstr{i.Op.Copy().(*RegisterExpr)}
}

func (*PopInstr) instr() {}
func (i *PopInstr) Repr() string {
	return fmt.Sprintf("POP %v", i.Op)
}
func (i *PopInstr) Copy() Instr {
	return &PopInstr{i.Op.Copy().(*RegisterExpr)}
}

func (CheckNullDereferenceInstr) instr() {}
func (i CheckNullDereferenceInstr) Repr() string {
	return fmt.Sprintf("CHECK NULL %v", i.Ptr.Repr())
}
func (i CheckNullDereferenceInstr) Copy() Instr {
	return &CheckNullDereferenceInstr{i.Ptr.Copy()}
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

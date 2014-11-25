package main

import (
	"fmt"
	"strconv"
)

type IFExpr interface {
	ifExpr()
	Repr() string
}

type ConstExpr struct {
	Value int
}

type NameExpr struct {
	Label string
}

type TempExpr struct {
	Id int
}

type BinOpExpr struct {
	Left  IFExpr
	Right IFExpr
}

func (ConstExpr) ifExpr()        {}
func (e ConstExpr) Repr() string { return fmt.Sprintf("CONST %d", e.Value) }

func (NameExpr) ifExpr()        {}
func (e NameExpr) Repr() string { return "NAME " + e.Label }

func (TempExpr) ifExpr()        {}
func (e TempExpr) Repr() string { return fmt.Sprintf("t%d", e.Id) }

func (BinOpExpr) ifExpr()        {}
func (e BinOpExpr) Repr() string { return fmt.Sprintf("BINOP %s %s", e.Left.Repr(), e.Right.Repr()) }

type InstrList struct {
	Head Instr
	Tail *InstrList
}

type Instr interface {
	instr()
	Repr() string
}

type NoOpInstr struct {
}

type LabelInstr struct {
	Label string
}

type SysCallInstr struct {
	SysCallID int
}

type MoveInstr struct {
	Src IFExpr
	Dst IFExpr
}

func (NoOpInstr) instr()       {}
func (NoOpInstr) Repr() string { return "NOOP" }

func (LabelInstr) instr() {}
func (i LabelInstr) Repr() string {
	return fmt.Sprintf("LABEL %s", i.Label)
}

func (SysCallInstr) instr() {}
func (i SysCallInstr) Repr() string {
	return fmt.Sprintf("SYSCALL %d", i.SysCallID)
}

func (MoveInstr) instr() {}
func (i MoveInstr) Repr() string {
	return fmt.Sprintf("MOVE %s %s", i.Src.Repr(), i.Dst.Repr())
}

type IFContext struct {
	labels       map[string]Instr
	instructions []Instr
	nextTemp     int
}

func (cxt *IFContext) addInstr(i Instr) {
	cxt.instructions = append(cxt.instructions, i)
}

func (cxt *IFContext) newTemp() *TempExpr {
	cxt.nextTemp++
	return &TempExpr{cxt.nextTemp}
}

func (cxt *IFContext) generateExpr(expr Expr) IFExpr {
	switch expr := expr.(type) {
	case *BasicLit:
		value, _ := strconv.Atoi(expr.Value)
		return &ConstExpr{value}

	case *BinaryExpr:
		return &BinOpExpr{cxt.generateExpr(expr.Left), cxt.generateExpr(expr.Right)}

	default:
		panic(fmt.Sprintf("Unhandled expression %T", expr))
	}
}

func (cxt *IFContext) generate(node Stmt) Instr {
	switch node := node.(type) {
	case *ProgStmt:
		cxt.addInstr(&LabelInstr{"main"})
		for _, n := range node.Body {
			cxt.generate(n)
		}

	case *ExitStmt:
		cxt.addInstr(&SysCallInstr{0})

	case *SkipStmt:
		cxt.addInstr(&NoOpInstr{})

	case *DeclStmt:
		cxt.addInstr(&MoveInstr{cxt.generateExpr(node.Right), &NameExpr{node.Ident.Name}})

	default:
		panic(fmt.Sprintf("Unhandled statement %T", node))
	}

	return nil
}

func GenerateIntermediateForm(program *ProgStmt) *IFContext {
	cxt := new(IFContext)

	//cxt.generate(program)

	for _, i := range cxt.instructions {
		fmt.Println(i.Repr())
	}

	return cxt
}

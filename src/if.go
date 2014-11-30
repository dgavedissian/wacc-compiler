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

type InstrNode struct {
	Instr Instr
	Next  *InstrNode
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
	return fmt.Sprintf("MOVE (%s) (%s)", i.Src.Repr(), i.Dst.Repr())
}

type IFContext struct {
	labels   map[string]Instr
	first    *InstrNode
	current  *InstrNode
	nextTemp int
}

func (ctx *IFContext) addInstr(i Instr) {
	if ctx.first == nil {
		ctx.first = &InstrNode{i, nil}
		ctx.current = ctx.first
	} else {
		ctx.current.Next = &InstrNode{i, nil}
		ctx.current = ctx.current.Next
	}
}

func (ctx *IFContext) newTemp() *TempExpr {
	ctx.nextTemp++
	return &TempExpr{ctx.nextTemp}
}

func (ctx *IFContext) generateExpr(expr Expr) IFExpr {
	switch expr := expr.(type) {
	case *BasicLit:
		value, _ := strconv.Atoi(expr.Value)
		return &ConstExpr{value}

	case *BinaryExpr:
		return &BinOpExpr{ctx.generateExpr(expr.Left), ctx.generateExpr(expr.Right)}

	default:
		panic(fmt.Sprintf("Unhandled expression %T", expr))
	}
}

func (ctx *IFContext) generate(node Stmt) {
	switch node := node.(type) {
	case *ProgStmt:
		ctx.addInstr(&LabelInstr{"main"})
		for _, n := range node.Body {
			ctx.generate(n)
		}

	case *SkipStmt:
		ctx.addInstr(&NoOpInstr{})

	case *DeclStmt:
		ctx.addInstr(&MoveInstr{ctx.generateExpr(node.Right), &NameExpr{node.Ident.Name}})

	case *AssignStmt:
		ctx.addInstr(&MoveInstr{ctx.generateExpr(node.Right), &NameExpr{node.Left.(*IdentExpr).Name}})

	// Read

	// Free

	case *ExitStmt:
		ctx.addInstr(&SysCallInstr{0})

		// Return

		// Print

		// If

	//case *WhileStmt:
	//beginWhile := &InstrNode{JumpZeroInstr{}, nil}

	// Scope

	default:
		panic(fmt.Sprintf("Unhandled statement %T", node))
	}
}

func printGraph(node *InstrNode) {
	for node != nil {
		fmt.Printf("| %s\n", node.Instr.Repr())
		node = node.Next
	}
}

func GenerateIntermediateForm(program *ProgStmt) *IFContext {
	ctx := new(IFContext)
	ctx.generate(program)
	printGraph(ctx.first)
	return ctx
}

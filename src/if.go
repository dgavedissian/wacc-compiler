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

type TestInstr struct {
	Cond IFExpr
}

type JmpInstr struct {
	Dest *InstrNode
}

type JmpZeroInstr struct {
	Dest *InstrNode
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

func (TestInstr) instr() {}
func (i TestInstr) Repr() string {
	return fmt.Sprintf("TEST (%s)", i.Cond.Repr())
}

func (JmpInstr) instr() {}
func (i JmpInstr) Repr() string {
	return fmt.Sprintf("JMP (%s)", i.Dest.Instr.(*LabelInstr).Repr())
}

func (JmpZeroInstr) instr() {}
func (i JmpZeroInstr) Repr() string {
	return fmt.Sprintf("JZ (%s)", i.Dest.Instr.(*LabelInstr).Repr())
}

type IFContext struct {
	labels   map[string]Instr
	first    *InstrNode
	current  *InstrNode
	nextTemp int
}

func (ctx *IFContext) makeNode(i Instr) *InstrNode {
	return &InstrNode{i, nil}
}

func (ctx *IFContext) appendNode(n *InstrNode) {
	if ctx.first == nil {
		ctx.first = n
		ctx.current = ctx.first
	} else {
		ctx.current.Next = n
		ctx.current = ctx.current.Next
	}
}

func (ctx *IFContext) addInstr(i Instr) *InstrNode {
	newNode := ctx.makeNode(i)
	ctx.appendNode(newNode)
	return newNode
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

	case *IfStmt:
		startElse := ctx.makeNode(&LabelInstr{"else_begin"})
		endIfElse := ctx.makeNode(&LabelInstr{"ifelse_end"})

		ctx.addInstr(&TestInstr{ctx.generateExpr(node.Cond)})
		ctx.addInstr(&JmpZeroInstr{startElse})

		// Build main branch
		for _, n := range node.Body {
			ctx.generate(n)
		}

		ctx.addInstr(&JmpInstr{endIfElse})
		ctx.appendNode(startElse)

		// Build else branch
		for _, n := range node.Else {
			ctx.generate(n)
		}

		// Build end
		ctx.appendNode(endIfElse)

	case *WhileStmt:
		beginWhile := ctx.makeNode(&LabelInstr{"while_begin"})
		endWhile := ctx.makeNode(&LabelInstr{"while_end"})

		// Build condition
		ctx.appendNode(beginWhile)
		ctx.addInstr(&TestInstr{ctx.generateExpr(node.Cond)})
		ctx.addInstr(&JmpZeroInstr{endWhile})

		// Build body
		for _, n := range node.Body {
			ctx.generate(n)
		}

		// Build end
		ctx.addInstr(&JmpInstr{beginWhile})
		ctx.appendNode(endWhile)

	// Scope

	default:
		panic(fmt.Sprintf("Unhandled statement %T", node))
	}
}

func GenerateIF(program *ProgStmt) *IFContext {
	ctx := new(IFContext)
	ctx.generate(program)
	return ctx
}

func DrawIFGraph(iform *IFContext) {
	// Transform into a list
	node := iform.first
	var list []Instr
	for node != nil {
		list = append(list, node.Instr)
		node = node.Next
	}
	instrCount := len(list)

	// To whoever reads this code... I'm sorry

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
					if jmp.Dest.Instr == instr {
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

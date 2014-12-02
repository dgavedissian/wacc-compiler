package backend

import (
	"fmt"
	"strconv"
	"strings"

	"../frontend"
)

type IFExpr interface {
	ifExpr()
	Repr() string
}

type IntConstExpr struct {
	Value int
}

type CharConstExpr struct {
	Value string
}

type ArrayExpr struct {
	Type  frontend.BasicType
	Elems []IFExpr
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

func (IntConstExpr) ifExpr()        {}
func (e IntConstExpr) Repr() string { return fmt.Sprintf("CONST %v", e.Value) }

func (CharConstExpr) ifExpr()        {}
func (e CharConstExpr) Repr() string { return fmt.Sprintf("CONST %v", strconv.QuoteToASCII(e.Value)) }

func (NameExpr) ifExpr()        {}
func (e NameExpr) Repr() string { return "NAME " + e.Label }

func (TempExpr) ifExpr()        {}
func (e TempExpr) Repr() string { return fmt.Sprintf("t%d", e.Id) }

func (ArrayExpr) ifExpr() {}
func (e ArrayExpr) Repr() string {
	rs := make([]string, len(e.Elems))
	for i, v := range e.Elems {
		rs[i] = v.Repr()
	}
	return "ARRAYCONST [" + strings.Join(rs, ", ") + "]"
}

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

type ReadInstr struct {
	Dst IFExpr // LValueExpr
}

type FreeInstr struct {
	Object IFExpr // LValueExpr
}

type ExitInstr struct {
	Expr IFExpr
}

type PrintInstr struct {
	Expr IFExpr
}

type MoveInstr struct {
	Src IFExpr
	Dst IFExpr // LValueExpr
}

type TestInstr struct {
	Cond IFExpr
}

type JmpInstr struct {
	Dst *InstrNode
}

type JmpZeroInstr struct {
	Dst *InstrNode
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
	return fmt.Sprintf("MOVE (%s) (%s)", i.Src.Repr(), i.Dst.Repr())
}

func (TestInstr) instr() {}
func (i TestInstr) Repr() string {
	return fmt.Sprintf("TEST (%s)", i.Cond.Repr())
}

func (JmpInstr) instr() {}
func (i JmpInstr) Repr() string {
	return fmt.Sprintf("JMP (%s)", i.Dst.Instr.(*LabelInstr).Repr())
}

func (JmpZeroInstr) instr() {}
func (i JmpZeroInstr) Repr() string {
	return fmt.Sprintf("JZ (%s)", i.Dst.Instr.(*LabelInstr).Repr())
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

func (ctx *IFContext) generateExpr(expr frontend.Expr) IFExpr {
	switch expr := expr.(type) {
	case *frontend.BasicLit:
		if expr.Type.Equals(frontend.BasicType{frontend.INT}) {
			value, _ := strconv.Atoi(expr.Value)
			return &IntConstExpr{value}
		}

		if expr.Type.Equals(frontend.BasicType{frontend.BOOL}) {
			if expr.Value == "true" {
				return &IntConstExpr{1}
			}
			return &IntConstExpr{0}
		}

		if expr.Type.Equals(frontend.BasicType{frontend.CHAR}) {
			return &CharConstExpr{expr.Value}
		}

		// Null
		if expr.Type.Equals(frontend.BasicType{frontend.PAIR}) {
			return &IntConstExpr{0}
		}

		panic(fmt.Sprintf("Unhandled BasicLit %s", expr.Type.Repr()))

	case *frontend.IdentExpr:
		return &NameExpr{expr.Name}

	case *frontend.BinaryExpr:
		return &BinOpExpr{ctx.generateExpr(expr.Left), ctx.generateExpr(expr.Right)}

	case *frontend.ArrayLit:
		a := ArrayExpr{}
		a.Elems = make([]IFExpr, len(expr.Values))
		for i, e := range expr.Values {
			a.Elems[i] = ctx.generateExpr(e)
		}
		return a

	default:
		panic(fmt.Sprintf("Unhandled expression %T", expr))
	}
}

func (ctx *IFContext) generate(node frontend.Stmt) {
	switch node := node.(type) {
	case *frontend.ProgStmt:
		ctx.addInstr(&LabelInstr{"main"})
		for _, n := range node.Body {
			ctx.generate(n)
		}

	case *frontend.SkipStmt:
		ctx.addInstr(&NoOpInstr{})

	case *frontend.DeclStmt:
		ctx.addInstr(&MoveInstr{ctx.generateExpr(node.Right), &NameExpr{node.Ident.Name}})

	case *frontend.AssignStmt:
		ctx.addInstr(&MoveInstr{ctx.generateExpr(node.Right), &NameExpr{node.Left.(*frontend.IdentExpr).Name}})

	case *frontend.ReadStmt:
		ctx.addInstr(&ReadInstr{ctx.generateExpr(node.Dst)})

	case *frontend.FreeStmt:
		ctx.addInstr(&FreeInstr{ctx.generateExpr(node.Object)})

	case *frontend.ExitStmt:
		ctx.addInstr(&ExitInstr{ctx.generateExpr(node.Result)})

	case *frontend.PrintStmt:
		ctx.addInstr(&PrintInstr{ctx.generateExpr(node.Right)})
		if node.NewLine {
			ctx.addInstr(&PrintInstr{CharConstExpr{"\n"}})
		}

	// Return

	case *frontend.IfStmt:
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

	case *frontend.WhileStmt:
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

func GenerateIF(program *frontend.ProgStmt) *IFContext {
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

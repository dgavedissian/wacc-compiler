package backend

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"

	"../frontend"
)

type Expr interface {
	expr()
	Repr() string

	replaceVar(*RegisterAllocatorContext) Expr
	generateCode(*GeneratorContext) int
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

type InstrNode struct {
	Instr Instr
	Next  *InstrNode
}

type Instr interface {
	instr()
	Repr() string

	replaceVar(*RegisterAllocatorContext)
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

func (ctx *IFContext) generateExpr(expr frontend.Expr) Expr {
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
			value, size := utf8.DecodeRuneInString(expr.Value)
			return &CharConstExpr{value, size}
		}

		// Null
		if expr.Type.Equals(frontend.BasicType{frontend.PAIR}) {
			return &IntConstExpr{0}
		}

		panic(fmt.Sprintf("Unhandled BasicLit %s", expr.Type.Repr()))

	case *frontend.IdentExpr:
		return &VarExpr{expr.Name}

	case *frontend.BinaryExpr:
		return &BinOpExpr{ctx.generateExpr(expr.Left), ctx.generateExpr(expr.Right)}

	case *frontend.ArrayLit:
		a := &ArrayExpr{}
		a.Elems = make([]Expr, len(expr.Values))
		for i, e := range expr.Values {
			a.Elems[i] = ctx.generateExpr(e)
		}
		return a

	case *frontend.UnaryExpr:
		op := expr.Operator
		switch op {
		case Not:
			/*
				Fold binaries in an optimisation step.
			*/
			return &NotExpr{ctx.generateExpr(expr.Operand)}
		case Ord:
			/* Fold chars in an optimisation step
			if x, ok := expr.Operand.(*BasicLit); ok {
				if x.Type.Equals(frontend.BasicType{frontend.CHAR}) {
					r, size := utf8.DecodeRuneInString(x.Value)
					return &IntConstExpr{r}
				}
			}*/
			return &OrdExpr{ctx.generateExpr(expr.Operand)}
		case Chr:
			/* Fold ints in an optimisation step */
			return &ChrExpr{ctx.generateExpr(expr.Operand)}
		case Neg:
			/* Fold negating ints in optimisation step
			if x, ok := expr.Operand.(*BasicLit); ok {
				if x.Type.Equals(frontend.BasicType{frontend.INT}) {
					n, _ := strconv.Atoi(x.Value)
					return &IntConstExpr{-n}
				}
			}*/
			return &NegExpr{ctx.generateExpr(expr.Operand)}
		case Len:
			/* Constant fold on strings and array literals */
			return &LenExpr{ctx.generateExpr(expr.Operand)}
		default:
			panic(fmt.Sprintf("Unhandled unary operator %v", expr.Operator))
		}

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
		ctx.addInstr(
			&MoveInstr{
				Dst: &VarExpr{node.Ident.Name},
				Src: ctx.generateExpr(node.Right)})

	case *frontend.AssignStmt:
		var dst Expr
		switch leftExpr := node.Left.(type) {
		case *frontend.IdentExpr:
			dst = &VarExpr{leftExpr.Name}
		case *frontend.PairElemExpr:
			panic("TODO: Pair locations")
		case *frontend.ArrayElemExpr:
			panic("TODO: Array locations")
		default:
			panic(fmt.Sprintf("Missing lhs %T", leftExpr))
		}
		ctx.addInstr(
			&MoveInstr{
				Dst: dst,
				Src: ctx.generateExpr(node.Right)})

	case *frontend.ReadStmt:
		ctx.addInstr(&ReadInstr{ctx.generateExpr(node.Dst)})

	case *frontend.FreeStmt:
		ctx.addInstr(&FreeInstr{ctx.generateExpr(node.Object)})

	case *frontend.ExitStmt:
		ctx.addInstr(&ExitInstr{ctx.generateExpr(node.Result)})

	case *frontend.PrintStmt:
		ctx.addInstr(&PrintInstr{ctx.generateExpr(node.Right)})
		if node.NewLine {
			ctx.addInstr(&PrintInstr{&CharConstExpr{'\n', 1}})
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

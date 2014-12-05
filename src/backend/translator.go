package backend

import (
	"fmt"
	"strconv"
	"unicode/utf8"

	"../frontend"
)

func TranslateToIF(program *frontend.ProgStmt) *IFContext {
	ctx := new(IFContext)
	ctx.functions = make(map[string]*InstrNode)
	ctx.translate(program)
	return ctx
}

func (ctx *IFContext) makeNode(i Instr) *InstrNode {
	return &InstrNode{i, nil}
}

func (ctx *IFContext) beginFunction(name string) {
	ctx.functions[name] = ctx.makeNode(&LabelInstr{name})
	ctx.current = ctx.functions[name]
}

func (ctx *IFContext) beginMain() {
	ctx.main = ctx.makeNode(&LabelInstr{"main"})
	ctx.current = ctx.main
}

func (ctx *IFContext) appendNode(n *InstrNode) {
	ctx.current.Next = n
	ctx.current = ctx.current.Next
}

func (ctx *IFContext) addInstr(i Instr) *InstrNode {
	newNode := ctx.makeNode(i)
	ctx.appendNode(newNode)
	return newNode
}

func (ctx *IFContext) translateExpr(expr frontend.Expr) Expr {
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
		return &BinOpExpr{
			Operator: expr.Operator,
			Left:     ctx.translateExpr(expr.Left),
			Right:    ctx.translateExpr(expr.Right)}

	case *frontend.ArrayLit:
		a := &ArrayExpr{}
		a.Elems = make([]Expr, len(expr.Values))
		for i, e := range expr.Values {
			a.Elems[i] = ctx.translateExpr(e)
		}
		return a

	case *frontend.UnaryExpr:
		op := expr.Operator
		switch op {
		case Not:
			/*
				Fold binaries in an optimisation step.
			*/
			return &NotExpr{ctx.translateExpr(expr.Operand)}
		case Ord:
			/* Fold chars in an optimisation step
			if x, ok := expr.Operand.(*BasicLit); ok {
				if x.Type.Equals(frontend.BasicType{frontend.CHAR}) {
					r, size := utf8.DecodeRuneInString(x.Value)
					return &IntConstExpr{r}
				}
			}*/
			return &OrdExpr{ctx.translateExpr(expr.Operand)}
		case Chr:
			/* Fold ints in an optimisation step */
			return &ChrExpr{ctx.translateExpr(expr.Operand)}
		case Neg:
			/* Fold negating ints in optimisation step
			if x, ok := expr.Operand.(*BasicLit); ok {
				if x.Type.Equals(frontend.BasicType{frontend.INT}) {
					n, _ := strconv.Atoi(x.Value)
					return &IntConstExpr{-n}
				}
			}*/
			return &NegExpr{ctx.translateExpr(expr.Operand)}
		case Len:
			/* Constant fold on strings and array literals */
			return &LenExpr{ctx.translateExpr(expr.Operand)}
		default:
			panic(fmt.Sprintf("Unhandled unary operator %v", expr.Operator))
		}

	case *frontend.NewPairCmd:
		return &NewPairExpr{
			Left:  ctx.translateExpr(expr.Left),
			Right: ctx.translateExpr(expr.Right)}

	case *frontend.CallCmd:
		translatedArgs := make([]Expr, len(expr.Args))
		for i, arg := range expr.Args {
			translatedArgs[i] = ctx.translateExpr(arg)
		}
		return &CallExpr{Label: &LocationExpr{expr.Ident.Name}, Args: translatedArgs}

	default:
		panic(fmt.Sprintf("Unhandled expression %T", expr))
	}
}

func (ctx *IFContext) translate(node frontend.Stmt) {
	switch node := node.(type) {
	case *frontend.ProgStmt:
		// Functions
		for _, f := range node.Funcs {
			ctx.beginFunction(f.Ident.Name)
			for _, n := range f.Body {
				ctx.translate(n)
			}
		}

		// Main
		ctx.beginMain()
		for _, n := range node.Body {
			ctx.translate(n)
		}

	case *frontend.SkipStmt:
		ctx.addInstr(&NoOpInstr{})

	case *frontend.DeclStmt:
		ctx.addInstr(
			&MoveInstr{
				Dst: &VarExpr{node.Ident.Name},
				Src: ctx.translateExpr(node.Right)})

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
				Src: ctx.translateExpr(node.Right)})

	case *frontend.ReadStmt:
		ctx.addInstr(&ReadInstr{ctx.translateExpr(node.Dst)})

	case *frontend.FreeStmt:
		ctx.addInstr(&FreeInstr{ctx.translateExpr(node.Object)})

	case *frontend.ReturnStmt:
		ctx.addInstr(&ReturnInstr{Expr: ctx.translateExpr(node.Result)})

	case *frontend.ExitStmt:
		ctx.addInstr(&ExitInstr{ctx.translateExpr(node.Result)})

	case *frontend.PrintStmt:
		ctx.addInstr(&PrintInstr{ctx.translateExpr(node.Right)})
		if node.NewLine {
			ctx.addInstr(&PrintInstr{&CharConstExpr{'\n', 1}})
		}

	// Return

	case *frontend.IfStmt:
		startElse := ctx.makeNode(&LabelInstr{"else_begin"})
		endIfElse := ctx.makeNode(&LabelInstr{"ifelse_end"})

		ctx.addInstr(&TestInstr{ctx.translateExpr(node.Cond)})
		ctx.addInstr(&JmpEqualInstr{startElse})

		// Build main branch
		for _, n := range node.Body {
			ctx.translate(n)
		}

		ctx.addInstr(&JmpInstr{endIfElse})
		ctx.appendNode(startElse)

		// Build else branch
		for _, n := range node.Else {
			ctx.translate(n)
		}

		// Build end
		ctx.appendNode(endIfElse)

	case *frontend.WhileStmt:
		beginWhile := ctx.makeNode(&LabelInstr{"while_begin"})
		endWhile := ctx.makeNode(&LabelInstr{"while_end"})

		// Build condition
		ctx.appendNode(beginWhile)
		ctx.addInstr(&TestInstr{ctx.translateExpr(node.Cond)})
		ctx.addInstr(&JmpEqualInstr{endWhile})

		// Build body
		for _, n := range node.Body {
			ctx.translate(n)
		}

		// Build end
		ctx.addInstr(&JmpInstr{beginWhile})
		ctx.appendNode(endWhile)

	// Scope

	default:
		panic(fmt.Sprintf("Unhandled statement %T", node))
	}
}

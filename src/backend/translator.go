package backend

import (
	"fmt"
	"strconv"
	"unicode/utf8"

	"../frontend"
)

type IFContext struct {
	// Variable scoping
	scope          []map[string]frontend.Type
	scopePushInstr []*PushScopeInstr
	depth          int

	// Labels
	labels map[string]Instr

	// Functions
	main      *InstrNode
	functions map[string]*InstrNode
	current   *InstrNode

	// Data Store
	dataStore      map[string]*StringConstExpr
	currentCounter int
}

func TranslateToIF(program *frontend.Program) *IFContext {
	ctx := new(IFContext)
	ctx.functions = make(map[string]*InstrNode)
	ctx.translate(program)
	return ctx
}

func (ctx *IFContext) makeNode(i Instr) *InstrNode {
	return &InstrNode{i, 0, nil, nil}
}

func (ctx *IFContext) beginFunction(name string) {
	ctx.functions[name] = ctx.makeNode(&LabelInstr{name})
	ctx.current = ctx.functions[name]
}

func (ctx *IFContext) beginMain() {
	ctx.main = ctx.makeNode(&LabelInstr{"main"})
	ctx.current = ctx.main
	ctx.addInstr(&LocaleInstr{})
}

func (ctx *IFContext) appendNode(n *InstrNode) {
	n.Prev = ctx.current
	ctx.current.Next = n
	ctx.current = ctx.current.Next
}

func (ctx *IFContext) removeNode(n *InstrNode) {
	n.Prev.Next = n.Next
	n.Next.Prev = n.Prev
}

func (ctx *IFContext) addInstr(i Instr) *InstrNode {
	newNode := ctx.makeNode(i)
	ctx.appendNode(newNode)
	return newNode
}

func (ctx *IFContext) addType(v string, t frontend.Type) {
	ctx.scope[ctx.depth-1][v] = t
}

func (ctx *IFContext) pushScope() {
	pushScopeInstr := &PushScopeInstr{StackSize: 0}
	ctx.scope = append(ctx.scope, make(map[string]frontend.Type))
	ctx.scopePushInstr = append(ctx.scopePushInstr, pushScopeInstr)
	ctx.depth++
	ctx.addInstr(pushScopeInstr)
}

func (ctx *IFContext) popScope() {
	// Determine stack size
	stackSize := len(ctx.scope[ctx.depth-1]) * regWidth

	// Ensure stack is double-word aligned (5.2.1.2)
	if (stackSize % 8) != 0 {
		stackSize += 8 - (stackSize % 8)
	}

	ctx.addInstr(&PopScopeInstr{stackSize})
	ctx.scopePushInstr[ctx.depth-1].StackSize = stackSize
	ctx.scope = ctx.scope[:ctx.depth-1]
	ctx.scopePushInstr = ctx.scopePushInstr[:ctx.depth-1]
	ctx.depth--
}

func (ctx *IFContext) getType(expr Expr) frontend.Type {
	switch expr := expr.(type) {
	case *IntConstExpr:
		return frontend.BasicType{frontend.INT}

	case *CharConstExpr:
		return frontend.BasicType{frontend.CHAR}

	case *ArrayConstExpr:
		return expr.Type

	case *StringConstExpr:
		return frontend.BasicType{frontend.STRING}

	case *BoolConstExpr:
		return frontend.BasicType{frontend.BOOL}

	case *PointerConstExpr:
		return frontend.BasicType{frontend.PAIR}

	case *UnaryExpr:
		return expr.Type

	case *BinaryExpr:
		return expr.Type

	case *VarExpr:
		// Search scopes for type
		for i := ctx.depth - 1; i >= 0; i-- {
			if t, ok := ctx.scope[i][expr.Name]; ok {
				return t
			}
		}

		// Cant find the variable, this should never happen
		panic(fmt.Sprintf("Cannot find variable %s", expr.Name))

	default:
		return nil
	}
}

func (ctx *IFContext) translateExpr(expr frontend.Expr) Expr {
	switch expr := expr.(type) {
	case *frontend.BasicLit:
		if expr.Type.Equals(frontend.BasicType{frontend.INT}) {
			value, _ := strconv.Atoi(expr.Value)
			return &IntConstExpr{value}
		}

		if expr.Type.Equals(frontend.BasicType{frontend.FLOAT}) {
			value, _ := strconv.ParseFloat(expr.Value, 32)
			return &FloatConstExpr{float32(value)}
		}

		if expr.Type.Equals(frontend.BasicType{frontend.BOOL}) {
			if expr.Value == "true" {
				return &BoolConstExpr{true}
			}
			return &BoolConstExpr{false}
		}

		if expr.Type.Equals(frontend.BasicType{frontend.CHAR}) {
			value, size := utf8.DecodeRuneInString(expr.Value)
			return &CharConstExpr{value, size}
		}

		if expr.Type.Equals(frontend.BasicType{frontend.STRING}) {
			return &StringConstExpr{expr.Value}
		}

		// Null
		if expr.Type.Equals(frontend.BasicType{frontend.PAIR}) {
			return &PointerConstExpr{0}
		}

		panic(fmt.Sprintf("Unhandled BasicLit %s", expr.Type.Repr()))

	case *frontend.IdentExpr:
		return &VarExpr{expr.Name}

	case *frontend.ArrayElemExpr:
		return &ArrayElemExpr{ctx.translateExpr(expr.Volume), ctx.translateExpr(expr.Index)}

	case *frontend.PairElemExpr:
		return &PairElemExpr{
			expr.SelectorType == frontend.FST,
			&VarExpr{expr.Operand.Name}}

	case *frontend.StructElemExpr:
		return &StructElemExpr{
			&VarExpr{expr.StructIdent.Name},
			&VarExpr{expr.ElemIdent.Name},
			expr.ElemNum * regWidth,
		}

	case *frontend.UnaryExpr:
		/* Fold negated constants */
		if expr.Operator == Neg {
			if x, ok := expr.Operand.(*frontend.BasicLit); ok {
				if x.Type.Equals(frontend.BasicType{frontend.INT}) {
					n, _ := strconv.Atoi(x.Value)
					return &IntConstExpr{-n}
				}
				if x.Type.Equals(frontend.BasicType{frontend.FLOAT}) {
					n, _ := strconv.ParseFloat(x.Value, 32)
					return &FloatConstExpr{-float32(n)}
				}
			}
		}

		return &UnaryExpr{
			Operator: expr.Operator,
			Operand:  ctx.translateExpr(expr.Operand),
			Type:     expr.Type}

	case *frontend.BinaryExpr:
		return &BinaryExpr{
			Operator: expr.Operator,
			Left:     ctx.translateExpr(expr.Left),
			Right:    ctx.translateExpr(expr.Right),
			Type:     expr.Type}

	case *frontend.ArrayLit:
		a := &ArrayConstExpr{Type: expr.Type}
		a.Elems = make([]Expr, len(expr.Values))
		for i, e := range expr.Values {
			a.Elems[i] = ctx.translateExpr(e)
		}
		return a

	case *frontend.NewStructCmd:
		translatedArgs := make([]Expr, len(expr.Args))
		for i, arg := range expr.Args {
			translatedArgs[i] = ctx.translateExpr(arg)
		}
		return &NewStructExpr{
			Label: &LocationExpr{expr.Ident.Name},
			Args:  translatedArgs}

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
	case *frontend.Program:
		// Functions
		for _, f := range node.Funcs {
			if !f.External {
				ctx.beginFunction(f.Ident.Name)
				ctx.pushScope()

				for regNum, p := range f.Params {
					if regNum < 4 {
						ctx.addType(p.Ident.Name, p.Type)
						ctx.addInstr(&DeclareInstr{&VarExpr{p.Ident.Name}, p.Type})
						ctx.addInstr(&MoveInstr{Dst: &VarExpr{p.Ident.Name}, Src: &RegisterExpr{regNum}})
					} else {
						ctx.addType(p.Ident.Name, p.Type)
						ctx.addInstr(&DeclareInstr{&VarExpr{p.Ident.Name}, p.Type})
						ctx.addInstr(&MoveInstr{
							Dst: &VarExpr{p.Ident.Name},
							Src: &StackArgumentExpr{(len(f.Params) - 5) - (regNum - 4)},
						})
					}
				}

				// Translate body
				for _, n := range f.Body {
					ctx.translate(n)
				}
				ctx.popScope()
			}
		}

		// Main
		ctx.beginMain()
		ctx.pushScope()
		for _, n := range node.Body {
			ctx.translate(n)
		}
		ctx.popScope()

	case *frontend.SkipStmt:
		ctx.addInstr(&NoOpInstr{})

	case *frontend.EvalStmt:
		ctx.addInstr(&EvalInstr{ctx.translateExpr(node.Expr)})

	case *frontend.DeclStmt:
		v := &VarExpr{node.Ident.Name}
		ctx.addType(v.Name, node.Type)
		ctx.addInstr(&DeclareInstr{v, node.Type})
		ctx.addInstr(&MoveInstr{Dst: v, Src: ctx.translateExpr(node.Right)})

	case *frontend.AssignStmt:
		ctx.addInstr(
			&MoveInstr{
				Dst: ctx.translateExpr(node.Left),
				Src: ctx.translateExpr(node.Right)})

	case *frontend.ReadStmt:
		ctx.addInstr(&ReadInstr{ctx.translateExpr(node.Dst), node.Type})

	case *frontend.FreeStmt:
		ctx.addInstr(&FreeInstr{ctx.translateExpr(node.Object)})

	case *frontend.ReturnStmt:
		ctx.addInstr(&ReturnInstr{Expr: ctx.translateExpr(node.Result)})

	case *frontend.ExitStmt:
		ctx.addInstr(&ExitInstr{ctx.translateExpr(node.Result)})

	case *frontend.PrintStmt:
		right := ctx.translateExpr(node.Right)
		ctx.addInstr(&PrintInstr{Expr: right, Type: node.Type})
		if node.NewLine {
			ctx.addInstr(&PrintInstr{
				Expr: &CharConstExpr{'\n', 1},
				Type: frontend.BasicType{frontend.CHAR}})
		}

	case *frontend.IfStmt:
		n := ctx.currentCounter
		startElse := ctx.makeNode(&LabelInstr{fmt.Sprintf("_else_begin%d", n)})
		endIfElse := ctx.makeNode(&LabelInstr{fmt.Sprintf("_ifelse_end%d", n)})
		ctx.currentCounter += 1

		trexpr := ctx.translateExpr(node.Cond)
		ctx.addInstr(&JmpCondInstr{startElse, &UnaryExpr{
			Operator: Not,
			Operand:  trexpr,
			Type:     frontend.BasicType{frontend.BOOL}}})

		// Build main branch
		ctx.pushScope()
		for _, n := range node.Body {
			ctx.translate(n)
		}
		ctx.popScope()

		ctx.addInstr(&JmpInstr{endIfElse})
		ctx.appendNode(startElse)

		// Build else branch
		ctx.pushScope()
		for _, n := range node.Else {
			ctx.translate(n)
		}
		ctx.popScope()

		// Build end
		ctx.appendNode(endIfElse)

	case *frontend.WhileStmt:
		n := ctx.currentCounter
		beginWhile := ctx.makeNode(&LabelInstr{fmt.Sprintf("_while_begin%d", n)})
		endWhile := ctx.makeNode(&LabelInstr{fmt.Sprintf("_while_end%d", n)})
		ctx.currentCounter += 1

		// Build condition
		ctx.appendNode(beginWhile)

		trexpr := ctx.translateExpr(node.Cond)
		ctx.addInstr(&JmpCondInstr{endWhile, &UnaryExpr{
			Operator: Not,
			Operand:  trexpr,
			Type:     frontend.BasicType{frontend.BOOL}}})

		// Build body
		ctx.pushScope()
		for _, n := range node.Body {
			ctx.translate(n)
		}
		ctx.popScope()

		// Build end
		ctx.addInstr(&JmpInstr{beginWhile})
		ctx.appendNode(endWhile)

	// Scope
	case *frontend.ScopeStmt:
		ctx.pushScope()
		for _, stmt := range node.Body {
			ctx.translate(stmt)
		}
		ctx.popScope()

	default:
		panic(fmt.Sprintf("Unhandled statement %T", node))
	}
}

package backend

type RegisterAllocatorContext struct {
	variableMap map[string]*RegisterExpr
	index       int
}

func (ctx *RegisterAllocatorContext) getRegister(v *VarExpr) *RegisterExpr {
	if reg, ok := ctx.variableMap[v.Name]; ok {
		return reg
	} else {
		reg := &RegisterExpr{ctx.index}
		ctx.index++
		ctx.variableMap[v.Name] = reg
		return reg
	}
}

func (e *IntConstExpr) replaceVar(*RegisterAllocatorContext) Expr  { return e }
func (e *CharConstExpr) replaceVar(*RegisterAllocatorContext) Expr { return e }
func (e *ArrayExpr) replaceVar(ctx *RegisterAllocatorContext) Expr {
	for _, elem := range e.Elems {
		elem = elem.replaceVar(ctx)
	}
	return e
}
func (e *LocationExpr) replaceVar(*RegisterAllocatorContext) Expr { return e }
func (e *VarExpr) replaceVar(ctx *RegisterAllocatorContext) Expr {
	return ctx.getRegister(e)
}
func (e *RegisterExpr) replaceVar(*RegisterAllocatorContext) Expr { return e }
func (e *BinOpExpr) replaceVar(ctx *RegisterAllocatorContext) Expr {
	e.Left = e.Left.replaceVar(ctx)
	e.Right = e.Right.replaceVar(ctx)
	return e
}

func (e *NotExpr) replaceVar(ctx *RegisterAllocatorContext) Expr {
	e.Operand = e.Operand.replaceVar(ctx)
	return e
}

func (e *OrdExpr) replaceVar(ctx *RegisterAllocatorContext) Expr {
	e.Operand = e.Operand.replaceVar(ctx)
	return e
}

func (e *ChrExpr) replaceVar(ctx *RegisterAllocatorContext) Expr {
	e.Operand = e.Operand.replaceVar(ctx)
	return e
}

func (e *NegExpr) replaceVar(ctx *RegisterAllocatorContext) Expr {
	e.Operand = e.Operand.replaceVar(ctx)
	return e
}

func (e *LenExpr) replaceVar(ctx *RegisterAllocatorContext) Expr {
	e.Operand = e.Operand.replaceVar(ctx)
	return e
}

func (e *NewPairExpr) replaceVar(ctx *RegisterAllocatorContext) Expr {
	e.Left = e.Left.replaceVar(ctx)
	e.Right = e.Right.replaceVar(ctx)
	return e
}

func (e *CallExpr) replaceVar(ctx *RegisterAllocatorContext) Expr {
	e.Ident = e.Ident.replaceVar(ctx)
	for i, arg := range e.Args {
		e.Args[i] = arg.replaceVar(ctx)
	}
	return e
}

func (i *NoOpInstr) replaceVar(*RegisterAllocatorContext)  {}
func (i *LabelInstr) replaceVar(*RegisterAllocatorContext) {}
func (i *ReadInstr) replaceVar(*RegisterAllocatorContext)  {}
func (i *FreeInstr) replaceVar(*RegisterAllocatorContext)  {}
func (i *ExitInstr) replaceVar(ctx *RegisterAllocatorContext) {
	i.Expr = i.Expr.replaceVar(ctx)
}
func (i *PrintInstr) replaceVar(ctx *RegisterAllocatorContext) {
	i.Expr = i.Expr.replaceVar(ctx)
}
func (i *MoveInstr) replaceVar(ctx *RegisterAllocatorContext) {
	i.Src = i.Src.replaceVar(ctx)
	i.Dst = i.Dst.replaceVar(ctx)
}
func (i *TestInstr) replaceVar(*RegisterAllocatorContext)    {}
func (i *JmpInstr) replaceVar(*RegisterAllocatorContext)     {}
func (i *JmpZeroInstr) replaceVar(*RegisterAllocatorContext) {}

func AllocateRegisters(ifCtx *IFContext) {
	ctx := &RegisterAllocatorContext{make(map[string]*RegisterExpr), 0}
	VisitInstructions(ifCtx, func(i Instr) {
		i.replaceVar(ctx)
	})
}

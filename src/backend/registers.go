package backend

type RegisterAllocatorContext struct {
	variableMap map[string]*RegisterExpr
	index       int
}

func (ctx *RegisterAllocatorContext) getRegister(v *VarExpr) *RegisterExpr {
	if reg, ok := ctx.variableMap[v.Name]; ok {
		return reg
	} else {
		ctx.index++
		reg := &RegisterExpr{ctx.index}
		ctx.variableMap[v.Name] = reg
		return reg
	}
}

func (e *IntConstExpr) replaceVar(*RegisterAllocatorContext) IFExpr  { return e }
func (e *CharConstExpr) replaceVar(*RegisterAllocatorContext) IFExpr { return e }
func (e *ArrayExpr) replaceVar(*RegisterAllocatorContext) IFExpr     { return e }
func (e *LocationExpr) replaceVar(*RegisterAllocatorContext) IFExpr  { return e }
func (e *VarExpr) replaceVar(ctx *RegisterAllocatorContext) IFExpr {
	return ctx.getRegister(e)
}
func (e *RegisterExpr) replaceVar(*RegisterAllocatorContext) IFExpr { return e }
func (e *BinOpExpr) replaceVar(ctx *RegisterAllocatorContext) IFExpr {
	e.Left = e.Left.replaceVar(ctx)
	e.Right = e.Right.replaceVar(ctx)
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

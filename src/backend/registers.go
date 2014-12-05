package backend

type VariableScope struct {
	variableMap map[string]*RegisterExpr
	start       int
}

type RegisterAllocatorContext struct {
	scope       []VariableScope
	prevNode    *InstrNode
	currentNode *InstrNode
}

func (ctx *RegisterAllocatorContext) lookupVariable(v *VarExpr) *RegisterExpr {
	if reg, ok := ctx.scope[0].variableMap[v.Name]; ok {
		return reg
	} else {
		panic("Using variable without initialising")
	}
}

func (ctx *RegisterAllocatorContext) lookupOrCreateVariable(v *VarExpr) *RegisterExpr {
	if reg, ok := ctx.scope[0].variableMap[v.Name]; ok {
		return reg
	} else {
		index := ctx.scope[0].start
		ctx.scope[0].start++
		reg := &RegisterExpr{index}
		ctx.scope[0].variableMap[v.Name] = reg
		return reg
	}
}

func (ctx *RegisterAllocatorContext) pushInstr(i Instr) {
	ctx.prevNode.Next = &InstrNode{i, ctx.currentNode}
	ctx.prevNode = ctx.prevNode.Next
}

func (e *IntConstExpr) allocateRegisters(ctx *RegisterAllocatorContext, r int) {
	ctx.pushInstr(&MoveInstr{
		Dst: &RegisterExpr{r},
		Src: e})
}
func (e *CharConstExpr) allocateRegisters(ctx *RegisterAllocatorContext, r int) {}
func (e *ArrayExpr) allocateRegisters(ctx *RegisterAllocatorContext, r int)     {}
func (e *LocationExpr) allocateRegisters(ctx *RegisterAllocatorContext, r int)  {}
func (e *VarExpr) allocateRegisters(ctx *RegisterAllocatorContext, r int) {
	ctx.pushInstr(&MoveInstr{
		Dst: &RegisterExpr{r},
		Src: ctx.lookupVariable(e)})
}
func (e *RegisterExpr) allocateRegisters(ctx *RegisterAllocatorContext, r int) {}
func (e *BinOpExpr) allocateRegisters(ctx *RegisterAllocatorContext, r int) {
	e.Left.allocateRegisters(ctx, r)
	e.Right.allocateRegisters(ctx, r+1)
	ctx.pushInstr(&AddInstr{
		Dst: &RegisterExpr{r},
		Op1: &RegisterExpr{r},
		Op2: &RegisterExpr{r + 1}})
}
func (e *NotExpr) allocateRegisters(ctx *RegisterAllocatorContext, r int)     {}
func (e *OrdExpr) allocateRegisters(ctx *RegisterAllocatorContext, r int)     {}
func (e *ChrExpr) allocateRegisters(ctx *RegisterAllocatorContext, r int)     {}
func (e *NegExpr) allocateRegisters(ctx *RegisterAllocatorContext, r int)     {}
func (e *LenExpr) allocateRegisters(ctx *RegisterAllocatorContext, r int)     {}
func (e *NewPairExpr) allocateRegisters(ctx *RegisterAllocatorContext, r int) {}
func (e *CallExpr) allocateRegisters(ctx *RegisterAllocatorContext, r int)    {}

func (i *NoOpInstr) allocateRegisters(ctx *RegisterAllocatorContext)  {}
func (i *LabelInstr) allocateRegisters(ctx *RegisterAllocatorContext) {}
func (i *ReadInstr) allocateRegisters(ctx *RegisterAllocatorContext)  {}
func (i *FreeInstr) allocateRegisters(ctx *RegisterAllocatorContext)  {}
func (i *ExitInstr) allocateRegisters(ctx *RegisterAllocatorContext)  {}
func (i *PrintInstr) allocateRegisters(ctx *RegisterAllocatorContext) {}
func (i *MoveInstr) allocateRegisters(ctx *RegisterAllocatorContext) {
	r := ctx.scope[0].start
	i.Src.allocateRegisters(ctx, r)
	i.Src = &RegisterExpr{r}
	i.Dst = ctx.lookupOrCreateVariable(i.Dst.(*VarExpr))
}
func (i *TestInstr) allocateRegisters(ctx *RegisterAllocatorContext)    {}
func (i *JmpInstr) allocateRegisters(ctx *RegisterAllocatorContext)     {}
func (i *JmpZeroInstr) allocateRegisters(ctx *RegisterAllocatorContext) {}

// Second stage IF instructions should never do anything
func (*AddInstr) allocateRegisters(ctx *RegisterAllocatorContext) {}
func (*AddInstr) generateCode(*GeneratorContext)                  {}

func AllocateRegisters(ifCtx *IFContext) {
	ctx := new(RegisterAllocatorContext)
	ctx.scope = make([]VariableScope, 1)
	ctx.scope[0] = VariableScope{variableMap: make(map[string]*RegisterExpr), start: 0}
	ctx.currentNode = ifCtx.first

	// Iterate through nodes in the IF
	for ctx.currentNode != nil {
		ctx.currentNode.Instr.allocateRegisters(ctx)
		ctx.prevNode = ctx.currentNode
		ctx.currentNode = ctx.currentNode.Next
	}
}

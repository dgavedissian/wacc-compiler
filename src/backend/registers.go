package backend

import (
	"fmt"
)

type VariableScope struct {
	variableMap map[string]*RegisterExpr
	start       int
}

type RegisterAllocatorContext struct {
	scope              []VariableScope
	prevNode           *InstrNode
	currentNode        *InstrNode
	dataStore          map[string]Expr
	dataStoreAllocator map[string]int
}

func (ctx *RegisterAllocatorContext) lookupVariable(v *VarExpr) *RegisterExpr {
	if reg, ok := ctx.scope[0].variableMap[v.Name]; ok {
		return reg
	} else {
		panic(fmt.Sprintf("Using variable '%s' without initialising", v.Name))
	}
}

func (ctx *RegisterAllocatorContext) lookupOrCreateVariable(v *VarExpr) *RegisterExpr {
	if reg, ok := ctx.scope[0].variableMap[v.Name]; ok {
		return reg
	} else {
		index := ctx.scope[0].start
		ctx.scope[0].start++
		if ctx.scope[0].start > 12 {
			panic("exceeded maximum registers")
		}
		reg := &RegisterExpr{index}
		ctx.scope[0].variableMap[v.Name] = reg
		return reg
	}
}

func (ctx *RegisterAllocatorContext) pushInstr(i Instr) {
	ctx.prevNode.Next = &InstrNode{i, ctx.currentNode}
	ctx.prevNode = ctx.prevNode.Next
}

//
// Expressions
//
func (e *IntConstExpr) allocateRegisters(ctx *RegisterAllocatorContext, r int) {
	ctx.pushInstr(&MoveInstr{
		Dst: &RegisterExpr{r},
		Src: e})
}
func (e *BoolConstExpr) allocateRegisters(ctx *RegisterAllocatorContext, r int) {
	ctx.pushInstr(&MoveInstr{
		Dst: &RegisterExpr{r},
		Src: e})
}
func (e *CharConstExpr) allocateRegisters(ctx *RegisterAllocatorContext, r int) {
	ctx.pushInstr(&MoveInstr{
		Dst: &RegisterExpr{r},
		Src: e,
	})
}
func (e *ArrayExpr) allocateRegisters(ctx *RegisterAllocatorContext, r int) {
	ctx.pushInstr(&MoveInstr{
		Dst: &RegisterExpr{r},
		Src: &LocationExpr{ctx.pushDataStore(e)},
	})
}
func (e *LocationExpr) allocateRegisters(ctx *RegisterAllocatorContext, r int) {}
func (e *VarExpr) allocateRegisters(ctx *RegisterAllocatorContext, r int) {
	ctx.pushInstr(&MoveInstr{
		Dst: &RegisterExpr{r},
		Src: ctx.lookupVariable(e)})
}
func (e *RegisterExpr) allocateRegisters(ctx *RegisterAllocatorContext, r int) {
	ctx.pushInstr(&MoveInstr{
		Dst: &RegisterExpr{r},
		Src: e,
	})
}
func (e *BinOpExpr) allocateRegisters(ctx *RegisterAllocatorContext, r int) {
	e.Left.allocateRegisters(ctx, r)
	e.Right.allocateRegisters(ctx, r+1)
	dst := &RegisterExpr{r}
	op1 := dst
	op2 := &RegisterExpr{r + 1}

	switch e.Operator {
	case Add:
		ctx.pushInstr(&AddInstr{Dst: dst, Op1: op1, Op2: op2})
	case Sub:
		ctx.pushInstr(&SubInstr{Dst: dst, Op1: op1, Op2: op2})
	case Mul:
		ctx.pushInstr(&MulInstr{Dst: dst, Op1: op1, Op2: op2})
	case Div:
		ctx.pushInstr(&DivInstr{Dst: dst, Op1: op1, Op2: op2})
	case And:
		ctx.pushInstr(&AndInstr{Dst: dst, Op1: op1, Op2: op2})
	case Or:
		ctx.pushInstr(&OrInstr{Dst: dst, Op1: op1, Op2: op2})
	case Mod:
		op3 := &RegisterExpr{r + 2}
		ctx.pushInstr(&DivInstr{Dst: op3, Op1: op1, Op2: op2})
		ctx.pushInstr(&MulInstr{Dst: op3, Op1: op3, Op2: op2})
		ctx.pushInstr(&SubInstr{Dst: dst, Op1: op1, Op2: op3})
	case LT, GT, LE, GE, EQ, NE:
		ctx.pushInstr(&CmpInstr{Dst: dst, Left: op1, Right: op2, Operator: e.Operator})
		ctx.pushInstr(&DeclareTypeInstr{dst, &BoolConstExpr{}})
	default:
		panic(fmt.Sprintf("Unknown operator %v", e.Operator))
	}
}
func (e *NotExpr) allocateRegisters(ctx *RegisterAllocatorContext, r int) {
	e.Operand.allocateRegisters(ctx, r)
	dst := &RegisterExpr{r}
	dst2 := &RegisterExpr{r + 1}
	ctx.pushInstr(&NotInstr{dst, dst})
	ctx.pushInstr(&MoveInstr{dst2, &IntConstExpr{1}})
	ctx.pushInstr(&AndInstr{Dst: dst, Op1: dst, Op2: dst2})
}
func (e *OrdExpr) allocateRegisters(ctx *RegisterAllocatorContext, r int) {
	e.Operand.allocateRegisters(ctx, r)
	ctx.pushInstr(&DeclareTypeInstr{&RegisterExpr{r}, &IntConstExpr{}})
}
func (e *ChrExpr) allocateRegisters(ctx *RegisterAllocatorContext, r int) {
	e.Operand.allocateRegisters(ctx, r)
	ctx.pushInstr(&DeclareTypeInstr{&RegisterExpr{r}, &CharConstExpr{}})
}
func (e *NegExpr) allocateRegisters(ctx *RegisterAllocatorContext, r int)     {}
func (e *LenExpr) allocateRegisters(ctx *RegisterAllocatorContext, r int)     {}
func (e *NewPairExpr) allocateRegisters(ctx *RegisterAllocatorContext, r int) {}

func (e *CallExpr) allocateRegisters(ctx *RegisterAllocatorContext, r int) {
	// save registers
	ctx.pushInstr(&PushScopeInstr{})

	// Move arguments into r0-r4
	if len(e.Args) > 4 {
		panic("unimplemented!")
	}
	for n, arg := range e.Args {
		arg.allocateRegisters(ctx, n)
	}

	// Call function
	ctx.pushInstr(&CallInstr{Label: e.Label})

	// Copy result into r
	ctx.pushInstr(&MoveInstr{
		Dst: &RegisterExpr{r},
		Src: &RegisterExpr{0}})

	// restore registers
	ctx.pushInstr(&PopScopeInstr{})
}

//
// Instructions
//
func (i *NoOpInstr) allocateRegisters(ctx *RegisterAllocatorContext)  {}
func (i *LabelInstr) allocateRegisters(ctx *RegisterAllocatorContext) {}
func (i *ReadInstr) allocateRegisters(ctx *RegisterAllocatorContext) {
	i.Dst = ctx.lookupOrCreateVariable(i.Dst.(*VarExpr))
}
func (i *FreeInstr) allocateRegisters(ctx *RegisterAllocatorContext) {}

func (i *ReturnInstr) allocateRegisters(ctx *RegisterAllocatorContext) {
	r := ctx.scope[0].start
	i.Expr.allocateRegisters(ctx, r)
	i.Expr = &RegisterExpr{r}
}

func (i *ExitInstr) allocateRegisters(ctx *RegisterAllocatorContext) {
	r := ctx.scope[0].start
	i.Expr.allocateRegisters(ctx, r)
	i.Expr = &RegisterExpr{r}
}

func (i *PrintInstr) allocateRegisters(ctx *RegisterAllocatorContext) {
	r := ctx.scope[0].start
	i.Expr.allocateRegisters(ctx, r)
	i.Expr = &RegisterExpr{r}
}

func (i *MoveInstr) allocateRegisters(ctx *RegisterAllocatorContext) {
	r := ctx.scope[0].start
	i.Src.allocateRegisters(ctx, r)
	i.Src = &RegisterExpr{r}
	i.Dst = ctx.lookupOrCreateVariable(i.Dst.(*VarExpr))
}

func (i *NotInstr) allocateRegisters(ctx *RegisterAllocatorContext) {
	r := ctx.scope[0].start
	i.Src.allocateRegisters(ctx, r)
	i.Src = &RegisterExpr{r}
	i.Dst = ctx.lookupOrCreateVariable(i.Dst.(*VarExpr))
}

func (i *CmpInstr) allocateRegisters(ctx *RegisterAllocatorContext) {
	r := ctx.scope[0].start
	i.Left.allocateRegisters(ctx, r)
	i.Right.allocateRegisters(ctx, r+1)
	i.Left = &RegisterExpr{r}
	i.Right = &RegisterExpr{r + 1}
}

func (i *JmpInstr) allocateRegisters(ctx *RegisterAllocatorContext) {}
func (i *JmpCondInstr) allocateRegisters(ctx *RegisterAllocatorContext) {
	r := ctx.scope[0].start
	i.Cond.allocateRegisters(ctx, r)
	i.Cond = &RegisterExpr{r}
}
func (*PushScopeInstr) allocateRegisters(ctx *RegisterAllocatorContext) {}
func (*PopScopeInstr) allocateRegisters(ctx *RegisterAllocatorContext)  {}
func (i *DeclareTypeInstr) allocateRegisters(ctx *RegisterAllocatorContext) {
	i.Dst = ctx.lookupVariable(i.Dst.(*VarExpr))
}

// Second stage IF instructions should never do anything
func (*AddInstr) allocateRegisters(ctx *RegisterAllocatorContext)  {}
func (*SubInstr) allocateRegisters(ctx *RegisterAllocatorContext)  {}
func (*MulInstr) allocateRegisters(ctx *RegisterAllocatorContext)  {}
func (*DivInstr) allocateRegisters(ctx *RegisterAllocatorContext)  {}
func (*AndInstr) allocateRegisters(ctx *RegisterAllocatorContext)  {}
func (*OrInstr) allocateRegisters(ctx *RegisterAllocatorContext)   {}
func (*CallInstr) allocateRegisters(ctx *RegisterAllocatorContext) {}

func (ctx *RegisterAllocatorContext) allocateRegistersForBranch(n *InstrNode) {
	ctx.currentNode = n
	for ctx.currentNode != nil {
		ctx.currentNode.Instr.allocateRegisters(ctx)
		ctx.prevNode = ctx.currentNode
		ctx.currentNode = ctx.currentNode.Next
	}
}
func (ctx *RegisterAllocatorContext) pushDataStore(e Expr) string {
	var nextIndex int
	var ok bool

	dataStorePrefix := ctx.dataStorePrefix(e)
	nextIndex, ok = ctx.dataStoreAllocator[dataStorePrefix]
	if !ok {
		nextIndex = 0
	}

	label := fmt.Sprintf("%s%d", dataStorePrefix, nextIndex)

	ctx.dataStore[label] = e
	ctx.dataStoreAllocator[dataStorePrefix] = nextIndex + 1

	return label
}
func (ctx *RegisterAllocatorContext) dataStorePrefix(e Expr) string {
	switch e.(type) {
	case *ArrayExpr:
		return "arraylit"
	default:
		panic("Unknown item attempted to store data?!?")
	}
}

func AllocateRegisters(ifCtx *IFContext) {
	ctx := new(RegisterAllocatorContext)
	ctx.scope = make([]VariableScope, 1)
	ctx.scope[0] = VariableScope{variableMap: make(map[string]*RegisterExpr), start: 4}
	ctx.dataStore = make(map[string]Expr)
	ctx.dataStoreAllocator = make(map[string]int)

	// Iterate through nodes in the IF
	for _, f := range ifCtx.functions {
		ctx.allocateRegistersForBranch(f)
	}
	ctx.allocateRegistersForBranch(ifCtx.main)

	ifCtx.dataStore = ctx.dataStore
}

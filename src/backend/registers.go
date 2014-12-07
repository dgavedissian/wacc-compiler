package backend

import (
	"fmt"

	"../frontend"
)

type VariableScope struct {
	variableMap        map[string]*RegisterExpr
	stashedVariableMap map[string]int
	next               int
	registerUseMap     map[RegisterExpr]bool
}

type RegisterAllocatorContext struct {
	scope          []VariableScope
	prevNode       *InstrNode
	currentNode    *InstrNode
	dataStore      map[string]*StringConstExpr
	dataStoreIndex int
}

func (ctx *RegisterAllocatorContext) tryAllocateRegister() (*RegisterExpr, bool) {
	var r *RegisterExpr
	for reg, inUse := range ctx.scope[0].registerUseMap {
		if !inUse {
			r = &reg
			break
		}
	}

	if r == nil {
		if !ctx.stashLiveVariables() {
			return nil, false
		}
		return ctx.allocateRegister(), true
	}
	ctx.scope[0].registerUseMap[*r] = true
	return r, true
}

func (ctx *RegisterAllocatorContext) allocateRegister() *RegisterExpr {
	r, ok := ctx.tryAllocateRegister()
	if !ok {
		panic("unable to relieve memory pressure by moving variables to stack")
	}
	return r
}

func (ctx *RegisterAllocatorContext) freeRegister(r *RegisterExpr) {
	if !ctx.scope[0].registerUseMap[*r] {
		panic("Freeing register not in use?!?")
	}
	ctx.scope[0].registerUseMap[*r] = false
}

func (ctx *RegisterAllocatorContext) innerLookupVariable(v *VarExpr) (*RegisterExpr, bool) {
	if reg, ok := ctx.scope[0].variableMap[v.Name]; ok {
		return reg, true
	} else if stash, ok := ctx.scope[0].stashedVariableMap[v.Name]; ok {
		x := ctx.allocateRegister()
		ctx.scope[0].variableMap[v.Name] = x
		ctx.pushInstr(&MoveInstr{
			Src: &StackLocationExpr{stash},
			Dst: x,
		})
		// we need to emit typing instructions, too
		return x, true
	} else {
		return nil, false
	}
}

func (ctx *RegisterAllocatorContext) lookupVariable(v *VarExpr) *RegisterExpr {
	if reg, ok := ctx.innerLookupVariable(v); ok {
		return reg
	} else {
		panic(fmt.Sprintf("Using variable '%s' without initialising", v.Name))
	}
}

func (ctx *RegisterAllocatorContext) lookupOrCreateVariable(v *VarExpr) *RegisterExpr {
	if reg, ok := ctx.innerLookupVariable(v); ok {
		return reg
	} else {
		reg := ctx.allocateRegister()
		ctx.scope[0].variableMap[v.Name] = reg
		return reg
	}
}

func (ctx *RegisterAllocatorContext) translateLValue(e Expr) Expr {
	switch expr := e.(type) {
	case *VarExpr:
		return ctx.lookupOrCreateVariable(expr)

	case *PairElemExpr:
		var offset int
		if expr.Fst {
			offset = 0
		} else {
			offset = regWidth
		}
		return &MemExpr{ctx.lookupVariable(expr.Operand), offset}

	default:
		panic(fmt.Sprintf("Unhandled lvalue %T", expr))
	}
}

func (ctx *RegisterAllocatorContext) stashLiveVariables() bool {
	if len(ctx.scope[0].variableMap) == 0 {
		return false
	}

	for vname, expr := range ctx.scope[0].variableMap {
		var n int
		if previousN, ok := ctx.scope[0].stashedVariableMap[vname]; ok {
			n = previousN
		} else {
			n = ctx.scope[0].next
			ctx.scope[0].next++
		}
		ctx.pushInstr(&MoveInstr{
			Dst: &StackLocationExpr{n},
			Src: expr,
		})
		// we need to emit typing instructions, too
		ctx.scope[0].stashedVariableMap[vname] = n
		ctx.freeRegister(expr)
	}
	ctx.scope[0].variableMap = make(map[string]*RegisterExpr)
	return true
}

func (ctx *RegisterAllocatorContext) pushInstr(i Instr) {
	ctx.prevNode.Next = &InstrNode{i, 0, ctx.currentNode}
	ctx.prevNode = ctx.prevNode.Next
}

//
// Expressions
//
func (e *TypeExpr) allocateRegisters(*RegisterAllocatorContext, int) {}

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
		Src: e})
}

func (e *StringConstExpr) allocateRegisters(ctx *RegisterAllocatorContext, r int) {
	ctx.pushInstr(&MoveInstr{
		Dst: &RegisterExpr{r},
		Src: &LocationExpr{ctx.pushDataStore(e)}})
}

func (e *PointerConstExpr) allocateRegisters(ctx *RegisterAllocatorContext, r int) {
	ctx.pushInstr(&MoveInstr{
		Dst: &RegisterExpr{r},
		Src: e})
}

func (e *ArrayConstExpr) allocateRegisters(ctx *RegisterAllocatorContext, r int) {
	// Allocate space on the heap
	length := len(e.Elems)
	ctx.pushInstr(&HeapAllocInstr{&RegisterExpr{r}, (length + 1) * regWidth})
	ctx.pushInstr(&MoveInstr{
		&RegisterExpr{r + 1},
		&IntConstExpr{length}})
	ctx.pushInstr(&MoveInstr{
		Dst: &MemExpr{&RegisterExpr{r}, 0},
		Src: &RegisterExpr{r + 1}})

	// Copy each element into the array
	for i, e := range e.Elems {
		e.allocateRegisters(ctx, r+1)
		ctx.pushInstr(&MoveInstr{
			Dst: &MemExpr{&RegisterExpr{r}, (i + 1) * regWidth},
			Src: &RegisterExpr{r + 1}})
	}
}

func (e *LocationExpr) allocateRegisters(ctx *RegisterAllocatorContext, r int) {}

func (e *VarExpr) allocateRegisters(ctx *RegisterAllocatorContext, r int) {
	ctx.pushInstr(&MoveInstr{
		Dst: &RegisterExpr{r},
		Src: ctx.lookupVariable(e)})
}

func (e *MemExpr) allocateRegisters(ctx *RegisterAllocatorContext, r int) {
	/*e.Expr.allocateRegisters(ctx, r)
	ctx.pushInstr(&MoveInstr{
		Dst: &RegisterExpr{r},
		Src: e})*/
}

func (e *RegisterExpr) allocateRegisters(ctx *RegisterAllocatorContext, r int) {
	ctx.pushInstr(&MoveInstr{
		Dst: &RegisterExpr{r},
		Src: e,
	})
}

func (e *StackLocationExpr) allocateRegisters(ctx *RegisterAllocatorContext, r int) {
	ctx.pushInstr(&MoveInstr{
		Dst: &RegisterExpr{r},
		Src: e,
	})
}

func (e *ArrayElemExpr) allocateRegisters(ctx *RegisterAllocatorContext, r int) {
	e.Array.allocateRegisters(ctx, r+1)
	e.Index.allocateRegisters(ctx, r+2)
	ctx.pushInstr(&AddInstr{
		Dst:      &RegisterExpr{r + 1},
		Op1:      &RegisterExpr{r + 1},
		Op2:      &IntConstExpr{4},
		Op2Shift: nil})
	ctx.pushInstr(&AddInstr{
		Dst:      &RegisterExpr{r + 1},
		Op1:      &RegisterExpr{r + 1},
		Op2:      &RegisterExpr{r + 2},
		Op2Shift: &LSL{2}})
	ctx.pushInstr(&MoveInstr{
		Dst: &RegisterExpr{r},
		Src: &MemExpr{&RegisterExpr{r + 1}, 0}})
	ctx.pushInstr(&DeclareTypeInstr{
		&RegisterExpr{r},
		&TypeExpr{frontend.BasicType{frontend.INT}}})
}

func (e *PairElemExpr) allocateRegisters(ctx *RegisterAllocatorContext, r int) {
	var offset int
	if e.Fst {
		offset = 0
	} else {
		offset = regWidth
	}

	ctx.pushInstr(&MoveInstr{
		Dst: &RegisterExpr{r},
		Src: &MemExpr{ctx.lookupVariable(e.Operand), offset}})
}

func (e *BinOpExpr) allocateRegisters(ctx *RegisterAllocatorContext, r int) {
	var op1, op2, r2 *RegisterExpr
	dst := &RegisterExpr{r}
	if e.Left.Weight() > e.Right.Weight() {
		e.Left.allocateRegisters(ctx, r)
		r2 = ctx.allocateRegister()
		e.Right.allocateRegisters(ctx, r2.Id)

		op1 = dst
		op2 = r2
	} else {
		e.Right.allocateRegisters(ctx, r)
		r2 = ctx.allocateRegister()
		e.Left.allocateRegisters(ctx, r2.Id)

		op1 = r2
		op2 = dst
	}

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
		op3 := ctx.allocateRegister()
		ctx.pushInstr(&DivInstr{Dst: op3, Op1: op1, Op2: op2})
		ctx.pushInstr(&MulInstr{Dst: op3, Op1: op3, Op2: op2})
		ctx.pushInstr(&SubInstr{Dst: dst, Op1: op1, Op2: op3})
		ctx.freeRegister(op3)

	case LT, GT, LE, GE, EQ, NE:
		ctx.pushInstr(&CmpInstr{Dst: dst, Left: op1, Right: op2, Operator: e.Operator})
		ctx.pushInstr(&DeclareTypeInstr{dst, &TypeExpr{frontend.BasicType{frontend.BOOL}}})

	default:
		panic(fmt.Sprintf("Unknown operator %v", e.Operator))
	}
	ctx.freeRegister(r2)
}

func (e *NotExpr) allocateRegisters(ctx *RegisterAllocatorContext, r int) {
	e.Operand.allocateRegisters(ctx, r)
	dst := &RegisterExpr{r}
	dst2 := ctx.allocateRegister()
	ctx.pushInstr(&NotInstr{dst, dst})
	ctx.pushInstr(&MoveInstr{dst2, &IntConstExpr{1}})
	ctx.pushInstr(&AndInstr{Dst: dst, Op1: dst, Op2: dst2})
	ctx.freeRegister(dst2)
}

func (e *OrdExpr) allocateRegisters(ctx *RegisterAllocatorContext, r int) {
	e.Operand.allocateRegisters(ctx, r)
	ctx.pushInstr(&DeclareTypeInstr{&RegisterExpr{r}, &TypeExpr{frontend.BasicType{frontend.INT}}})
}

func (e *ChrExpr) allocateRegisters(ctx *RegisterAllocatorContext, r int) {
	e.Operand.allocateRegisters(ctx, r)
	ctx.pushInstr(&DeclareTypeInstr{&RegisterExpr{r}, &TypeExpr{frontend.BasicType{frontend.INT}}})
}

func (e *NegExpr) allocateRegisters(ctx *RegisterAllocatorContext, r int) {}

func (e *LenExpr) allocateRegisters(ctx *RegisterAllocatorContext, r int) {
	ctx.pushInstr(&MoveInstr{
		&RegisterExpr{r},
		&MemExpr{ctx.lookupVariable(e.Operand.(*VarExpr)), 0}})
	ctx.pushInstr(&DeclareTypeInstr{
		&RegisterExpr{r},
		&TypeExpr{frontend.BasicType{frontend.INT}}})
}

func (e *NewPairExpr) allocateRegisters(ctx *RegisterAllocatorContext, r int) {
	ctx.pushInstr(&HeapAllocInstr{&RegisterExpr{r}, 2 * regWidth})
	e.Left.allocateRegisters(ctx, r+1)
	ctx.pushInstr(&MoveInstr{
		&MemExpr{&RegisterExpr{r}, 0},
		&RegisterExpr{r + 1}})
	e.Right.allocateRegisters(ctx, r+1)
	ctx.pushInstr(&MoveInstr{
		&MemExpr{&RegisterExpr{r}, regWidth},
		&RegisterExpr{r + 1}})
}

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
func (i *NoOpInstr) allocateRegisters(ctx *RegisterAllocatorContext) {}

func (i *LabelInstr) allocateRegisters(ctx *RegisterAllocatorContext) {}

func (i *ReadInstr) allocateRegisters(ctx *RegisterAllocatorContext) {
	i.Dst = ctx.lookupOrCreateVariable(i.Dst.(*VarExpr))
}

func (i *FreeInstr) allocateRegisters(ctx *RegisterAllocatorContext) {}

func (i *ReturnInstr) allocateRegisters(ctx *RegisterAllocatorContext) {
	r := ctx.allocateRegister()
	i.Expr.allocateRegisters(ctx, r.Id)
	i.Expr = r
	ctx.freeRegister(r)
}

func (i *ExitInstr) allocateRegisters(ctx *RegisterAllocatorContext) {
	r := ctx.allocateRegister()
	i.Expr.allocateRegisters(ctx, r.Id)
	i.Expr = r
	ctx.freeRegister(r)
}

func (i *PrintInstr) allocateRegisters(ctx *RegisterAllocatorContext) {
	if _, ok := i.Expr.(*CharConstExpr); ok {
		return
	}
	if _, ok := i.Expr.(*BoolConstExpr); ok {
		return
	}
	if _, ok := i.Expr.(*IntConstExpr); ok {
		return
	}
	r := ctx.allocateRegister()
	i.Expr.allocateRegisters(ctx, r.Id)
	i.Expr = r
	ctx.freeRegister(r)
}

func (i *MoveInstr) allocateRegisters(ctx *RegisterAllocatorContext) {
	r := ctx.allocateRegister()
	i.Src.allocateRegisters(ctx, r.Id)
	i.Src = r
	i.Dst = ctx.translateLValue(i.Dst)
	ctx.freeRegister(r)
}

func (i *NotInstr) allocateRegisters(ctx *RegisterAllocatorContext) {
	r := ctx.allocateRegister()
	i.Src.allocateRegisters(ctx, r.Id)
	i.Src = r
	i.Dst = ctx.translateLValue(i.Dst)
	ctx.freeRegister(r)
}

func (i *CmpInstr) allocateRegisters(ctx *RegisterAllocatorContext) {
	r := ctx.allocateRegister()
	r2 := ctx.allocateRegister()
	i.Left.allocateRegisters(ctx, r.Id)
	i.Right.allocateRegisters(ctx, r2.Id)
	i.Left = r
	i.Right = r2
	ctx.freeRegister(r)
	ctx.freeRegister(r2)
}

func (i *JmpInstr) allocateRegisters(ctx *RegisterAllocatorContext) {}

func (i *JmpCondInstr) allocateRegisters(ctx *RegisterAllocatorContext) {
	r := ctx.allocateRegister()
	i.Cond.allocateRegisters(ctx, r.Id)
	i.Cond = r
	ctx.freeRegister(r)
}

func (*PushScopeInstr) allocateRegisters(ctx *RegisterAllocatorContext) {}

func (*PopScopeInstr) allocateRegisters(ctx *RegisterAllocatorContext) {}

func (i *DeclareTypeInstr) allocateRegisters(ctx *RegisterAllocatorContext) {
	i.Dst = ctx.lookupVariable(i.Dst.(*VarExpr))
}

// Second stage IF instructions should never do anything
func (*AddInstr) allocateRegisters(ctx *RegisterAllocatorContext)       {}
func (*SubInstr) allocateRegisters(ctx *RegisterAllocatorContext)       {}
func (*MulInstr) allocateRegisters(ctx *RegisterAllocatorContext)       {}
func (*DivInstr) allocateRegisters(ctx *RegisterAllocatorContext)       {}
func (*AndInstr) allocateRegisters(ctx *RegisterAllocatorContext)       {}
func (*OrInstr) allocateRegisters(ctx *RegisterAllocatorContext)        {}
func (*CallInstr) allocateRegisters(ctx *RegisterAllocatorContext)      {}
func (*HeapAllocInstr) allocateRegisters(ctx *RegisterAllocatorContext) {}

func (ctx *RegisterAllocatorContext) allocateRegistersForBranch(n *InstrNode) {
	ctx.currentNode = n
	for ctx.currentNode != nil {
		ctx.currentNode.Instr.allocateRegisters(ctx)
		ctx.prevNode = ctx.currentNode
		ctx.currentNode = ctx.currentNode.Next
	}
	n.stackSpace = ctx.scope[0].next
}

func (ctx *RegisterAllocatorContext) pushDataStore(e *StringConstExpr) string {
	label := fmt.Sprintf("stringlit%d", ctx.dataStoreIndex)
	ctx.dataStore[label] = e
	ctx.dataStoreIndex = ctx.dataStoreIndex + 1
	return label
}

func makeStartRegisterMap() map[RegisterExpr]bool {
	x := make(map[RegisterExpr]bool)
	for n := 4; n <= 12; n++ {
		x[RegisterExpr{n}] = false
	}
	return x
}

func AllocateRegisters(ifCtx *IFContext) {
	ctx := new(RegisterAllocatorContext)
	ctx.scope = make([]VariableScope, 1)
	ctx.scope[0] = VariableScope{
		variableMap:        make(map[string]*RegisterExpr),
		registerUseMap:     makeStartRegisterMap(),
		stashedVariableMap: make(map[string]int)}
	ctx.dataStore = make(map[string]*StringConstExpr)
	ctx.dataStoreIndex = 0

	// Iterate through nodes in the IF
	for _, f := range ifCtx.functions {
		ctx.allocateRegistersForBranch(f)
	}
	ctx.allocateRegistersForBranch(ifCtx.main)
	ctx.pushInstr(&MoveInstr{&RegisterExpr{0}, &IntConstExpr{0}})

	ifCtx.dataStore = ctx.dataStore
}

package backend

import (
	"fmt"
	"log"

	"../frontend"
)

type Variable struct {
	stack       int
	typeInfo    frontend.Type
	initialised bool
}

type VariableScope struct {
	variableMap map[string]*Variable
	next        int
}

type RegisterAllocatorContext struct {
	// Variable scoping
	scope []VariableScope
	depth int

	// Registers in use
	registerUseList [12]bool

	// Current location in the list
	prevNode    *InstrNode
	currentNode *InstrNode

	// String data store
	dataStore      map[string]*StringConstExpr
	dataStoreIndex int
}

func (ctx *RegisterAllocatorContext) allocateRegister() *RegisterExpr {
	var reg *RegisterExpr
	for k, inUse := range ctx.registerUseList {
		if k < 4 {
			continue
		}
		if !inUse {
			reg = &RegisterExpr{k}
			break
		}
	}

	if reg == nil {
		panic("Ran out of registers - need to spill")
	}
	ctx.registerUseList[reg.Id] = true
	return reg
}

func (ctx *RegisterAllocatorContext) freeRegister(r *RegisterExpr) {
	if !ctx.registerUseList[r.Id] {
		panic("Freeing register not in use")
	}
	ctx.registerUseList[r.Id] = false
}

func (ctx *RegisterAllocatorContext) innerLookupVariable(v *VarExpr) (*Variable, bool) {
	// Search all scopes from the top most for this variable
	for i := ctx.depth - 1; i >= 0; i-- {
		if v, ok := ctx.scope[i].variableMap[v.Name]; ok {
			if v.initialised {
				return v, true
			}
		}
	}

	// Give up
	return nil, false
}

func (ctx *RegisterAllocatorContext) initialiseVariable(v *VarExpr) {
	// Search all scopes from the top most for this variable
	for i := ctx.depth - 1; i >= 0; i-- {
		if v, ok := ctx.scope[i].variableMap[v.Name]; ok {
			v.initialised = true
			return
		}
	}

	// Give up
	panic(fmt.Sprintf("Trying to initialise non-existent variable '%s'", v.Name))
}

func (ctx *RegisterAllocatorContext) lookupVariable(v *VarExpr) *StackLocationExpr {
	if innerVar, ok := ctx.innerLookupVariable(v); ok {
		return &StackLocationExpr{innerVar.stack}
	} else {
		panic(fmt.Sprintf("Trying to access non-existent variable '%s'", v.Name))
	}
}

func (ctx *RegisterAllocatorContext) lookupType(v *VarExpr) frontend.Type {
	if innerVar, ok := ctx.innerLookupVariable(v); ok {
		return innerVar.typeInfo
	} else {
		panic(fmt.Sprintf("Trying to get type of non-existent variable '%s'", v.Name))
	}
}

func (ctx *RegisterAllocatorContext) createVariable(d *DeclareInstr) *StackLocationExpr {
	n := ctx.scope[ctx.depth-1].next
	log.Printf("New variable at scope %d (stack pos %d)\n", ctx.depth-1, n)
	ctx.scope[ctx.depth-1].variableMap[d.Var.Name] = &Variable{n, d.Type, false}
	ctx.scope[ctx.depth-1].next++
	return &StackLocationExpr{n}
}

func (ctx *RegisterAllocatorContext) pushInstr(i Instr) {
	ctx.prevNode.Next = &InstrNode{i, 0, ctx.currentNode}
	ctx.prevNode = ctx.prevNode.Next
}

func (ctx *RegisterAllocatorContext) allocateRegistersForBranch(n *InstrNode) {
	ctx.pushScope()
	ctx.currentNode = n
	for ctx.currentNode != nil {
		ctx.currentNode.Instr.allocateRegisters(ctx)
		ctx.prevNode = ctx.currentNode
		ctx.currentNode = ctx.currentNode.Next
	}
	n.stackSpace = ctx.scope[0].next
	ctx.popScope()
}

func (ctx *RegisterAllocatorContext) pushDataStore(e *StringConstExpr) string {
	label := fmt.Sprintf("stringlit%d", ctx.dataStoreIndex)
	ctx.dataStore[label] = e
	ctx.dataStoreIndex = ctx.dataStoreIndex + 1
	return label
}

func (ctx *RegisterAllocatorContext) pushScope() {
	// Create a new scope and start at the next available stack address of the
	// parent scope
	newScope := VariableScope{variableMap: make(map[string]*Variable)}
	if ctx.depth > 0 {
		newScope.next = ctx.scope[ctx.depth-1].next
	}

	// Add to the top of the scope stack
	ctx.scope = append(ctx.scope, newScope)
	ctx.depth++
}

func (ctx *RegisterAllocatorContext) popScope() {
	ctx.scope = ctx.scope[:ctx.depth-1]
	ctx.depth--
}

func AllocateRegisters(ifCtx *IFContext) {
	ctx := new(RegisterAllocatorContext)
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

//
// Expression Utils
//
func (ctx *RegisterAllocatorContext) translateLValue(e Expr, r *RegisterExpr) Expr {
	switch expr := e.(type) {
	case *VarExpr:
		ctx.initialiseVariable(expr)
		return ctx.lookupVariable(expr)

	case *ArrayElemExpr:
		arrayPtr := r
		index := ctx.allocateRegister()

		expr.Array.allocateRegisters(ctx, arrayPtr)
		expr.Index.allocateRegisters(ctx, index)

		// Runtime safety check
		ctx.pushInstr(&PushInstr{&RegisterExpr{0}})
		ctx.pushInstr(&PushInstr{&RegisterExpr{1}})
		ctx.pushInstr(&MoveInstr{Dst: &RegisterExpr{1}, Src: arrayPtr})
		ctx.pushInstr(&MoveInstr{Dst: &RegisterExpr{0}, Src: index})
		ctx.pushInstr(&CallInstr{Label: &LocationExpr{RuntimeCheckArrayBoundsLabel}})
		ctx.pushInstr(&CheckNullDereferenceInstr{arrayPtr})
		ctx.pushInstr(&PopInstr{&RegisterExpr{1}})
		ctx.pushInstr(&PopInstr{&RegisterExpr{0}})

		ctx.pushInstr(&AddInstr{
			Dst:      arrayPtr,
			Op1:      arrayPtr,
			Op2:      index,
			Op2Shift: &LSL{2}})

		ctx.freeRegister(index)
		return &MemExpr{arrayPtr, 4}

	case *PairElemExpr:
		var offset int
		if expr.Fst {
			offset = 0
		} else {
			offset = regWidth
		}

		v := ctx.lookupVariable(expr.Operand)
		ctx.pushInstr(&MoveInstr{r, v})
		ctx.pushInstr(&CheckNullDereferenceInstr{r})
		return &MemExpr{r, offset}

	default:
		panic(fmt.Sprintf("Unhandled lvalue %T", expr))
	}
}

//
// Expressions
//
func (e *TypeExpr) allocateRegisters(*RegisterAllocatorContext, int) {}

func (e *IntConstExpr) allocateRegisters(ctx *RegisterAllocatorContext, r *RegisterExpr) {
	ctx.pushInstr(&MoveInstr{Dst: r, Src: e})
}

func (e *BoolConstExpr) allocateRegisters(ctx *RegisterAllocatorContext, r *RegisterExpr) {
	ctx.pushInstr(&MoveInstr{Dst: r, Src: e})
}

func (e *CharConstExpr) allocateRegisters(ctx *RegisterAllocatorContext, r *RegisterExpr) {
	ctx.pushInstr(&MoveInstr{Dst: r, Src: e})
}

func (e *StringConstExpr) allocateRegisters(ctx *RegisterAllocatorContext, r *RegisterExpr) {
	ctx.pushInstr(&MoveInstr{Dst: r, Src: &LocationExpr{ctx.pushDataStore(e)}})
}

func (e *PointerConstExpr) allocateRegisters(ctx *RegisterAllocatorContext, r *RegisterExpr) {
	ctx.pushInstr(&MoveInstr{Dst: r, Src: e})
}

func (e *ArrayConstExpr) allocateRegisters(ctx *RegisterAllocatorContext, r *RegisterExpr) {
	helperReg := ctx.allocateRegister()

	// Allocate space on the heap
	length := len(e.Elems)
	ctx.pushInstr(&HeapAllocInstr{r, (length + 1) * regWidth})
	ctx.pushInstr(&MoveInstr{
		helperReg,
		&IntConstExpr{length}})
	ctx.pushInstr(&MoveInstr{
		Dst: &MemExpr{r, 0},
		Src: helperReg})

	// Copy each element into the array
	for i, e := range e.Elems {
		e.allocateRegisters(ctx, helperReg)
		ctx.pushInstr(&MoveInstr{
			Dst: &MemExpr{r, (i + 1) * regWidth},
			Src: helperReg})
	}

	ctx.freeRegister(helperReg)
}

func (e *LocationExpr) allocateRegisters(ctx *RegisterAllocatorContext, r *RegisterExpr) {}

func (e *VarExpr) allocateRegisters(ctx *RegisterAllocatorContext, r *RegisterExpr) {
	variable := ctx.lookupVariable(e)
	ctx.pushInstr(&MoveInstr{Dst: r, Src: variable})
}

func (e *MemExpr) allocateRegisters(ctx *RegisterAllocatorContext, r *RegisterExpr) {
}

func (e *RegisterExpr) allocateRegisters(ctx *RegisterAllocatorContext, r *RegisterExpr) {
	ctx.pushInstr(&MoveInstr{Dst: r, Src: e})
}

func (e *StackLocationExpr) allocateRegisters(ctx *RegisterAllocatorContext, r *RegisterExpr) {
	ctx.pushInstr(&MoveInstr{Dst: r, Src: e})
}

func (e *StackArgumentExpr) allocateRegisters(ctx *RegisterAllocatorContext, r *RegisterExpr) {
	ctx.pushInstr(&MoveInstr{Dst: r, Src: e})
}

func (e *ArrayElemExpr) allocateRegisters(ctx *RegisterAllocatorContext, r *RegisterExpr) {
	helperReg := ctx.allocateRegister()
	ctx.pushInstr(&MoveInstr{Dst: r, Src: ctx.translateLValue(e, helperReg)})
	ctx.freeRegister(helperReg)
}

func (e *PairElemExpr) allocateRegisters(ctx *RegisterAllocatorContext, r *RegisterExpr) {
	helperReg := ctx.allocateRegister()
	ctx.pushInstr(&MoveInstr{Dst: r, Src: ctx.translateLValue(e, helperReg)})
	ctx.freeRegister(helperReg)
}

func (e *UnaryExpr) allocateRegisters(ctx *RegisterAllocatorContext, r *RegisterExpr) {
	e.Operand.allocateRegisters(ctx, r)
	switch e.Operator {
	case Not:
		dst := r
		dst2 := ctx.allocateRegister()
		ctx.pushInstr(&NotInstr{dst, dst})
		ctx.pushInstr(&MoveInstr{dst2, &IntConstExpr{1}})
		ctx.pushInstr(&AndInstr{Dst: dst, Op1: dst, Op2: dst2})
		ctx.freeRegister(dst2)

	case Ord:
		ctx.pushInstr(&DeclareTypeInstr{r, &TypeExpr{frontend.BasicType{frontend.INT}}})

	case Chr:
		ctx.pushInstr(&DeclareTypeInstr{r, &TypeExpr{frontend.BasicType{frontend.CHAR}}})

	case Neg:
		dst := r
		ctx.pushInstr(&NegInstr{dst})

	case Len:
		ctx.pushInstr(&MoveInstr{r, &MemExpr{r, 0}})
		ctx.pushInstr(&DeclareTypeInstr{
			r,
			&TypeExpr{frontend.BasicType{frontend.INT}}})

	default:
		panic("Unhandled unary operator")
	}
}

func (e *BinaryExpr) allocateRegisters(ctx *RegisterAllocatorContext, r *RegisterExpr) {
	var op1, op2, r2 *RegisterExpr
	dst := r
	if e.Left.Weight() > e.Right.Weight() {
		e.Left.allocateRegisters(ctx, r)
		r2 = ctx.allocateRegister()
		e.Right.allocateRegisters(ctx, r2)

		op1 = dst
		op2 = r2
	} else {
		e.Right.allocateRegisters(ctx, r)
		r2 = ctx.allocateRegister()
		e.Left.allocateRegisters(ctx, r2)

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

func (e *NewPairExpr) allocateRegisters(ctx *RegisterAllocatorContext, r *RegisterExpr) {
	helperReg := ctx.allocateRegister()
	ctx.pushInstr(&HeapAllocInstr{r, 2 * regWidth})
	e.Left.allocateRegisters(ctx, helperReg)
	ctx.pushInstr(&MoveInstr{
		&MemExpr{r, 0},
		helperReg})
	e.Right.allocateRegisters(ctx, helperReg)
	ctx.pushInstr(&MoveInstr{
		&MemExpr{r, regWidth},
		helperReg})
	ctx.freeRegister(helperReg)
}

func (e *CallExpr) allocateRegisters(ctx *RegisterAllocatorContext, r *RegisterExpr) {
	// Move arguments into r0-r4
	for n, arg := range e.Args {
		//arg := e.Args[n]
		if n < 4 {
			arg.allocateRegisters(ctx, &RegisterExpr{n})
		} else {
			freeReg := ctx.allocateRegister()
			arg.allocateRegisters(ctx, freeReg)
			ctx.pushInstr(&PushInstr{
				Op: freeReg,
			})
			ctx.freeRegister(freeReg)
		}
	}

	// Call function
	ctx.pushInstr(&CallInstr{Label: e.Label})

	// Copy result into r
	ctx.pushInstr(&MoveInstr{
		Dst: r,
		Src: &RegisterExpr{0}})

	// Get rid of arguments
	if len(e.Args) > 4 {
		freeReg := ctx.allocateRegister()
		for n := len(e.Args) - 1; n >= 4; n-- {
			ctx.pushInstr(&PopInstr{
				Op: freeReg,
			})
		}
		ctx.freeRegister(freeReg)
	}
}

//
// Instructions
//
func (i *NoOpInstr) allocateRegisters(ctx *RegisterAllocatorContext) {}

func (i *LabelInstr) allocateRegisters(ctx *RegisterAllocatorContext) {}

func (i *ReadInstr) allocateRegisters(ctx *RegisterAllocatorContext) {
	if v, ok := i.Dst.(*VarExpr); ok {
		i.Dst = ctx.lookupVariable(v)

	} else if array, ok := i.Dst.(*ArrayElemExpr); ok {
		r := ctx.allocateRegister()
		i.Dst = ctx.translateLValue(array, r)
		ctx.freeRegister(r)

	} else if pair, ok := i.Dst.(*PairElemExpr); ok {
		r := ctx.allocateRegister()
		i.Dst = ctx.translateLValue(pair, r)
		ctx.freeRegister(r)

	} //panic(fmt.Sprintf("Cannot read into %v", i.Dst.Repr()))
}

func (i *FreeInstr) allocateRegisters(ctx *RegisterAllocatorContext) {
	r := ctx.allocateRegister()
	i.Object.allocateRegisters(ctx, r)
	i.Object = r
	ctx.freeRegister(r)
}

func (i *ReturnInstr) allocateRegisters(ctx *RegisterAllocatorContext) {
	r := ctx.allocateRegister()
	i.Expr.allocateRegisters(ctx, r)
	i.Expr = r
	ctx.freeRegister(r)
}

func (i *ExitInstr) allocateRegisters(ctx *RegisterAllocatorContext) {
	r := ctx.allocateRegister()
	i.Expr.allocateRegisters(ctx, r)
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
	i.Expr.allocateRegisters(ctx, r)
	i.Expr = r

	// If the type is nil, we have an issue
	if i.Type == nil {
		panic(fmt.Sprintf("Type not defined for %T", i))
	}

	ctx.freeRegister(r)
}

func (i *MoveInstr) allocateRegisters(ctx *RegisterAllocatorContext) {
	src := ctx.allocateRegister()
	dst := ctx.allocateRegister()
	i.Src.allocateRegisters(ctx, src)
	i.Src = src
	i.Dst = ctx.translateLValue(i.Dst, dst)
	ctx.freeRegister(dst)
	ctx.freeRegister(src)
}

func (i *NotInstr) allocateRegisters(ctx *RegisterAllocatorContext) {
	r := ctx.allocateRegister()
	i.Src.allocateRegisters(ctx, r)
	i.Src = r
	i.Dst = ctx.translateLValue(i.Dst, r)
	ctx.freeRegister(r)
}

func (i *NegInstr) allocateRegisters(ctx *RegisterAllocatorContext) {
	r := ctx.allocateRegister()
	i.Expr.allocateRegisters(ctx, r)
	i.Expr = r
	ctx.freeRegister(r)
}

func (i *CmpInstr) allocateRegisters(ctx *RegisterAllocatorContext) {
	r := ctx.allocateRegister()
	r2 := ctx.allocateRegister()
	i.Left.allocateRegisters(ctx, r)
	i.Right.allocateRegisters(ctx, r2)
	i.Left = r
	i.Right = r2
	ctx.freeRegister(r)
	ctx.freeRegister(r2)
}

func (i *JmpInstr) allocateRegisters(ctx *RegisterAllocatorContext) {}

func (i *JmpCondInstr) allocateRegisters(ctx *RegisterAllocatorContext) {
	r := ctx.allocateRegister()
	i.Cond.allocateRegisters(ctx, r)
	i.Cond = r
	ctx.freeRegister(r)
}

func (i *DeclareInstr) allocateRegisters(ctx *RegisterAllocatorContext) {
	ctx.createVariable(i)
}

func (i *PushInstr) allocateRegisters(ctx *RegisterAllocatorContext) {}
func (i *PopInstr) allocateRegisters(ctx *RegisterAllocatorContext)  {}

func (*PushScopeInstr) allocateRegisters(ctx *RegisterAllocatorContext) {
	ctx.pushScope()
}

func (*PopScopeInstr) allocateRegisters(ctx *RegisterAllocatorContext) {
	ctx.popScope()
}

func (i *DeclareTypeInstr) allocateRegisters(ctx *RegisterAllocatorContext) {
	i.Dst = ctx.lookupVariable(i.Dst.(*VarExpr))
}

func (*CheckNullDereferenceInstr) allocateRegisters(ctx *RegisterAllocatorContext) {}

// Second stage IF instructions should never do anything
func (*AddInstr) allocateRegisters(ctx *RegisterAllocatorContext)       {}
func (*SubInstr) allocateRegisters(ctx *RegisterAllocatorContext)       {}
func (*MulInstr) allocateRegisters(ctx *RegisterAllocatorContext)       {}
func (*DivInstr) allocateRegisters(ctx *RegisterAllocatorContext)       {}
func (*AndInstr) allocateRegisters(ctx *RegisterAllocatorContext)       {}
func (*OrInstr) allocateRegisters(ctx *RegisterAllocatorContext)        {}
func (*CallInstr) allocateRegisters(ctx *RegisterAllocatorContext)      {}
func (*HeapAllocInstr) allocateRegisters(ctx *RegisterAllocatorContext) {}

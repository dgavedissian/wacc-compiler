package backend

import (
	"fmt"
	"strings"

	"../frontend"
)

type Optimizer interface {
	Optimize(*IFContext)
}

type fpWhileUnrollerContext struct {
	ifCtx *IFContext

	loopVariable *VarExpr

	lvStart     int
	lvIncrement int
	lvEnd       int

	loopEnd string
}

func (ctx *fpWhileUnrollerContext) conditionalIsSimple(expr Expr) (*VarExpr, int, bool) {
	switch expr := expr.(type) {
	case *UnaryExpr:
		if expr.Operator != "!" {
			return nil, 0, false
		}
		return ctx.conditionalIsSimple(expr.Operand)
	case *BinaryExpr:
		right, rightOk := expr.Right.(*IntConstExpr)
		left, leftOk := expr.Left.(*VarExpr)
		if !rightOk || !leftOk {
			return nil, 0, false
		}
		rightValue := right.Value
		if expr.Operator == "<=" {
			rightValue = rightValue + 1
		}
		return left, rightValue, expr.Operator == "<" || expr.Operator == "<="
	default:
		return nil, 0, false
	}
}

func (ctx *fpWhileUnrollerContext) checkInitialisesLoopVariable(node *InstrNode) (int, bool) {
	for node != nil {
		if instr, ok := node.Instr.(*MoveInstr); ok {
			if left, ok := instr.Dst.(*VarExpr); ok {
				if ctx.loopVariable.Name == left.Name {
					lvRight, ok := instr.Src.(*IntConstExpr)
					if !ok {
						return 0, false
					}
					return lvRight.Value, true
				}
			}
		}
		node = node.Prev
	}

	return 0, false
}

func (ctx *fpWhileUnrollerContext) checkLoopVariableIncrements(node *InstrNode) (incrementsBy int, incrementsOk bool) {
	incrementsBy, incrementsOk = 0, true

	for node != nil {
		if instr, ok := node.Instr.(*LabelInstr); ok {
			if instr.Label == ctx.loopEnd {
				return
			}
		}

		if instr, ok := node.Instr.(*MoveInstr); ok {
			if dst, ok := instr.Dst.(*VarExpr); ok && dst.Name == ctx.loopVariable.Name {
				if src, ok := instr.Src.(*BinaryExpr); ok {
					if src.Operator != "+" {
						incrementsOk = false
						return
					}

					var variable *VarExpr
					var increment *IntConstExpr
					var leftIsVariable bool
					var rightIsVariable bool

					variable, leftIsVariable = src.Left.(*VarExpr)
					if !leftIsVariable {
						variable, rightIsVariable = src.Right.(*VarExpr)
						increment, ok = src.Left.(*IntConstExpr)
						if !ok {
							incrementsOk = false
							return
						}
					} else {
						increment, ok = src.Right.(*IntConstExpr)
						if !ok {
							incrementsOk = false
							return
						}
					}

					if !leftIsVariable && !rightIsVariable {
						incrementsOk = false
						return
					}

					if variable.Name != ctx.loopVariable.Name {
						incrementsOk = false
						return
					}

					incrementsBy += increment.Value
				} else {
					incrementsOk = false
					return
				}
			}
		}
		node = node.Next
	}
	return
}

func (ctx *fpWhileUnrollerContext) optimizeLoop(node *InstrNode, whileCond *JmpCondInstr, endPoint *LabelInstr) {
	ctx.loopEnd = endPoint.Label

	var ok bool

	// firstly, check to see whether the conditional is a "simple" conditional
	if ctx.loopVariable, ctx.lvEnd, ok = ctx.conditionalIsSimple(whileCond.Cond); !ok {
		return
	}

	// now check whether or not the loop variable is initialised before the loop
	if ctx.lvStart, ok = ctx.checkInitialisesLoopVariable(node); !ok {
		return
	}

	if ctx.lvIncrement, ok = ctx.checkLoopVariableIncrements(node); !ok || ctx.lvIncrement == 0 {
		return
	}

	loopLength := (ctx.lvEnd - ctx.lvStart) / ctx.lvIncrement

	if loopLength > OPTIMISER_LOOPUNROLL_MAX {
		return
	}

	// now we unroll it!
	// first, remove the conditional jump
	node.Prev.Next, node.Next.Prev = node.Next, node.Prev
	node = node.Next

	// now remove the jump back and compile a list of instructions
	var instrList []Instr
	n := node
	for n != nil {
		if instr, ok := n.Instr.(*LabelInstr); ok {
			if instr.Label == ctx.loopEnd {
				break
			}
		}

		instrList = append(instrList, n.Instr)

		n = n.Next
	}
	instrList = instrList[1 : len(instrList)-2]
	knownLabels := make(map[string]bool)
	for _, x := range instrList {
		if labelInstr, ok := x.(*LabelInstr); ok {
			knownLabels[labelInstr.Label] = true
		}
	}

	n = n.Prev
	n.Prev.Next, n.Next.Prev = n.Next, n.Prev // omit the jump
	n = n.Prev                                // n is the popscope

	node.Next, n.Prev = n, node // remove all the instructions

	loopStart := node
	loopEnd := n

	var lastNode *InstrNode
	lastNode = loopStart
	for i := ctx.lvStart; i < ctx.lvEnd; i += ctx.lvIncrement {
		for _, instr := range instrList {
			instr = instr.Copy()
			if jmpInstr, ok := instr.(*JmpInstr); ok {
				if _, isIn := knownLabels[jmpInstr.Dst.Instr.(*LabelInstr).Label]; isIn {
					jmpInstr.Dst.Instr = &LabelInstr{
						Label: fmt.Sprintf("%s_loop_%d", jmpInstr.Dst.Instr.(*LabelInstr).Label, i),
					}
				}
			}
			if jmpCondInstr, ok := instr.(*JmpCondInstr); ok {
				if _, isIn := knownLabels[jmpCondInstr.Dst.Instr.(*LabelInstr).Label]; isIn {
					jmpCondInstr.Dst.Instr = &LabelInstr{
						Label: fmt.Sprintf("%s_loop_%d", jmpCondInstr.Dst.Instr.(*LabelInstr).Label, i),
					}
				}
			}
			if labelInstr, ok := instr.(*LabelInstr); ok {
				instr.(*LabelInstr).Label = fmt.Sprintf("%s_loop_%d", labelInstr.Label, i)
			}
			thisNode := new(InstrNode)
			thisNode.Instr = instr
			thisNode.Prev, lastNode.Next = lastNode, thisNode
			lastNode = thisNode
		}
	}
	lastNode.Next, loopEnd.Prev = loopEnd, lastNode

}

func (ctx *fpWhileUnrollerContext) optimizePath(initialInstr *InstrNode) {
	node := initialInstr
	for node != nil {
		if instr, ok := node.Instr.(*JmpCondInstr); ok {
			if endPoint, ok := instr.Dst.Instr.(*LabelInstr); ok {
				ctx.optimizeLoop(node, instr, endPoint)
			}
		}
		node = node.Next
	}
}

func (ctx *fpWhileUnrollerContext) Optimize(ifCtx *IFContext) {
	ctx.ifCtx = ifCtx

	for _, path := range ifCtx.functions {
		ctx.optimizePath(path)
	}

	ctx.optimizePath(ifCtx.main)
}

type fpInlinerContext struct {
	ifCtx             *IFContext
	replacementCode   map[string][]Instr
	functionLabels    map[string]map[string]bool
	functionArguments []fpInlinerFuncArg
	inlineCount       int
}

type fpInlinerFuncArg struct {
	Type  frontend.Type
	Name  string
	Node  *InstrNode
	extra *DeclareInstr
}

func (ctx *fpInlinerContext) exprDoesCall(expr Expr) bool {
	switch expr := expr.(type) {
	case *CallExpr:
		return true
	case *VarExpr:
		return false
	case *UnaryExpr:
		return ctx.exprDoesCall(expr.Operand)
	case *BinaryExpr:
		return ctx.exprDoesCall(expr.Left) || ctx.exprDoesCall(expr.Right)
	case *NewPairExpr:
		return ctx.exprDoesCall(expr.Left) || ctx.exprDoesCall(expr.Right)
	case *NewStructExpr:
		doesCall := true
		for _, e := range expr.Args {
			doesCall = doesCall || ctx.exprDoesCall(e)
		}
		return doesCall
	case *PairElemExpr:
		return ctx.exprDoesCall(expr.Operand)
	case *CharConstExpr, *StringConstExpr, *ArrayConstExpr, *IntConstExpr, *BoolConstExpr, *PointerConstExpr:
		return false
	case *RegisterExpr, *StackArgumentExpr, *StackLocationExpr:
		return false
	default:
		panic(fmt.Sprintf("Can't work out whether this calls: %T", expr))
	}
}

func (ctx *fpInlinerContext) moveDoesCall(instr *MoveInstr) bool {
	return ctx.exprDoesCall(instr.Src)
}

func (ctx *fpInlinerContext) checkInlinable(initNode *InstrNode) {
	funcName := initNode.Instr.(*LabelInstr).Label
	ctx.functionLabels[funcName] = make(map[string]bool)
	ctx.functionLabels[funcName][fmt.Sprintf("_%s_end", funcName)] = true

	nodeCount := 0
	node := initNode
	var firstNode *InstrNode
	stillLookingForArguments := true
	var buildingInstr *fpInlinerFuncArg
	var extraDeclarations []*DeclareInstr
	for node != nil {
		if stillLookingForArguments {
			switch instr := node.Instr.(type) {
			case *DeclareInstr:
				if buildingInstr != nil {
					extraDeclarations = append(extraDeclarations, buildingInstr.Node.Instr.(*DeclareInstr))
				}
				buildingInstr = &fpInlinerFuncArg{
					Type: instr.Type,
					Name: instr.Var.Name,
					Node: node,
				}
			case *MoveInstr:
				if ctx.moveDoesCall(instr) {
					// abort!
					return
				}

				if registerExpr, ok := instr.Src.(*RegisterExpr); ok {
					if registerExpr.Id < 4 {
						ctx.functionArguments = append(ctx.functionArguments, *buildingInstr)
						buildingInstr = nil
					} else {
						firstNode = buildingInstr.Node
						stillLookingForArguments = false
					}
				} else if _, ok := instr.Src.(*StackArgumentExpr); ok {

				} else {
					firstNode = buildingInstr.Node
					stillLookingForArguments = false
				}
			case *LabelInstr, *PushScopeInstr:
				// noop
			default:
				firstNode = node
				stillLookingForArguments = false
			}
		}
		if !stillLookingForArguments {
			switch instr := node.Instr.(type) {
			case *LabelInstr:
				ctx.functionLabels[funcName][instr.Label] = true
			}
		}

		nodeCount += 1
		node = node.Next
	}

	if nodeCount > OPTIMISER_INLINER_MAX {
		return
	}

	instrList := make([]Instr, 0)
	node = initNode
	seenFirstNode := false
	argCount := 0
	for node != nil {
		instr := node.Instr.Copy()
		if node == firstNode {
			seenFirstNode = true
		}
		if !seenFirstNode {
			if moveInstr, ok := instr.(*MoveInstr); ok {
				moveInstr.Src = &VarExpr{fmt.Sprintf("_%s_arg_%d", funcName, argCount)}
				argCount += 1
			}
		}

		if returnInstr, ok := instr.(*ReturnInstr); ok {
			// replace!
			instr = &MoveInstr{
				Dst: &VarExpr{fmt.Sprintf("_%s_retVal", funcName)},
				Src: returnInstr.Expr,
			}
			instrList = append(instrList, instr)

			instr = &JmpInstr{
				Dst: &InstrNode{
					Instr: &LabelInstr{fmt.Sprintf("_%s_end", funcName)},
				},
			}
		}

		instrList = append(instrList, instr)
		node = node.Next
	}

	instrList = instrList[1:] // chop off function name

	// walk backwards to place the end label
	endPoint := len(instrList) - 1
	endLabel := &LabelInstr{
		Label: fmt.Sprintf("_%s_end", funcName),
	}
	lastInstr := instrList[endPoint]
	instrList = append(instrList[:endPoint], endLabel, lastInstr)

	ctx.replacementCode[funcName] = instrList
}

func (ctx *fpInlinerContext) fixLabels(funcName string, prefix string, instr Instr) Instr {
	switch instr := instr.(type) {
	case *LabelInstr:
		instr.Label = prefix + instr.Label
	case *MoveInstr:
		instr.Src = ctx.fixLabelsExpr(funcName, prefix, instr.Src)
		instr.Dst = ctx.fixLabelsExpr(funcName, prefix, instr.Dst)
	case *JmpInstr:
		lbl := instr.Dst.Instr.(*LabelInstr).Label
		if _, ok := ctx.ifCtx.functions[lbl]; !ok {
			instr.Dst = &InstrNode{
				Instr: &LabelInstr{
					Label: prefix + lbl,
				},
			}
		}
	case *JmpCondInstr:
		lbl := instr.Dst.Instr.(*LabelInstr).Label
		if _, ok := ctx.ifCtx.functions[lbl]; !ok {
			instr.Dst = &InstrNode{
				Instr: &LabelInstr{
					Label: prefix + lbl,
				},
			}
		}
		instr.Cond = ctx.fixLabelsExpr(funcName, prefix, instr.Cond)
	case *PrintInstr:
		instr.Expr = ctx.fixLabelsExpr(funcName, prefix, instr.Expr)
	case *ReadInstr:
		instr.Dst = ctx.fixLabelsExpr(funcName, prefix, instr.Dst)
	case *PushScopeInstr, *PopScopeInstr, *NoOpInstr:
	case *FreeInstr:
		instr.Object = ctx.fixLabelsExpr(funcName, prefix, instr.Object)
	case *DeclareInstr:
		instr.Var.Name = prefix + instr.Var.Name
	default:
		panic(fmt.Sprintf("Unrecognized widget %#v", instr))
	}

	return instr
}

func (ctx *fpInlinerContext) fixLabelsExpr(funcName string, prefix string, expr Expr) Expr {
	switch expr := expr.(type) {
	case *CallExpr:
		if _, ok := ctx.ifCtx.functions[expr.Label.Label]; !strings.HasPrefix(expr.Label.Label, fmt.Sprintf("_%s_", funcName)) && !ok {
			expr.Label.Label = prefix + expr.Label.Label
		}
		newArgs := make([]Expr, len(expr.Args))
		for i, arg := range expr.Args {
			newArgs[i] = ctx.fixLabelsExpr(funcName, prefix, arg)
		}
	case *LocationExpr:
		if !strings.HasPrefix(expr.Label, fmt.Sprintf("_%s_", funcName)) {
			expr.Label = prefix + expr.Label
		}
	case *VarExpr:
		if !strings.HasPrefix(expr.Name, fmt.Sprintf("_%s_", funcName)) {
			expr.Name = prefix + expr.Name
		}
	case *UnaryExpr:
		expr.Operand = ctx.fixLabelsExpr(funcName, prefix, expr.Operand)
	case *BinaryExpr:
		expr.Left = ctx.fixLabelsExpr(funcName, prefix, expr.Left)
		expr.Right = ctx.fixLabelsExpr(funcName, prefix, expr.Right)
	case *NewPairExpr:
		expr.Left = ctx.fixLabelsExpr(funcName, prefix, expr.Left)
		expr.Right = ctx.fixLabelsExpr(funcName, prefix, expr.Right)
	case *PairElemExpr:
		expr.Operand = ctx.fixLabelsExpr(funcName, prefix, expr.Operand).(*VarExpr)
	case *ArrayElemExpr:
		expr.Array = ctx.fixLabelsExpr(funcName, prefix, expr.Array)
		expr.Index = ctx.fixLabelsExpr(funcName, prefix, expr.Index)
	case *CharConstExpr, *StringConstExpr, *ArrayConstExpr, *IntConstExpr, *BoolConstExpr, *PointerConstExpr:
	case *RegisterExpr, *StackArgumentExpr, *StackLocationExpr:
	default:
		panic(fmt.Sprintf("Unrecognized exprwidget %#v", expr))
	}

	return expr
}

func (ctx *fpInlinerContext) inlineInPath(node *InstrNode) {
	pushScopeStack := make([]*PushScopeInstr, 0)

	for node != nil {
		if pushScopeInstr, ok := node.Instr.(*PushScopeInstr); ok {
			pushScopeStack = append(pushScopeStack, pushScopeInstr)
		}
		if popScopeInstr, ok := node.Instr.(*PopScopeInstr); ok {
			popScopeInstr.StackSize = pushScopeStack[len(pushScopeStack)-1].StackSize
			pushScopeStack = pushScopeStack[:len(pushScopeStack)-1]
		}

		if instr, ok := node.Instr.(*MoveInstr); ok {
			if callExpr, ok := instr.Src.(*CallExpr); ok {
				cnt := ctx.inlineCount
				ctx.inlineCount += 1

				replacementCode, ok := ctx.replacementCode[callExpr.Label.Label]
				if !ok {
					node = node.Next
					continue
				}

				backNode := node.Prev

				// TODO: unbreak usage of stack...
				pushScopeStack[len(pushScopeStack)-1].StackSize += 48 * regWidth

				for argNum, argExpr := range callExpr.Args {
					varExpr := &VarExpr{fmt.Sprintf("_%s_arg_%d", callExpr.Label.Label, argNum)}
					newNode := &InstrNode{
						Instr: &DeclareInstr{
							Var:  varExpr,
							Type: ctx.functionArguments[argNum].Type,
						},
					}
					pushScopeStack[len(pushScopeStack)-1].StackSize += regWidth

					backNode.Next, newNode.Prev = newNode, backNode
					newNode.Next, node.Prev = node, newNode
					backNode = newNode

					newNode = &InstrNode{
						Instr: &MoveInstr{
							Dst: varExpr,
							Src: argExpr,
						},
					}

					backNode.Next, newNode.Prev = newNode, backNode
					newNode.Next, node.Prev = node, newNode
					backNode = newNode
				}

				resultVar := instr.Dst
				for _, instr := range replacementCode {
					instr = ctx.fixLabels(callExpr.Label.Label, fmt.Sprintf("_%s_ins%d_", callExpr.Label.Label, cnt), instr.Copy())

					if moveInstr, ok := instr.(*MoveInstr); ok {
						if varExpr, ok := moveInstr.Dst.(*VarExpr); ok && varExpr.Name == fmt.Sprintf("_%s_retVal", callExpr.Label.Label) {
							// replace this with the output
							moveInstr.Dst = resultVar
						}
					}

					newNode := &InstrNode{
						Instr: instr,
					}
					backNode.Next, newNode.Prev = newNode, backNode
					newNode.Next, node.Prev = node, newNode
					backNode = newNode
				}

				// remove old call node
				node.Prev.Next, node.Next.Prev = node.Next, node.Prev
			}
		}
		node = node.Next
	}
}

func (ctx *fpInlinerContext) Optimize(ifCtx *IFContext) {
	ctx.ifCtx = ifCtx
	ctx.replacementCode = make(map[string][]Instr)
	ctx.functionLabels = make(map[string]map[string]bool)

	for _, path := range ifCtx.functions {
		ctx.checkInlinable(path)
	}

	for _, path := range ifCtx.functions {
		ctx.inlineInPath(path)
	}
	ctx.inlineInPath(ifCtx.main)
}

func OptimiseFirstPassIF(ifCtx *IFContext) {
	new(fpWhileUnrollerContext).Optimize(ifCtx)
	new(fpInlinerContext).Optimize(ifCtx)
}

func OptimiseSecondPassIF(ifCtx *IFContext) {
}

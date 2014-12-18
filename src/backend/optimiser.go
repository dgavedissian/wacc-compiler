package backend

import (
	"fmt"
	"log"
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
	log.Printf("%#v", expr)

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
	log.Printf("ATTEMPTING TO OPTIMIZE %#v until %#v", whileCond, endPoint)

	ctx.loopEnd = endPoint.Label

	var ok bool

	// firstly, check to see whether the conditional is a "simple" conditional
	if ctx.loopVariable, ctx.lvEnd, ok = ctx.conditionalIsSimple(whileCond.Cond); !ok {
		log.Println("Conditional decreed not-simple.")
		return
	}

	// now check whether or not the loop variable is initialised before the loop
	if ctx.lvStart, ok = ctx.checkInitialisesLoopVariable(node); !ok {
		log.Println("Loop variable not inited to constant.")
		return
	}

	if ctx.lvIncrement, ok = ctx.checkLoopVariableIncrements(node); !ok || ctx.lvIncrement == 0 {
		log.Println("Loop variable not solely incremented.")
		return
	}

	log.Println("Loop variable increments by", ctx.lvIncrement)
	loopLength := (ctx.lvEnd - ctx.lvStart) / ctx.lvIncrement

	if loopLength > OPTIMISER_LOOPUNROLL_MAX {
		log.Println("Loop is %d long - too long to unroll.", loopLength)
		return
	}

	// now we unroll it!
	// first, remove the conditional jump
	node.Prev.Next, node.Next.Prev = node.Next, node.Prev
	node = node.Next
	log.Printf("%#v", node)

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

	log.Printf("%#v", n) // n is the label
	n = n.Prev
	log.Printf("%#v", n) // n is the jump
	log.Println(knownLabels)

	n.Prev.Next, n.Next.Prev = n.Next, n.Prev // omit the jump
	n = n.Prev                                // n is the popscope

	node.Next, n.Prev = n, node // remove all the instructions

	loopStart := node
	loopEnd := n

	var lastNode *InstrNode
	lastNode = loopStart
	for i := ctx.lvStart; i < ctx.lvEnd; i += ctx.lvIncrement {
		for _, instr := range instrList {
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
				instr = new(LabelInstr)
				instr.(*LabelInstr).Label = fmt.Sprintf("%s_loop_%d", labelInstr.Label, i)
			}
			thisNode := new(InstrNode)
			thisNode.Instr = instr
			thisNode.Prev, lastNode.Next = lastNode, thisNode
			lastNode = thisNode
		}
	}
	lastNode.Next, loopEnd.Prev = loopEnd, lastNode

	log.Println(instrList)

}

func (ctx *fpWhileUnrollerContext) optimizePath(initialInstr *InstrNode) {
	log.Println("LOOKING FOR WHILES")

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
	log.Println("WHILE UNROLLING START")
	ctx.ifCtx = ifCtx

	for _, path := range ifCtx.functions {
		ctx.optimizePath(path)
	}

	ctx.optimizePath(ifCtx.main)
}

func OptimiseFirstPassIF(ifCtx *IFContext) {
	new(fpWhileUnrollerContext).Optimize(ifCtx)
}

func OptimiseSecondPassIF(ifCtx *IFContext) {
}

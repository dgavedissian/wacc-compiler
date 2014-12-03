package backend

import "fmt"

type GeneratorContext struct {
	out     string
	inLabel bool
}

func (ctx *GeneratorContext) pushCode(s string) {
	if ctx.inLabel {
		ctx.out += "    "
	}
	ctx.out += s + "\n"
}

func (ctx *GeneratorContext) generateMovImm(value int, dst string) string {
	// MOV only supports immediate values which are 1 byte. Constants which
	// are larger require LDR (which is unfortunately slower)
	if value > 255 {
		return fmt.Sprintf("ldr %v, =%v", dst, value)
	} else {
		return fmt.Sprintf("mov %v, #%v", dst, value)
	}
}

func (ctx *GeneratorContext) generateInstr(instr Instr) {
	switch instr := instr.(type) {
	case *LabelInstr:
		ctx.pushCode(instr.Label + ":")
		ctx.inLabel = true

	case *MoveInstr:
		dst := "r0"
		if expr, ok := instr.Src.(*IntConstExpr); ok {
			ctx.pushCode(ctx.generateMovImm(expr.Value, dst))
		}

	case *ExitInstr:
		ctx.pushCode(ctx.generateMovImm(instr.Expr.(*IntConstExpr).Value, "r0"))
		ctx.pushCode("bl exit")

	default:
		panic(fmt.Sprintf("Unhandled instruction: %T", instr))
	}
}

func VisitInstructions(ifCtx *IFContext, f func(Instr)) {
	// Start at the first node after the label
	node := ifCtx.first.Next
	for node != nil {
		f(node.Instr)
		node = node.Next
	}
}

func GenerateCode(ifCtx *IFContext) string {
	ctx := &GeneratorContext{"", false}

	// Generate header
	ctx.out += ".text\n"

	// Add the label of each function to the global list
	ctx.out += ".global main\n"

	// Generate program code
	// TODO: For each function
	ctx.generateInstr(ifCtx.first.Instr)
	ctx.pushCode("push {lr}")
	VisitInstructions(ifCtx, func(i Instr) {
		ctx.generateInstr(i)
	})
	ctx.pushCode("pop {pc}")

	return ctx.out
}

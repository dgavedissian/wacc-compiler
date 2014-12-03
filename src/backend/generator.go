package backend

import "fmt"

type GeneratorContext struct {
	out     string
	inLabel bool
}

func (ctx *GeneratorContext) pushCode(s string) {
	if ctx.inLabel {
		ctx.out += "  "
	}
	ctx.out += s + "\n"
}

func (ctx *GeneratorContext) generateInstr(instr Instr) {
	switch instr := instr.(type) {
	case *LabelInstr:
		ctx.pushCode(instr.Label + ":")
		ctx.inLabel = true

	case *MoveInstr:
		dst := "r0"
		src := "r0"
		ctx.pushCode(fmt.Sprintf("mov %v, %v", dst, src))

	default:
		panic(fmt.Sprintf("Unhandled instruction: %T", instr))
	}
}

func VisitInstructions(ifCtx *IFContext, f func(Instr)) {
	node := ifCtx.first
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
	VisitInstructions(ifCtx, func(i Instr) {
		ctx.generateInstr(i)
	})

	return ctx.out
}

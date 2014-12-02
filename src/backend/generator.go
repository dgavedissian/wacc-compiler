package backend

import "fmt"

type GeneratorContext struct {
	scope int
}

func (ctx GeneratorContext) generateInstr(instr Instr) string {
	switch instr := instr.(type) {
	case *LabelInstr:
		return instr.Label + ":\n"

	case *NoOpInstr:
		return ""

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

func GenerateCode(iform *IFContext) string {
	ctx := &GeneratorContext{0}
	out := ""

	// Generate header

	// Perform a pass through the code to find global labels
	VisitInstructions(iform, func(i Instr) {
		fmt.Printf("Found %v\n", i.Repr())
	})

	// Generate program code
	VisitInstructions(iform, func(i Instr) {
		out += ctx.generateInstr(i)
	})

	return out
}

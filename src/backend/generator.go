package backend

import (
	"fmt"
	"strconv"
)

type GeneratorContext struct {
	out     string
	inLabel bool
}

func (ctx *GeneratorContext) pushCode(s string) {
	if ctx.inLabel {
		ctx.out += "\t"
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
		switch src := instr.Src.(type) {
		case *IntConstExpr:
			ctx.pushCode(ctx.generateMovImm(src.Value, dst))

		case *CharConstExpr:
			ctx.pushCode(ctx.generateMovImm(int(src.Value), dst))

		case *LocationExpr:
			ctx.pushCode(fmt.Sprintf("ldr %v, =%v", dst, src.Label))

		default:
			panic(fmt.Sprintf("Unimplemented MoveInstr for src %T", src))
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

	// Data section
	ctx.out += ".data\n"

	// Search for any string array literals and replace them with labels in the
	// data section
	stringCounter := 0
	VisitInstructions(ifCtx, func(i Instr) {
		switch instr := i.(type) {
		case *MoveInstr:
			// Is the source an array of ascii chars?
			if expr, ok := instr.Src.(*ArrayExpr); ok {
				if elem, ok := expr.Elems[0].(*CharConstExpr); ok {
					if elem.Size == 1 {
						// Build a string from the char array
						str := ""
						for _, e := range expr.Elems {
							str += string(e.(*CharConstExpr).Value)
						}

						// Generate a label
						label := fmt.Sprintf("stringlit%v", stringCounter)
						stringCounter += 1

						// Record the string in the data section
						ctx.out += fmt.Sprintf("%v:\n", label)
						ctx.out += fmt.Sprintf("\t.word %v\n", len(str))
						ctx.out += fmt.Sprintf("\t.ascii %v\n", strconv.QuoteToASCII(str))

						// Replace src with a label
						instr.Src = &LocationExpr{label}
					}
				}
			}
		}
	})

	// Program code section
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

package backend

import "fmt"

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

func (ctx *GeneratorContext) generateExpr(expr IFExpr) int {
	switch expr := expr.(type) {
	case *BinOpExpr:
		left := ctx.generateExpr(expr.Left)
		right := ctx.generateExpr(expr.Right)
		result := 2
		ctx.pushCode(fmt.Sprintf("add r%v, r%v, r%v", result, left, right))
		return result

	case *IntConstExpr:
		result := 2
		ctx.pushCode(fmt.Sprintf("ldr r%v, =%v", result, expr.Value))
		return result

	case *RegisterExpr:
		return expr.Id

	default:
		panic(fmt.Sprintf("Unhandled Expr: %T", expr))
	}
}

func (ctx *GeneratorContext) generateInstr(instr Instr) {
	switch instr := instr.(type) {
	case *LabelInstr:
		ctx.pushCode(instr.Label + ":")
		ctx.inLabel = true

		// Read

		// Free

	case *ExitInstr:
		exitCode := instr.Expr.(*IntConstExpr).Value
		ctx.pushCode(fmt.Sprintf("ldr r0, =%v", exitCode))
		ctx.pushCode("bl exit")

	case *PrintInstr:
		// save regs r0 and r1

		//
		switch obj := instr.Expr.(type) {
		case *IntConstExpr:
			value := instr.Expr.(*IntConstExpr).Value
			ctx.pushCode("ldr r0, =printf_fmt_int")
			ctx.pushCode(fmt.Sprintf("ldr r1, =%v", value))
			ctx.pushCode("bl printf")

		case *CharConstExpr:
			value := int(instr.Expr.(*CharConstExpr).Value)
			ctx.pushCode("ldr r0, =printf_fmt_char")
			ctx.pushCode(fmt.Sprintf("ldr r1, =%v", value))
			ctx.pushCode("bl printf")

		case *LocationExpr:
			ctx.pushCode("ldr r0, =printf_fmt_str")
			ctx.pushCode(fmt.Sprintf("ldr r1, =%v", obj.Label))
			ctx.pushCode("bl printf")

		default:
			result := ctx.generateExpr(obj)
			ctx.pushCode("ldr r0, =printf_fmt_int")
			ctx.pushCode(fmt.Sprintf("mov r1, r%v", result))
			ctx.pushCode("bl printf")
		}

		// Flush output stream
		ctx.pushCode("mov r0, #0")
		ctx.pushCode("bl fflush")

		// load regs r0 and r1

	case *MoveInstr:
		dst := instr.Dst.(*RegisterExpr).Repr()
		switch src := instr.Src.(type) {
		case *IntConstExpr:
			ctx.pushCode(fmt.Sprintf("ldr %v, =%v", dst, src.Value))

		case *CharConstExpr:
			ctx.pushCode(fmt.Sprintf("ldr %v, =%v", dst, int(src.Value)))

		case *LocationExpr:
			ctx.pushCode(fmt.Sprintf("ldr %v, =%v", dst, src.Label))

		default:
			panic(fmt.Sprintf("Unimplemented MoveInstr for src %T", src))
		}

		// Test

		// Jmp

		// JmpZero

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

	// Printf format strings
	ctx.out += "printf_fmt_int:\n\t.asciz \"%d\"\n"
	ctx.out += "printf_fmt_char:\n\t.asciz \"%c\"\n"
	ctx.out += "printf_fmt_str:\n\t.asciz \"%s\"\n"

	// Search for any string array literals and replace them with labels in the
	// data section
	stringCounter := 0
	tryReplaceString := func(expr IFExpr) IFExpr {
		// Is the source an array of ascii chars?
		if expr, ok := expr.(*ArrayExpr); ok {
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
					ctx.out += fmt.Sprintf("\t.asciz \"%v\"\n", str)

					// Replace src with a label
					return &LocationExpr{label}
				}
			}
		}
		return expr
	}

	VisitInstructions(ifCtx, func(i Instr) {
		switch instr := i.(type) {
		case *MoveInstr:
			instr.Src = tryReplaceString(instr.Src)

		case *PrintInstr:
			instr.Expr = tryReplaceString(instr.Expr)
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

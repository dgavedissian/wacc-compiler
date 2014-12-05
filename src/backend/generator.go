package backend

import "fmt"

type GeneratorContext struct {
	stringCounter int
	data          string
	text          string
}

// Search for any string array literals and replace them with labels in the
// data section
func (ctx *GeneratorContext) handleString(expr Expr) Expr {
	// Is the expr an array
	if expr, ok := expr.(*ArrayExpr); ok {
		// Is the source an array of ascii chars?
		if elem, ok := expr.Elems[0].(*CharConstExpr); ok {
			if elem.Size == 1 {
				// Build a string from the char array
				str := ""
				for _, e := range expr.Elems {
					str += string(e.(*CharConstExpr).Value)
				}

				// Generate a label
				label := fmt.Sprintf("stringlit%v", ctx.stringCounter)
				ctx.stringCounter += 1

				// Record the string in the data section
				ctx.data += fmt.Sprintf("%v:\n", label)
				ctx.data += fmt.Sprintf("\t.asciz \"%v\"\n", str)

				// Replace src with a label
				return &LocationExpr{label}
			}
		}
	}
	return expr
}

func (ctx *GeneratorContext) pushLabel(label string) {
	ctx.text += fmt.Sprintf("%v:\n", label)
}

func (ctx *GeneratorContext) pushCode(s string, a ...interface{}) {
	ctx.text += "\t" + fmt.Sprintf(s, a...) + "\n"
}

func (e *IntConstExpr) generateCode(*GeneratorContext) int  { return 0 }
func (e *CharConstExpr) generateCode(*GeneratorContext) int { return 0 }
func (e *ArrayExpr) generateCode(*GeneratorContext) int     { return 0 }
func (e *LocationExpr) generateCode(*GeneratorContext) int  { return 0 }
func (e *VarExpr) generateCode(*GeneratorContext) int {
	panic("Trying to generate code for a variable")
	return 0
}
func (e *RegisterExpr) generateCode(*GeneratorContext) int { return 0 }
func (e *BinOpExpr) generateCode(*GeneratorContext) int    { return 0 }
func (e *NotExpr) generateCode(*GeneratorContext) int      { return 0 }
func (e *OrdExpr) generateCode(*GeneratorContext) int      { return 0 }
func (e *ChrExpr) generateCode(*GeneratorContext) int      { return 0 }
func (e *NegExpr) generateCode(*GeneratorContext) int      { return 0 }
func (e *LenExpr) generateCode(*GeneratorContext) int      { return 0 }
func (e *NewPairExpr) generateCode(*GeneratorContext) int  { return 0 }
func (e *CallExpr) generateCode(*GeneratorContext) int     { return 0 }

func (i *NoOpInstr) generateCode(*GeneratorContext) {}
func (i *LabelInstr) generateCode(ctx *GeneratorContext) {
	ctx.pushLabel(i.Label)
}

func (i *ReadInstr) generateCode(*GeneratorContext) {}
func (i *FreeInstr) generateCode(*GeneratorContext) {}
func (i *ExitInstr) generateCode(ctx *GeneratorContext) {
	ctx.pushCode("mov r0, %v", i.Expr.Repr())
	ctx.pushCode("bl exit")
}

func (i *PrintInstr) generateCode(ctx *GeneratorContext) {
	// save regs r0 and r1

	i.Expr = ctx.handleString(i.Expr)

	// Printf depending on type
	switch obj := i.Expr.(type) {
	case *IntConstExpr:
		value := i.Expr.(*IntConstExpr).Value
		ctx.pushCode("ldr r1, =%v", value)
		ctx.pushCode("ldr r0, =printf_fmt_int")
		ctx.pushCode("bl printf")

	case *CharConstExpr:
		value := int(i.Expr.(*CharConstExpr).Value)
		ctx.pushCode("ldr r1, =%v", value)
		ctx.pushCode("ldr r0, =printf_fmt_char")
		ctx.pushCode("bl printf")

	case *LocationExpr:
		ctx.pushCode("ldr r1, =%v", obj.Label)
		ctx.pushCode("ldr r0, =printf_fmt_str")
		ctx.pushCode("bl printf")

	case *RegisterExpr:
		ctx.pushCode("mov r1, %v", obj.Repr())
		ctx.pushCode("ldr r0, =printf_fmt_int")
		ctx.pushCode("bl printf")

	default:
		panic(fmt.Sprintf("Cannot print an object of type %T", obj))
	}

	// Flush output stream
	ctx.pushCode("mov r0, #0")
	ctx.pushCode("bl fflush")

	// load regs r0 and r1
}

func (i *MoveInstr) generateCode(ctx *GeneratorContext) {
	dst := i.Dst.(*RegisterExpr).Repr()
	i.Src = ctx.handleString(i.Src)

	// Optimisation step: If we're moving from a constant, just load
	switch src := i.Src.(type) {
	case *IntConstExpr:
		ctx.pushCode("ldr %v, =%v", dst, src.Value)

	case *CharConstExpr:
		ctx.pushCode("ldr %v, =%v", dst, int(src.Value))

	case *LocationExpr:
		ctx.pushCode("ldr %v, =%v", dst, src.Label)

	case *RegisterExpr:
		ctx.pushCode("mov %v, %v", dst, src.Repr())

	default:
		panic(fmt.Sprintf("Unhandled src type of mov %T", src))
	}
}

func (i *TestInstr) generateCode(*GeneratorContext)    {}
func (i *JmpInstr) generateCode(*GeneratorContext)     {}
func (i *JmpZeroInstr) generateCode(*GeneratorContext) {}

func (i *AddInstr) generateCode(ctx *GeneratorContext) {
	ctx.pushCode("add %v, %v, %v", i.Dst.Repr(), i.Op1.Repr(), i.Op2.Repr())
}

func (i *PushScopeInstr) generateCode(ctx *GeneratorContext) {}
func (i *PopScopeInstr) generateCode(ctx *GeneratorContext)  {}

func VisitInstructions(ifCtx *IFContext, f func(Instr)) {
	// Start at the first node after the label
	node := ifCtx.first.Next
	for node != nil {
		f(node.Instr)
		node = node.Next
	}
}

func GenerateCode(ifCtx *IFContext) string {
	ctx := new(GeneratorContext)

	// Printf format strings
	ctx.data += "printf_fmt_int:\n\t.asciz \"%d\"\n"
	ctx.data += "printf_fmt_char:\n\t.asciz \"%c\"\n"
	ctx.data += "printf_fmt_str:\n\t.asciz \"%s\"\n"
	ctx.data += "printf_fmt_addr:\n\t.asciz \"%p\"\n"

	// Add the label of each function to the global list
	ctx.text += ".global main\n"

	// Generate program code
	// TODO: For each function
	ifCtx.first.Instr.generateCode(ctx)
	ctx.pushCode("push {lr}")
	VisitInstructions(ifCtx, func(i Instr) {
		i.generateCode(ctx)
	})
	ctx.pushCode("pop {pc}")

	// Combine data and text sections
	return ".data\n" + ctx.data + ".text\n" + ctx.text
}

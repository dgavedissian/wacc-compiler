package backend

import (
	"fmt"
	"log"
)

type GeneratorContext struct {
	stringCounter    int
	data             string
	text             string
	registerContents []map[string]Expr
	dataContents     map[string]Expr
	funcStack        []string
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

func (i *NoOpInstr) generateCode(*GeneratorContext) {}
func (i *LabelInstr) generateCode(ctx *GeneratorContext) {
	ctx.pushLabel(i.Label)
}

func (i *ReadInstr) generateCode(ctx *GeneratorContext) {
	// save regs r0 and r1
	ctx.pushCode("push {r0,r1}")

	log.Printf("%#v\n", i.Dst)

	ctx.pushCode("add r1, sp, #0")

	// Printf depending on type
	printfType := getPrintfTypeForExpr(ctx, i.Dst)
	if printfType == "__BOOL__" {
		panic("???")
	} else {
		ctx.pushCode("ldr r0, =%s", printfType)
		ctx.pushCode("bl scanf")
	}

	// Move output to destination
	ctx.pushCode("ldr %s, [sp]", i.Dst.Repr())

	// load regs r0 and r1
	ctx.pushCode("pop {r0,r1}")
}
func (i *FreeInstr) generateCode(*GeneratorContext) {}

func (i *ReturnInstr) generateCode(ctx *GeneratorContext) {
	ctx.pushCode("mov r0, %v", i.Expr.Repr())
	ctx.pushCode("b _" + ctx.funcStack[0] + "_end")
}

func (i *ExitInstr) generateCode(ctx *GeneratorContext) {
	ctx.pushCode("mov r0, %v", i.Expr.Repr())
	ctx.pushCode("bl exit")
}

func getPrintfTypeForExpr(ctx *GeneratorContext, expr Expr) string {
	switch obj := expr.(type) {
	case *IntConstExpr:
		return "printf_fmt_int"
	case *CharConstExpr:
		return "printf_fmt_char"
	case *ArrayExpr:
		return "printf_fmt_str"
	case *BoolConstExpr:
		return "__BOOL__"
	case *RegisterExpr:
		log.Println("REG", obj.Repr())
		return getPrintfTypeForExpr(ctx, ctx.registerContents[0][obj.Repr()])
	case *LocationExpr:
		return getPrintfTypeForExpr(ctx, ctx.dataContents[obj.Label])
	default:
		panic(fmt.Sprintf("Attempted to get printf type for unknown thing %#v", expr))
	}
}

func (i *PrintInstr) generateCode(ctx *GeneratorContext) {
	// save regs r0 and r1
	ctx.pushCode("push {r0,r1}")

	// Printf depending on type
	switch obj := i.Expr.(type) {
	case *IntConstExpr:
		value := i.Expr.(*IntConstExpr).Value
		ctx.pushCode("ldr r1, =%v", value)

	case *BoolConstExpr:
		value := i.Expr.(*BoolConstExpr).Value
		ctx.pushCode("ldr r1, =%d", value)

	case *CharConstExpr:
		value := int(i.Expr.(*CharConstExpr).Value)
		ctx.pushCode("ldr r1, =%v", value)

	case *LocationExpr:
		ctx.pushCode("ldr r1, =%v", obj.Label)

	case *RegisterExpr:
		ctx.pushCode("mov r1, %v", obj.Repr())

	default:
		panic(fmt.Sprintf("Cannot print an object of type %T", obj))
	}

	printfType := getPrintfTypeForExpr(ctx, i.Expr)
	if printfType == "__BOOL__" {
		ctx.pushCode("mov r0, r1")
		ctx.pushCode("bl _wacc_printBool")
	} else {
		ctx.pushCode("ldr r0, =%s", printfType)
		ctx.pushCode("bl printf")
	}

	// Flush output stream
	ctx.pushCode("mov r0, #0")
	ctx.pushCode("bl fflush")

	// load regs r0 and r1
	ctx.pushCode("pop {r0,r1}")
}

func (i *MoveInstr) generateCode(ctx *GeneratorContext) {
	dst := i.Dst.(*RegisterExpr).Repr()

	// Optimisation step: If we're moving from a constant, just load
	switch src := i.Src.(type) {
	case *IntConstExpr:
		ctx.pushCode("ldr %v, =%v", dst, src.Value)
		ctx.registerContents[0][dst] = src

	case *BoolConstExpr:
		n := 0
		if src.Value {
			n = 1
		}
		ctx.pushCode("ldr %v, =%v", dst, n)
		ctx.registerContents[0][dst] = src

	case *CharConstExpr:
		ctx.pushCode("ldr %v, =%v", dst, int(src.Value))
		ctx.registerContents[0][dst] = src

	case *LocationExpr:
		ctx.pushCode("ldr %v, =%v", dst, src.Label)
		ctx.registerContents[0][dst] = ctx.dataContents[src.Label]

	case *RegisterExpr:
		ctx.pushCode("mov %v, %v", dst, src.Repr())
		ctx.registerContents[0][dst] = ctx.registerContents[0][src.Repr()]

	default:
		panic(fmt.Sprintf("Unhandled src type of mov %T", src))
	}
}

func (i *NotInstr) generateCode(ctx *GeneratorContext) {
	dst := i.Dst.(*RegisterExpr).Repr()
	src := i.Src.(*RegisterExpr).Repr()

	ctx.pushCode("mvn %v, %v", dst, src)
}

func (i *CmpInstr) generateCode(ctx *GeneratorContext) {
	cc := "al"
	switch i.Operator {
	case EQ:
		cc = "eq"
	case NE:
		cc = "ne"
	case LT:
		cc = "lt"
	case GT:
		cc = "gt"
	case LE:
		cc = "le"
	case GE:
		cc = "ge"
	}
	ctx.pushCode("cmp %v, %v", i.Left.Repr(), i.Right.Repr())
	ctx.pushCode("mov %v, #0", i.Dst.Repr())
	ctx.pushCode("mov%s %v, #1", cc, i.Dst.Repr())
}

func (i *JmpInstr) generateCode(ctx *GeneratorContext) {
	ctx.pushCode("b %v", i.Dst.Instr.(*LabelInstr).Label)
}

func (i *JmpCondInstr) generateCode(ctx *GeneratorContext) {
	if _, ok := i.Cond.(*RegisterExpr); !ok {
		panic("condition is not a register, abort")
	}
	ctx.pushCode("cmp %v, #0", i.Cond.Repr())
	ctx.pushCode("bne %v", i.Dst.Instr.(*LabelInstr).Label)
}

func (i *AddInstr) generateCode(ctx *GeneratorContext) {
	/*
		say we had: add r32, r32, r3

		ldr r4, [r32pos]
		ldr r5, [r33pos]
		add r4, r4, r5
		str r4, [r32pos]
	*/
	ctx.pushCode("add %v, %v, %v", i.Dst.Repr(), i.Op1.Repr(), i.Op2.Repr())
}

func (i *SubInstr) generateCode(ctx *GeneratorContext) {
	ctx.pushCode("sub %v, %v, %v", i.Dst.Repr(), i.Op1.Repr(), i.Op2.Repr())
}

func (i *MulInstr) generateCode(ctx *GeneratorContext) {
	ctx.pushCode("mul %v, %v, %v", i.Dst.Repr(), i.Op1.Repr(), i.Op2.Repr())
}

func (i *DivInstr) generateCode(ctx *GeneratorContext) {
	ctx.pushCode("mov r0, %v", i.Op1.Repr())
	ctx.pushCode("mov r1, %v", i.Op2.Repr())
	ctx.pushCode("bl __aeabi_idiv")
	ctx.pushCode("mov %v, r0", i.Dst.Repr())
}

func (i *AndInstr) generateCode(ctx *GeneratorContext) {
	ctx.pushCode("and %v, %v, %v", i.Dst.Repr(), i.Op1.Repr(), i.Op2.Repr())
}

func (i *OrInstr) generateCode(ctx *GeneratorContext) {
	ctx.pushCode("orr %v, %v, %v", i.Dst.Repr(), i.Op1.Repr(), i.Op2.Repr())
}

func (i *PushScopeInstr) generateCode(ctx *GeneratorContext) {
	ctx.pushCode("push {r4-r12}")
	ctx.registerContents = append([]map[string]Expr{make(map[string]Expr)}, ctx.registerContents...)
	for k, v := range ctx.registerContents[1] {
		ctx.registerContents[0][k] = v
	}
}
func (i *PopScopeInstr) generateCode(ctx *GeneratorContext) {
	ctx.pushCode("pop {r4-r12}")
	ctx.registerContents = ctx.registerContents[1:]
}
func (i *DeclareTypeInstr) generateCode(ctx *GeneratorContext) {
	dst := i.Dst.Repr()
	ctx.registerContents[0][dst] = i.Type
}

func (i *CallInstr) generateCode(ctx *GeneratorContext) {
	ctx.pushCode("bl %v", i.Label.Label)
}

func (ctx *GeneratorContext) generateData(ifCtx *IFContext) {
	for k, v := range ifCtx.dataStore {
		av := v.(*ArrayExpr)
		// Build a string from the char array
		str := ""
		for _, e := range av.Elems {
			str += string(e.(*CharConstExpr).Value)
		}
		ctx.data += fmt.Sprintf("%s:\n\t.asciz \"%s\"\n", k, str)
	}
}

func GenerateCode(ifCtx *IFContext) string {
	ctx := new(GeneratorContext)
	ctx.registerContents = []map[string]Expr{make(map[string]Expr)}
	ctx.dataContents = ifCtx.dataStore

	// Printf format strings
	ctx.data += "printf_fmt_int:\n\t.asciz \"%d\"\n"
	ctx.data += "printf_fmt_char:\n\t.asciz \"%c\"\n"
	ctx.data += "printf_fmt_str:\n\t.asciz \"%s\"\n"
	ctx.data += "printf_fmt_addr:\n\t.asciz \"%p\"\n"
	ctx.generateData(ifCtx)

	// Add the label of each function to the global list
	for _, f := range ifCtx.functions {
		ctx.text += fmt.Sprintf(".global %v\n", f.Instr.(*LabelInstr).Label)
	}
	ctx.text += ".global main\n"

	// Generate function code
	for _, f := range ifCtx.functions {
		ctx.funcStack = append([]string{f.Instr.(*LabelInstr).Label}, ctx.funcStack...)
		f.Instr.generateCode(ctx)
		ctx.pushCode("push {r4-r12,lr}")
		node := f.Next
		for node != nil {
			node.Instr.generateCode(ctx)
			node = node.Next
		}
		ctx.pushLabel("_" + ctx.funcStack[0] + "_end")
		ctx.pushCode("pop {r4-r12,pc}")
		ctx.funcStack = ctx.funcStack[1:]
	}

	// Generate program code
	ifCtx.main.Instr.generateCode(ctx)
	ctx.pushCode("push {lr}")
	node := ifCtx.main.Next
	for node != nil {
		node.Instr.generateCode(ctx)
		node = node.Next
	}
	ctx.pushCode("ldr r0, =0") // default exit code
	ctx.pushCode("pop {pc}")

	// Combine data and text sections
	return ".data\n" + ctx.data + ".text\n" + ctx.text
}

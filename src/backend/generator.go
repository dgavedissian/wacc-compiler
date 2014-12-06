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
	case *PointerConstExpr:
		return "printf_fmt_addr"
	case *RegisterExpr:
		//TODO: log.Println("REG", obj.Repr())
		return getPrintfTypeForExpr(ctx, ctx.registerContents[0][obj.Repr()])
	case *LocationExpr:
		return getPrintfTypeForExpr(ctx, ctx.dataContents[obj.Label])
	default:
		panic(fmt.Sprintf("Attempted to get printf type for unknown thing %#v", expr))
	}
}

func (i *PrintInstr) generateCode(ctx *GeneratorContext) {
	// save regs r0 and r1
	//ctx.pushCode("push {r0,r1}")

	if v, ok := i.Expr.(*CharConstExpr); ok && v.Value == '\n' {
		ctx.pushCode("bl _wacc_print_nl")
		return
	}

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
		ctx.pushCode("bl _wacc_print_bool")
		return
	} else if printfType == "printf_fmt_int" {
		ctx.pushCode("bl _wacc_print_int")
		return
	} else if printfType == "printf_fmt_char" {
		ctx.pushCode("bl _wacc_print_char")
		return
	} else if printfType == "printf_fmt_addr" {
		ctx.pushCode("bl _wacc_print_addr")
		return
	} else {
		ctx.pushCode("ldr r0, =%s", printfType)
		ctx.pushCode("bl printf")
	}

	// Flush output stream
	ctx.pushCode("mov r0, #0")
	ctx.pushCode("bl fflush")

	// load regs r0 and r1
	//ctx.pushCode("pop {r0,r1}")
}

func (i *MoveInstr) generateCode(ctx *GeneratorContext) {
	switch dst := i.Dst.(type) {
	case *MemExpr:
		if dst.Offset == 0 {
			ctx.pushCode("str %v, [%v]", i.Src.Repr(), dst.Address.Repr())
		} else {
			ctx.pushCode("str %v, [%v, #%v]", i.Src.Repr(), dst.Address.Repr(), dst.Offset)
		}

	case *StackLocationExpr:
		ctx.pushCode("str %v, [sp, #%v]", i.Src.(*RegisterExpr).Repr(), i.Dst.(*StackLocationExpr).Id*4)

	case *RegisterExpr:
		// Optimisation step: If we're moving from a constant, just load
		switch src := i.Src.(type) {
		case *IntConstExpr:
			ctx.pushCode("ldr %v, =%v", dst.Repr(), src.Value)
			ctx.registerContents[0][dst.Repr()] = src

		case *BoolConstExpr:
			n := 0
			if src.Value {
				n = 1
			}
			ctx.pushCode("ldr %v, =%v", dst.Repr(), n)
			ctx.registerContents[0][dst.Repr()] = src

		case *CharConstExpr:
			ctx.pushCode("ldr %v, =%v", dst.Repr(), int(src.Value))
			ctx.registerContents[0][dst.Repr()] = src

		case *PointerConstExpr:
			ctx.pushCode("ldr %v, =%v", dst.Repr(), src.Value)
			ctx.registerContents[0][dst.Repr()] = src

		case *LocationExpr:
			ctx.pushCode("ldr %v, =%v", dst.Repr(), src.Label)
			ctx.registerContents[0][dst.Repr()] = ctx.dataContents[src.Label]

		case *RegisterExpr:
			ctx.pushCode("mov %v, %v", dst.Repr(), src.Repr())
			ctx.registerContents[0][dst.Repr()] = ctx.registerContents[0][src.Repr()]

		case *StackLocationExpr:
			ctx.pushCode("ldr %v, [sp, #%v]", dst.Repr(), src.Id*4)

		default:
			panic(fmt.Sprintf("Unhandled src type of mov %T", src))
		}

	default:
		panic(fmt.Sprintf("Unhandled dst type of mov %T", dst))
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
	ctx.pushCode("adds %v, %v, %v", i.Dst.Repr(), i.Op1.Repr(), i.Op2.Repr())
	ctx.pushCode("blvs _wacc_throw_overflow_error")
}

func (i *SubInstr) generateCode(ctx *GeneratorContext) {
	ctx.pushCode("subs %v, %v, %v", i.Dst.Repr(), i.Op1.Repr(), i.Op2.Repr())
	ctx.pushCode("blvs _wacc_throw_overflow_error")
}

func (i *MulInstr) generateCode(ctx *GeneratorContext) {
	ctx.pushCode("smull %v, %v, %v, %v", i.Dst.Repr(), i.Op1.Repr(), i.Op2.Repr(), i.Op1.Repr())
	ctx.pushCode("CMP %v, %v, ASR #31", i.Op1.Repr(), i.Dst.Repr())
	ctx.pushCode("BLNE _wacc_throw_overflow_error")
}

func (i *DivInstr) generateCode(ctx *GeneratorContext) {
	ctx.pushCode("mov r0, %v", i.Op1.Repr())
	ctx.pushCode("mov r1, %v", i.Op2.Repr())
	ctx.pushCode("bl _wacc_check_divide_by_zero")
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
	ctx.pushCode("push {r4-r11}")
	ctx.registerContents = append([]map[string]Expr{make(map[string]Expr)}, ctx.registerContents...)
	for k, v := range ctx.registerContents[1] {
		ctx.registerContents[0][k] = v
	}
}

func (i *PopScopeInstr) generateCode(ctx *GeneratorContext) {
	ctx.pushCode("pop {r4-r11}")
	ctx.registerContents = ctx.registerContents[1:]
}

func (i *DeclareTypeInstr) generateCode(ctx *GeneratorContext) {
	dst := i.Dst.Repr()
	ctx.registerContents[0][dst] = i.Type
}

func (i *CallInstr) generateCode(ctx *GeneratorContext) {
	ctx.pushCode("bl %v", i.Label.Label)
}

func (i *HeapAllocInstr) generateCode(ctx *GeneratorContext) {
	ctx.pushCode("ldr r0, =%v", i.Size)
	ctx.pushCode("bl malloc")
	ctx.pushCode("mov %v, r0", i.Dst.Repr())
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

func (ctx *GeneratorContext) generateFunction(x *InstrNode) {
	ctx.funcStack = append([]string{x.Instr.(*LabelInstr).Label}, ctx.funcStack...)
	x.Instr.generateCode(ctx)
	ctx.pushCode("push {r4-r11,lr}")

	stackSpace := x.stackSpace * 4
	for stackSpace > 0 {
		thisTime := stackSpace
		if thisTime > 1024 {
			thisTime = 1024
		}
		ctx.pushCode("sub sp, sp, #%d", thisTime)
		stackSpace -= thisTime
	}

	node := x.Next
	for node != nil {
		node.Instr.generateCode(ctx)
		node = node.Next
	}

	ctx.pushLabel("_" + ctx.funcStack[0] + "_end")
	stackSpace = x.stackSpace * 4
	for stackSpace > 0 {
		thisTime := stackSpace
		if thisTime > 1024 {
			thisTime = 1024
		}
		ctx.pushCode("add sp, sp, #%d", thisTime)
		stackSpace -= thisTime
	}
	ctx.pushCode("pop {r4-r11,pc}")
	ctx.funcStack = ctx.funcStack[1:]
}

func GenerateCode(ifCtx *IFContext) string {
	ctx := new(GeneratorContext)
	ctx.registerContents = []map[string]Expr{make(map[string]Expr)}
	ctx.dataContents = ifCtx.dataStore

	// Printf format strings
	ctx.data += `
printf_fmt_int:
	.asciz "%d"
printf_fmt_char:
	.asciz "%c"
printf_fmt_str:
	.asciz "%s"
printf_fmt_addr:
	.asciz "%p"
printf_true:
	.asciz "true"
printf_false:
	.asciz "false"
_wacc_overflow_error_msg:
	.asciz	"OverflowError: the result is too small/large to store in a 4-byte signed-integer.\n"
_wacc_divide_by_zero_msg:
	.asciz "DivideByZeroError: divide or modulo by zero\n"
`
	ctx.generateData(ifCtx)

	// Add the label of each function to the global list
	for _, f := range ifCtx.functions {
		ctx.text += fmt.Sprintf(".global %v\n", f.Instr.(*LabelInstr).Label)
	}
	ctx.text += ".global main\n"

	// Generate function code
	for _, f := range ifCtx.functions {
		ctx.generateFunction(f)
	}

	// Generate program code
	ctx.generateFunction(ifCtx.main)

	// Combine data and text sections
	return ".data\n" + ctx.data + ".text\n" + ctx.text + `
_wacc_check_divide_by_zero:
	PUSH {lr}
	CMP r1, #0
	LDREQ r1, =_wacc_divide_by_zero_msg
	BLEQ _wacc_throw_runtime_error
	POP {pc}
_wacc_throw_overflow_error:
	ldr r1, =_wacc_overflow_error_msg
	bl _wacc_throw_runtime_error
_wacc_throw_runtime_error:
	bl _wacc_print_str
	mov r0, #-1
	bl exit
_wacc_print_bool:
	push {lr}
	cmp r1, #0
	beq _wacc_print_bool_false
	ldr r0, =printf_true
	b _wacc_print_bool_done
_wacc_print_bool_false:
	ldr r0, =printf_false
_wacc_print_bool_done:
	bl printf
	mov r0, #0
	bl fflush
	pop {pc}
_wacc_print_int:
	push {lr}
	ldr r0, =printf_fmt_int
	bl printf
	mov r0, #0
	bl fflush
	pop {pc}
_wacc_print_char:
	push {lr}
	ldr r0, =printf_fmt_char
	bl printf
	mov r0, #0
	bl fflush
	pop {pc}
_wacc_print_str:
	push {lr}
	ldr r0, =printf_fmt_str
	bl printf
	mov r0, #0
	bl fflush
	pop {pc}
_wacc_print_addr:
	push {lr}
	ldr r0, =printf_fmt_addr
	bl printf
	mov r0, #0
	bl fflush
	pop {pc}
_wacc_print_nl:
	push {lr}
	mov r0, #'\n'
	bl putchar
	pop {pc}
`
}

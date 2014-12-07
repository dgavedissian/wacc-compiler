package backend

import (
	"fmt"
	"log"
	"unicode/utf8"

	"../frontend"
)

type GeneratorContext struct {
	stringCounter    int
	data             string
	text             string
	registerContents []map[string]Expr
	dataContents     map[string]*StringConstExpr
	funcStack        []string
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

	// Get scanf format string depending on type
	t := getTypeForExpr(ctx, i.Dst).Type
	var fmtString string
	if t.Equals(frontend.BasicType{frontend.INT}) {
		fmtString = "scanf_fmt_int"
	} else if t.Equals(frontend.BasicType{frontend.CHAR}) {
		fmtString = "scanf_fmt_char"
	}
	ctx.pushCode("ldr r0, =%s", fmtString)
	ctx.pushCode("bl wscanf")

	// Move output to destination
	ctx.pushCode("ldr %s, [sp]", i.Dst.Repr())

	// load regs r0 and r1
	ctx.pushCode("pop {r0,r1}")
}

func (i *FreeInstr) generateCode(ctx *GeneratorContext) {
	ctx.pushCode("mov r0, %v", i.Object.(*RegisterExpr).Repr())
	ctx.pushCode("bl free")
}

func (i *ReturnInstr) generateCode(ctx *GeneratorContext) {
	ctx.pushCode("mov r0, %v", i.Expr.Repr())
	ctx.pushCode("b _" + ctx.funcStack[0] + "_end")
}

func (i *ExitInstr) generateCode(ctx *GeneratorContext) {
	ctx.pushCode("mov r0, %v", i.Expr.Repr())
	ctx.pushCode("bl exit")
}

func getTypeForExpr(ctx *GeneratorContext, expr Expr) *TypeExpr {
	switch obj := expr.(type) {
	case *TypeExpr:
		return obj

	case *IntConstExpr:
		return &TypeExpr{frontend.BasicType{frontend.INT}}

	case *CharConstExpr:
		return &TypeExpr{frontend.BasicType{frontend.CHAR}}

	case *ArrayConstExpr:
		return &TypeExpr{obj.Type}

	case *StringConstExpr:
		return &TypeExpr{frontend.BasicType{frontend.STRING}}

	case *BoolConstExpr:
		return &TypeExpr{frontend.BasicType{frontend.BOOL}}

	case *PointerConstExpr:
		return &TypeExpr{frontend.BasicType{frontend.PAIR}}

	case *RegisterExpr:
		log.Println("REG", obj.Repr())
		return getTypeForExpr(ctx, ctx.registerContents[0][obj.Repr()])

	case *LocationExpr:
		return getTypeForExpr(ctx, ctx.dataContents[obj.Label])

	default:
		panic(fmt.Sprintf("Attempted to get printf type for unknown thing %#v", expr))
	}
}

func getFormatStringFromType(t *TypeExpr) string {
	internalType := t.Type
	if internalType.Equals(frontend.BasicType{frontend.BOOL}) {
		return "printf_fmt_int"
	} else if internalType.Equals(frontend.BasicType{frontend.INT}) {
		return "printf_fmt_int"
	} else if internalType.Equals(frontend.BasicType{frontend.CHAR}) {
		return "printf_fmt_char"
	} else if internalType.Equals(frontend.BasicType{frontend.STRING}) {
		return "printf_fmt_wstr"
	} else {
		return "printf_fmt_addr"
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
		if value {
			ctx.pushCode("ldr r1, =1")
		} else {
			ctx.pushCode("ldr r1, =0")
		}

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

	derivedType := getTypeForExpr(ctx, i.Expr).Type
	if derivedType.Equals(frontend.BasicType{frontend.BOOL}) {
		ctx.pushCode("bl _wacc_print_bool")
	} else if derivedType.Equals(frontend.BasicType{frontend.INT}) {
		ctx.pushCode("bl _wacc_print_int")
	} else if derivedType.Equals(frontend.BasicType{frontend.CHAR}) {
		ctx.pushCode("bl _wacc_print_char")
	} else if derivedType.Equals(frontend.BasicType{frontend.STRING}) {
		ctx.pushCode("bl _wacc_print_wstr")
	} else if derivedType.Equals(frontend.ArrayType{frontend.BasicType{frontend.CHAR}}) {
		ctx.pushCode("bl _wacc_print_wstr")
	} else {
		ctx.pushCode("bl _wacc_print_addr")
	}

	// load regs r0 and r1
	//ctx.pushCode("pop {r0,r1}")
}

func (i *MoveInstr) generateCode(ctx *GeneratorContext) {
	switch dst := i.Dst.(type) {
	case *MemExpr:
		if dst.Offset == 0 {
			ctx.pushCode("str %v, [%v]", i.Src.(*RegisterExpr).Repr(), dst.Address.Repr())
		} else {
			ctx.pushCode("str %v, [%v, #%v]", i.Src.(*RegisterExpr).Repr(), dst.Address.Repr(), dst.Offset)
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

		case *MemExpr:
			if src.Offset == 0 {
				ctx.pushCode("ldr %v, [%v]", dst.Repr(), src.Address.Repr())
			} else {
				ctx.pushCode("ldr %v, [%v, #%v]", dst.Repr(), src.Address.Repr(), src.Offset)
			}

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

func (i *NegInstr) generateCode(ctx *GeneratorContext) {
	arg := i.Expr.(*RegisterExpr).Repr()

	ctx.pushCode("rsbs %v, %v, #0", arg, arg)
	ctx.pushCode("blvs " + RuntimeOverflowLabel)
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
	shift := ""
	if i.Op2Shift != nil {
		shift = ", " + i.Op2Shift.Repr()
	}

	// ADD can have either an immediate or register as the 2nd operand
	if imm, ok := i.Op2.(*IntConstExpr); ok {
		ctx.pushCode("adds %v, %v, #%v%v", i.Dst.Repr(), i.Op1.Repr(), imm.Value, shift)
	} else {
		ctx.pushCode("adds %v, %v, %v%v", i.Dst.Repr(), i.Op1.Repr(), i.Op2.(*RegisterExpr).Repr(), shift)
	}
	ctx.pushCode("blvs " + RuntimeOverflowLabel)
}

func (i *SubInstr) generateCode(ctx *GeneratorContext) {
	shift := ""
	if i.Op2Shift != nil {
		shift = ", " + i.Op2Shift.Repr()
	}

	// SUB can have either an immediate or register as the 2nd operand
	if imm, ok := i.Op2.(*IntConstExpr); ok {
		ctx.pushCode("subs %v, %v, #%v%v", i.Dst.Repr(), i.Op1.Repr(), imm.Value, shift)
	} else {
		ctx.pushCode("subs %v, %v, %v%v", i.Dst.Repr(), i.Op1.Repr(), i.Op2.(*RegisterExpr).Repr(), shift)
	}
	ctx.pushCode("blvs " + RuntimeOverflowLabel)
}

func (i *MulInstr) generateCode(ctx *GeneratorContext) {
	ctx.pushCode("smull %v, %v, %v, %v", i.Dst.Repr(), i.Op1.Repr(), i.Op2.Repr(), i.Op1.Repr())
	ctx.pushCode("cmp %v, %v, ASR #31", i.Op1.Repr(), i.Dst.Repr())
	ctx.pushCode("blne " + RuntimeOverflowLabel)
}

func (i *DivInstr) generateCode(ctx *GeneratorContext) {
	ctx.pushCode("mov r0, %v", i.Op1.Repr())
	ctx.pushCode("mov r1, %v", i.Op2.Repr())
	ctx.pushCode("bl " + RuntimeCheckDivZeroLabel)
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
		wideString := ""
		for _, c := range v.Value {
			bs := make([]byte, 4)
			utf8.EncodeRune(bs, c)
			for _, b := range bs {
				wideString += fmt.Sprintf("\\%03o", b)
			}
		}
		ctx.data += fmt.Sprintf("%s:\n\t.word %v\n\t.ascii \"%s\"\n", k, len(v.Value), wideString)
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
	.ascii "%\000\000\000d\000\000\000\000\000\000\000"
scanf_fmt_int:
	.ascii "%\000\000\000d\000\000\000\000\000\000\000"
printf_fmt_char:
	.ascii "%\000\000\000c\000\000\000\000\000\000\000"
scanf_fmt_char:
	.ascii " \000\000\000%\000\000\000c\000\000\000\000\000\000\000"
printf_fmt_str:
	.ascii "%\000\000\000s\000\000\000\000\000\000\000"
printf_fmt_wstr:
	.ascii "%\000\000\000.\000\000\000*\000\000\000l\000\000\000s\000\000\000\000\000\000\000"
printf_fmt_addr:
	.ascii "%\000\000\000p\000\000\000\000\000\000\000"
printf_true:
	.asciz "true"
printf_false:
	.asciz "false"
_wacc_overflow_error_msg:
	.asciz "OverflowError: the result is too small/large to store in a 4-byte signed-integer.\n"
_wacc_divide_by_zero_msg:
	.asciz "DivideByZeroError: divide or modulo by zero\n"
_wacc_array_index_negative_msg:
	.asciz "ArrayIndexOutOfBoundsError: negative index\n"
_wacc_array_index_large_msg:
	.asciz "ArrayIndexOutOfBoundsError: index too large\n"
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
` + RuntimeCheckArrayBoundsLabel + `:
	push {lr}
	cmp r0, #0
	ldrlt r1, =_wacc_array_index_negative_msg
	bllt _wacc_throw_runtime_error
	ldr r1, [r1]
	cmp r0, r1
	ldrge r1, =_wacc_array_index_large_msg
	blge _wacc_throw_runtime_error
	pop {pc}
` + RuntimeCheckDivZeroLabel + `:
	push {lr}
	cmp r1, #0
	ldreq r1, =_wacc_divide_by_zero_msg
	bleq _wacc_throw_runtime_error
	pop {pc}
` + RuntimeOverflowLabel + `:
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
	ldr r1, =printf_true
	b _wacc_print_bool_done
_wacc_print_bool_false:
	ldr r1, =printf_false
_wacc_print_bool_done:
	bl _wacc_print_str
	pop {pc}
_wacc_print_int:
	push {lr}
	ldr r0, =printf_fmt_int
	bl wprintf
	mov r0, #0
	bl fflush
	pop {pc}
_wacc_print_char:
	push {lr}
	ldr r0, =printf_fmt_char
	bl wprintf
	mov r0, #0
	bl fflush
	pop {pc}
_wacc_print_str:
	push {lr}
	ldr r0, =printf_fmt_str
	bl wprintf
	mov r0, #0
	bl fflush
	pop {pc}
_wacc_print_wstr:
	push {lr}
	add r2, r1, #4
	ldr r1, [r1]
	ldr r0, =printf_fmt_wstr
	bl wprintf
	mov r0, #0
	bl fflush
	pop {pc}
_wacc_print_addr:
	push {lr}
	ldr r0, =printf_fmt_addr
	bl wprintf
	mov r0, #0
	bl fflush
	pop {pc}
_wacc_print_nl:
	push {lr}
	mov r0, #'\n'
	bl putwchar
	pop {pc}
`
}

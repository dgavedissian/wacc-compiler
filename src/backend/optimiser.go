package backend

func AllocateRegisters(ifCtx *IFContext) {
	variableMap := make(map[*VarExpr]*RegisterExpr)

	index := 0
	VisitInstructions(ifCtx, func(i Instr) {
		switch instr := i.(type) {
		case *MoveInstr:
			if v, ok := instr.Dst.(*VarExpr); ok {
				if reg, ok := variableMap[v]; ok {
					instr.Dst = reg
				} else {
					reg := &RegisterExpr{index}
					instr.Dst = reg
					variableMap[v] = reg
					index++
				}
			}
		}
	})
}

func OptimiseIF(ifCtx *IFContext) {
}

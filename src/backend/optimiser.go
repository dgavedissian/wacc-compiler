package backend

func AllocateRegisters(ifCtx *IFContext) {
	variableMap := make(map[string]*RegisterExpr)

	index := 0
	replaceWithReg := func(v *VarExpr) *RegisterExpr {
		if reg, ok := variableMap[v.Name]; ok {
			return reg
		} else {
			index++
			reg := &RegisterExpr{index}
			variableMap[v.Name] = reg
			return reg
		}
	}

	VisitInstructions(ifCtx, func(i Instr) {
		switch instr := i.(type) {
		case *MoveInstr:
			if v, ok := instr.Dst.(*VarExpr); ok {
				instr.Dst = replaceWithReg(v)
			}
		}
	})
}

func OptimiseIF(ifCtx *IFContext) {
}

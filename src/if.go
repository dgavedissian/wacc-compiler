package main

import (
	"fmt"
)

type Instr interface {
	instrNode()
}

type NoopInstr struct {
	Next Instr
}

func (NoopInstr) instrNode() {}

type IfContext struct {
	labels        map[string]Instr
	startNode     Instr
	nextTemporary int
}

func (ctx *IfContext) generateIf(nodes []Stmt) Instr {
	if len(nodes) == 0 {
		panic("zero-length array passed to generateIf")
	}

	var nextInstr Instr

	for i := len(nodes) - 1; i >= 0; i -= 1 {
		switch node := nodes[i].(type) {
		case *ProgStmt:
			if len(node.Funcs) != 0 {
				panic("not yet implemented")
			}
			return ctx.generateIf(node.Body)
		case *SkipStmt:
			thisInstr := new(NoopInstr)
			thisInstr.Next = nextInstr
			nextInstr = thisInstr
		default:
			panic(fmt.Sprintf("what is a %s?", node.Repr()))
		}
	}

	if nextInstr == nil {
		panic("reached end of function with no head. I'm headless")
	}
	return nextInstr
}

func GenerateIntermediateForm(program *ProgStmt) *IfContext {
	ctx := new(IfContext)

	ctx.startNode = ctx.generateIf([]Stmt{program})

	return ctx
}

package main

import (
	"fmt"
)

func VerifySemantics(program *ProgStmt) {
	// Verify functions
	for _, f := range program.Funcs {
		VerifyFunctionSemantics(f)
	}

	// Verify statements
	for _, s := range program.Body {
		VerifyStatementSemantics(s)
	}
}

func VerifyFunctionSemantics(function Func) {
}

func VerifyStatementSemantics(statement Stmt) {
	switch statement.(type) {
	case *DeclStmt:
		declStatement := statement.(*DeclStmt)
		fmt.Printf("%d == %d\n", declStatement.Kind, GetKind(declStatement.Right))
		if declStatement.Kind != GetKind(declStatement.Right) {
			SemanticError("semantic error - Right hand side of variable declaration doesn't match the type of the variable")
		}
	}
}

func GetKind(expr Expr) int {
	switch expr.(type) {
	case *BasicLit:
		basicLit := expr.(*BasicLit)
		return basicLit.Kind

	default:
		return -1
	}
}

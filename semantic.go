package main

import ()

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
		if declStatement.Kind != GetKind(declStatement.Right) {
			SemanticError("semantic error - Right hand side of variable declaration doesn't match the type of the variable")
		}
	}
}

func GetKind(expr Expr) int {
	switch expr.(type) {
	case *BasicLit:
		basicLit := expr.(*BasicLit)

		// Get type of *_LIT
		switch basicLit.Kind {
		case INT_LIT:
			return INT
		case BOOL_LIT:
			return BOOL
		case CHAR_LIT:
			return CHAR
		case STRING_LIT:
			return STRING
		default:
			return -1
		}

	default:
		return -1
	}
}

package main

func VerifySemantics(program ProgStmt) {
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
}

func GetKind(expr *Expr) int {
	return INT
}

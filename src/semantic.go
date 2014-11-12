package main

type Context struct {
	store map[string]int
}

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
		if !declStatement.Kind.Equals(GetKind(declStatement.Right)) {
			SemanticError("semantic error - Right hand side of variable declaration (%r) doesn't match the type of the variable (%r)")
		}

	case *AssignStmt:
		// TODO: Check

	case *IfStmt:
		ifStmt := statement.(*IfStmt)

		// Check for boolean condition
		if !GetKind(ifStmt.Cond).Equals(BasicType{BOOL}) {
			SemanticError("semantic error - Condition is not a bool")
		}

		// Verify branches
		for _, s := range ifStmt.Body {
			VerifyStatementSemantics(s)
		}
		for _, s := range ifStmt.Else {
			VerifyStatementSemantics(s)
		}

	}
}

func GetKind(expr Expr) Type {
	switch expr := expr.(type) {
	case *BasicLit:
		return expr.Kind
	default:
		panic("WTF I DON'T KNOW WHAT THIS IS HELP")
	}
}

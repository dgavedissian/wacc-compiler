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
	VerifyStatementListSemantics(program.Body)
}

func VerifyStatementListSemantics(statementList []Stmt) {
	for _, s := range statementList {
		VerifyStatementSemantics(s)
	}
}

func VerifyFunctionSemantics(function Func) {
}

func VerifyStatementSemantics(statement Stmt) {
	switch statement := statement.(type) {
	case *DeclStmt:
		if !statement.Kind.Equals(GetKind(statement.Right)) {
			SemanticError(0, "semantic error - Right hand side of variable declaration doesn't match the type of the variable")
		}

	case *AssignStmt:
		// TODO: Check

	case *IfStmt:
		// Check for boolean condition
		if !GetKind(statement.Cond).Equals(BasicType{BOOL}) {
			SemanticError(0, "semantic error - Condition '%s' is not a bool", statement.Cond.Repr())
		}

		// Verify branches
		VerifyStatementListSemantics(statement.Body)
		VerifyStatementListSemantics(statement.Else)

	case *WhileStmt:
		// Check the condition
		if !GetKind(statement.Cond).Equals(BasicType{BOOL}) {
			SemanticError(0, "semantic error - Condition '%s' is not a bool", statement.Cond.Repr())
		}

		// Verfy body
		VerifyStatementListSemantics(statement.Body)
	}
}

func GetKind(expr Expr) Type {
	switch expr := expr.(type) {
	case *BasicLit:
		return expr.Kind

	case *UnaryExpr:
		t := GetKind(expr.Operand)

		// TODO: Check whether unary operator supports the operand type
		// Refer to the table in the spec
		return t

	case *BinaryExpr:
		t1, t2 := GetKind(expr.Left), GetKind(expr.Right)

		// TODO: Check whether binary operator supports the operand types
		// Refer to the table in the spec

		if !t1.Equals(t2) {
			SemanticError(0, "semantic error - Types of binary expression operands do not match")
		}

		return t1

	default:
		panic("WTF I DON'T KNOW WHAT THIS IS HELP")
	}
}

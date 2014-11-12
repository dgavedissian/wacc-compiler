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
	switch statement.(type) {
	case *DeclStmt:
		declStatement := statement.(*DeclStmt)
		if declStatement.Kind != GetKind(declStatement.Right) {
			SemanticError(0, "semantic error - Right hand side of variable declaration doesn't match the type of the variable")
		}

	case *AssignStmt:
		// TODO: Check

	case *IfStmt:
		ifStmt := statement.(*IfStmt)

		// Check for boolean condition
		if GetKind(ifStmt.Cond) != BOOL {
			SemanticError(0, "semantic error - Condition '%s' is not a bool", ifStmt.Cond.Repr())
		}

		// Verify branches
		VerifyStatementListSemantics(ifStmt.Body)
		VerifyStatementListSemantics(ifStmt.Else)

	case *WhileStmt:
		whileStmt := statement.(*WhileStmt)

		// Check the condition
		if GetKind(whileStmt.Cond) != BOOL {
			SemanticError(0, "semantic error - Condition '%s' is not a bool", whileStmt.Cond.Repr())
		}

		// Verfy body
		VerifyStatementListSemantics(whileStmt.Body)
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

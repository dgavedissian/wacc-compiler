package main

import (
	"strconv"
)

// Verification of a function body
func VerifyStatementReturns(stmt Stmt) bool {
	switch stmt.(type) {
	case *IfStmt:
		ifStmt := stmt.(*IfStmt)
		return VerifyStatementReturns(ifStmt.Body[len(ifStmt.Body)-1]) &&
			VerifyStatementReturns(ifStmt.Else[len(ifStmt.Else)-1])

	case *WhileStmt:
		whileStmt := stmt.(*WhileStmt)
		return VerifyStatementReturns(whileStmt.Body[len(whileStmt.Body)-1])

	case *ExitStmt:
		return true

	case *ReturnStmt:
		return true

	default:
		return false
	}
}

func IntLiteralToIntConst(basicLit BasicLit) int64 {
	n, _ := strconv.ParseInt(basicLit.Value, 10, 64)
	return n
}

func StaticUnaryMinusOverflows(unaryExpr UnaryExpr) bool {
	operand := unaryExpr.Operand

	switch operand.(type) {
	case *BasicLit:
		basicLit := operand.(*BasicLit)
		if basicLit.Kind == INT_LIT {
			n := IntLiteralToIntConst(*basicLit)
			// Smallest 32bit literal is -(1<<31)
			// The lexer always generates positive literals
			return n > (1 << 31)
		}
		return false

	default:
		return StaticExprOverflows(operand)
	}
}

func StaticExprOverflows(expr Expr) bool {
	switch expr.(type) {
	case *UnaryExpr:
		unaryExpr := expr.(*UnaryExpr)
		if unaryExpr.Operator == "-" {
			return StaticUnaryMinusOverflows(*unaryExpr)
		}
		return StaticExprOverflows(unaryExpr.Operand)

	case *BasicLit:
		basicLit := expr.(*BasicLit)
		if basicLit.Kind == INT_LIT {
			n := IntLiteralToIntConst(*basicLit)
			return n > ((1 << 31) - 1)
		}
		return false

	default:
		return false
	}
}

func VerifyNoOverflows(expr Expr) {
	if StaticExprOverflows(expr) {
		lex.Error("syntax error - Int literal overflow")
	}
}

// We're only concerned with the very last statement
func VerifyFunctionReturns(stmtList []Stmt) {
	if !VerifyStatementReturns(stmtList[len(stmtList)-1]) {
		lex.Error("syntax error - Function has no return statement on every control path or doesn't end in an exit statement")
	}
}

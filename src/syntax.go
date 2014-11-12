package main

import (
	"strconv"
)

// Verification of a function body
func VerifyStatementReturns(stmt Stmt) bool {
	switch stmt := stmt.(type) {
	case *IfStmt:
		return VerifyStatementReturns(stmt.Body[len(stmt.Body)-1]) &&
			VerifyStatementReturns(stmt.Else[len(stmt.Else)-1])

	case *WhileStmt:
		return VerifyStatementReturns(stmt.Body[len(stmt.Body)-1])

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

	switch operand := operand.(type) {
	case *BasicLit:
		if operand.Type.Equals(BasicType{INT}) {
			n := IntLiteralToIntConst(*operand)
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
	switch expr := expr.(type) {
	case *UnaryExpr:
		if expr.Operator == "-" {
			return StaticUnaryMinusOverflows(*expr)
		}
		return StaticExprOverflows(expr.Operand)

	case *BasicLit:
		if expr.Type.Equals(BasicType{INT}) {
			n := IntLiteralToIntConst(*expr)
			return n > ((1 << 31) - 1)
		}
		return false

	default:
		return false
	}
}

func VerifyNoOverflows(expr Expr) {
	if StaticExprOverflows(expr) {
		// TODO: Use token.Position.Line rather than lex.l
		SyntaxError(lex.l, "syntax error - Int literal overflow")
	}
}

// We're only concerned with the very last statement
func VerifyFunctionReturns(stmtList []Stmt) {
	endStmt := stmtList[len(stmtList)-1]
	if !VerifyStatementReturns(endStmt) {
		// TODO: Use token.Position.Line rather than lex.l
		SyntaxError(lex.l, "syntax error - Function has no return statement on every control path or doesn't end in an exit statement")
	}
}

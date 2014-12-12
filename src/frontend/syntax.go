package frontend

import (
	"strconv"
)

const LOWER_BOUND = -(1 << 31)
const UPPER_BOUND = (1 << 31) - 1

func VerifyAnyStatementsReturn(stmts []Stmt) bool {
	for i := len(stmts) - 1; i >= 0; i-- {
		if VerifyStatementReturns(stmts[i]) {
			return true
		}
	}
	return false
}

// Verification of a statement
func VerifyStatementReturns(stmt Stmt) bool {
	switch stmt := stmt.(type) {
	case *IfStmt:
		return VerifyAnyStatementsReturn(stmt.Body) &&
			VerifyAnyStatementsReturn(stmt.Else)

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
			// Negate n as the lexer always generates abs(n)
			// It is possible to just check the abs, but this adds clarity.
			n = -n
			return n < LOWER_BOUND
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
			return n > UPPER_BOUND
		}
		return false

	default:
		return false
	}
}

func VerifyNoOverflows(expr Expr) {
	if StaticExprOverflows(expr) {
		SyntaxError(expr.Pos(), "integer literal does not fit in an int variable")
	}
}

// Iterate in reverse through body. If any of the top level statements return,
// it returns on all code paths. If none of the top level statements return,
// error.
func VerifyFunctionReturns(stmtList []Stmt) {
	if !VerifyAnyStatementsReturn(stmtList) {
		SyntaxError(stmtList[0].Pos(), "function does not have a return or exit statement on every control path")
	}
}

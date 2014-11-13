package main

// Context (Variable Store)
type Context struct {
	types map[string]Type
}

func (cxt Context) Add(expr LValueExpr, t Type) {
	switch expr := expr.(type) {
	case *IdentExpr:
		cxt.types[expr.Name] = t

	default:
		SemanticError(0, "IMPLEMENT_ME: Context.Add not defined for type %T", expr)
	}
}

func (cxt Context) LookupType(expr LValueExpr) Type {
	switch expr := expr.(type) {
	case *IdentExpr:
		t, ok := cxt.types[expr.Name]
		if !ok {
			SemanticError(0, "semantic error - Variable '%s' is not in the variable store", expr.Name)
			return ErrorType{}
		} else {
			return t
		}

	case *ArrayElemExpr:
		t := cxt.LookupType(expr.Volume).(ArrayType)
		return t.BaseType

	default:
		SemanticError(0, "IMPLEMENT_ME: Context.LookupType not defined for type %T", expr)
		return ErrorType{}
	}
}

// Semantic Checking
func VerifySemantics(program *ProgStmt) {
	cxt := &Context{make(map[string]Type)}
	for _, f := range program.Funcs {
		VerifyFunctionSemantics(cxt, f)
	}
	VerifyStatementListSemantics(cxt, program.Body)
}

func VerifyFunctionSemantics(cxt *Context, function Function) {
}

func VerifyStatementListSemantics(cxt *Context, statementList []Stmt) {
	for _, s := range statementList {
		VerifyStatementSemantics(cxt, s)
	}
}

func VerifyStatementSemantics(cxt *Context, statement Stmt) {
	switch statement := statement.(type) {
	case *DeclStmt:
		t1, t2 := statement.Type, DeriveType(cxt, statement.Right)
		if !t1.Equals(t2) {
			SemanticError(0, "semantic error - Value being used to initialise '%s' does not match it's type (%s != %s)",
				statement.Ident.Name, t1.Repr(), t2.Repr())
		} else {
			cxt.Add(statement.Ident, statement.Type)
		}

	case *AssignStmt:
		t1, t2 := cxt.LookupType(statement.Left), DeriveType(cxt, statement.Right)
		if !t1.Equals(t2) {
			SemanticError(0, "semantic error - Cannot assign rvalue to lvalue with a different type (%s != %s)", t1.Repr(), t2.Repr())
		}

	case *IfStmt:
		// Check for boolean condition
		t := DeriveType(cxt, statement.Cond)
		if !t.Equals(BasicType{BOOL}) {
			SemanticError(0, "semantic error - Condition '%s' is not a bool (actual type: %s)", statement.Cond.Repr(), t.Repr())
		}

		// Verify branches
		VerifyStatementListSemantics(cxt, statement.Body)
		VerifyStatementListSemantics(cxt, statement.Else)

	case *WhileStmt:
		// Check the condition
		t := DeriveType(cxt, statement.Cond)
		if !t.Equals(BasicType{BOOL}) {
			SemanticError(0, "semantic error - Condition '%s' is not a bool (actual type: %s)", statement.Cond.Repr(), t.Repr())
		}

		// Verfy body
		VerifyStatementListSemantics(cxt, statement.Body)
	}
}

func DeriveType(cxt *Context, expr Expr) Type {
	switch expr := expr.(type) {
	case *BasicLit:
		return expr.Type

	case *ArrayLit:
		// Check if the array has any elements
		if len(expr.Values) == 0 {
			SemanticError(0, "semantic error - Array literal cannot be empty")
			return ErrorType{}
		}

		// Check that all the types match
		t := DeriveType(cxt, expr.Values[0])
		for i := 1; i < len(expr.Values); i++ {
			if !t.Equals(DeriveType(cxt, expr.Values[i])) {
				SemanticError(0, "semantic error - All expressions in the array literal must have the same type")
				return ErrorType{}
			}
		}

		// Just return the first elements type
		return ArrayType{t}

	case *IdentExpr:
		return cxt.LookupType(expr)

	case *UnaryExpr:
		t := DeriveType(cxt, expr.Operand)

		// TODO: Check whether unary operator supports the operand type
		// Refer to the table in the spec
		return t

	case *BinaryExpr:
		t1, t2 := DeriveType(cxt, expr.Left), DeriveType(cxt, expr.Right)

		// TODO: Check whether binary operator supports the operand types
		// Refer to the table in the spec

		if !t1.Equals(t2) {
			SemanticError(0, "semantic error - Types of binary expression operands do not match (%s != %s)", t1.Repr(), t2.Repr())
			return ErrorType{}
		} else {
			return t1
		}

	case *NewPairCmd:
		return PairType{DeriveType(cxt, expr.Left), DeriveType(cxt, expr.Right)}

	default:
		SemanticError(0, "IMPLEMENT_ME: Unhandled type in DeriveType - Type: %T", expr)
		return ErrorType{}
	}
}

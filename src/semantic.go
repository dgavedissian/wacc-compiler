package main

// Context (Variable Store)
type Context struct {
	types map[string]Type
}

func (cxt *Context) Add(expr LValueExpr, t Type) {
	switch expr := expr.(type) {
	case *IdentExpr:
		cxt.types[expr.Name] = t

	default:
		SemanticError(0, "IMPLEMENT_ME: Context.Add not defined for type %T", expr)
	}
}

func (cxt *Context) DeriveType(expr Expr) Type {
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
		t := cxt.DeriveType(expr.Volume).(ArrayType)
		return t.BaseType

	case *PairElemExpr:
		t := cxt.DeriveType(expr.Operand)
		if pair, ok := t.(PairType); ok {
			switch expr.SelectorType {
			case FST:
				return pair.Fst
			case SND:
				return pair.Snd
			default:
				panic("expr.SelectorType must be either FST or SND")
			}
		} else {
			SemanticError(0, "semantic error - Operand of pair selector must be a pair type (actual: %s)", t.Repr())
			return ErrorType{}
		}

	case *BasicLit:
		return expr.Type

	case *ArrayLit:
		// Check if the array has any elements
		if len(expr.Values) == 0 {
			return ArrayType{AnyType{}}
		}

		// Check that all the types match
		t := cxt.DeriveType(expr.Values[0])
		for i := 1; i < len(expr.Values); i++ {
			if !t.Equals(cxt.DeriveType(expr.Values[i])) {
				SemanticError(0, "semantic error - All expressions in the array literal must have the same type")
				return ErrorType{}
			}
		}

		// Just return the first elements type
		return ArrayType{t}

	case *UnaryExpr:
		t := cxt.DeriveType(expr.Operand)

		// TODO: Check whether unary operator supports the operand type
		// Refer to the table in the spec
		return t

	case *BinaryExpr:
		t1, t2 := cxt.DeriveType(expr.Left), cxt.DeriveType(expr.Right)

		// TODO: Check whether binary operator supports the operand types
		// Refer to the table in the spec

		if !t1.Equals(t2) {
			SemanticError(0, "semantic error - Types of binary expression operands do not match (%s != %s)", t1.Repr(), t2.Repr())
			return ErrorType{}
		} else {
			return t1
		}

	case *NewPairCmd:
		return PairType{cxt.DeriveType(expr.Left), cxt.DeriveType(expr.Right)}

	default:
		SemanticError(0, "IMPLEMENT_ME: Unhandled type in DeriveType - Type: %T", expr)
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
		t1, t2 := statement.Type, cxt.DeriveType(statement.Right)
		if !t1.Equals(t2) {
			SemanticError(0, "semantic error - Value being used to initialise '%s' does not match it's type (%s != %s)",
				statement.Ident.Name, t1.Repr(), t2.Repr())
		} else {
			cxt.Add(statement.Ident, statement.Type)
		}

	case *AssignStmt:
		t1, t2 := cxt.DeriveType(statement.Left), cxt.DeriveType(statement.Right)
		if !t1.Equals(t2) {
			SemanticError(0, "semantic error - Cannot assign rvalue to lvalue with a different type (%s != %s)", t1.Repr(), t2.Repr())
		}

	case *IfStmt:
		// Check for boolean condition
		t := cxt.DeriveType(statement.Cond)
		if !t.Equals(BasicType{BOOL}) {
			SemanticError(0, "semantic error - Condition '%s' is not a bool (actual type: %s)", statement.Cond.Repr(), t.Repr())
		}

		// Verify branches
		VerifyStatementListSemantics(cxt, statement.Body)
		VerifyStatementListSemantics(cxt, statement.Else)

	case *WhileStmt:
		// Check the condition
		t := cxt.DeriveType(statement.Cond)
		if !t.Equals(BasicType{BOOL}) {
			SemanticError(0, "semantic error - Condition '%s' is not a bool (actual type: %s)", statement.Cond.Repr(), t.Repr())
		}

		// Verfy body
		VerifyStatementListSemantics(cxt, statement.Body)
	}
}

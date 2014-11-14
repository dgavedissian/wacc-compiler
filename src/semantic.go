package main

// Context
type Context struct {
	functions       map[string]*Function
	currentFunction *Function
	types           map[string]Type
}

//
// Functions
//
func (cxt *Context) LookupFunction(ident *IdentExpr) (*Function, bool) {
	f, ok := cxt.functions[ident.Name]
	return f, ok
}

func (cxt *Context) AddFunction(f *Function) {
	if _, ok := cxt.LookupFunction(f.Ident); ok {
		SemanticError(0, "semantic error -- function '%s' already exists in this program", f.Ident.Name)
	} else {
		cxt.functions[f.Ident.Name] = f
	}
}

//
// Scope
//
func (cxt *Context) PushScope() {
}

func (cxt *Context) LookupVariable(ident *IdentExpr) (Type, bool) {
	t, ok := cxt.types[ident.Name]
	return t, ok
}

func (cxt *Context) AddVariable(t Type, ident *IdentExpr) {
	if _, ok := cxt.LookupVariable(ident); ok {
		SemanticError(0, "semantic error -- variable '%s' already exists in this scope", ident.Name)
	} else {
		cxt.types[ident.Name] = t
	}
}

func (cxt *Context) PopScope() {
}

//
// Derive Type
//
func (cxt *Context) DeriveType(expr Expr) Type {
	switch expr := expr.(type) {
	case *IdentExpr:
		if t, ok := cxt.LookupVariable(expr); !ok {
			SemanticError(0, "semantic error -- use of undeclared variable '%s'", expr.Name)
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
			SemanticError(0, "semantic error -- operand of pair selector must be a pair type (actual: %s)", t.Repr())
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
				SemanticError(0, "semantic error -- all expressions in the array literal must have the same type")
				return ErrorType{}
			}
		}

		// Just return the first elements type
		return ArrayType{t}

	case *UnaryExpr:
		t := cxt.DeriveType(expr.Operand)

		switch expr.Operator {
		case "!":
			expected := BasicType{BOOL}
			if !t.Equals(expected) {
				SemanticError(0, "semantic error -- unexpected operand type (expected: %s, actual: %s)", expected.Repr(), t.Repr())
				return ErrorType{}
			}
			return BasicType{BOOL}

		case "-":
			expected := BasicType{INT}
			if !t.Equals(expected) {
				SemanticError(0, "semantic error -- unexpected operand type (expected: %s, actual: %s)", expected.Repr(), t.Repr())
				return ErrorType{}
			}
			return BasicType{INT}

		case "len":
			expected := ArrayType{AnyType{}}
			if !t.Equals(expected) {
				SemanticError(0, "semantic error -- unexpected operand type (expected: %s, actual: %s)", expected.Repr(), t.Repr())
				return ErrorType{}
			}
			return BasicType{INT}

		case "ord":
			expected := BasicType{CHAR}
			if !t.Equals(expected) {
				SemanticError(0, "semantic error -- unexpected operand type (expected: %s, actual: %s)", expected.Repr(), t.Repr())
				return ErrorType{}
			}
			return BasicType{INT}

		case "chr":
			expected := BasicType{INT}
			if !t.Equals(expected) {
				SemanticError(0, "semantic error -- unexpected operand type (expected: %s, actual: %s)", expected.Repr(), t.Repr())
				return ErrorType{}
			}
			return BasicType{CHAR}

		default:
			SemanticError(0, "IMPLEMENT_ME - operator '%s' unhandled", expr.Operator)
			return ErrorType{}
		}

	case *BinaryExpr:
		t1, t2 := cxt.DeriveType(expr.Left), cxt.DeriveType(expr.Right)

		switch expr.Operator {
		case "*", "/", "%", "+", "-":
			if !t1.Equals(BasicType{INT}) {
				SemanticError(0, "semantic error -- invalid type on left of operator '%s' (expected: INT, actual: %s)", expr.Operator, t1.Repr())
				return ErrorType{}
			}
			if !t2.Equals(BasicType{INT}) {
				SemanticError(0, "semantic error -- invalid type on right of operator '%s' (expected: INT, actual: %s)", expr.Operator, t2.Repr())
				return ErrorType{}
			}
			return BasicType{INT}

		case ">", ">=", "<", "<=":
			if !t1.Equals(BasicType{INT}) && !t1.Equals(BasicType{CHAR}) {
				SemanticError(0, "semantic error -- invalid type on left of operator '%s' (expected: {INT, CHAR}, actual: %s)", expr.Operator, t1.Repr())
				return ErrorType{}
			}
			if !t2.Equals(BasicType{INT}) && !t1.Equals(BasicType{CHAR}) {
				SemanticError(0, "semantic error -- invalid type on right of operator '%s' (expected: {INT, CHAR}, actual: %s)", expr.Operator, t2.Repr())
				return ErrorType{}
			}
			if !t1.Equals(t2) {
				SemanticError(0, "semantic error -- types of the operands of the binary operator '%s' do not match (%s != %s)", expr.Operator, t1.Repr(), t2.Repr())
				return ErrorType{}
			}
			return BasicType{BOOL}

		case "==", "!=":
			if !t1.Equals(t2) {
				SemanticError(0, "semantic error -- types of the operands of the binary operator '%s' do not match (%s != %s)", expr.Operator, t1.Repr(), t2.Repr())
				return ErrorType{}
			}
			return BasicType{BOOL}

		case "&&", "||":
			if !t1.Equals(BasicType{BOOL}) {
				SemanticError(0, "semantic error -- invalid type on left of operator '%s' (expected: BOOL, actual: %s)", expr.Operator, t1.Repr())
				return ErrorType{}
			}
			if !t2.Equals(BasicType{BOOL}) {
				SemanticError(0, "semantic error -- invalid type on right of operator '%s' (expected: BOOL, actual: %s)", expr.Operator, t2.Repr())
				return ErrorType{}
			}
			return BasicType{BOOL}

		default:
			SemanticError(0, "IMPLEMENT_ME - operator '%s' unhandled", expr.Operator)
			return ErrorType{}
		}

	case *NewPairCmd:
		return PairType{cxt.DeriveType(expr.Left), cxt.DeriveType(expr.Right)}

	case *CallCmd:
		if f, ok := cxt.LookupFunction(expr.Ident); ok {
			// Verify number of arguments
			argsLen, paramLen := len(expr.Args), len(f.Params)
			if argsLen != paramLen {
				SemanticError(0, "semantic error -- wrong number of arguments specified (expected: %d, actual: %d)", argsLen, paramLen)
				return ErrorType{}
			}

			// Verify argument types
			for i := 0; i < argsLen; i++ {
				argType, paramType := cxt.DeriveType(expr.Args[i]), f.Params[i].Type
				if !argType.Equals(paramType) {
					SemanticError(0, "semantic error -- parameter type mismatch (expected: %s, actual: %s)", paramType.Repr(), argType.Repr())
					return ErrorType{}
				}
			}

			// Return function type
			return f.Type
		} else {
			SemanticError(0, "semantic error -- use of undefined function '%s'", expr.Ident.Name)
			return ErrorType{}
		}

	default:
		SemanticError(0, "IMPLEMENT_ME: unhandled type in DeriveType - type: %T", expr)
		return ErrorType{}
	}
}

// Semantic Checking
func VerifySemantics(program *ProgStmt) {
	cxt := &Context{make(map[string]*Function), nil, make(map[string]Type)}
	for _, f := range program.Funcs {
		VerifyFunctionSemantics(cxt, f)
	}
	VerifyStatementListSemantics(cxt, program.Body)
}

func VerifyFunctionSemantics(cxt *Context, f *Function) {
	// Add this function
	cxt.AddFunction(f)

	// Verify this function
	cxt.PushScope()
	cxt.currentFunction = f
	VerifyStatementListSemantics(cxt, f.Body)
	cxt.currentFunction = nil
	cxt.PopScope()
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
			SemanticError(0, "semantic error -- value being used to initialise '%s' does not match it's type (%s != %s)",
				statement.Ident.Name, t1.Repr(), t2.Repr())
		} else {
			cxt.AddVariable(statement.Type, statement.Ident)
		}

	case *AssignStmt:
		t1, t2 := cxt.DeriveType(statement.Left), cxt.DeriveType(statement.Right)
		if !t1.Equals(t2) {
			SemanticError(0, "semantic error -- cannot assign rvalue to lvalue with a different type (%s != %s)", t1.Repr(), t2.Repr())
		}

	case *IfStmt:
		// Check the condition
		t := cxt.DeriveType(statement.Cond)
		if !t.Equals(BasicType{BOOL}) {
			SemanticError(0, "semantic error -- condition '%s' is not a bool (actual type: %s)", statement.Cond.Repr(), t.Repr())
		}

		// Verify branches
		VerifyStatementListSemantics(cxt, statement.Body)
		VerifyStatementListSemantics(cxt, statement.Else)

	case *WhileStmt:
		// Check the condition
		t := cxt.DeriveType(statement.Cond)
		if !t.Equals(BasicType{BOOL}) {
			SemanticError(0, "semantic error -- condition '%s' is not a bool (actual type: %s)", statement.Cond.Repr(), t.Repr())
		}

		// Verfy body
		VerifyStatementListSemantics(cxt, statement.Body)
	}
}

package frontend

import "fmt"

// Context
type Context struct {
	functions       map[string]*Function
	currentFunction *Function
	types           []map[string]Type
	depth           int
}

//
// Semantic Checking
//
func verifyProgram(program *Program) {
	ctx := &Context{make(map[string]*Function), nil, nil, 0}

	// Verify functions
	// This needs to be done in two passes. Firstly, add the functions to the
	// function list, then verify the functions afterwards. This is to allow
	// mutual recursion
	for _, f := range program.Funcs {
		ctx.AddFunction(f)
	}
	for _, f := range program.Funcs {
		if !f.External {
			ctx.PushScope()
			ctx.currentFunction = f
			ctx.VerifyStatementList(f.Body)
			ctx.currentFunction = nil
			ctx.PopScope()
		}
	}

	// Verify main
	ctx.PushScope()
	ctx.VerifyStatementList(program.Body)
	ctx.PopScope()
}

//
// Functions
//
func (ctx *Context) LookupFunction(ident *IdentExpr) (*Function, bool) {
	f, ok := ctx.functions[ident.Name]
	return f, ok
}

func (ctx *Context) AddFunction(f *Function) {
	if _, ok := ctx.LookupFunction(f.Ident); ok {
		SemanticError(f.Pos(), "function '%s' already exists in this program", f.Ident.Name)
	} else {
		ctx.functions[f.Ident.Name] = f
	}
}

//
// Scope
//
func (ctx *Context) PushScope() {
	ctx.types = append(ctx.types, make(map[string]Type))
	ctx.depth++
}

func (ctx *Context) LookupVariable(ident *IdentExpr) (Type, bool) {
	var t Type
	var ok bool

	// Search for variables in each scope
	for i := ctx.depth - 1; i >= 0; i-- {
		if t, ok = ctx.types[i][ident.Name]; ok {
			break
		}
	}

	// If the variable does not exist in this scope and we're in a function,
	// then search for a parameter
	if !ok && ctx.currentFunction != nil {
		// Search for a function parameter
		for _, param := range ctx.currentFunction.Params {
			if param.Ident.Name == ident.Name {
				return param.Type, true
			}
		}

		// Give up otherwise
		return nil, false
	} else {
		return t, ok
	}
}

func (ctx *Context) AddVariable(t Type, ident *IdentExpr) {
	if _, ok := ctx.types[ctx.depth-1][ident.Name]; ok {
		SemanticError(ident.Pos(), "variable '%s' already exists in this scope", ident.Name)
	} else {
		ctx.types[ctx.depth-1][ident.Name] = t
	}
}

func (ctx *Context) PopScope() {
	ctx.types = ctx.types[:ctx.depth-1]
	ctx.depth--
}

//
// Derive Type
//
func (ctx *Context) DeriveType(expr Expr) Type {
	switch expr := expr.(type) {
	case *IdentExpr:
		if t, ok := ctx.LookupVariable(expr); !ok {
			SemanticError(expr.Pos(), "use of undeclared variable '%s'", expr.Name)
			return ErrorType{}
		} else {
			return t
		}

	case *ArrayElemExpr:
		t := ctx.DeriveType(expr.Volume) // given a[i] - find a
		if t.Equals(BasicType{STRING}) {
			return BasicType{CHAR}
		}
		if array, ok := t.(ArrayType); ok {
			return array.BaseType
		} else {
			SemanticError(expr.Pos(), "cannot index a value which isn't an array (actual: %s)", t.Repr())
			return ErrorType{}
		}

	case *PairElemExpr:
		t := ctx.DeriveType(expr.Operand)
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
			SemanticError(expr.Pos(), "operand of pair selector must be a pair type (actual: %s)", t.Repr())
			return ErrorType{}
		}

	case *BasicLit:
		return expr.Type

	case *ArrayLit:
		// Check if we've already calculated the type (strings).
		if expr.Type != nil {
			return expr.Type
		}
		// Check if the array has any elements
		if len(expr.Values) == 0 {
			expr.Type = ArrayType{AnyType{}}
			return expr.Type
		}

		// Check that all the types match
		t := ctx.DeriveType(expr.Values[0])
		for i := 1; i < len(expr.Values); i++ {
			if !t.Equals(ctx.DeriveType(expr.Values[i])) {
				SemanticError(expr.Pos(), "all expressions in the array literal must have the same type")
				return ErrorType{}
			}
		}

		// Just return the first elements type
		expr.Type = ArrayType{t}
		return expr.Type

	case *UnaryExpr:
		t := ctx.DeriveType(expr.Operand)
		switch expr.Operator {
		case "!":
			expected := BasicType{BOOL}
			if !t.Equals(expected) {
				SemanticError(expr.Pos(), "unexpected operand type (expected: %s; actual: %s)", expected.Repr(), t.Repr())
				return ErrorType{}
			}
			expr.Type = BasicType{BOOL}
			return expr.Type

		case "-":
			if !t.Equals(BasicType{INT}) && !t.Equals(BasicType{FLOAT}) {
				SemanticError(expr.Pos(), "unexpected operand type (expected: int, float; actual: %s)", t.Repr())
				return ErrorType{}
			}
			expr.Type = t
			return expr.Type

		case "len":
			expected := ArrayType{AnyType{}}
			if !t.Equals(expected) {
				SemanticError(expr.Pos(), "unexpected operand type (expected: %s; actual: %s)", expected.Repr(), t.Repr())
				return ErrorType{}
			}
			expr.Type = BasicType{INT}
			return expr.Type

		case "ord":
			expected := BasicType{CHAR}
			if !t.Equals(expected) {
				SemanticError(expr.Pos(), "unexpected operand type (expected: %s; actual: %s)", expected.Repr(), t.Repr())
				return ErrorType{}
			}
			expr.Type = BasicType{INT}
			return expr.Type

		case "chr":
			expected := BasicType{INT}
			if !t.Equals(expected) {
				SemanticError(expr.Pos(), "unexpected operand type (expected: %s; actual: %s)", expected.Repr(), t.Repr())
				return ErrorType{}
			}
			expr.Type = BasicType{CHAR}
			return expr.Type

		default:
			SemanticError(expr.Pos(), "IMPLEMENT_ME - operator '%s' unhandled", expr.Operator)
			return ErrorType{}
		}

	case *BinaryExpr:
		t1, t2 := ctx.DeriveType(expr.Left), ctx.DeriveType(expr.Right)

		switch expr.Operator {
		case "*", "/", "%", "+", "-":
			if !t1.Equals(BasicType{INT}) && !t1.Equals(BasicType{FLOAT}) {
				SemanticError(expr.Pos(), "invalid type on left of operator '%s' (expected: int, float; actual: %s)", expr.Operator, t1.Repr())
				return ErrorType{}
			}
			if !t2.Equals(BasicType{INT}) && !t1.Equals(BasicType{FLOAT}) {
				SemanticError(expr.Pos(), "invalid type on right of operator '%s' (expected: int, float; actual: %s)", expr.Operator, t2.Repr())
				return ErrorType{}
			}
			expr.Type = t1
			return expr.Type

		case ">", ">=", "<", "<=":
			if !t1.Equals(BasicType{INT}) && !t1.Equals(BasicType{FLOAT}) && !t1.Equals(BasicType{CHAR}) {
				SemanticError(expr.Pos(), "invalid type on left of operator '%s' (expected: int, float, char; actual: %s)", expr.Operator, t1.Repr())
				return ErrorType{}
			}
			if !t2.Equals(BasicType{INT}) && !t2.Equals(BasicType{FLOAT}) && !t2.Equals(BasicType{CHAR}) {
				SemanticError(expr.Pos(), "invalid type on right of operator '%s' (expected: int, float, char; actual: %s)", expr.Operator, t2.Repr())
				return ErrorType{}
			}
			if !t1.Equals(t2) {
				SemanticError(expr.Pos(), "operand types for '%s' do not match (%s does not match %s)", expr.Operator, t1.Repr(), t2.Repr())
				return ErrorType{}
			}
			expr.Type = BasicType{BOOL}
			return expr.Type

		case "==", "!=":
			if !t1.Equals(t2) {
				SemanticError(expr.Pos(), "operand types for '%s' do not match (%s does not match %s)", expr.Operator, t1.Repr(), t2.Repr())
				return ErrorType{}
			}
			expr.Type = BasicType{BOOL}
			return expr.Type

		case "&&", "||":
			if !t1.Equals(BasicType{BOOL}) {
				SemanticError(expr.Pos(), "invalid type on left of operator '%s' (expected: bool; actual: %s)", expr.Operator, t1.Repr())
				return ErrorType{}
			}
			if !t2.Equals(BasicType{BOOL}) {
				SemanticError(expr.Pos(), "invalid type on right of operator '%s' (expected: bool; actual: %s)", expr.Operator, t2.Repr())
				return ErrorType{}
			}
			expr.Type = BasicType{BOOL}
			return expr.Type

		default:
			SemanticError(expr.Pos(), "IMPLEMENT_ME - operator '%s' unhandled", expr.Operator)
			return ErrorType{}
		}

	case *NewPairCmd:
		return PairType{ctx.DeriveType(expr.Left), ctx.DeriveType(expr.Right)}

	case *NewStructCmd:
		return StructType{expr.Ident.Name}

	case *CallCmd:
		if f, ok := ctx.LookupFunction(expr.Ident); ok {
			// Verify number of arguments
			argsLen, paramLen := len(expr.Args), len(f.Params)
			if argsLen != paramLen {
				SemanticError(expr.Pos(), "wrong number of arguments to '%s' specified (expected: %d; actual: %d)", f.Ident.Name, argsLen, paramLen)
				return ErrorType{}
			}

			// Verify argument types
			for i := 0; i < argsLen; i++ {
				argType, paramType := ctx.DeriveType(expr.Args[i]), f.Params[i].Type
				if !argType.Equals(paramType) {
					SemanticError(expr.Pos(), "parameter type mismatch (expected: %s; actual: %s)", paramType.Repr(), argType.Repr())
					return ErrorType{}
				}
			}

			// Return function type
			return f.Type
		} else {
			SemanticError(expr.Pos(), "use of undefined function '%s'", expr.Ident.Name)
			return ErrorType{}
		}

	default:
		SemanticError(expr.Pos(), "IMPLEMENT_ME: unhandled type in DeriveType - type: %T", expr)
		return ErrorType{}
	}
}

//
// Verify Statements
//
func (ctx *Context) VerifyStatementList(statementList []Stmt) {
	for _, s := range statementList {
		ctx.VerifyStatement(s)
	}
}

func (ctx *Context) VerifyStatement(statement Stmt) {
	switch statement := statement.(type) {
	case *SkipStmt:
		// Do nothing

	case *EvalStmt:
		ctx.DeriveType(statement.Expr)

	case *DeclStmt:
		t1, t2 := statement.Type, ctx.DeriveType(statement.Right)
		if !t1.Equals(t2) {
			SemanticError(statement.Pos(), "value being used to initialise '%s' does not match its declared type (%s does not match %s)",
				statement.Ident.Name, t1.Repr(), t2.Repr())
		} else {
			ctx.AddVariable(statement.Type, statement.Ident)
		}

	case *AssignStmt:
		t1, t2 := ctx.DeriveType(statement.Left), ctx.DeriveType(statement.Right)
		if !t1.Equals(t2) {
			SemanticError(statement.Pos(), "cannot assign rvalue to lvalue with a different type (%s does not match %s)", t1.Repr(), t2.Repr())
		}

	case *ReadStmt:
		t := ctx.DeriveType(statement.Dst)
		if !t.Equals(BasicType{INT}) && !t.Equals(BasicType{CHAR}) {
			SemanticError(statement.Dst.Pos(), "destination of read has incorrect type (expected: int or char; actual: %s)", t.Repr())
		}
		statement.Type = t

	case *FreeStmt:
		t := ctx.DeriveType(statement.Object)
		if !t.Equals(PairType{AnyType{}, AnyType{}}) && !t.Equals(ArrayType{AnyType{}}) {
			SemanticError(statement.Object.Pos(), "object being freed must be either a pair or an array (actual: %s)", t.Repr())
		}

	case *ExitStmt:
		t := ctx.DeriveType(statement.Result)
		if !t.Equals(BasicType{INT}) {
			SemanticError(statement.Result.Pos(), "incorrect type in exit statement (expected: int; actual: %s)", t.Repr())
		}

	case *ReturnStmt:
		// Check if we're in a function
		if ctx.currentFunction == nil {
			SemanticError(statement.Pos(), "cannot call return in the program body")
		} else {
			// Check if the type of the operand matches the return type
			t := ctx.DeriveType(statement.Result)
			if !t.Equals(ctx.currentFunction.Type) {
				SemanticError(statement.Result.Pos(), "type in return statement must match the return type of the function (expected: %s; actual: %s)",
					ctx.currentFunction.Type.Repr(), t.Repr())
			}
		}

	case *PrintStmt:
		// Verify expression by attempting to derive the type
		statement.Type = ctx.DeriveType(statement.Right)

	case *IfStmt:
		// Check the condition
		t := ctx.DeriveType(statement.Cond)
		if !t.Equals(BasicType{BOOL}) {
			SemanticError(statement.Cond.Pos(), "condition type is incorrect (expected: bool; actual: %s)", t.Repr())
		}

		// Verify true branch
		ctx.PushScope()
		ctx.VerifyStatementList(statement.Body)
		ctx.PopScope()

		// Verify false branch
		ctx.PushScope()
		ctx.VerifyStatementList(statement.Else)
		ctx.PopScope()

	case *WhileStmt:
		// Check the condition
		t := ctx.DeriveType(statement.Cond)
		if !t.Equals(BasicType{BOOL}) {
			SemanticError(statement.Cond.Pos(), "condition type is incorrect (expected: bool; actual: %s)", t.Repr())
		}

		// Verfy body
		ctx.PushScope()
		ctx.VerifyStatementList(statement.Body)
		ctx.PopScope()

	case *ScopeStmt:
		ctx.PushScope()
		ctx.VerifyStatementList(statement.Body)
		ctx.PopScope()

	default:
		panic(fmt.Sprintf("IMPLEMENT_ME: Unchecked statement: %T", statement))
	}
}

package frontend

import (
	"fmt"
	"strings"
)

// Context
type Context struct {
	structs         map[string]*Struct
	functions       map[string]*Function
	currentFunction *Function
	types           []map[string]Type
	depth           int
	err             bool
}

//
// Name manging
//
func (ctx *Context) encodeType(t Type) string {
	switch t := t.(type) {
	case BasicType:
		if t.TypeId == STRING {
			// Match char[]
			return "ac"
		} else {
			return string([]rune(t.Repr())[0])
		}

	case ArrayType:
		return "a" + ctx.encodeType(t.BaseType)

	case PairType:
		// We can't encode the sub-pair types here in case null is given
		return "p"

	default:
		panic(fmt.Sprintf("Unhandled type in encodeType: %T", t))
	}
}

func (ctx *Context) encodeFunctionName(ident *IdentExpr, types []Type) string {
	name := ident.Name
	for _, t := range types {
		name += ctx.encodeType(t)
	}
	return name
}

func (ctx *Context) paramsToTypes(params []Param) []Type {
	out := []Type{}
	for _, p := range params {
		out = append(out, p.Type)
	}
	return out
}

func (ctx *Context) genTypeSignature(ident string, types []Type) string {
	typeList := []string{}
	for _, t := range types {
		typeList = append(typeList, t.Repr())
	}
	return fmt.Sprintf("%v(%v)", ident, strings.Join(typeList, ", "))
}

//
// Semantic Checking
//
func VerifyProgram(program *Program) bool {
	ctx := &Context{make(map[string]*Struct), make(map[string]*Function), nil, nil, 0, false}

	// Add structs to the context to ensureeeach struct has a unique identifier
	// and so we can lookup structs later
	for _, s := range program.Structs {
		ctx.AddStruct(s)
	}

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

	// Return true if okay
	return !ctx.err
}

//
// Structs
//

func (ctx *Context) LookupStruct(name string) (*Struct, bool) {
	s, ok := ctx.structs[name]
	return s, ok
}

func (ctx *Context) AddStruct(s *Struct) {
	name := s.Ident.Name
	if _, ok := ctx.LookupStruct(name); ok {
		SemanticError(s.Pos(), "struct '%v' already exists in this program", name)
		ctx.err = true
	} else {
		ctx.structs[name] = s
	}
}

//
// Functions
//
func (ctx *Context) LookupFunction(name string) (*Function, bool) {
	f, ok := ctx.functions[name]
	return f, ok
}

func (ctx *Context) AddFunction(f *Function) {
	types := ctx.paramsToTypes(f.Params)
	originalName := f.Ident.Name
	if !f.External {
		f.Ident.Name = ctx.encodeFunctionName(f.Ident, types)
	}

	// Lookup function
	var ok bool
	if !f.External {
		_, ok = ctx.LookupFunction(f.Ident.Name)
	} else {
		_, ok = ctx.LookupFunction(f.Ident.Name)
	}

	if ok {
		SemanticError(f.Pos(), "function '%v' already exists in this program", ctx.genTypeSignature(originalName, types))
		ctx.err = true
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
		SemanticError(ident.Pos(), "variable '%v' already exists in this scope", ident.Name)
		ctx.err = true
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
			SemanticError(expr.Pos(), "use of undeclared variable '%v'", expr.Name)
			ctx.err = true
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
			SemanticError(expr.Pos(), "cannot index a value which isn't an array (actual: %v)", t.Repr())
			ctx.err = true
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
			SemanticError(expr.Pos(), "operand of pair selector must be a pair type (actual: %v)", t.Repr())
			ctx.err = true
			return ErrorType{}
		}

	case *StructElemExpr:
		// StructElemExpr is of the form x.y
		t := ctx.DeriveType(expr.StructIdent)

		// Check x has a struct type
		if st, ok := t.(StructType); ok {

			// Check that x is a struct that exists
			if s, ok := ctx.LookupStruct(st.TypeId); ok {

				// Check that y is a member of x's
				for i, m := range s.Members {
					if m.Ident.Name == expr.ElemIdent.Name {
						expr.ElemNum = i
						return m.Type
					}
				}
				SemanticError(expr.ElemIdent.Pos(), "the struct %v does not contain member %v",
					expr.Repr(), expr.ElemIdent.Repr())
				ctx.err = true
				return ErrorType{}

			} else {
				SemanticError(expr.Pos(), "no such struct exists: %v", expr.Repr())
				ctx.err = true
				return ErrorType{}
			}
		} else {
			SemanticError(expr.Pos(), "can only access members from struct type (actual: %v)", t.Repr())
			ctx.err = true
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
				ctx.err = true
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
				SemanticError(expr.Pos(), "unexpected operand type (expected: %v; actual: %v)", expected.Repr(), t.Repr())
				ctx.err = true
				return ErrorType{}
			}
			expr.Type = BasicType{BOOL}
			return expr.Type

		case "-":
			if !t.Equals(BasicType{INT}) && !t.Equals(BasicType{FLOAT}) {
				SemanticError(expr.Pos(), "unexpected operand type (expected: int, float; actual: %v)", t.Repr())
				ctx.err = true
				return ErrorType{}
			}
			expr.Type = t
			return expr.Type

		case "len":
			expected := ArrayType{AnyType{}}
			if !t.Equals(expected) {
				SemanticError(expr.Pos(), "unexpected operand type (expected: %v; actual: %v)", expected.Repr(), t.Repr())
				ctx.err = true
				return ErrorType{}
			}
			expr.Type = BasicType{INT}
			return expr.Type

		case "ord":
			expected := BasicType{CHAR}
			if !t.Equals(expected) {
				SemanticError(expr.Pos(), "unexpected operand type (expected: %v; actual: %v)", expected.Repr(), t.Repr())
				ctx.err = true
				return ErrorType{}
			}
			expr.Type = BasicType{INT}
			return expr.Type

		case "chr":
			expected := BasicType{INT}
			if !t.Equals(expected) {
				SemanticError(expr.Pos(), "unexpected operand type (expected: %v; actual: %v)", expected.Repr(), t.Repr())
				ctx.err = true
				return ErrorType{}
			}
			expr.Type = BasicType{CHAR}
			return expr.Type

		default:
			SemanticError(expr.Pos(), "IMPLEMENT_ME - operator '%v' unhandled", expr.Operator)
			ctx.err = true
			return ErrorType{}
		}

	case *BinaryExpr:
		t1, t2 := ctx.DeriveType(expr.Left), ctx.DeriveType(expr.Right)

		switch expr.Operator {
		case "*", "/", "%", "+", "-":
			if !t1.Equals(BasicType{INT}) && !t1.Equals(BasicType{FLOAT}) {
				SemanticError(expr.Pos(), "invalid type on left of operator '%v' (expected: int, float; actual: %v)", expr.Operator, t1.Repr())
				ctx.err = true
				return ErrorType{}
			}
			if !t2.Equals(t1) {
				SemanticError(expr.Pos(), "invalid type on right of operator '%v' (expected: %v; actual: %v)", expr.Operator, t1.Repr(), t2.Repr())
				ctx.err = true
				return ErrorType{}
			}
			expr.Type = t1
			return expr.Type

		case ">", ">=", "<", "<=":
			if !t1.Equals(BasicType{INT}) && !t1.Equals(BasicType{FLOAT}) && !t1.Equals(BasicType{CHAR}) {
				SemanticError(expr.Pos(), "invalid type on left of operator '%v' (expected: int, float, char; actual: %v)", expr.Operator, t1.Repr())
				ctx.err = true
				return ErrorType{}
			}
			if !t2.Equals(t1) {
				SemanticError(expr.Pos(), "invalid type on right of operator '%v' (expected: %v; actual: %v)", expr.Operator, t1.Repr(), t2.Repr())
				ctx.err = true
				return ErrorType{}
			}
			expr.Type = BasicType{BOOL}
			return expr.Type

		case "==", "!=":
			if !t1.Equals(t2) {
				SemanticError(expr.Pos(), "operand types for '%v' do not match (%v does not match %v)", expr.Operator, t1.Repr(), t2.Repr())
				ctx.err = true
				return ErrorType{}
			}
			expr.Type = BasicType{BOOL}
			return expr.Type

		case "&&", "||":
			if !t1.Equals(BasicType{BOOL}) {
				SemanticError(expr.Pos(), "invalid type on left of operator '%v' (expected: bool; actual: %v)", expr.Operator, t1.Repr())
				ctx.err = true
				return ErrorType{}
			}
			if !t2.Equals(BasicType{BOOL}) {
				SemanticError(expr.Pos(), "invalid type on right of operator '%v' (expected: bool; actual: %v)", expr.Operator, t2.Repr())
				ctx.err = true
				return ErrorType{}
			}
			expr.Type = BasicType{BOOL}
			return expr.Type

		default:
			SemanticError(expr.Pos(), "IMPLEMENT_ME - operator '%v' unhandled", expr.Operator)
			ctx.err = true
			return ErrorType{}
		}

	case *NewPairCmd:
		return PairType{ctx.DeriveType(expr.Left), ctx.DeriveType(expr.Right)}

	case *NewStructCmd:
		return StructType{expr.Ident.Name}

	case *CallCmd:
		// Derive parameter types
		paramTypes := []Type{}
		for _, e := range expr.Args {
			t := ctx.DeriveType(e)
			// If the error type was returned, bail
			if !t.Equals(t) {
				return ErrorType{}
			}
			paramTypes = append(paramTypes, t)
		}

		// Encode function name
		originalName := expr.Ident.Name
		encodedName := ctx.encodeFunctionName(expr.Ident, paramTypes)
		f, ok := ctx.LookupFunction(encodedName)
		if ok {
			expr.Ident.Name = encodedName
		} else {
			// If not found, try again with original name
			f, ok = ctx.LookupFunction(expr.Ident.Name)
		}
		if ok {
			// Verify number of arguments
			argsLen, paramLen := len(expr.Args), len(f.Params)
			if argsLen != paramLen {
				SemanticError(expr.Pos(), "wrong number of arguments to '%v' specified (expected: %v; actual: %v)", originalName, argsLen, paramLen)
				ctx.err = true
				return ErrorType{}
			}

			// Verify argument types
			for i := 0; i < argsLen; i++ {
				argType, paramType := paramTypes[i], f.Params[i].Type
				if !argType.Equals(paramType) {
					SemanticError(expr.Pos(), "parameter type mismatch (expected: %v; actual: %v)", paramType.Repr(), argType.Repr())
					ctx.err = true
					return ErrorType{}
				}
			}

			// Return function type
			return f.Type
		} else {
			SemanticError(expr.Pos(), "use of undefined function '%v'", ctx.genTypeSignature(originalName, paramTypes))
			ctx.err = true
			return ErrorType{}
		}

	default:
		SemanticError(expr.Pos(), "IMPLEMENT_ME: unhandled type in DeriveType - type: %T", expr)
		ctx.err = true
		return ErrorType{}
	}
	return ErrorType{}
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
			SemanticError(statement.Pos(), "value being used to initialise '%v' does not match its declared type (%v does not match %v)",
				statement.Ident.Name, t1.Repr(), t2.Repr())
			ctx.err = true
		} else {
			ctx.AddVariable(statement.Type, statement.Ident)
		}

	case *AssignStmt:
		t1, t2 := ctx.DeriveType(statement.Left), ctx.DeriveType(statement.Right)
		if !t1.Equals(t2) {
			SemanticError(statement.Pos(), "cannot assign rvalue to lvalue with a different type (%v does not match %v)", t1.Repr(), t2.Repr())
			ctx.err = true
		}

	case *ReadStmt:
		t := ctx.DeriveType(statement.Dst)
		if !t.Equals(BasicType{INT}) && !t.Equals(BasicType{CHAR}) {
			SemanticError(statement.Dst.Pos(), "destination of read has incorrect type (expected: int or char; actual: %v)", t.Repr())
			ctx.err = true
		}
		statement.Type = t

	case *FreeStmt:
		t := ctx.DeriveType(statement.Object)
		if !t.Equals(PairType{AnyType{}, AnyType{}}) && !t.Equals(ArrayType{AnyType{}}) {
			SemanticError(statement.Object.Pos(), "object being freed must be either a pair or an array (actual: %v)", t.Repr())
			ctx.err = true
		}

	case *ExitStmt:
		t := ctx.DeriveType(statement.Result)
		if !t.Equals(BasicType{INT}) {
			SemanticError(statement.Result.Pos(), "incorrect type in exit statement (expected: int; actual: %v)", t.Repr())
			ctx.err = true
		}

	case *ReturnStmt:
		// Check if we're in a function
		if ctx.currentFunction == nil {
			SemanticError(statement.Pos(), "cannot call return in the program body")
			ctx.err = true
		} else {
			// Check if the type of the operand matches the return type
			t := ctx.DeriveType(statement.Result)
			if !t.Equals(ctx.currentFunction.Type) {
				SemanticError(statement.Result.Pos(), "type in return statement must match the return type of the function (expected: %v; actual: %v)",
					ctx.currentFunction.Type.Repr(), t.Repr())
				ctx.err = true
			}
		}

	case *PrintStmt:
		// Verify expression by attempting to derive the type
		statement.Type = ctx.DeriveType(statement.Right)

	case *IfStmt:
		// Check the condition
		t := ctx.DeriveType(statement.Cond)
		if !t.Equals(BasicType{BOOL}) {
			SemanticError(statement.Cond.Pos(), "condition type is incorrect (expected: bool; actual: %v)", t.Repr())
			ctx.err = true
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
			SemanticError(statement.Cond.Pos(), "condition type is incorrect (expected: bool; actual: %v)", t.Repr())
			ctx.err = true
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

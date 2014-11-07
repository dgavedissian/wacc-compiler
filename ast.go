package main

import (
	"go/token"
	"strconv"
	"strings"
)

type Pos token.Pos
type Position token.Position

type Node interface {
	Pos() Pos     // Position of first character belonging to the node
	End() Pos     // Position of first character immediately after the node
	Repr() string // Representation
}

type Expr interface {
	Node
	exprNode()
}

type Stmt interface {
	Node
	stmtNode()
}

type Prog interface {
	Node
	progNode()
}

type Ident struct {
	NamePos Pos
	Name    string
}

type BasicLit struct {
	ValuePos Pos // Literal position
	Kind     int // Token kind (e.g. INT_LITER, CHAR_LITER)
	Value    string
}

type ArrayLit struct {
	ValuesPos Pos
	Values    []Expr
}

type PairExpr struct {
	ValuePos  Pos
	LeftKind  int
	LeftExpr  Expr
	RightKind int
	RightExpr Expr
}

type UnaryExpr struct {
	OperatorPos Pos
	Operator    string
	Operand     Expr
}

type BinaryExpr struct {
	Left        Expr
	OperatorPos Pos
	Operator    string
	Right       Expr
}

type ArrayIndexExpr struct {
	IdentPos Pos
	Ident    Ident
	Index    Expr
}

type CallExpr struct {
	Call  Pos
	Ident Ident
	Args  []Expr
}

type ProgStmt struct {
	BeginKw Pos // position of "begin" keyword
	Funcs   []Func
	Body    []Stmt
	EndKw   Pos // position of "end keyword
}

type SkipStmt struct {
	Skip Pos // position of "skip" keyword
}

type DeclStmt struct {
	TypeKw Pos // Position of the type keyword
	Kind   int
	Ident  Ident
	Right  Expr
}

type AssignStmt struct {
	Ident Ident
	Right Expr
}

/* TODO:
 * Pair element assign e.g. fst p = 1
 * Array element assign e.g. xs[0] = 1
 */

type ExitStmt struct {
	Exit   Pos  // position of "exit" keyword
	Result Expr // result expression
}

type ReturnStmt struct {
	Return Pos
	Result Expr
}

type PrintStmt struct {
	Print   Pos // position of print keyword
	Right   Expr
	NewLine bool
}

type Func struct {
	Func   Pos
	Kind   int
	Ident  Ident
	Params []Param
	Stmts  []Stmt
}

type Param struct {
	Start  Pos
	Kind   int
	Ident  Ident
	Finish Pos
}

type IfStmt struct {
	If   Pos
	Cond Expr
	Body []Stmt
	Else []Stmt
	Fi   Pos
}

type WhileStmt struct {
	While Pos
	Cond  Expr
	Body  []Stmt
	Done  Pos
}

// Repr helpers
// David: Can't make this general to []Node :( @Luke help?
func ReprFuncs(funcList []Func) string {
	funcs := []string{}
	for _, f := range funcList {
		funcs = append(funcs, f.Repr())
	}
	return strings.Join(funcs, ", ")
}

func ReprParams(paramList []Param) string {
	params := []string{}
	for _, p := range paramList {
		params = append(params, p.Repr())
	}
	return strings.Join(params, ", ")
}

func ReprStmts(stmtList []Stmt) string {
	statements := []string{}
	for _, s := range stmtList {
		if s != nil {
			statements = append(statements, s.Repr())
		}
	}
	return strings.Join(statements, ", ")
}

func ReprExprs(exprList []Expr) string {
	exprs := []string{}
	for _, x := range exprList {
		if x != nil {
			exprs = append(exprs, x.Repr())
		}
	}
	return strings.Join(exprs, ", ")
}

// Identifier
func (*Ident) exprNode()  {}
func (x *Ident) Pos() Pos { return x.NamePos }
func (x *Ident) End() Pos { return Pos(int(x.NamePos) + len(x.Name)) }
func (x *Ident) Repr() string {
	if x.Name == "" {
		return "Ident(Bad name)"
	}
	return "Ident(" + x.Name + ")"
}

// Basic Literal
func (*BasicLit) exprNode()  {}
func (x *BasicLit) Pos() Pos { return x.ValuePos }
func (x *BasicLit) End() Pos { return Pos(int(x.ValuePos) + len(x.Value)) }
func (x *BasicLit) Repr() string {
	return "Lit(" + strconv.Itoa(x.Kind) + ", " + x.Value + ")"
}

// Array literal
func (*ArrayLit) exprNode()  {}
func (x *ArrayLit) Pos() Pos { return x.ValuesPos }
func (x *ArrayLit) End() Pos {
	if x.Values == nil {
		return Pos(int(x.ValuesPos) + 1) /* CLose bracket */
	}
	return Pos(int(x.Values[len(x.Values)-1].End()) + 1)
}
func (x *ArrayLit) Repr() string {
	if x.Values == nil {
		return "ArrayLit([])"
	}
	return "ArrayLit([" + ReprExprs(x.Values) + "])"
}

// Pairs
func (*PairExpr) exprNode()  {}
func (x *PairExpr) Pos() Pos { return x.ValuePos }
func (x *PairExpr) End() Pos {
	return Pos(int(x.ValuePos) + len(x.RightExpr.Repr()) + 1) // Right bracket
}
func (x *PairExpr) Repr() string {
	if x.LeftExpr == nil || x.RightExpr == nil {
		return "Pair()"
	}
	return "Pair(" + strconv.Itoa(x.LeftKind) + ", " + x.LeftExpr.Repr() + strconv.Itoa(x.RightKind) + ", " + x.RightExpr.Repr() + ")"
}

// Unary Expression
func (*UnaryExpr) exprNode()  {}
func (x *UnaryExpr) Pos() Pos { return x.OperatorPos }
func (x *UnaryExpr) End() Pos { return x.Operand.End() }
func (x *UnaryExpr) Repr() string {
	if x.Operand == nil {
		return "Unary(" + x.Operator + ", )"
	}
	return "Unary(" + x.Operator + ", " + x.Operand.Repr() + ")"
}

// Binary Expression
func (*BinaryExpr) exprNode()  {}
func (x *BinaryExpr) Pos() Pos { return x.Left.Pos() }
func (x *BinaryExpr) End() Pos { return x.Right.End() }
func (x *BinaryExpr) Repr() string {
	if x.Left == nil || x.Right == nil {
		return "Binary(" + x.Operator + ", , )"
	}
	return "Binary(" + x.Operator + ", " +
		x.Left.Repr() + ", " + x.Right.Repr() + ")"
}

// Array index expression
func (*ArrayIndexExpr) exprNode()  {}
func (x *ArrayIndexExpr) Pos() Pos { return x.IdentPos }
func (x *ArrayIndexExpr) End() Pos {
	return x.Index.End() + 1 /* Close bracket*/
}
func (x *ArrayIndexExpr) Repr() string {
	return "ArrayIndex(" + x.Ident.Repr() + ", " + x.Index.Repr() + ")"
}

// Function call
func (*CallExpr) exprNode()  {}
func (x *CallExpr) Pos() Pos { return x.Call }
func (x *CallExpr) End() Pos {
	return x.Call /* TODO */
}
func (x *CallExpr) Repr() string {
	return "Call(" + x.Ident.Repr() + ", " + ReprExprs(x.Args) + ")"
}

// Program Statement
func (*ProgStmt) stmtNode()  {}
func (s *ProgStmt) Pos() Pos { return s.BeginKw }
func (s *ProgStmt) End() Pos {
	return s.EndKw + Pos(len("end"))
}
func (s *ProgStmt) Repr() string {
	return "Prog(" + ReprFuncs(s.Funcs) + ")(" +
		ReprStmts(s.Body) + ")"
}

// Skip Statement
func (*SkipStmt) stmtNode()  {}
func (s *SkipStmt) Pos() Pos { return s.Skip }
func (s *SkipStmt) End() Pos {
	return s.Skip + Pos(len("skip"))
}
func (s *SkipStmt) Repr() string { return "Skip" }

// Declaration statement
func (s *DeclStmt) stmtNode() {}
func (s *DeclStmt) Pos() Pos  { return s.TypeKw }
func (s *DeclStmt) End() Pos  { return s.Pos() } // TODO
func (s *DeclStmt) Repr() string {
	if s.Right == nil {
		return "Decl(" + strconv.Itoa(s.Kind) + " " + s.Ident.Repr() + ", )"
	}
	return "Decl(" + strconv.Itoa(s.Kind) + " " + s.Ident.Repr() + ", " + s.Right.Repr() + ")"
}

// Assign Statement
func (s *AssignStmt) stmtNode() {}
func (s *AssignStmt) Pos() Pos  { return s.Ident.Pos() }
func (s *AssignStmt) End() Pos  { return s.Pos() } // TODO
func (s *AssignStmt) Repr() string {
	if s.Right == nil {
		return "Assign(" + s.Ident.Repr() + ", )"
	}
	return "Assign(" + s.Ident.Repr() + ", " + s.Right.Repr() + ")"
}

// Exit Statement
func (*ExitStmt) stmtNode()  {}
func (s *ExitStmt) Pos() Pos { return s.Exit }
func (s *ExitStmt) End() Pos {
	if s.Result != nil {
		return s.Result.End()
	}
	return s.Exit + Pos(len("exit"))
}
func (s *ExitStmt) Repr() string {
	if s.Result != nil {
		return "Exit(" + s.Result.Repr() + ")"
	} else {
		return "Exit()"
	}
}

// Return Statement
func (*ReturnStmt) stmtNode()  {}
func (s *ReturnStmt) Pos() Pos { return s.Return }
func (s *ReturnStmt) End() Pos {
	if s.Result != nil {
		return s.Result.End()
	}
	return s.Return + Pos(len("return"))
}
func (s *ReturnStmt) Repr() string {
	if s.Result != nil {
		return "Return(" + s.Result.Repr() + ")"
	} else {
		return "Return()"
	}
}

// Print Statement
func (*PrintStmt) stmtNode()  {}
func (s *PrintStmt) Pos() Pos { return s.Print }
func (s *PrintStmt) End() Pos { return s.Pos() }
func (s *PrintStmt) Repr() string {
	var v string
	if s.Right == nil {
		v = ""
	} else {
		v = s.Right.Repr()
	}
	if s.NewLine {
		return "Println(" + v + ")"
	} else {
		return "Print(" + v + ")"
	}
}

// If Statement
func (*IfStmt) stmtNode()  {}
func (s *IfStmt) Pos() Pos { return s.If }
func (s *IfStmt) End() Pos {
	return s.Fi + Pos(len("Fi"))
}
func (s *IfStmt) Repr() string {
	return "If(" + s.Cond.Repr() +
		")Then(" + ReprStmts(s.Body) +
		")Else(" + ReprStmts(s.Else) + ")"
}

// While Statement
func (*WhileStmt) stmtNode()  {}
func (s *WhileStmt) Pos() Pos { return s.While }
func (s *WhileStmt) End() Pos {
	return s.Done + Pos(len("Done"))
}
func (s *WhileStmt) Repr() string {
	return "While(" + s.Cond.Repr() +
		")Do(" + ReprStmts(s.Body) +
		")Done"
}

// Function Statement
func (s *Func) Pos() Pos { return s.Func }
func (s *Func) End() Pos {
	return s.Stmts[len(s.Stmts)-1].End()
}
func (s *Func) Repr() string {
	return "Func(type:" + strconv.Itoa(s.Kind) +
		", name:" + s.Ident.Repr() +
		", params:(" + ReprParams(s.Params) +
		"), body:(" + ReprStmts(s.Stmts) + ")"
}

// Function Parameter
func (s *Param) Pos() Pos { return s.Start }
func (s *Param) End() Pos { return s.Finish }
func (s *Param) Repr() string {
	return "Param(" + strconv.Itoa(s.Kind) + ", " + s.Ident.Repr() + ")"
}

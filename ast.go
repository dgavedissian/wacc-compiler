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

type BasicLit struct {
	ValuePos Pos // Literal position
	Kind     int // Token kind (e.g. INT_LITER, CHAR_LITER)
	Value    string
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
	Type   string
	Ident  string
	Right  Expr
}

type AssignStmt struct {
	IdentKw Pos // Position of the identifier
	Ident   string
	Right   Expr
}

type ExitStmt struct {
	Exit   Pos  // position of "exit" keyword
	Result Expr // result expression
}

type Func struct {
	Func       Pos
	ReturnType string
	Name       string
	Params     []Param
	Stmts      []Stmt
}

type Param struct {
	Start  Pos
	Type   string
	Name   string
	Finish Pos
}

type IfStmt struct {
	If   Pos
	Cond Expr
	Body []Stmt
	Else []Stmt
	Fi   Pos
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
		statements = append(statements, s.Repr())
	}
	return strings.Join(statements, ", ")
}

// Basic Literal
func (*BasicLit) exprNode()  {}
func (x *BasicLit) Pos() Pos { return x.ValuePos }
func (x *BasicLit) End() Pos { return Pos(int(x.ValuePos) + len(x.Value)) }
func (x *BasicLit) Repr() string {
	return "Lit(" + strconv.Itoa(x.Kind) + ", " + x.Value + ")"
}

// Unary Expression
func (*UnaryExpr) exprNode()  {}
func (x *UnaryExpr) Pos() Pos { return x.OperatorPos }
func (x *UnaryExpr) End() Pos { return x.Operand.End() }
func (x *UnaryExpr) Repr() string {
	return "Unary(" + x.Operator + ", " + x.Operand.Repr()
}

// Binary Expression
func (*BinaryExpr) exprNode()  {}
func (x *BinaryExpr) Pos() Pos { return x.Left.Pos() }
func (x *BinaryExpr) End() Pos { return x.Right.End() }
func (x *BinaryExpr) Repr() string {
	return "Binary(" + x.Operator + ", " +
		x.Left.Repr() + ", " + x.Right.Repr()
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
func (self *DeclStmt) stmtNode() {}
func (self *DeclStmt) Pos() Pos  { return self.TypeKw }
func (self *DeclStmt) End() Pos  { return self.Pos() } // TODO
func (self *DeclStmt) Repr() string {
	return "Decl(" + self.Type + " " + self.Ident + ", " + self.Right.Repr() + ")"
}

// Assign Statement
func (self *AssignStmt) stmtNode() {}
func (self *AssignStmt) Pos() Pos  { return self.IdentKw }
func (self *AssignStmt) End() Pos  { return self.Pos() } // TODO
func (self *AssignStmt) Repr() string {
	return "Assign(" + self.Ident + ", " + self.Right.Repr() + ")"
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
	return "Exit(" + s.Result.Repr() + ")"
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

// Function Statement
func (s *Func) Pos() Pos { return s.Func }
func (s *Func) End() Pos {
	return s.Stmts[len(s.Stmts)-1].End()
}
func (s *Func) Repr() string {
	return "Func(name:" + s.Name +
		", params:(" + ReprParams(s.Params) +
		"), body:(" + ReprStmts(s.Stmts) + ")"
}

// Function Parameter
func (s *Param) Pos() Pos { return s.Start }
func (s *Param) End() Pos { return s.Finish }
func (s *Param) Repr() string {
	return "Param(" + s.Type + ", " + s.Name + ")"
}

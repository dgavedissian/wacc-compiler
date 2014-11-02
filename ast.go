package main

import "go/token"

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

type ExitStmt struct {
	Exit   Pos  // position of "exit" keyword
	Result Expr // result expression
}

type SkipStmt struct {
	Skip Pos // position of "skip" keyword
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

// Basic Literal
func (*BasicLit) exprNode()      {}
func (x *BasicLit) Pos() Pos     { return x.ValuePos }
func (x *BasicLit) End() Pos     { return Pos(int(x.ValuePos) + len(x.Value)) }
func (x *BasicLit) Repr() string { return "LITERAL" }

// Unary Expression
func (*UnaryExpr) exprNode()      {}
func (x *UnaryExpr) Pos() Pos     { return x.OperatorPos }
func (x *UnaryExpr) End() Pos     { return x.Operand.End() }
func (x *UnaryExpr) Repr() string { return "UNARY" }

// Binary Expression
func (*BinaryExpr) exprNode()      {}
func (x *BinaryExpr) Pos() Pos     { return x.Left.Pos() }
func (x *BinaryExpr) End() Pos     { return x.Right.End() }
func (x *BinaryExpr) Repr() string { return "BINARY" }

// Exit Statement
func (*ExitStmt) stmtNode()  {}
func (s *ExitStmt) Pos() Pos { return s.Exit }
func (s *ExitStmt) End() Pos {
	if s.Result != nil {
		return s.Result.End()
	}
	return s.Exit + Pos(len("exit"))
}
func (s *ExitStmt) Repr() string { return "EXIT" }

// Program Statement
func (*ProgStmt) stmtNode()  {}
func (s *ProgStmt) Pos() Pos { return s.BeginKw }
func (s *ProgStmt) End() Pos {
	return s.EndKw + Pos(len("end"))
}
func (s *ProgStmt) Repr() string { return "PROGRAM" }

// Skip Statement
func (*SkipStmt) stmtNode()  {}
func (s *SkipStmt) Pos() Pos { return s.Skip }
func (s *SkipStmt) End() Pos {
	return s.Skip + Pos(len("skip"))
}
func (s *SkipStmt) Repr() string { return "SKIP" }

// If Statement
func (*IfStmt) stmtNode()  {}
func (s *IfStmt) Pos() Pos { return s.If }
func (s *IfStmt) End() Pos {
	return s.Fi + Pos(len("Fi"))
}
func (s *IfStmt) Repr() string { return "IF" }

// Function Statement
func (s *Func) Pos() Pos { return s.Func }
func (s *Func) End() Pos {
	return s.Stmts[len(s.Stmts)-1].End()
}
func (s *Func) Repr() string { return "FUNC" }

// Param?
func (s *Param) Pos() Pos     { return s.Start }
func (s *Param) End() Pos     { return s.Finish }
func (s *Param) Repr() string { return "PARAM" }

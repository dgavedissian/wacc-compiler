package main

import (
	"go/token"
)

type Pos token.Pos
type Position token.Position

type Node interface {
	Pos() Pos // Position of first character belonging to the node
	End() Pos // Position of first character immediately after the node
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
	Body    []Stmt
	EndKw   Pos // position of "end keyword
}

type ExitStmt struct {
	Exit   Pos  // position of "exit" keyword
	Result Expr // result expression
}

type SkipStmt struct {
	Skip Pos
}

func (*BasicLit) exprNode() {}

func (x *BasicLit) Pos() Pos { return x.ValuePos }
func (x *BasicLit) End() Pos { return Pos(int(x.ValuePos) + len(x.Value)) }

func (*UnaryExpr) exprNode() {}

func (x *UnaryExpr) Pos() Pos { return x.OperatorPos }
func (x *UnaryExpr) End() Pos { return x.Operand.End() }

func (*BinaryExpr) exprNode() {}

func (x *BinaryExpr) Pos() Pos { return x.Left.Pos() }
func (x *BinaryExpr) End() Pos { return x.Right.End() }

func (*ExitStmt) stmtNode() {}

func (s *ExitStmt) Pos() Pos { return s.Result.Pos() }
func (s *ExitStmt) End() Pos {
	if s.Result != nil {
		return s.Result.End()
	}
	return s.Result.Pos() + Pos(len("exit"))
}

func (*ProgStmt) stmtNode() {}

func (s *ProgStmt) Pos() Pos { return s.BeginKw }
func (s *ProgStmt) End() Pos {
	return s.EndKw + Pos(len("end"))
}

func (*SkipStmt) stmtNode() {}

func (s *SkipStmt) Pos() Pos { return s.Skip }
func (s *SkipStmt) End() Pos {
	return s.Skip + Pos(len("skip"))
}

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

type ProgStmt struct {
	Begin Pos // position of "begin" keyword
	Body  []Stmt
	End   Pos // position of "end keyword
}

type ExitStmt struct {
	Exit   Pos  // position of "exit" keyword
	Result Expr // result expression
}

func (*BasicLit) exprNode() {}

func (x *BasicLit) Pos() Pos { return x.ValuePos }
func (x *BasicLit) End() Pos { return Pos(int(x.ValuePos) + len(x.Value)) }

func (*ExitStmt) stmtNode() {}

func (s *ExitStmt) Pos() Pos { return s.Result.Pos() }
func (s *ExitStmt) End() Pos {
	if s.Result != nil {
		return s.Result.End()
	}
	return s.Result.Pos() + 4 // len("exit")
}

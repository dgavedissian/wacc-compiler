package main

import (
	"fmt"
	"go/token"
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

type LValueExpr interface {
	Expr
	lvalueExprNode()
}

type Type interface {
	Equals(Type) bool
	Repr() string
}

//
// Types
//
type BasicType struct {
	TypeId int
}

type ArrayType struct {
	BaseType Type
}

type PairType struct {
	Fst Type
	Snd Type
}

// Special case used in empty array literals
type AnyType struct {
}

// Special case used to propagate errors
type ErrorType struct {
}

//
// Statements
//
type ProgStmt struct {
	BeginKw Pos // position of "begin" keyword
	Funcs   []Function
	Body    []Stmt
	EndKw   Pos // position of "end keyword
}

type SkipStmt struct {
	Skip Pos // position of "skip" keyword
}

type DeclStmt struct {
	TypeKw Pos // Position of the type keyword
	Type   Type
	Ident  *IdentExpr
	Right  Expr
}

type AssignStmt struct {
	Left  LValueExpr
	Right Expr
}

type ReadStmt struct {
	Read Pos
	Dest LValueExpr
}

type FreeStmt struct {
	Free   Pos
	Object Expr
}

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

type Function struct {
	Func   Pos
	Type   Type
	Ident  *IdentExpr
	Params []Param
	Stmts  []Stmt
}

type Param struct {
	Start  Pos
	Type   Type
	Ident  *IdentExpr
	Finish Pos
}

//
// LValue Expressions
//
type IdentExpr struct {
	NamePos Pos
	Name    string
}

type ArrayElemExpr struct {
	VolumePos Pos
	Volume    LValueExpr
	Index     Expr
}

// Technically not an expression - but can be used as an lvalue
type PairElemExpr struct {
	SelectorPos  Pos
	SelectorType int
	Operand      Expr
}

//
// Expressions
//
type BasicLit struct {
	ValuePos Pos  // Literal position
	Type     Type // Token kind (e.g. INT_LIT, CHAR_LIT)
	Value    string
}

type ArrayLit struct {
	ValuesPos Pos
	Values    []Expr
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

//
// Commands: Expressions which are only used in assignments
//
type NewPairCmd struct {
	ValuePos Pos
	Left     Expr
	Right    Expr
}

type CallCmd struct {
	Call  Pos
	Ident *IdentExpr
	Args  []Expr
}

//
// Repr Helpers
//
func ReprNodes(nodeList interface{}) string {
	realNodeList := make([]Node, 0)
	switch nodeList := nodeList.(type) {
	case []Node:
		return reprNodesInt(nodeList)

	case []Stmt:
		for _, n := range nodeList {
			realNodeList = append(realNodeList, n)
		}
		return reprNodesInt(realNodeList)

	case []Expr:
		for _, n := range nodeList {
			realNodeList = append(realNodeList, n)
		}
		return reprNodesInt(realNodeList)

	case []Function:
		for _, n := range nodeList {
			realNodeList = append(realNodeList, n)
		}
		return reprNodesInt(realNodeList)

	case []Param:
		for _, n := range nodeList {
			realNodeList = append(realNodeList, n)
		}
		return reprNodesInt(realNodeList)

	default:
		panic("nodeList is not of valid type")
	}
}

func reprNodesInt(nodeList []Node) string {
	nodes := []string{}
	for _, f := range nodeList {
		if f, ok := f.(Node); ok {
			nodes = append(nodes, f.(Node).Repr())
		} else {
			nodes = append(nodes, "<nil_node>")
		}
	}
	return strings.Join(nodes, ", ")
}

//
// Types
//

// Basic Type
func (bt BasicType) Equals(t2 Type) bool {
	if _, ok := t2.(PairType); ok && bt.TypeId == PAIR {
		return true
	} else if bt2, ok := t2.(BasicType); ok {
		return bt2.TypeId == bt.TypeId
	}
	return false
}
func (bt BasicType) Repr() string {
	switch bt.TypeId {
	case INT:
		return "INT"
	case BOOL:
		return "BOOL"
	case CHAR:
		return "CHAR"
	case STRING:
		return "STRING"
	case PAIR: // null
		return "PAIR"
	default:
		panic(fmt.Sprintf("BasicType.Repr: wtf is a %d?", bt.TypeId))
	}
}

// Array Type
func (at ArrayType) Equals(t2 Type) bool {
	if at2, ok := t2.(ArrayType); ok {
		return at2.BaseType.Equals(at.BaseType)
	}
	return false
}
func (at ArrayType) Repr() string {
	return at.BaseType.Repr() + "[]"
}

// Pair Type
func (pt PairType) Equals(t2 Type) bool {
	if bt2, ok := t2.(BasicType); ok {
		return bt2.TypeId == PAIR
	} else if pt2, ok := t2.(PairType); ok {
		return pt2.Fst.Equals(pt.Fst) && pt2.Snd.Equals(pt.Snd)
	}
	return false
}
func (pt PairType) Repr() string {
	return "PAIR(" + pt.Fst.Repr() + ", " + pt.Snd.Repr() + ")"
}

// Any Type
func (AnyType) Equals(Type) bool {
	return true
}
func (AnyType) Repr() string {
	return "ANY_TYPE"
}

// Error Type
func (ErrorType) Equals(Type) bool {
	return false
}
func (ErrorType) Repr() string {
	return "ERROR_TYPE"
}

//
// Statements
//

// Program Statement
func (ProgStmt) stmtNode()  {}
func (s ProgStmt) Pos() Pos { return s.BeginKw }
func (s ProgStmt) End() Pos {
	return s.EndKw + Pos(len("end"))
}
func (s ProgStmt) Repr() string {
	return "Prog(" + ReprNodes(s.Funcs) + ")(" +
		ReprNodes(s.Body) + ")"
}

// Skip Statement
func (SkipStmt) stmtNode()  {}
func (s SkipStmt) Pos() Pos { return s.Skip }
func (s SkipStmt) End() Pos {
	return s.Skip + Pos(len("skip"))
}
func (s SkipStmt) Repr() string { return "Skip" }

// Declaration statement
func (DeclStmt) stmtNode()  {}
func (s DeclStmt) Pos() Pos { return s.TypeKw }
func (s DeclStmt) End() Pos { return s.Pos() } // TODO
func (s DeclStmt) Repr() string {
	if s.Right == nil {
		return "Decl(" + s.Type.Repr() + ", " + s.Ident.Repr() + ", <missing rhs>)"
	}
	return "Decl(" + s.Type.Repr() + ", " + s.Ident.Repr() + ", " + s.Right.Repr() + ")"
}

// Assign Statement
func (AssignStmt) stmtNode()  {}
func (s AssignStmt) Pos() Pos { return s.Left.Pos() }
func (s AssignStmt) End() Pos { return s.Pos() } // TODO
func (s AssignStmt) Repr() string {
	if s.Right == nil {
		return "Assign(" + s.Left.Repr() + ", <missing rhs>)"
	}
	return "Assign(" + s.Left.Repr() + ", " + s.Right.Repr() + ")"
}

// Read Statement
func (ReadStmt) stmtNode()  {}
func (s ReadStmt) Pos() Pos { return s.Read }
func (s ReadStmt) End() Pos { return s.Pos() } // TODO
func (s ReadStmt) Repr() string {
	if s.Dest == nil {
		return "Read(<missing destination>)"
	}
	return "Read(" + s.Dest.Repr() + ")"
}

// Free Statement
func (FreeStmt) stmtNode()  {}
func (s FreeStmt) Pos() Pos { return s.Free }
func (s FreeStmt) End() Pos { return s.Pos() } // TODO
func (s FreeStmt) Repr() string {
	if s.Object == nil {
		return "Free(<missing object>)"
	}
	return "Free(" + s.Object.Repr() + ")"
}

// Exit Statement
func (ExitStmt) stmtNode()  {}
func (s ExitStmt) Pos() Pos { return s.Exit }
func (s ExitStmt) End() Pos {
	if s.Result != nil {
		return s.Result.End()
	}
	return s.Exit + Pos(len("exit"))
}
func (s ExitStmt) Repr() string {
	if s.Result != nil {
		return "Exit(" + s.Result.Repr() + ")"
	} else {
		return "Exit(<missing expr>)"
	}
}

// Return Statement
func (ReturnStmt) stmtNode()  {}
func (s ReturnStmt) Pos() Pos { return s.Return }
func (s ReturnStmt) End() Pos {
	if s.Result != nil {
		return s.Result.End()
	}
	return s.Return + Pos(len("return"))
}
func (s ReturnStmt) Repr() string {
	if s.Result != nil {
		return "Return(" + s.Result.Repr() + ")"
	} else {
		return "Return(<missing expr>)"
	}
}

// Print Statement
func (PrintStmt) stmtNode()  {}
func (s PrintStmt) Pos() Pos { return s.Print }
func (s PrintStmt) End() Pos { return s.Pos() }
func (s PrintStmt) Repr() string {
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
func (IfStmt) stmtNode()  {}
func (s IfStmt) Pos() Pos { return s.If }
func (s IfStmt) End() Pos {
	return s.Fi + Pos(len("Fi"))
}
func (s IfStmt) Repr() string {
	return "If(" + s.Cond.Repr() +
		")Then(" + ReprNodes(s.Body) +
		")Else(" + ReprNodes(s.Else) + ")"
}

// While Statement
func (WhileStmt) stmtNode()  {}
func (s WhileStmt) Pos() Pos { return s.While }
func (s WhileStmt) End() Pos {
	return s.Done + Pos(len("Done")) // TODO
}
func (s WhileStmt) Repr() string {
	return "While(" + s.Cond.Repr() +
		")Do(" + ReprNodes(s.Body) +
		")Done"
}

// Function Statement
func (s Function) Pos() Pos { return s.Func }
func (s Function) End() Pos {
	return s.Stmts[len(s.Stmts)-1].End()
}
func (s Function) Repr() string {
	return "Function(type:" + s.Type.Repr() +
		", name:" + s.Ident.Repr() +
		", params:(" + ReprNodes(s.Params) +
		"), body:(" + ReprNodes(s.Stmts) + ")"
}

// Function Parameter
func (s Param) Pos() Pos { return s.Start }
func (s Param) End() Pos { return s.Finish }
func (s Param) Repr() string {
	return "Param(" + s.Type.Repr() + ", " + s.Ident.Repr() + ")"
}

//
// LValue Expressions
//

// Identifier
func (IdentExpr) lvalueExprNode() {}
func (IdentExpr) exprNode()       {}
func (x IdentExpr) Pos() Pos      { return x.NamePos }
func (x IdentExpr) End() Pos      { return Pos(int(x.NamePos) + len(x.Name)) }
func (x IdentExpr) Repr() string {
	if x.Name == "" {
		return "IdentExpr(<missing name>)"
	}
	return "IdentExpr(" + x.Name + ")"
}

// Array element expression
func (ArrayElemExpr) lvalueExprNode() {}
func (ArrayElemExpr) exprNode()       {}
func (x ArrayElemExpr) Pos() Pos      { return x.VolumePos }
func (x ArrayElemExpr) End() Pos {
	return x.Index.End() + 1 /* Close bracket*/
}
func (x ArrayElemExpr) Repr() string {
	return "ArrayElem(" + x.Volume.Repr() + ", " + x.Index.Repr() + ")"
}

// Pair selector expressions
func (PairElemExpr) lvalueExprNode() {}
func (PairElemExpr) exprNode()       {}
func (x PairElemExpr) Pos() Pos      { return x.SelectorPos }
func (x PairElemExpr) End() Pos      { return Pos(int(x.SelectorPos) + 3) } // TODO: unbreak that
func (x PairElemExpr) Repr() string {
	return fmt.Sprintf("PairElem(%d, %f)", x.SelectorType, x.Operand)
}

//
// Expressions
//

// Basic Literal
func (BasicLit) exprNode()  {}
func (x BasicLit) Pos() Pos { return x.ValuePos }
func (x BasicLit) End() Pos { return Pos(int(x.ValuePos) + len(x.Value)) }
func (x BasicLit) Repr() string {
	return "Lit(" + x.Type.Repr() + ", " + x.Value + ")"
}

// Array literal
func (ArrayLit) exprNode()  {}
func (x ArrayLit) Pos() Pos { return x.ValuesPos }
func (x ArrayLit) End() Pos {
	if x.Values == nil {
		return Pos(int(x.ValuesPos) + 1) /* CLose bracket */
	}
	return Pos(int(x.Values[len(x.Values)-1].End()) + 1)
}
func (x ArrayLit) Repr() string {
	if x.Values == nil {
		return "ArrayLit([])"
	}
	return "ArrayLit([" + ReprNodes(x.Values) + "])"
}

// Unary Expression
func (UnaryExpr) exprNode()  {}
func (x UnaryExpr) Pos() Pos { return x.OperatorPos }
func (x UnaryExpr) End() Pos { return x.Operand.End() }
func (x UnaryExpr) Repr() string {
	if x.Operand == nil {
		return "Unary(" + x.Operator + ", <missing operand>)"
	}
	return "Unary(" + x.Operator + ", " + x.Operand.Repr() + ")"
}

// Binary Expression
func (BinaryExpr) exprNode()  {}
func (x BinaryExpr) Pos() Pos { return x.Left.Pos() }
func (x BinaryExpr) End() Pos { return x.Right.End() }
func (x BinaryExpr) Repr() string {
	if x.Left == nil || x.Right == nil {
		return "Binary(" + x.Operator + ", , )"
	}
	return "Binary(" + x.Operator + ", " +
		x.Left.Repr() + ", " + x.Right.Repr() + ")"
}

//
// Commands: Expressions which are only used in assignments
//

// Pairs
func (NewPairCmd) exprNode()  {}
func (x NewPairCmd) Pos() Pos { return x.ValuePos }
func (x NewPairCmd) End() Pos {
	return Pos(int(x.ValuePos) + len(x.Right.Repr()) + 1) // Right bracket
}
func (x NewPairCmd) Repr() string {
	if x.Left == nil || x.Right == nil {
		return "NewPair(<missing elements>)"
	}
	return "NewPair(" + x.Left.Repr() + ", " + x.Right.Repr() + ")"
}

// Function call
func (CallCmd) exprNode()  {}
func (x CallCmd) Pos() Pos { return x.Call }
func (x CallCmd) End() Pos {
	return x.Call /* TODO */
}
func (x CallCmd) Repr() string {
	return "Call(" + x.Ident.Repr() + ", " + ReprNodes(x.Args) + ")"
}

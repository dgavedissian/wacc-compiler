package frontend

import (
	"fmt"
	"strings"
)

type Node interface {
	Pos() *Position // *Position of first character belonging to the node
	End() *Position // *Position of first character immediately after the node
	Repr() string   // Representation
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

type StructType struct {
	TypeId string
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
type Program struct {
	BeginPos *Position // position of "begin" keyword
	Structs  []*Struct
	Funcs    []*Function
	Body     []Stmt
	EndPos   *Position // position of "end keyword
}

type Struct struct {
	Struct  *Position
	Ident   *IdentExpr
	Members []*StructMember
}

type StructMember struct {
	MemberPos *Position
	Type      Type
	Ident     *IdentExpr
}

type Function struct {
	Func     *Position
	Type     Type
	Ident    *IdentExpr
	Params   []Param
	Body     []Stmt
	External bool
}

type Param struct {
	Start  *Position
	Type   Type
	Ident  *IdentExpr
	Finish *Position
}

type SkipStmt struct {
	Skip *Position // position of "skip" keyword
}

type EvalStmt struct {
	Expr Expr
}

type DeclStmt struct {
	TypePos *Position // Position of the type keyword
	Type    Type
	Ident   *IdentExpr
	Right   Expr
}

type AssignStmt struct {
	Left  LValueExpr
	Right Expr
}

type ReadStmt struct {
	Read *Position
	Dst  LValueExpr
	Type Type
}

type FreeStmt struct {
	Free   *Position
	Object Expr
}

type ExitStmt struct {
	Exit   *Position // position of "exit" keyword
	Result Expr      // result expression
}

type ReturnStmt struct {
	Return *Position
	Result Expr
}

type PrintStmt struct {
	Print   *Position // position of print keyword
	Right   Expr
	NewLine bool
	Type    Type
}

type IfStmt struct {
	If   *Position
	Cond Expr
	Body []Stmt
	Else []Stmt
	Fi   *Position
}

type WhileStmt struct {
	While *Position
	Cond  Expr
	Body  []Stmt
	Done  *Position
}

type ScopeStmt struct {
	BeginPos *Position
	Body     []Stmt
	EndPos   *Position
}

//
// LValue Expressions
//
type IdentExpr struct {
	NamePos *Position
	Name    string
}

type ArrayElemExpr struct {
	VolumePos       *Position
	Volume          LValueExpr
	Index           Expr
	CloseBracketPos *Position
}

// Technically not an expression - but can be used as an lvalue
type PairElemExpr struct {
	SelectorPos  *Position
	SelectorType int
	Operand      *IdentExpr
	EndPos       *Position
}

//
// Expressions
//
type BasicLit struct {
	ValuePos *Position // Literal position
	Type     Type      // Token kind (e.g. INT_LIT, CHAR_LIT)
	Value    string
}

type ArrayLit struct {
	ValuesPos *Position
	Values    []Expr
	EndPos    *Position
	Type      Type
}

type UnaryExpr struct {
	OperatorPos *Position
	Operator    string
	Operand     Expr
	Type        Type
}

type BinaryExpr struct {
	Left        Expr
	OperatorPos *Position
	Operator    string
	Right       Expr
	Type        Type
}

//
// Commands: Expressions which are only used in assignments
//
type NewStructCmd struct {
	ValuePos     *Position
	Ident        *IdentExpr
	Args         []Expr
	RightBracket *Position
}

type NewPairCmd struct {
	ValuePos     *Position
	Left         Expr
	Right        Expr
	RightBracket *Position
}

type CallCmd struct {
	Call         *Position
	Ident        *IdentExpr
	Args         []Expr
	RightBracket *Position
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

	case []*Struct:
		for _, n := range nodeList {
			realNodeList = append(realNodeList, n)
		}
		return reprNodesInt(realNodeList)

	case []*StructMember:
		for _, n := range nodeList {
			realNodeList = append(realNodeList, n)
		}
		return reprNodesInt(realNodeList)

	case []*Function:
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
	if bt.Equals(BasicType{STRING}) {
		if array, ok := t2.(ArrayType); ok {
			return array.BaseType.Equals(BasicType{CHAR})
		}
	}
	return false
}
func (bt BasicType) Repr() string {
	switch bt.TypeId {
	case INT:
		return "int"
	case FLOAT:
		return "float"
	case BOOL:
		return "bool"
	case CHAR:
		return "char"
	case STRING:
		return "string"
	case PAIR: // null
		return "pair"
	case VOID:
		return "void"
	default:
		panic(fmt.Sprintf("BasicType.Repr: Undefined repr for %d?", bt.TypeId))
	}
}

// Struct Type
func (st StructType) Equals(t2 Type) bool {
	if st2, ok := t2.(StructType); ok {
		return st.TypeId == st2.TypeId
	}
	return false
}
func (st StructType) Repr() string { return st.TypeId }

// Array Type
func (at ArrayType) Equals(t2 Type) bool {
	if at2, ok := t2.(ArrayType); ok {
		return at2.BaseType.Equals(at.BaseType)
	}
	if at.BaseType.Equals(BasicType{CHAR}) {
		if bt2, ok := t2.(BasicType); ok {
			return bt2.TypeId == STRING
		}
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
	return "pair(" + pt.Fst.Repr() + ", " + pt.Snd.Repr() + ")"
}

// Any Type
func (AnyType) Equals(Type) bool {
	return true
}
func (AnyType) Repr() string {
	return "T"
}

// Error Type
func (ErrorType) Equals(Type) bool {
	return false
}
func (ErrorType) Repr() string {
	return "ERROR"
}

//
// Statements
//

// Program
func (Program) stmtNode()        {}
func (s Program) Pos() *Position { return s.BeginPos }
func (s Program) End() *Position {
	return s.EndPos.End()
}
func (s Program) Repr() string {
	return fmt.Sprintf("Prog(%v, %v, %v)",
		ReprNodes(s.Structs), ReprNodes(s.Funcs), ReprNodes(s.Body))
}

// Struct Statement
func (s Struct) Pos() *Position { return s.Struct }
func (s Struct) End() *Position {
	return s.Members[len(s.Members)-1].End()
}
func (s Struct) Repr() string {
	return fmt.Sprintf("Struct(%v, %v)",
		s.Ident.Repr(), ReprNodes(s.Members))
}

// Struct Member
func (s StructMember) Pos() *Position { return s.MemberPos }
func (s StructMember) End() *Position { return s.MemberPos }
func (s StructMember) Repr() string {
	return fmt.Sprintf("Member(%v, %v)",
		s.Type.Repr(), s.Ident.Repr())
}

// Function Statement
func (s Function) Pos() *Position { return s.Func }
func (s Function) End() *Position {
	return s.Body[len(s.Body)-1].End()
}
func (s Function) Repr() string {
	if s.External {
		return fmt.Sprintf("Function(%v, %v)(%v)(external)",
			s.Type.Repr(), s.Ident.Repr(), ReprNodes(s.Params))
	} else {
		return fmt.Sprintf("Function(%v, %v)(%v)(%v)",
			s.Type.Repr(), s.Ident.Repr(), ReprNodes(s.Params), ReprNodes(s.Body))
	}
}

// Function Parameter
func (s Param) Pos() *Position { return s.Start }
func (s Param) End() *Position { return s.Finish.End() }
func (s Param) Repr() string {
	return "Param(" + s.Type.Repr() + ", " + s.Ident.Repr() + ")"
}

// Skip Statement
func (SkipStmt) stmtNode()        {}
func (s SkipStmt) Pos() *Position { return s.Skip }
func (s SkipStmt) End() *Position {
	return s.Skip.End()
}
func (s SkipStmt) Repr() string { return "Skip" }

// Eval Statement
func (EvalStmt) stmtNode()        {}
func (s EvalStmt) Pos() *Position { return s.Expr.Pos() }
func (s EvalStmt) End() *Position { return s.Expr.End() }
func (s EvalStmt) Repr() string   { return fmt.Sprintf("Eval(%v)", s.Expr.Repr()) }

// Declaration statement
func (DeclStmt) stmtNode()        {}
func (s DeclStmt) Pos() *Position { return s.TypePos }
func (s DeclStmt) End() *Position { return s.Pos() } // TODO
func (s DeclStmt) Repr() string {
	if s.Right == nil {
		return "Decl(" + s.Type.Repr() + ", " + s.Ident.Repr() + ", <missing rhs>)"
	}
	return "Decl(" + s.Type.Repr() + ", " + s.Ident.Repr() + ", " + s.Right.Repr() + ")"
}

// Assign Statement
func (AssignStmt) stmtNode()        {}
func (s AssignStmt) Pos() *Position { return s.Left.Pos() }
func (s AssignStmt) End() *Position { return s.Right.Pos().End() }
func (s AssignStmt) Repr() string {
	if s.Right == nil {
		return "Assign(" + s.Left.Repr() + ", <missing rhs>)"
	}
	return "Assign(" + s.Left.Repr() + ", " + s.Right.Repr() + ")"
}

// Read Statement
func (ReadStmt) stmtNode()        {}
func (s ReadStmt) Pos() *Position { return s.Read }
func (s ReadStmt) End() *Position { return s.Dst.Pos().End() }
func (s ReadStmt) Repr() string {
	if s.Dst == nil {
		return "Read(<missing destination>)"
	}
	return "Read(" + s.Dst.Repr() + ")"
}

// Free Statement
func (FreeStmt) stmtNode()        {}
func (s FreeStmt) Pos() *Position { return s.Free }
func (s FreeStmt) End() *Position { return s.Object.Pos().End() }
func (s FreeStmt) Repr() string {
	if s.Object == nil {
		return "Free(<missing object>)"
	}
	return "Free(" + s.Object.Repr() + ")"
}

// Exit Statement
func (ExitStmt) stmtNode()        {}
func (s ExitStmt) Pos() *Position { return s.Exit }
func (s ExitStmt) End() *Position {
	if s.Result != nil {
		return s.Result.End()
	}
	return s.Exit.End()
}
func (s ExitStmt) Repr() string {
	if s.Result != nil {
		return "Exit(" + s.Result.Repr() + ")"
	} else {
		return "Exit(<missing expr>)"
	}
}

// Return Statement
func (ReturnStmt) stmtNode()        {}
func (s ReturnStmt) Pos() *Position { return s.Return }
func (s ReturnStmt) End() *Position {
	if s.Result != nil {
		return s.Result.End()
	}
	return s.Return.End()
}
func (s ReturnStmt) Repr() string {
	if s.Result != nil {
		return "Return(" + s.Result.Repr() + ")"
	} else {
		return "Return(<missing expr>)"
	}
}

// Print Statement
func (PrintStmt) stmtNode()        {}
func (s PrintStmt) Pos() *Position { return s.Print }
func (s PrintStmt) End() *Position { return s.Right.End() }
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
func (IfStmt) stmtNode()        {}
func (s IfStmt) Pos() *Position { return s.If }
func (s IfStmt) End() *Position {
	return s.Fi.End()
}
func (s IfStmt) Repr() string {
	return "If(" + s.Cond.Repr() +
		")Then(" + ReprNodes(s.Body) +
		")Else(" + ReprNodes(s.Else) + ")"
}

// While Statement
func (WhileStmt) stmtNode()        {}
func (s WhileStmt) Pos() *Position { return s.While }
func (s WhileStmt) End() *Position {
	return s.Done.End()
}
func (s WhileStmt) Repr() string {
	return "While(" + s.Cond.Repr() +
		")Do(" + ReprNodes(s.Body) +
		")Done"
}

// Scope Statement
func (ScopeStmt) stmtNode()        {}
func (s ScopeStmt) Pos() *Position { return s.BeginPos }
func (s ScopeStmt) End() *Position {
	return s.EndPos.End()
}
func (s ScopeStmt) Repr() string {
	return "Scope(" + ReprNodes(s.Body) + ")"
}

//
// LValue Expressions
//

// Identifier
func (IdentExpr) lvalueExprNode()  {}
func (IdentExpr) exprNode()        {}
func (x IdentExpr) Pos() *Position { return x.NamePos }
func (x IdentExpr) End() *Position { return x.NamePos.End() }
func (x IdentExpr) Repr() string {
	if x.Name == "" {
		return "IdentExpr(<missing name>)"
	}
	return "IdentExpr(" + x.Name + ")"
}

// Array element expression
func (ArrayElemExpr) lvalueExprNode()  {}
func (ArrayElemExpr) exprNode()        {}
func (x ArrayElemExpr) Pos() *Position { return x.VolumePos }
func (x ArrayElemExpr) End() *Position {
	return x.CloseBracketPos.End()
}
func (x ArrayElemExpr) Repr() string {
	return "ArrayElem(" + x.Volume.Repr() + ", " + x.Index.Repr() + ")"
}

// Pair selector expressions
func (PairElemExpr) lvalueExprNode()  {}
func (PairElemExpr) exprNode()        {}
func (x PairElemExpr) Pos() *Position { return x.SelectorPos }
func (x PairElemExpr) End() *Position { return x.EndPos.End() }
func (x PairElemExpr) Repr() string {
	return fmt.Sprintf("PairElem(%d, %v)", x.SelectorType, x.Operand)
}

//
// Expressions
//

// Basic Literal
func (BasicLit) exprNode()        {}
func (x BasicLit) Pos() *Position { return x.ValuePos }
func (x BasicLit) End() *Position { return x.ValuePos.End() }
func (x BasicLit) Repr() string {
	return fmt.Sprintf("Lit(%v, %v)", x.Type.Repr(), x.Value)
}

// Array literal
func (ArrayLit) exprNode()        {}
func (x ArrayLit) Pos() *Position { return x.ValuesPos }
func (x ArrayLit) End() *Position {
	return x.EndPos.End()
}
func (x ArrayLit) Repr() string {
	if x.Values == nil {
		return "ArrayLit([])"
	}
	return "ArrayLit([" + ReprNodes(x.Values) + "])"
}

// Unary Expression
func (UnaryExpr) exprNode()        {}
func (x UnaryExpr) Pos() *Position { return x.OperatorPos }
func (x UnaryExpr) End() *Position { return x.Operand.End() }
func (x UnaryExpr) Repr() string {
	var t string
	if x.Type != nil {
		t = x.Type.Repr()
	}
	return fmt.Sprintf("Unary(%v, %v, %v)", x.Operator, x.Operand.Repr(), t)
}

// Binary Expression
func (BinaryExpr) exprNode()        {}
func (x BinaryExpr) Pos() *Position { return x.Left.Pos() }
func (x BinaryExpr) End() *Position { return x.Right.End() }
func (x BinaryExpr) Repr() string {
	var t string
	if x.Type != nil {
		t = x.Type.Repr()
	}
	return fmt.Sprintf("Binary(%v, %v, %v, %v)", x.Operator, x.Left.Repr(), x.Right.Repr(), t)
}

//
// Commands: Expressions which are only used in assignments
//

// Structs
func (NewStructCmd) exprNode()        {}
func (x NewStructCmd) Pos() *Position { return x.ValuePos }
func (x NewStructCmd) End() *Position {
	return x.RightBracket.End()
}
func (x NewStructCmd) Repr() string {
	return fmt.Sprintf("NewStruct(%v, %v)",
		x.Ident.Repr(), ReprNodes(x.Args))
}

// Pairs
func (NewPairCmd) exprNode()        {}
func (x NewPairCmd) Pos() *Position { return x.ValuePos }
func (x NewPairCmd) End() *Position {
	return x.RightBracket.End()
}
func (x NewPairCmd) Repr() string {
	if x.Left == nil || x.Right == nil {
		return "NewPair(<missing elements>)"
	}
	return "NewPair(" + x.Left.Repr() + ", " + x.Right.Repr() + ")"
}

// Function call
func (CallCmd) exprNode()        {}
func (x CallCmd) Pos() *Position { return x.Call }
func (x CallCmd) End() *Position {
	return x.RightBracket.End()
}
func (x CallCmd) Repr() string {
	return "Call(" + x.Ident.Repr() + ", " + ReprNodes(x.Args) + ")"
}

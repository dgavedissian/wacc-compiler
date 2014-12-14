/* File: parser.y */

%{
  package frontend
%}

%union {
  Expr   Expr
  Value  string

  Structs []*Struct
  Struct *Struct
  StructMembers []*StructMember
  StructMember *StructMember

  Funcs  []*Function
  Func   *Function 
  Params []Param
  Param  Param
  
  Stmts  []Stmt
  Stmt   Stmt
  
  Type   Type
  lines  int
  Exprs  []Expr

  Position *Position
}

%{
  var top yySymType
%}

%token BEGIN END
%token INT_LIT FLOAT_LIT BOOL_LIT CHAR_LIT STRING_LIT PAIR_LIT
%token IDENT
%token UNARY_OPER BINARY_OPER
%token SKIP READ FREE RETURN EXIT PRINT PRINTLN NEWPAIR NEWSTRUCT CALL
%token INT FLOAT BOOL CHAR STRING PAIR
%token IS EXTERNAL STRUCT
%token IF THEN ELSE FI
%token WHILE DO DONE
%token LEN ORD CHR FST SND
%token LE GE EQ NE AND OR
%%

top
    : program { top = $1 }
    ;

program
    : BEGIN struct_list END { $$.Stmt = &ProgStmt{$1.Position, $2.Structs, $2.Funcs, $2.Stmts, $3.Position} }
    ;

struct_list
    : struct struct_list {
        $$.Stmts = $2.Stmts
        $$.Funcs = $2.Funcs
        $$.Structs = append([]*Struct{$1.Struct}, $2.Structs...)
      }
    | body {
        $$.Stmts = $1.Stmts
        $$.Funcs = $1.Funcs
      }
    ;

body
    : func body {
        $$.Stmts = $2.Stmts
        $$.Funcs = append([]*Function{$1.Func}, $2.Funcs...)
      }
    | statement_list { $$.Stmts = $1.Stmts }
    ;

/* Structs */
struct
    : STRUCT identifier IS struct_member_list END {
        $$.Struct = &Struct{$1.Position, $2.Expr.(*IdentExpr), $4.StructMembers}
      }
    ;

struct_member_list
    : struct_member ';' struct_member_list {
        $$.StructMembers = append([]*StructMember{$1.StructMember}, $3.StructMembers...)
      }
    | struct_member { $$.StructMembers = []*StructMember{$1.StructMember} }
    ;

struct_member
    : type identifier { $$.StructMember = &StructMember{$1.Position, $1.Type, $2.Expr.(*IdentExpr)} }
    ;

/* Functions */
func
    : type identifier '(' optional_param_list ')' IS statement_list END {
        VerifyFunctionReturns($7.Stmts)
        $$.Func = &Function{$1.Position, $1.Type, $2.Expr.(*IdentExpr), $4.Params, $7.Stmts, false}
      }
    | type identifier '(' optional_param_list ')' IS EXTERNAL {
        $$.Func = &Function{$1.Position, $1.Type, $2.Expr.(*IdentExpr), $4.Params, nil, true} 
      }
    ;

optional_param_list
    : param_list { $$.Params = $1.Params }
    |
    ;

param_list
    : param ',' param_list { $$.Params = append([]Param{$1.Param}, $3.Params...) }
    | param { $$.Params = []Param{$1.Param} }
    ;

param
    : type identifier { $$.Param = Param{$1.Position, $1.Type, $2.Expr.(*IdentExpr), $2.Position} }
    ;

/* Statements */
statement_list
    : statement { $$.Stmts = []Stmt{$1.Stmt} }
    | statement ';' statement_list { $$.Stmts = append([]Stmt{$1.Stmt}, $3.Stmts...) }
    | error ';' statement_list
    ;

statement
    : SKIP                            { $$.Stmt = &SkipStmt{$1.Position} }
    | type identifier '=' assign_rhs  { $$.Stmt = &DeclStmt{$1.Position, $1.Type, $2.Expr.(*IdentExpr), $4.Expr} }
    | assign_lhs '=' assign_rhs       { $$.Stmt = &AssignStmt{$1.Expr.(LValueExpr), $3.Expr} }
    | READ assign_lhs                 { $$.Stmt = &ReadStmt{$1.Position, $2.Expr.(LValueExpr), nil} }
    | FREE expression                 { $$.Stmt = &FreeStmt{$1.Position, $2.Expr} }
    | RETURN expression               { $$.Stmt = &ReturnStmt{$1.Position, $2.Expr} }
    | EXIT expression                 { $$.Stmt = &ExitStmt{$1.Position, $2.Expr} }
    | PRINT expression                { $$.Stmt = &PrintStmt{$1.Position, $2.Expr, false, nil} }
    | PRINTLN expression              { $$.Stmt = &PrintStmt{$1.Position, $2.Expr, true, nil} }
    | BEGIN statement_list END        { $$.Stmt = &ScopeStmt{$1.Position, $2.Stmts, $3.Position} }
    | IF expression THEN statement_list ELSE statement_list FI {
        $$.Stmt = &IfStmt{$1.Position, $2.Expr, $4.Stmts, $6.Stmts, $7.Position}
      }
    | WHILE expression DO statement_list DONE {
        $$.Stmt = &WhileStmt{ $1.Position, $2.Expr, $4.Stmts , $5.Position }
      }
    ;

assign_lhs
    : identifier     { $$.Expr = $1.Expr }
    | identifier '[' expression ']' { $$.Expr = &ArrayElemExpr{$1.Position, $1.Expr.(LValueExpr), $3.Expr, $4.Position} }
    | pair_elem      { $$.Expr = $1.Expr }
    ;

assign_rhs
    : expression {$$.Expr = $1.Expr}
    | NEWSTRUCT '(' identifier ',' optional_arg_list ')' {
        $$.Expr = &NewStructCmd{$1.Position, $3.Expr.(*IdentExpr), $5.Exprs, $6.Position}
      }
    | NEWPAIR '(' expression ',' expression ')' {
        $$.Expr = &NewPairCmd{$1.Position, $3.Expr, $5.Expr, $6.Position}
      }
    | CALL identifier '(' optional_arg_list ')' { $$.Expr = &CallCmd{$1.Position, $2.Expr.(*IdentExpr), $4.Exprs, $5.Position} }
    | '[' array_liter ']' { $$.Expr = &ArrayLit{$1.Position, $2.Exprs, $3.Position, nil} }
    | pair_elem
    ;

identifier
    : IDENT { $$.Expr = &IdentExpr{$1.Position, $1.Value} }
    ;

optional_arg_list
    : arg_list { $$.Exprs = $1.Exprs }
    |
    ;

arg_list
    : expression ',' arg_list {
        $$.Exprs = append([]Expr{$1.Expr}, $3.Exprs...)
      }
    | expression { $$.Exprs = []Expr{$1.Expr} }
    ;


/* Types */
type
    : base_type
    | pair_type
    | type '[' ']' { $$.Type = ArrayType{$1.Type} }
    ;

base_type
    : INT    { $$.Type = BasicType{INT} }
    | FLOAT  { $$.Type = BasicType{FLOAT} }
    | BOOL   { $$.Type = BasicType{BOOL} }
    | CHAR   { $$.Type = BasicType{CHAR} }
    | STRING { $$.Type = BasicType{STRING} }
    ;

array_type
    : type '[' ']' { $$.Type = ArrayType{$1.Type} }
    ;

pair_type
    : PAIR '(' pair_elem_type ',' pair_elem_type ')' { $$.Type = PairType{$3.Type, $5.Type} }
    ;

pair_elem_type
    : base_type
    | array_type
    | PAIR        { $$.Type = BasicType{PAIR} }
    ;

pair_elem
    : FST expression { $$.Expr = &PairElemExpr{$1.Position, FST, $2.Expr.(*IdentExpr), $2.Position} }
    | SND expression { $$.Expr = &PairElemExpr{$1.Position, SND, $2.Expr.(*IdentExpr), $2.Position} }

array_liter
    : array_contents
    |
    ;

array_contents
    : expression ',' array_contents {
      $$.Exprs = append([]Expr{$1.Expr}, $3.Exprs...)
    }
    | expression { $$.Exprs =  []Expr{$1.Expr} }
    ;

/* Expression */
array_expression
    : identifier '[' expression ']' { $$.Expr = &ArrayElemExpr{$1.Position, $1.Expr.(LValueExpr), $3.Expr, $4.Position} }
    | array_expression '[' expression ']' { $$.Expr = &ArrayElemExpr{$1.Position, $1.Expr.(LValueExpr), $3.Expr, $4.Position} }
    ;

primary_expression
    : identifier          { $$.Expr = &IdentExpr{$1.Position, $1.Value} }
    | INT_LIT             { $$.Expr = &BasicLit{$1.Position, BasicType{INT}, $1.Value} }
    | FLOAT_LIT           { $$.Expr = &BasicLit{$1.Position, BasicType{FLOAT}, $1.Value} }
    | BOOL_LIT            { $$.Expr = &BasicLit{$1.Position, BasicType{BOOL}, $1.Value} }
    | CHAR_LIT            { $$.Expr = &BasicLit{$1.Position, BasicType{CHAR}, $1.Value} }
    | STRING_LIT          { $$.Expr = &BasicLit{$1.Position, BasicType{STRING}, $1.Value} }
    | PAIR_LIT            { $$.Expr = &BasicLit{$1.Position, BasicType{PAIR}, $1.Value} }
    | '(' expression ')'  { $$.Expr = $2.Expr }
    | array_expression
    ;

unary_expression
    : primary_expression   { $$.Expr = $1.Expr }
    | '!' unary_expression { $$.Expr = &UnaryExpr{$1.Position, "!", $2.Expr, nil} }
    | '+' unary_expression { $$.Expr = &UnaryExpr{$1.Position, "+", $2.Expr, nil} }
    | '-' unary_expression { $$.Expr = &UnaryExpr{$1.Position, "-", $2.Expr, nil} }
    | LEN unary_expression { $$.Expr = &UnaryExpr{$1.Position, "len", $2.Expr, nil} }
    | ORD unary_expression { $$.Expr = &UnaryExpr{$1.Position, "ord", $2.Expr, nil} }
    | CHR unary_expression { $$.Expr = &UnaryExpr{$1.Position, "chr", $2.Expr, nil} }
    ;

multiplicative_expression
    : unary_expression {
        VerifyNoOverflows($1.Expr)
        $$.Expr = $1.Expr
    }
    | multiplicative_expression '*' unary_expression { $$.Expr = &BinaryExpr{$1.Expr, $2.Position, "*", $3.Expr, nil} }
    | multiplicative_expression '/' unary_expression { $$.Expr = &BinaryExpr{$1.Expr, $2.Position, "/", $3.Expr, nil} }
    | multiplicative_expression '%' unary_expression { $$.Expr = &BinaryExpr{$1.Expr, $2.Position, "%", $3.Expr, nil} }
    ;

additive_expression
    : multiplicative_expression { $$.Expr = $1.Expr }
    | additive_expression '+' multiplicative_expression { $$.Expr = &BinaryExpr{$1.Expr, $2.Position, "+", $3.Expr, nil} }
    | additive_expression '-' multiplicative_expression { $$.Expr = &BinaryExpr{$1.Expr, $2.Position, "-", $3.Expr, nil} }
    ;

relational_expression
    : additive_expression { $$.Expr = $1.Expr }
    | relational_expression '<' additive_expression { $$.Expr = &BinaryExpr{$1.Expr, $2.Position, "<", $3.Expr, nil} }
    | relational_expression '>' additive_expression { $$.Expr = &BinaryExpr{$1.Expr, $2.Position, ">", $3.Expr, nil} }
    | relational_expression LE additive_expression { $$.Expr = &BinaryExpr{$1.Expr, $2.Position, "<=", $3.Expr, nil} }
    | relational_expression GE additive_expression { $$.Expr = &BinaryExpr{$1.Expr, $2.Position, ">=", $3.Expr, nil} }
    ;

equality_expression
    : relational_expression { $$.Expr = $1.Expr }
    | equality_expression EQ relational_expression { $$.Expr = &BinaryExpr{$1.Expr, $2.Position, "==", $3.Expr, nil} }
    | equality_expression NE relational_expression { $$.Expr = &BinaryExpr{$1.Expr, $2.Position, "!=", $3.Expr, nil} }
    ;

logical_and_expression
    : equality_expression { $$.Expr = $1.Expr }
    | logical_and_expression AND equality_expression { $$.Expr = &BinaryExpr{$1.Expr, $2.Position, "&&", $3.Expr, nil} }
    ;

logical_or_expression
    : logical_and_expression { $$.Expr = $1.Expr }
    | logical_or_expression OR logical_and_expression { $$.Expr = &BinaryExpr{$1.Expr, $2.Position, "||", $3.Expr, nil} }
    ;

expression
    : logical_or_expression { $$.Expr = $1.Expr }
    ;

%%

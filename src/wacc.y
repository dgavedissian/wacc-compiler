/* File: wacc.y */

%{
  package main
%}

%union {
  Expr   Expr
  Value  string
  Funcs  []Function
  Func   *Function
  Stmts  []Stmt
  Stmt   Stmt
  Params []Param
  Param  Param
  Type   Type
  lines  int
  Exprs  []Expr
}

%{
  var top yySymType
%}

%token BEGIN END
%token INT_LIT BOOL_LIT CHAR_LIT STRING_LIT PAIR_LIT
%token IDENT
%token UNARY_OPER BINARY_OPER
%token SKIP READ FREE RETURN EXIT PRINT PRINTLN NEWPAIR CALL
%token INT BOOL CHAR STRING PAIR
%token FUNC_IS
%token IF THEN ELSE FI
%token WHILE DO DONE
%token LEN ORD CHR FST SND
%token LE GE EQ NE AND OR
%%

top
    : program { top = $1 }
    ;

program
    : BEGIN body END { $$.Stmt = &ProgStmt{0, $2.Funcs, $2.Stmts, 0} }
    ;

body
    : func body {
        $$.Stmts = $2.Stmts
        $$.Funcs = append([]Function{*$1.Func}, $2.Funcs...)
      }
    | statement_list { $$.Stmts = $1.Stmts }
    ;

/* Functions */
func
    : type identifier '(' optional_param_list ')' FUNC_IS statement_list END {
        VerifyFunctionReturns($7.Stmts)
        $$.Func = &Function{0, $1.Type, $2.Expr.(*IdentExpr), $4.Params, $7.Stmts}
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
    : type identifier { $$.Param = Param{0, $1.Type, $2.Expr.(*IdentExpr), 0} }
    ;

/* Statements */
statement_list
    : statement { $$.Stmts = []Stmt{$1.Stmt} }
    | statement ';' statement_list { $$.Stmts = append([]Stmt{$1.Stmt}, $3.Stmts...) }
    | error ';' statement_list
    ;

statement
    : SKIP                            { $$.Stmt = &SkipStmt{0} }
    | type identifier '=' assign_rhs  { $$.Stmt = &DeclStmt{0, $1.Type, $2.Expr.(*IdentExpr), $4.Expr} }
    | assign_lhs '=' assign_rhs       { $$.Stmt = &AssignStmt{$1.Expr.(LValueExpr), $3.Expr} }
    | READ assign_lhs                 { $$.Stmt = &ReadStmt{0, $2.Expr.(LValueExpr)} }
    | FREE expression                 { $$.Stmt = &FreeStmt{0, $2.Expr} }
    | RETURN expression               { $$.Stmt = &ReturnStmt{0, $2.Expr} }
    | EXIT expression                 { $$.Stmt = &ExitStmt{0, $2.Expr} }
    | PRINT expression                { $$.Stmt = &PrintStmt{0, $2.Expr, false} }
    | PRINTLN expression              { $$.Stmt = &PrintStmt{0, $2.Expr, true} }
    | BEGIN statement_list END        { $$.Stmts = $2.Stmts }
    | IF expression THEN statement_list ELSE statement_list FI {
        $$.Stmt = &IfStmt{0, $2.Expr, $4.Stmts, $6.Stmts, 0}
      }
    | WHILE expression DO statement_list DONE {
        $$.Stmt = &WhileStmt{ 0, $2.Expr, $4.Stmts ,0 }
      }
    ;

assign_lhs
    : identifier     { $$.Expr = $1.Expr }
    | identifier '[' expression ']' { $$.Expr = &ArrayElemExpr{0, $1.Expr, $3.Expr} }
    | pair_elem      { $$.Expr = $1.Expr }
    ;

assign_rhs
    : expression {$$.Expr = $1.Expr}
    | NEWPAIR '(' expression ',' expression ')' {
        $$.Expr = &NewPairCmd{0, $3.Expr, $5.Expr}
      }
    | CALL identifier '(' optional_arg_list ')' { $$.Expr = &CallCmd{0, $2.Expr.(*IdentExpr), $4.Exprs} }
    | '[' array_liter ']' { $$.Expr = &ArrayLit{0, $2.Exprs} }
    | pair_elem
    ;

identifier
    : IDENT { $$.Expr = &IdentExpr{0, $1.Value} }
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
    : INT    { $$.Type = $1.Type }
    | BOOL   { $$.Type = $1.Type }
    | CHAR   { $$.Type = $1.Type }
    | STRING { $$.Type = $1.Type }
    ;

array_type
    : type '[' ']' { $$.Type = ArrayType{$1.Type} }
    ;

pair_type
    : PAIR '(' pair_elem_type ',' pair_elem_type ')' { $$.Type = PairType{$3.Type, $5.Type} }
    ;

pair_elem_type
    : base_type   { $$.Expr = $1.Expr }
    | array_type  { $$.Expr = $1.Expr }
    | PAIR        { $$.Expr = $1.Expr }
    ;

pair_elem
    : FST expression { $$.Expr = &PairElemExpr{0, FST, $2.Expr} }
    | SND expression { $$.Expr = &PairElemExpr{0, SND, $2.Expr} }

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
    : identifier '[' expression ']' { $$.Expr = &ArrayElemExpr{0, $1.Expr, $3.Expr} }
    | array_expression '[' expression ']' { $$.Expr = &ArrayElemExpr{0, $1.Expr, $3.Expr} }
    ;

primary_expression
    : identifier          { $$.Expr = &IdentExpr{0, $1.Value} }
    | INT_LIT             { $$.Expr = &BasicLit{0, BasicType{INT}, $1.Value} }
    | BOOL_LIT            { $$.Expr = &BasicLit{0, BasicType{BOOL}, $1.Value} }
    | CHAR_LIT            { $$.Expr = &BasicLit{0, BasicType{CHAR}, $1.Value} }
    | STRING_LIT          { $$.Expr = &BasicLit{0, BasicType{STRING}, $1.Value} }
    | PAIR_LIT            { $$.Expr = &BasicLit{0, BasicType{PAIR}, $1.Value} }
    | '(' expression ')'  { $$.Expr = $2.Expr }
    | array_expression
    ;

unary_expression
    : primary_expression   { $$.Expr = $1.Expr }
    | '!' unary_expression { $$.Expr = &UnaryExpr{0, "!", $2.Expr} }
    | '+' unary_expression { $$.Expr = &UnaryExpr{0, "+", $2.Expr} }
    | '-' unary_expression { $$.Expr = &UnaryExpr{0, "-", $2.Expr} }
    | LEN unary_expression { $$.Expr = &UnaryExpr{0, "len", $2.Expr} }
    | ORD unary_expression { $$.Expr = &UnaryExpr{0, "ord", $2.Expr} }
    | CHR unary_expression { $$.Expr = &UnaryExpr{0, "chr", $2.Expr} }
    ;

multiplicative_expression
    : unary_expression {
        VerifyNoOverflows($1.Expr)
        $$.Expr = $1.Expr
    }
    | multiplicative_expression '*' unary_expression { $$.Expr = &BinaryExpr{$1.Expr, 0, "*", $3.Expr} }
    | multiplicative_expression '/' unary_expression { $$.Expr = &BinaryExpr{$1.Expr, 0, "/", $3.Expr} }
    | multiplicative_expression '%' unary_expression { $$.Expr = &BinaryExpr{$1.Expr, 0, "%", $3.Expr} }
    ;

additive_expression
    : multiplicative_expression { $$.Expr = $1.Expr }
    | additive_expression '+' multiplicative_expression { $$.Expr = &BinaryExpr{$1.Expr, 0, "+", $3.Expr} }
    | additive_expression '-' multiplicative_expression { $$.Expr = &BinaryExpr{$1.Expr, 0, "-", $3.Expr} }
    ;

relational_expression
    : additive_expression { $$.Expr = $1.Expr }
    | relational_expression '<' additive_expression { $$.Expr = &BinaryExpr{$1.Expr, 0, "<", $3.Expr} }
    | relational_expression '>' additive_expression { $$.Expr = &BinaryExpr{$1.Expr, 0, ">", $3.Expr} }
    | relational_expression LE additive_expression { $$.Expr = &BinaryExpr{$1.Expr, 0, "<=", $3.Expr} }
    | relational_expression GE additive_expression { $$.Expr = &BinaryExpr{$1.Expr, 0, ">=", $3.Expr} }
    ;

equality_expression
    : relational_expression { $$.Expr = $1.Expr }
    | equality_expression EQ relational_expression { $$.Expr = &BinaryExpr{$1.Expr, 0, "==", $3.Expr} }
    | equality_expression NE relational_expression { $$.Expr = &BinaryExpr{$1.Expr, 0, "!=", $3.Expr} }
    ;

logical_and_expression
    : equality_expression { $$.Expr = $1.Expr }
    | logical_and_expression AND equality_expression { $$.Expr = &BinaryExpr{$1.Expr, 0, "&&", $3.Expr} }
    ;

logical_or_expression
    : logical_and_expression { $$.Expr = $1.Expr }
    | logical_or_expression OR logical_and_expression { $$.Expr = &BinaryExpr{$1.Expr, 0, "||", $3.Expr} }
    ;

expression
    : logical_or_expression { $$.Expr = $1.Expr }
    ;

%%

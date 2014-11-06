/* File: wacc.y */

%{
  package main
%}

%union {
    Expr Expr
    Value string
    Funcs []Func
    Func  *Func
    Stmts []Stmt
    Stmt  Stmt
    Params []Param
    Param Param
    Kind  int
    Ident Ident
    lines int
}

%{
  var top yySymType
%}

%token BEGIN END
%token INT_LITER BOOL_LITER CHAR_LITER STR_LITER PAIR_LITER
%token INT_SIGN
%token IDENT
%token UNARY_OPER BINARY_OPER
%token SKIP READ FREE RETURN EXIT PRINT PRINTLN NEWPAIR CALL
%token INT BOOL CHAR STRING
%token PAIR
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
        $$.Funcs = append([]Func{*$1.Func}, $2.Funcs...)
      }
    | statement_list { $$.Stmts = $1.Stmts }
    | 
    ;

/* Functions */
func
    : type identifier '(' param_list ')' FUNC_IS statement_list END {
        $$.Func = &Func{0, $1.Kind, $2.Ident, $4.Params, $7.Stmts}
      }
    ;

param_list
    : param ',' param_list { $$.Params = append([]Param{$1.Param}, $2.Params...) }
    | param { $$.Params = []Param{$1.Param} }
    |
    ;

param
    : type identifier { $$.Param = Param{0, $1.Kind, $2.Ident, 0} }
    ;

/* Statements */
statement_list
    : statement { $$.Stmts = []Stmt{$1.Stmt} }
    | statement ';' statement_list = { $$.Stmts = append([]Stmt{$1.Stmt}, $3.Stmts...) }
    ;

statement
    : SKIP { $$.Stmt = &SkipStmt{0} }
    | type identifier '=' assign_rhs { $$.Stmt = &DeclStmt{0, $1.Kind, $2.Ident, $4.Expr} }
    | assign_lhs '=' assign_rhs { $$.Stmt = &AssignStmt{$1.Ident, $3.Expr} }
    | READ assign_lhs {}
    | FREE expression {}
    | RETURN expression { $$.Stmt = &ReturnStmt{0, $2.Expr} }
    | EXIT expression { $$.Stmt = &ExitStmt{0, $2.Expr} }
    | PRINT expression { $$.Stmt = &PrintStmt{0, $2.Expr, false} }
    | PRINTLN expression { $$.Stmt = &PrintStmt{0, $2.Expr, true} }
    | BEGIN statement_list END { $$.Stmts = $2.Stmts }
    | IF expression THEN statement_list ELSE statement_list FI {
        $$.Stmt = &IfStmt{0, $2.Expr, $4.Stmts, $6.Stmts, 0}
      }
    | WHILE expression DO statement_list DONE {
        $$.Stmt = &WhileStmt{ 0, $2.Expr, $4.Stmts ,0 }
      }
    ;

assign_lhs
    : identifier
    | identifier '[' expression ']' /* TODO: REPLACE WITH array_elem */
    | FST identifier
    | SND identifier
    ;

assign_rhs
    : expression {}
    | NEWPAIR '(' expression ',' expression ')' {}
    | CALL identifier '(' arg_list ')' {}
    ;

identifier
    : IDENT { $$.Ident = Ident{0, $1.Value} }
    ;

arg_list
    : expression ',' arg_list {}
    | expression {}
    |
    ;

/* Types */
type
    : base_type
    | array_type
    | pair_type
    ;

base_type
    : INT { $$.Kind = $1.Kind }
    | BOOL { $$.Kind = $1.Kind }
    | CHAR { $$.Kind = $1.Kind }
    | STRING { $$.Kind = $1.Kind }
    ;

array_type
    : type '[' ']'
    ;

pair_type
    : PAIR '(' pair_elem_type ',' pair_elem_type ')'
    ;

pair_elem_type
    : base_type
    | array_type
    | PAIR
    ;

/* Expression */
primary_expression
    : identifier
    | INT_LITER           { $$.Expr = &BasicLit{0, INT_LITER, $1.Value} }
    | BOOL_LITER          { $$.Expr = &BasicLit{0, BOOL_LITER, $1.Value} }
    | CHAR_LITER          { $$.Expr = &BasicLit{0, CHAR_LITER, $1.Value} }
    | STR_LITER           { $$.Expr = &BasicLit{0, STR_LITER, $1.Value} }
    | PAIR_LITER          { $$.Expr = &BasicLit{0, PAIR_LITER, $1.Value} }
    | '(' expression ')'  { $$.Expr = $2.Expr }
    ;

unary_expression
    : primary_expression   { $$.Expr = $1.Expr }
    | '!' unary_expression { $$.Expr = &UnaryExpr{0, "!", $2.Expr} }
    | '+' unary_expression { $$.Expr = &UnaryExpr{0, "+", $2.Expr} }
    | '-' unary_expression { $$.Expr = &UnaryExpr{0, "-", $2.Expr} }
    | LEN unary_expression { $$.Expr = &UnaryExpr{0, "len", $2.Expr} }
    | ORD unary_expression { $$.Expr = &UnaryExpr{0, "ord", $2.Expr} }
    | CHR unary_expression { $$.Expr = &UnaryExpr{0, "chr", $2.Expr} }
    | FST unary_expression { $$.Expr = &UnaryExpr{0, "fst", $2.Expr} }
    | SND unary_expression { $$.Expr = &UnaryExpr{0, "snd", $2.Expr} }
    ;

multiplicative_expression
    : unary_expression { $$.Expr = $1.Expr }
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
/*
expression
    : expression BINARY_OPER expression { $$.Expr = &BinaryExpr{$1.Expr, 0, $2.Value, $3.Expr} }
    | UNARY_OPER expression { $$.Expr = &UnaryExpr{0, $1.Value, $2.Expr} }
    | INT_LITER    { $$.Expr = &BasicLit{0, INT_LITER, $1.Value} }
    | BOOL_LITER   { $$.Expr = &BasicLit{0, BOOL_LITER, $1.Value} }
    | CHAR_LITER   { $$.Expr = &BasicLit{0, CHAR_LITER, $1.Value} }
    | STR_LITER    { $$.Expr = &BasicLit{0, STR_LITER, $1.Value} }
    | PAIR_LITER
    | IDENT
    | array_elem
    | '(' expression ')' { $$.Expr = $2.Expr }
    ;

array_elem
    : IDENT '[' expression ']'
    ;
*/

%%

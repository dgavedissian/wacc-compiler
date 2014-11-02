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
}

%{
  var top yySymType
%}
%token BEGIN END EXIT
%token INT_LITER BOOL_LITER CHAR_LITER STR_LITER PAIR_LITER
%token IDENT
%token UNARY_OPER BINARY_OPER
%token STATEMENT_SEPARATOR
%token SKIP
%token SQUARE_BRACKET_OPEN SQUARE_BRACKET_CLOSE
%token ROUND_BRACKET_OPEN ROUND_BRACKET_CLOSE
%token INT BOOL CHAR STRING
%token PAIR COMMA
%token BASE_TYPE
%token FUNC_IS
%token IF THEN ELSE FI
%%

top
    : program { top = $1 }
    ;

program
    : BEGIN func_list statement_list END { $$.Stmt = &ProgStmt{0, $2.Funcs, $3.Stmts, 0} }
    ;

func_list
    : func func_list { $$.Funcs = append([]Func{*$1.Func}, $2.Funcs...) }
    |
    ;

func
    : type IDENT ROUND_BRACKET_OPEN param_list ROUND_BRACKET_CLOSE FUNC_IS statement_list END {
        $$.Func = &Func{0, $1.Value, $2.Value, $4.Params, $7.Stmts}
    }
    ;

param_list
    : param COMMA param_list { $$.Params = append([]Param{$1.Param}, $2.Params...) }
    | param { $$.Params = []Param{$1.Param} }
    |
    ;

param
    : type IDENT { $$.Param = Param{0, $1.Value, $2.Value, 0} }
    ;

type
    : BASE_TYPE
    | array_type
    | pair_type
    ;

array_type
    : type SQUARE_BRACKET_OPEN SQUARE_BRACKET_CLOSE
    ;

pair_type
    : PAIR ROUND_BRACKET_OPEN pair_elem_type COMMA pair_elem_type ROUND_BRACKET_CLOSE
    ;

pair_elem_type
    : BASE_TYPE
    | array_type
    | PAIR
    ;

statement_list
    : statement { $$.Stmts = []Stmt{$1.Stmt} }
    | statement STATEMENT_SEPARATOR statement_list = { $$.Stmts = append([]Stmt{$1.Stmt}, $3.Stmts...) }
    ;

statement
    : SKIP { $$.Stmt = &SkipStmt{0} }
    | EXIT expression { $$.Stmt = &ExitStmt{0, $2.Expr} }
    | program { $$.Stmt = $1.Stmt }
    | IF expr THEN statement_list ELSE statement_list FI {
        $$.Stmt = &IfStmt{0, $2.Expr, $4.Stmts, $6.Stmts, 0}
      }
    ;

expression
    : expr BINARY_OPER expression { $$.Expr = &BinaryExpr{$1.Expr, 0, $2.Value, $3.Expr} }
    | UNARY_OPER expression { $$.Expr = &UnaryExpr{0, $1.Value, $2.Expr} }
    | expr { $$.Expr = $1.Expr }
    | array_elem
    | ROUND_BRACKET_OPEN expression ROUND_BRACKET_CLOSE
    ;

expr
    : INT_LITER    { $$.Expr = &BasicLit{0, INT_LITER, $1.Value} }
    | BOOL_LITER   { $$.Expr = &BasicLit{0, BOOL_LITER, $1.Value} }
    | CHAR_LITER   { $$.Expr = &BasicLit{0, CHAR_LITER, $1.Value} }
    | STR_LITER    { $$.Expr = &BasicLit{0, STR_LITER, $1.Value} }
    | PAIR_LITER
    | IDENT
    ;

array_elem
    : IDENT SQUARE_BRACKET_OPEN expression SQUARE_BRACKET_CLOSE
    ;
%%

/* File: wacc.y */

%{
  package main
%}

%union {
    BasicLit BasicLit
    ProgStmt ProgStmt
    ExitStmt ExitStmt
    Value string
    Stmts []Stmt
    Stmt  Stmt
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
%%

top
    : program { top = $1 }
    ;

program
    : BEGIN func_list statement_list END { $$.ProgStmt = ProgStmt{0, $2.Stmts, 0} }
    ;

func_list
    : func func_list
    | func
    |
    ;
func
    : type IDENT ROUND_BRACKET_OPEN param_list ROUND_BRACKET_CLOSE FUNC_IS statement_list END
    ;
param_list
    : param COMMA param_list
    | param
    |
    ;
param
    : type IDENT
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
    : SKIP
    | EXIT expression { $$.Stmt = &ExitStmt{0, &$2.BasicLit} }
    | program
    ;

expression
    : expr BINARY_OPER expression
    | UNARY_OPER expression
    | expr
    | array_elem
    | ROUND_BRACKET_OPEN expression ROUND_BRACKET_CLOSE
    ;

expr
    : INT_LITER       { $$.BasicLit = BasicLit{0, INT_LITER, $1.Value} }
    | BOOL_LITER      { $$.BasicLit = BasicLit{0, BOOL_LITER, $1.Value }}
    | CHAR_LITER
    | STR_LITER
    | PAIR_LITER
    | IDENT
    ;

array_elem
    : IDENT SQUARE_BRACKET_OPEN expression SQUARE_BRACKET_CLOSE
    ;
%%

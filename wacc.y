/* File: wacc.y */

%{
  package main
  import ("fmt")
%}

%union {
  int_liter  int
  bool_liter bool
  char_liter byte
  str_liter  string
}
%token BEGIN END EXIT 
%token INT_LITER BOOL_LITER CHAR_LITER STR_LITER PAIR_LITER
%token IDENT
%token UNARY_OPER BINARY_OPER
%token STATEMENT_SEPARATOR
%token SKIP
%token SQUARE_BRACKET_OPEN SQUARE_BRACKET_CLOSE
%token ROUND_BRACKET_OPEN ROUND_BRACKET_CLOSE
%token INT BOOL CHAR STRING
%token PAIR PAIR_SEP
%token EQUALS
%%
program
    : BEGIN statement_list END
    ;

type
    : base_type
    | array_type
    | pair_type
    ;

base_type
    : INT
    | BOOL
    | CHAR
    | STRING
    ;
array_type
    : type SQUARE_BRACKET_OPEN SQUARE_BRACKET_CLOSE
    ;
pair_type
    : PAIR ROUND_BRACKET_OPEN pair_elem_type PAIR_SEP pair_elem_type ROUND_BRACKET_CLOSE
    ;
pair_elem_type
    : base_type
    | array_type
    | PAIR
    ;

statement_list
    : statement
    | statement STATEMENT_SEPARATOR statement_list
    ;

statement
    : SKIP
    | EXIT expression { fmt.Println($2.int_liter) }
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
    : INT_LITER       { fmt.Println($1.int_liter) }
    | BOOL_LITER      { fmt.Println($1.bool_liter) }
    | CHAR_LITER
    | STR_LITER
    | PAIR_LITER
    | IDENT
    ;

array_elem
    : IDENT SQUARE_BRACKET_OPEN expression SQUARE_BRACKET_CLOSE
    ;
%%

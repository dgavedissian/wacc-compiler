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
%token INT_LITER BOOL_LITER
%token SKIP
%%
program
    : BEGIN statement_list END
    ;

statement_list
    : statement
    ;

statement
    : SKIP
    | EXIT expression { fmt.Println($2.int_liter) }
    | program
    ;
expression
    : INT_LITER       { fmt.Println($1.int_liter) }
    | BOOL_LITER      { fmt.Println($1.bool_liter) }
    ;
%%

package main

import (
	"flag"
	"fmt"
	"os"
)

var exitFlag int = 0

func main() {
	enableDebug := flag.Bool("d", false, "Enable debug mode")
	flag.Parse()

	if *enableDebug {
		yyDebug = 20
	}

	// Parse the code, build the AST using the yacc file, and syntax-check.
	// We tend to think of the first line as line 1, not line 0
	lex = NewLexerWithInit(os.Stdin, func(l *Lexer) { l.l = 1 })
	yyParse(lex)
	if exitFlag != 0 {
		os.Exit(exitFlag)
	}

	// Semantic-check the tree
	VerifyProgram(top.Stmt.(*ProgStmt))
	if exitFlag != 0 {
		os.Exit(exitFlag)
	}

	// Print the representation of the tree
	fmt.Println("Compilation successful!")
	fmt.Println(top.Stmt.Repr())
}

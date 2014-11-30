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
	lex = NewLexerWithInit(SetUpErrorOutput(os.Stdin), func(l *Lexer) {
	})
	yyParse(lex)
	if exitFlag != 0 {
		os.Exit(exitFlag)
	}

	// Semantic-check the tree
	VerifyProgram(top.Stmt.(*ProgStmt))
	if exitFlag != 0 {
		os.Exit(exitFlag)
	}
	fmt.Println("Abstract Syntax Tree:")
	fmt.Println(top.Stmt.Repr())
	fmt.Println()

	// Generate the intermediate form
	fmt.Println("Generated intermediate form:")
	iform := GenerateIF(top.Stmt.(*ProgStmt))
	DrawIFGraph(iform)
	fmt.Println()

	// Generate code
	fmt.Println("Generated code:")
	code := GenerateCode(iform)
	fmt.Println(code)

}

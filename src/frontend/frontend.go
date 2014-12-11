package frontend

import "io"

func GenerateAST(input io.Reader) (*ProgStmt, bool) {
	// Generate AST
	lex = NewLexerWithInit(SetUpErrorOutput(input), func(l *Lexer) {})
	yyParse(lex)

	// Syntax errors will have
	if ExitCode() != 0 {
		return nil, true
	}
	program := top.Stmt.(*ProgStmt)

	return program, false
}

func VerifySemantics(ast *ProgStmt) bool {
	verifyProgram(ast)
	return ExitCode() == 0
}

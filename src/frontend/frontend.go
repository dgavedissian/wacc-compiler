package frontend

import "io"

func EnableDebug() {
	yyDebug = 20
}

func GenerateAST(input io.Reader) (*ProgStmt, bool) {
	lex = NewLexerWithInit(SetUpErrorOutput(input), func(l *Lexer) {})
	yyParse(lex)
	if ExitCode() != 0 {
		return nil, true
	}
	program := top.Stmt.(*ProgStmt)
	VerifyProgram(program)
	if ExitCode() != 0 {
		return nil, true
	}
	return program, false
}

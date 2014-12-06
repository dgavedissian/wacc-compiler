package frontend

import "io"

func EnableDebug() {
	yyDebug = 20
}

func GenerateAST(input io.Reader, checkSemantics bool) (*ProgStmt, bool) {
	// Generate AST
	lex = NewLexerWithInit(SetUpErrorOutput(input), func(l *Lexer) {})
	yyParse(lex)
	if ExitCode() != 0 {
		return nil, true
	}
	program := top.Stmt.(*ProgStmt)

	// Verify semantics
	if checkSemantics {
		VerifyProgram(program)
		if ExitCode() != 0 {
			return nil, true
		}
	}

	return program, false
}

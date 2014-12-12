package frontend

import "io"

var lex *Lexer

func processEscapedCharacters(s string) string {
	s = s[1 : len(s)-1]

	// Replace escaped characters with their unicode equivalent
	output := ""
	for i := 0; i < len(s); i++ {
		if s[i] == '\\' {
			i++
			switch s[i] {
			case '0':
				output += "\000"
			case 'b':
				output += "\b"
			case 't':
				output += "\t"
			case 'n':
				output += "\n"
			case 'f':
				output += "\f"
			case 'r':
				output += "\r"
			case '\047':
				output += "\047"
			case '\042':
				output += "\042"
			case '\\':
				output += "\\"
			default:
				panic("Encountered an unknown escape sequence, this should never happen")
			}
		} else {
			output += string(s[i])
		}
	}
	return output
}

// Error callback for nex
func (l *Lexer) Error(s string) {
	pos := NewPositionFromLexer(l)
	if len(l.stack) > 0 {
		unexpectedToken := l.Text()
		SyntaxError(pos, "unexpected '%s'", unexpectedToken)
	} else {
		SyntaxError(pos, "unexpected '<EOF>'")
	}
}

func GenerateAST(input io.Reader) (*ProgStmt, bool) {
	// Generate AST
	lex = NewLexer(SetUpErrorOutput(input))
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

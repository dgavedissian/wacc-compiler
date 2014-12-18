package frontend

import (
	"io"
	"unicode/utf8"
)

const INT_MIN = -(1 << 31)
const INT_MAX = (1 << 31) - 1

// Empty error callback for nex
func (l *Lexer) Error(s string) {}

// Decorate Nex Lexer
type WACCLexer struct {
	lexer   *Lexer
	program *Program
	err     bool
}

func (l *WACCLexer) Lex(lval *yySymType) int { return l.lexer.Lex(lval) }
func (l *WACCLexer) Error(e string) {
	pos := NewPositionFromLexer(l.lexer)
	if len(l.lexer.stack) > 0 {
		unexpectedToken := l.lexer.Text()
		SyntaxError(pos, "unexpected '%s'", unexpectedToken)
	} else {
		SyntaxError(pos, "unexpected '<EOF>'")
	}
	l.err = true
}

func GenerateAST(input io.Reader) (*Program, bool) {
	// Generate AST
	yyDebug = 2
	generateAST := func(input io.Reader) (*Program, bool) {
		lexer := &WACCLexer{NewLexer(SetUpErrorOutput(input)), nil, false}
		yyParse(lexer)
		if lexer.err {
			return nil, false
		}
		return lexer.program, true
	}
	program, ok := generateAST(input)
	if !ok {
		return nil, true
	}

	return program, false
}

func VerifySemantics(ast *Program) bool {
	verifyProgram(ast)
	return ExitCode() == 0
}

func processEscapedCharacters(s string) string {
	s = s[1 : len(s)-1]

	// Replace escaped characters with their unicode equivalent
	output := ""
	for i, w := 0, 0; i < len(s); i += w {
		runeValue, width := utf8.DecodeRuneInString(s[i:])
		if runeValue == '\\' {
			i += width
			runeValue, width = utf8.DecodeRuneInString(s[i:])
			switch runeValue {
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
			output += string(runeValue)
		}
		w = width
	}
	return output
}

package frontend

import (
	"io"
	"unicode/utf8"
)

const INT_MIN = -(1 << 31)
const INT_MAX = (1 << 31) - 1

// Error callback for nex
func (l *Lexer) Error(s string) {
	pos := NewPositionFromLexer(l)
	if len(l.stack) > 0 {
		unexpectedToken := l.Text()
		SyntaxError(pos, "unexpected '%s'", unexpectedToken)
	} else {
		SyntaxError(pos, "unexpected '<EOF>'")
	}
	l.err = true
}

func GenerateAST(input io.Reader) (*Program, bool) {
	// Generate AST
	generateAST := func(input io.Reader) (*Program, bool) {
		lexer := NewLexer(SetUpErrorOutput(input))
		yyParse(lexer)
		if lexer.err {
			return nil, false
		}
		return lexer.program, true
	}
	return generateAST(input)
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

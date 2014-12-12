package frontend

import (
	"bytes"
	"fmt"
	"io"
	"strings"

	"gitlab.doc.ic.ac.uk/np1813/ansi"
)

const SYNTAX_ERROR = 100
const SEMANTIC_ERROR = 200

const POINTER_LINE_CHAR = '~'
const POINTER_POINTER_CHAR = '^'

var ERROR_COLOUR = ansi.ColorCode("red+h")
var LINE_COLOUR = ansi.ColorCode("reset")
var BAD_LINE_COLOUR = ansi.ColorCode("yellow")
var POINTER_COLOUR = ansi.ColorCode("reset")
var RESET = ansi.ColorCode("reset")

const CONTEXT_TO_PRINT = 2

var exitCode int

func ExitCode() int {
	return exitCode
}

func SyntaxError(position *Position, s string, a ...interface{}) {
	errorStr := fmt.Sprintf(s, a...)
	fmt.Print(ERROR_COLOUR)
	fmt.Printf("%v:%v:%v: syntax error: %v\n", position.Name(), position.Line(), position.Column(), errorStr)
	dumpLineData(position)
	fmt.Print(RESET)
	exitCode = SYNTAX_ERROR
}

func SemanticError(position *Position, s string, a ...interface{}) {
	errorStr := fmt.Sprintf(s, a...)
	fmt.Print(ERROR_COLOUR)
	fmt.Printf("%v:%v:%v: semantic error: %v\n", position.Name(), position.Line(), position.Column(), errorStr)
	dumpLineData(position)
	fmt.Print(RESET)
	exitCode = SEMANTIC_ERROR
}

var errBuffer []string

func SetUpErrorOutput(r io.Reader) io.Reader {
	// we need to read the entire set of lines into a buffer so we can do
	// pretty error output
	buf := new(bytes.Buffer)
	if _, err := buf.ReadFrom(r); err != nil {
		panic(err)
	}

	strdata := string(buf.Bytes())
	errBuffer = strings.Split(strings.Replace(strdata, "\t", " ", -1), "\n")

	return buf
}

func dumpLineData(position *Position) {
	lineToPrint := position.Line() - 1

	if len(errBuffer) <= lineToPrint {
		return
	}

	for currentLine := lineToPrint - CONTEXT_TO_PRINT; currentLine < lineToPrint; currentLine += 1 {
		if currentLine < 0 {
			continue
		}
		fmt.Println(LINE_COLOUR + errBuffer[currentLine])
	}

	fmt.Println(BAD_LINE_COLOUR + errBuffer[lineToPrint])

	dashes := strings.Repeat(string(POINTER_LINE_CHAR), position.Column()-1)
	caret := string(POINTER_POINTER_CHAR)
	fmt.Println(POINTER_COLOUR + dashes + caret)

	for currentLine := lineToPrint + 1; currentLine <= lineToPrint+CONTEXT_TO_PRINT; currentLine += 1 {
		if currentLine >= len(errBuffer) {
			continue
		}
		fmt.Println(LINE_COLOUR + errBuffer[currentLine])
	}

	fmt.Print(RESET)
}

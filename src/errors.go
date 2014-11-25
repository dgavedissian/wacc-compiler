package main

import (
	"bytes"
	"fmt"
	"io"
	"strings"

	"github.com/mgutz/ansi"
)

const SYNTAX_ERROR = 100
const SEMANTIC_ERROR = 200

const POINTER_LINE_CHAR = '-'
const POINTER_POINTER_CHAR = '^'

var ERROR_COLOUR = ansi.ColorCode("red+h:black")
var LINE_COLOUR = ansi.ColorCode("yellow:black")
var BAD_LINE_COLOUR = ansi.ColorCode("yellow+h:black")
var POINTER_COLOUR = ansi.ColorCode("green+h:black")
var RESET = ansi.ColorCode("reset")

const CONTEXT_TO_PRINT = 3

func SyntaxError(lineNo int, s string, a ...interface{}) {
	errorStr := fmt.Sprintf(s, a...)
	fmt.Print(ERROR_COLOUR)
	if lineNo > 0 {
		fmt.Printf("Line %d: %s\n", lineNo, errorStr)
	} else {
		fmt.Printf("%s\n", errorStr)
	}
	fmt.Print(RESET)
	exitFlag = SYNTAX_ERROR
}

func SemanticError(position *Position, s string, a ...interface{}) {
	errorStr := fmt.Sprintf(s, a...)
	fmt.Print(ERROR_COLOUR)
	if position != nil {
		fmt.Printf("Line %d: %s\n", position.Line(), errorStr)
		dumpLineData(position)
	} else {
		fmt.Printf("%s\n", errorStr)
	}
	fmt.Print(RESET)
	exitFlag = SEMANTIC_ERROR
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

	for currentLine := lineToPrint + 1; currentLine < lineToPrint+CONTEXT_TO_PRINT; currentLine += 1 {
		if currentLine >= len(errBuffer) {
			continue
		}
		fmt.Println(LINE_COLOUR + errBuffer[currentLine])
	}

	fmt.Println(RESET)
}

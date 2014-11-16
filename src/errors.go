package main

import (
	"bytes"
	"fmt"
	"github.com/mgutz/ansi"
	"io"
	"strings"
)

const SYNTAX_ERROR = 100
const SEMANTIC_ERROR = 200

const POINTER_LINE_CHAR = '-'
const POINTER_POINTER_CHAR = '^'

var ERROR_COLOUR = ansi.ColorCode("red+h:black")
var BAD_LINE_COLOUR = ansi.ColorCode("yellow+h:black")
var POINTER_COLOUR = ansi.ColorCode("green+h:black")
var RESET = ansi.ColorCode("reset")

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
	if len(errBuffer) < position.Line() {
		return
	}

	fmt.Println(BAD_LINE_COLOUR + errBuffer[position.Line()-1] + RESET)

	dashes := strings.Repeat(string(POINTER_LINE_CHAR), position.Column()-1)
	caret := string(POINTER_POINTER_CHAR)
	fmt.Println(POINTER_COLOUR + dashes + caret + RESET)
	fmt.Println()
}

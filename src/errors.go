package main

import (
	"fmt"
)

const SYNTAX_ERROR = 100
const SEMANTIC_ERROR = 200

func SyntaxError(lineNo int, s string, a ...interface{}) {
	errorStr := fmt.Sprintf(s, a...)
	fmt.Printf("Line %d: %s\n", lineNo, errorStr)
	exitFlag = SYNTAX_ERROR
}

func SemanticError(lineNo int, s string, a ...interface{}) {
	errorStr := fmt.Sprintf(s, a...)
	fmt.Printf("Line %d: %s\n", lineNo, errorStr)
	exitFlag = SEMANTIC_ERROR
}

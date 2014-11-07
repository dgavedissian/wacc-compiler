# Locations
SHELL       := /bin/bash
SOURCE_DIR	:= src
OUTPUT_DIR	:= bin 

# Tools
FIND	:= find
RM	:= rm -rf
MKDIR	:= mkdir -p
NEX     := nex
GO      := GOPATH=$$HOME/go go


# the make rules

all: frontend

parser.go: go wacc.y
	$(GO) tool yacc -o parser.go wacc.y

lexer.go: go nex wacc.nex
	$$HOME/go/bin/nex -e=true -o lexer.go wacc.nex

frontend: parser.go lexer.go ast.go
	$(GO) build -o frontend $^

clean:
	$(GO) clean
	$(RM) parser.go lexer.go frontend y.output

nex:
	$(GO) get github.com/blynn/nex

go:
	./installgo.sh

test: testvalid testinvalid
	@echo "Tests complete"

testvalid: frontend
	@echo "Testing valid cases..."
	@find ./wacc_examples/valid/ -name *.wacc -exec ./compile -x {} ";" | awk '{run+=1; if ($$0 != 0){ failed+=1; }} END {print "VALID:", run - failed, "/", run, "tests passed";}'

testinvalid: frontend
	@echo "Testing invalid cases..."
	@find ./wacc_examples/invalid/syntaxErr -name *.wacc -exec ./compile -x {} ";" | awk '{run+=1; if ($$0 == 0){ failed+=1; }} END {print "INVALID:", run - failed, "/", run, "tests passed";}'

.PHONY: clean all nex test go testvalid testinvalid

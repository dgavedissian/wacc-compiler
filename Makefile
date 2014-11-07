# Sample Makefile for the WACC Compiler lab: edit this to build your own comiler
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

GOPACKGE = $(shell if [ -z "`dpkg -l | grep golang-go`" ]; then sudo apt-get install -y golang-go; fi)


# the make rules

all: go frontend

parser.go: go wacc.y
	$(GO) tool yacc -o parser.go wacc.y

lexer.go: go wacc.nex
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

# make test invalids=~/labs/wacc_examples/invalid/ valids=~/labs/wacc_examples/valid/
test:
	@echo "Running tests.."
	@[ -n "$(valids)" ] && \
		find $(valids) -name *.wacc -exec ./compile -x {} ";" | awk '{run+=1; if ($$0 == 100){ failed+=1; }} END {print "VALID:", run - failed, "/", run, "tests passed";}'; \
 	[ -n "$(invalids)" ] && \
		find $(invalids) -name *.wacc -exec ./compile -x {} ";" | awk '{run+=1; if ($$0 == 0){failed+=1;}} END {print "INVALID:", run - failed, "/", run, "tests passed";}'

.PHONY: clean all nex test go

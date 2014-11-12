# Locations
SHELL       := /bin/bash

BASE_DIR    := $(shell dirname $(realpath $(lastword $(MAKEFILE_LIST))))
SOURCE_DIR	:= src
SCRIPTS_DIR := scripts
TESTS_DIR   := tests

# Tools
FIND	:= find
RM	    := rm -rf
MKDIR	:= mkdir -p
NEX     := $$HOME/go/bin/nex
GO      := GOPATH=$$HOME/go go


# the make rules

all: frontend

frontend: $(SOURCE_DIR)/parser.go $(SOURCE_DIR)/lexer.go $(SOURCE_DIR)/ast.go $(SOURCE_DIR)/syntax.go $(SOURCE_DIR)/semantic.go
	$(GO) build -o frontend $^

$(SOURCE_DIR)/parser.go: go $(SOURCE_DIR)/wacc.y
	$(GO) tool yacc -o $(SOURCE_DIR)/parser.go -v y.output $(SOURCE_DIR)/wacc.y

$(SOURCE_DIR)/lexer.go: go nex $(SOURCE_DIR)/wacc.nex
	$(NEX) -e=true -o $(SOURCE_DIR)/lexer.go $(SOURCE_DIR)/wacc.nex


nex:
	$(GO) get github.com/blynn/nex

go:
	$(SCRIPTS_DIR)/installgo.sh

clean:
	$(GO) clean
	$(RM) $(SOURCE_DIR)/parser.go $(SOURCE_DIR)/lexer.go frontend


test: testvalid testinvalidsyntax testinvalidsemantic
	@echo "Tests complete"

testvalid: frontend
	@echo "Testing valid cases..."
	@find $(TESTS_DIR)/wacc_examples/valid/ -name *.wacc | xargs -n 1 -P 4 $(BASE_DIR)/compile -x | awk '{run+=1; if ($$0 != 0){ failed+=1; }} END {print "VALID:", run - failed, "/", run, "tests passed";}'

testinvalidsyntax: frontend
	@echo "Testing invalid syntax cases..."
	@find $(TESTS_DIR)/wacc_examples/invalid/syntaxErr -name *.wacc | xargs -n 1 -P 4 $(BASE_DIR)/compile -x | awk '{run+=1; if ($$0 != 100){ failed+=1; }} END {print "INVALID SYNTAX:", run - failed, "/", run, "tests passed";}'

testinvalidsemantic: frontend
	@echo "Testing invalid semantic cases..."
	@find $(TESTS_DIR)/wacc_examples/invalid/semanticErr -name *.wacc | xargs -n 1 -P 4 $(BASE_DIR)/compile -x | awk '{run+=1; if ($$0 != 200){ failed+=1; }} END {print "INVALID SEMANTIC:", run - failed, "/", run, "tests passed";}'


.PHONY: clean all nex test go testvalid testinvalidsyntax testinvalidsemantic

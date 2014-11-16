# Locations
SHELL        := /bin/bash

BASE_DIR     := $(shell dirname $(realpath $(lastword $(MAKEFILE_LIST))))
SOURCE_DIR	 := src
SCRIPTS_DIR  := scripts
EXAMPLES_DIR := examples

# Tools
FIND	:= find
RM	    := rm -rf
MKDIR	:= mkdir -p
NEX     := $$HOME/go/bin/nex
GO      := GOPATH=$$HOME/go go


# the make rules

all: frontend

frontend: $(SOURCE_DIR)/parser.go $(SOURCE_DIR)/lexer.go $(SOURCE_DIR)/ast.go $(SOURCE_DIR)/syntax.go $(SOURCE_DIR)/semantic.go $(SOURCE_DIR)/errors.go $(SOURCE_DIR)/main.go $(SOURCE_DIR)/position.go
	$(GO) build -o frontend $^

$(SOURCE_DIR)/parser.go: go $(SOURCE_DIR)/wacc.y
	$(GO) tool yacc -o $(SOURCE_DIR)/parser.go -v y.output $(SOURCE_DIR)/wacc.y

$(SOURCE_DIR)/lexer.go: go nex $(SOURCE_DIR)/wacc.nex $(SCRIPTS_DIR)/add_fields.sh
	$(NEX) -e=true -o $(SOURCE_DIR)/lexer.go $(SOURCE_DIR)/wacc.nex
	$(SCRIPTS_DIR)/add_fields.sh $(SOURCE_DIR)/lexer.go


nex:
	$(GO) get github.com/blynn/nex

go:
	$(SCRIPTS_DIR)/installgo.sh

clean:
	$(GO) clean
	$(RM) $(SOURCE_DIR)/parser.go $(SOURCE_DIR)/lexer.go frontend


test: frontend
	$(SCRIPTS_DIR)/test_examples.py

testvalid: frontend
	$(SCRIPTS_DIR)/test_examples.py "Valid"

testinvalidsyntax: frontend
	$(SCRIPTS_DIR)/test_examples.py "Invalid Syntax"

testinvalidsemantic: frontend
	$(SCRIPTS_DIR)/test_examples.py "Invalid Semantic"


.PHONY: clean all nex test go testvalid testinvalidsyntax testinvalidsemantic

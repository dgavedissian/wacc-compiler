# Locations
SHELL        := /bin/bash

BASE_DIR     := $(shell dirname $(realpath $(lastword $(MAKEFILE_LIST))))
SOURCE_DIR	 := src
SCRIPTS_DIR  := scripts
EXAMPLES_DIR := examples

# Tools
FIND	:= find
RM	    := rm -rf
NEX     := $$HOME/go/bin/nex
GO      := GOPATH=$$HOME/go go
GOGET   := $(GO) get

SOURCE_FILES := \
	$(SOURCE_DIR)/parser.go \
	$(SOURCE_DIR)/lexer.go \
	$(SOURCE_DIR)/ast.go \
	$(SOURCE_DIR)/errors.go \
	$(SOURCE_DIR)/generator.go \
	$(SOURCE_DIR)/if.go \
	$(SOURCE_DIR)/main.go \
	$(SOURCE_DIR)/position.go \
	$(SOURCE_DIR)/semantic.go \
	$(SOURCE_DIR)/syntax.go


# the make rules

all: compile

compile: godeps $(SOURCE_FILES)
	$(GO) build -o compile $(SOURCE_FILES)

$(SOURCE_DIR)/parser.go: go $(SOURCE_DIR)/wacc.y
	$(GO) tool yacc -o $(SOURCE_DIR)/parser.go -v y.output $(SOURCE_DIR)/wacc.y

$(SOURCE_DIR)/lexer.go: godeps $(SOURCE_DIR)/wacc.nex
	$(NEX) -e=true -o $(SOURCE_DIR)/lexer.go $(SOURCE_DIR)/wacc.nex


nex: go
	$(GOGET) gitlab.doc.ic.ac.uk/np1813/nex

ansi: go
	$(GOGET) gitlab.doc.ic.ac.uk/np1813/ansi

godeps: nex ansi

go:
	$(SCRIPTS_DIR)/installgo.sh

clean:
	$(GO) clean
	$(RM) $(SOURCE_DIR)/parser.go $(SOURCE_DIR)/lexer.go compile y.output


test: compile
	$(SCRIPTS_DIR)/test_examples.py

testvalid: compile
	$(SCRIPTS_DIR)/test_examples.py "Valid"

testinvalidsyntax: compile
	$(SCRIPTS_DIR)/test_examples.py "Invalid Syntax"

testinvalidsemantic: compile
	$(SCRIPTS_DIR)/test_examples.py "Invalid Semantic"


.PHONY: clean all nex test go testvalid testinvalidsyntax testinvalidsemantic

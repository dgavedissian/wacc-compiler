# Locations
SHELL        := /bin/bash

BASE_DIR     := $(shell dirname $(realpath $(lastword $(MAKEFILE_LIST))))
SOURCE_DIR	 := src
FRONTEND_DIR := $(SOURCE_DIR)/frontend
BACKEND_DIR  := $(SOURCE_DIR)/backend
ERRORS_DIR   := $(SOURCE_DIR)/errors
SCRIPTS_DIR  := scripts
EXAMPLES_DIR := examples

# Tools
FIND	:= find
RM	    := rm -rf
NEX     := $$HOME/go/bin/nex
GO      := GOPATH=$$HOME/go go
GOGET   := $(GO) get

FRONTEND_FILES := \
	$(FRONTEND_DIR)/parser.go \
	$(FRONTEND_DIR)/lexer.go \
	$(FRONTEND_DIR)/ast.go \
	$(FRONTEND_DIR)/semantic.go \
	$(FRONTEND_DIR)/syntax.go \
	$(FRONTEND_DIR)/position.go

ERRORS_FILES := \
	$(ERRORS_DIR)/errors.go

BACKEND_FILES := \
	$(BACKEND_DIR)/generator.go \
	$(BACKEND_DIR)/if.go \
	$(BACKEND_DIR)/optimiser.go \
	$(BACKEND_DIR)/registers.go

MAIN_FILES := \
	$(SOURCE_DIR)/main.go


# the make rules

all: compile

compile: godeps generated $(SOURCE_FILES)
	$(GO) build -o compile $(MAIN_FILES)

generated: $(FRONTEND_DIR)/parser.go $(FRONTEND_DIR)/lexer.go

$(FRONTEND_DIR)/parser.go: godeps $(FRONTEND_DIR)/parser.y
	$(GO) tool yacc -o $(FRONTEND_DIR)/parser.go -v y.output $(FRONTEND_DIR)/parser.y

$(FRONTEND_DIR)/lexer.go: godeps $(FRONTEND_DIR)/lexer.nex
	$(NEX) -e=true -o $(FRONTEND_DIR)/lexer.go $(FRONTEND_DIR)/lexer.nex


nex: go
	$(GOGET) gitlab.doc.ic.ac.uk/np1813/nex

ansi: go
	$(GOGET) gitlab.doc.ic.ac.uk/np1813/ansi

godeps: nex ansi

go:
	$(SCRIPTS_DIR)/installgo.sh

clean:
	$(GO) clean
	$(RM) $(FRONTEND_DIR)/parser.go $(FRONTEND_DIR)/lexer.go compile y.output


test: compile
	$(SCRIPTS_DIR)/test_examples.py

testvalid: compile
	$(SCRIPTS_DIR)/test_examples.py "Valid"

testinvalidsyntax: compile
	$(SCRIPTS_DIR)/test_examples.py "Invalid Syntax"

testinvalidsemantic: compile
	$(SCRIPTS_DIR)/test_examples.py "Invalid Semantic"


.PHONY: clean all nex test go testvalid testinvalidsyntax testinvalidsemantic generated

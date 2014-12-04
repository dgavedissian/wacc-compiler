package main

import (
	"flag"
	"fmt"
	"os"
	"path"

	"./backend"
	"./frontend"
)

var exitFlag int = 0

func main() {
	enableDebug := flag.Bool("d", false, "Enable debug mode")
	enableVerbose := flag.Bool("v", true, "Enable verbose logging")
	stopAtAST := flag.Bool("ast", false, "Stop the compile process once the AST has been generated")
	stopAtIF := flag.Bool("if", true, "Stop the compile process once the IF representation has been generated")
	flag.Parse()

	// Open file specified in the remaining argument
	filename := flag.Arg(0)
	input := os.Stdin
	useStdin := true
	if filename != "-" {
		f, err := os.Open(filename)
		if err != nil {
			panic(err)
		}
		input = f
		useStdin = false
	}

	if *enableDebug {
		frontend.EnableDebug()
	}

	// Generate the AST
	ast, asterr := frontend.GenerateAST(input)
	if asterr {
		os.Exit(frontend.ExitCode())
	}
	if *enableVerbose {
		fmt.Println("Abstract Syntax Tree:")
		fmt.Println(ast.Repr())
		fmt.Println()
	}

	if *stopAtAST {
		return
	}

	// Generate the intermediate form
	iform := backend.TranslateToIF(ast)
	if *enableVerbose {
		fmt.Println("First pass intermediate form:")
		backend.DrawIFGraph(iform)
		fmt.Println()
	}

	// Optimise the intermediate form
	backend.AllocateRegisters(iform)
	backend.OptimiseIF(iform)
	if *enableVerbose {
		fmt.Println("Second pass intermediate form:")
		backend.DrawIFGraph(iform)
		fmt.Println()
	}

	if *stopAtIF {
		return
	}

	// Generate code
	code := backend.GenerateCode(iform)
	if *enableVerbose {
		fmt.Println("Generated code:")
		fmt.Println(code)
	}

	outFile := "out"
	if useStdin == false {
		// Extract source code name from file
		basename := path.Base(filename)
		outFile = basename[:len(basename)-len(path.Ext(filename))]
	}

	// Generate final assembly file
	f, err := os.Create(outFile + ".s")
	if err != nil {
		panic("Unable to open output file")
	}
	f.WriteString(code)
	f.Close()
}

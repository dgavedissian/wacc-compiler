package main

import (
	"flag"
	"fmt"
	"os"

	"./backend"
	"./frontend"
)

var exitFlag int = 0

func main() {
	enableDebug := flag.Bool("d", false, "Enable debug mode")
	stopAtIF := flag.Bool("if", true, "Stop the compile process once IF is generated")
	flag.Parse()

	// Open file specified in the remaining argument
	filename := flag.Arg(0)
	input := os.Stdin
	if filename != "-" {
		f, err := os.Open(filename)
		if err != nil {
			panic(err)
		}
		input = f
	}

	if *enableDebug {
		frontend.EnableDebug()
	}

	ast, err := frontend.GenerateAST(input)
	if err {
		os.Exit(frontend.ExitCode())
	}

	fmt.Println("Abstract Syntax Tree:")
	fmt.Println(ast.Repr())
	fmt.Println()

	// Generate the intermediate form
	fmt.Println("Generated intermediate form:")
	iform := backend.GenerateIF(ast)
	backend.DrawIFGraph(iform)
	fmt.Println()

	// Generate code
	if *stopAtIF == false {
		fmt.Println("Generated code:")
		code := backend.GenerateCode(iform)
		fmt.Println(code)

		// Write final .s file
		f, err := os.Create("test.s")
		if err != nil {
			panic("Unable to open output file")
		}
		f.WriteString(code)
		f.Close()
	}
}

package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"./backend"
	"./frontend"
)

const INTERRUPT_CODE = 0
const OK_CODE = 1

func compilerFactory(verbose bool, astonly bool, ifonly bool, checkSemantics bool) func(string, *os.File) (string, int) {
	compiler := func(modulePath string, input *os.File) (string, int) {
		// Generate AST for input file
		ast, astOk := frontend.GenerateAST(modulePath, input)
		if !astOk {
			return "", frontend.SYNTAX_ERROR
		}

		// Perform semantic checks
		if checkSemantics {
			semanticOk := frontend.VerifyProgram(ast)
			if !semanticOk {
				return "", frontend.SEMANTIC_ERROR
			}
		}

		// Print the current AST
		if verbose {
			fmt.Println("Abstract Syntax Tree:")
			fmt.Println(ast.Repr())
			fmt.Println()
		}

		if astonly {
			return "", INTERRUPT_CODE
		}

		// Translate to intermediate form
		intermediateForm := backend.TranslateToIF(ast)
		if verbose {
			fmt.Println("First pass intermediate form")
			backend.DrawIFGraph(intermediateForm)
			fmt.Println()
		}

		// Perform optimisation and register-allocation passes over IF
		backend.OptimiseFirstPassIF(intermediateForm)
		backend.AllocateRegisters(intermediateForm)
		backend.OptimiseSecondPassIF(intermediateForm)

		if verbose {
			fmt.Println("Second pass intermediate form")
			backend.DrawIFGraph(intermediateForm)
			fmt.Println()
		}

		if ifonly {
			return "", INTERRUPT_CODE
		}

		// Generate final assembly code
		asm := backend.GenerateCode(intermediateForm)
		if verbose {
			fmt.Println("Assembly")
			fmt.Println(asm)
		}

		return asm, OK_CODE
	}

	return compiler
}

func main() {
	// Command line arguments
	verboseFlag := flag.Bool("v", false, "Enable verbose logging")
	astonlyFlag := flag.Bool("ast", false, "Stop the compile process once the AST has been generated")
	ifonlyFlag := flag.Bool("if", false, "Stop the compile process once the IF representation has been generated")
	disableSemanticFlag := flag.Bool("i-know-what-im-doing", false, "Disable semantic checking")
	outFile := flag.String("o", "out.s", "File to write asm to")
	modulePathFlag := flag.String("mp", "", "Module path")
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

	// This is our nod to Java, but with a Golang twist
	// Create a compile function with the flags we like
	compile := compilerFactory(*verboseFlag, *astonlyFlag, *ifonlyFlag, !*disableSemanticFlag)

	// Compile the source code
	modulePath := *modulePathFlag
	if modulePath == "" && useStdin == false {
		modulePath, _ = filepath.Abs(filepath.Dir(filename))
	}
	asm, compileCode := compile(modulePath, input)
	if compileCode != OK_CODE {
		os.Exit(compileCode)
	}

	// Grab the assembled code from the source name.
	if useStdin == false && *outFile == "out.s" {
		// Extract source code name from file
		basename := filepath.Base(filename)
		*outFile = basename[:len(basename)-len(filepath.Ext(filename))] + ".s"
	}

	// Save assembly to file
	f, err := os.Create(*outFile)
	if err != nil {
		panic("Unable to open output file")
	}

	f.WriteString(asm)
	f.Close()
}

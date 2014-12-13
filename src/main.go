package main

import (
	"flag"
	"fmt"
	"os"
	"path"

	"./backend"
	"./frontend"
)

const INTERRUPT_CODE = 0
const OK_CODE = 1

func compilerFactory(verbose bool, astonly bool, ifonly bool, checkSemantics bool) func(*os.File) (string, int) {
	compiler := func(input *os.File) (string, int) {
		ast, asterr := frontend.GenerateAST(input)

		if asterr {
			return "", frontend.SYNTAX_ERROR
		}

		semanticOk := frontend.VerifySemantics(ast)

		if checkSemantics && !semanticOk {
			return "", frontend.SEMANTIC_ERROR
		}

		if verbose {
			fmt.Println("Abstract Syntax Tree:")
			fmt.Println(ast.Repr())
			fmt.Println()
		}

		if astonly {
			return "", INTERRUPT_CODE
		}

		intermediateForm := backend.TranslateToIF(ast)

		if verbose {
			fmt.Println("First pass intermediate form")
			backend.DrawIFGraph(intermediateForm)
			fmt.Println()
		}

		backend.AllocateRegisters(intermediateForm)
		backend.OptimiseIF(intermediateForm)

		if verbose {
			fmt.Println("Second pass intermediate form")
			backend.DrawIFGraph(intermediateForm)
			fmt.Println()
		}

		if ifonly {
			return "", INTERRUPT_CODE
		}

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
	verboseFlag := flag.Bool("v", false, "Enable verbose logging")
	astonlyFlag := flag.Bool("ast", false, "Stop the compile process once the AST has been generated")
	ifonlyFlag := flag.Bool("if", false, "Stop the compile process once the IF representation has been generated")
	disableSemanticFlag := flag.Bool("i-know-what-im-doing", false, "Disable semantic checking")
	outFile := flag.String("o", "out.s", "File to write asm to")
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
	asm, compileCode := compile(input)

	if compileCode != OK_CODE {
		os.Exit(compileCode)
	}

	// Grab the assembled code from the source name.
	if useStdin == false && *outFile == "out.s" {
		// Extract source code name from file
		basename := path.Base(filename)
		*outFile = basename[:len(basename)-len(path.Ext(filename))] + ".s"
	}

	// Save assembly to file
	f, err := os.Create(*outFile)
	if err != nil {
		panic("Unable to open output file")
	}

	f.WriteString(asm)
	f.Close()
}

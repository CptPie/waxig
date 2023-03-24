package repl

import (
	"bufio"
	"fmt"
	"io"
	"waixg/interpreter/lexer"
	"waixg/interpreter/parser"
)

const PROMPT = ">> "

func Start(in io.Reader, out io.Writer) {
	scanner := bufio.NewScanner(in)

	for {
		_, _ = fmt.Fprintf(out, PROMPT)
		scanned := scanner.Scan()
		if !scanned {
			return
		}

		line := scanner.Text()
		l := lexer.New(line)
		p := parser.New(l)

		program := p.ParseProgram()

		if len(p.Errors()) != 0 {
			printParserErrors(out, p.Errors())
			continue
		}

		_, _ = io.WriteString(out, program.String())
		_, _ = io.WriteString(out, "\n")

	}
}

func printParserErrors(out io.Writer, errors []error) {
	for _, msg := range errors {
		_, _ = fmt.Fprintf(out, "\t%s\n", msg)
	}
}

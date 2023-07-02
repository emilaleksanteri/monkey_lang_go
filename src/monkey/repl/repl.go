package repl

import (
	"bufio"
	"fmt"
	"io"
	"monkey/lexer"
	"monkey/token"
)

const PROMT = ">>"

func Start(in io.Reader, out io.Writer) {
	scanner := bufio.NewScanner(in)

	for {
		fmt.Fprintf(out, PROMT)
		scanned := scanner.Scan()
		// exit loop when nothing to read
		if !scanned {
			return
		}

		line := scanner.Text()
		l := lexer.New(line)
		
		// tokenize till end of line, return print all
		for tok := l.NextToken(); tok.Type != token.EOF; tok = l.NextToken() {
			fmt.Fprintf(out, "%+v\n", tok)
		}
	}
}

package repl

import (
	"bufio"
	"fmt"
	"os"

	"github.com/Jonaires777/src/lexer"
	"github.com/Jonaires777/src/parser"
)

const prompt = `jwfs>> `

func Start() {
	scanner := bufio.NewScanner(os.Stdin)

	for {
		fmt.Print(prompt)
		scanned := scanner.Scan()
		if !scanned {
			return
		}

		line := scanner.Text()
		l := lexer.New(line)
		p := parser.New(l)

		program := p.ParseCommand()

		fmt.Println(program)
	}
}

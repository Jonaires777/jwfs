package repl

import (
	"fmt"
	"os"

	"github.com/chzyer/readline"

	"github.com/Jonaires777/src/lexer"
	"github.com/Jonaires777/src/parser"
)

const prompt = `jwfs>> `

func Start() {
	rl, err := readline.NewEx(&readline.Config{
		Prompt:                 prompt,
		HistoryFile:            "/tmp/readline.tmp",
		DisableAutoSaveHistory: false,
	})
	if err != nil {
		panic(err)
	}

	defer rl.Close()

	for {
		line, err := rl.Readline()
		if err != nil {
			break
		}

		if line == "" {
			continue
		}

		if line == "exit" {
			os.Exit(0)
		}

		l := lexer.New(line)
		p := parser.New(l)

		program := p.ParseCommand()

		fmt.Println(program)
	}
}

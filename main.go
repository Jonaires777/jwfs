package main

import (
	"fmt"
	"os/user"

	"github.com/Jonaires777/src/repl"
)

func main() {
	currentUser, err := user.Current()
	if err != nil {
		panic(err)
	}

	fmt.Printf("Hello, %s\n", currentUser.Username)
	fmt.Println("Welcome to the virtual file system implementation in Go\n")

	repl.Start()
}

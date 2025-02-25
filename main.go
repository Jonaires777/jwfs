package main

import (
	"fmt"
	"os/user"

	"github.com/Jonaires777/src/constants"
	"github.com/Jonaires777/src/filemanager"
	"github.com/Jonaires777/src/repl"
)

func main() {
	if !filemanager.CheckFileExistence(constants.VirtualDisk) {
		err := filemanager.CreateVirtualDisk(constants.VirtualDisk)
		if err != nil {
			panic(err)
		} 
	}

	currentUser, err := user.Current()
	if err != nil {
		panic(err)
	}

	fmt.Printf("Hello, %s\n", currentUser.Username)
	fmt.Println("Welcome to the virtual file system implementation in Go\n")

	repl.Start()
}

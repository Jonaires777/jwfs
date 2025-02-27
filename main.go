package main

import (
	"fmt"
	"os"
	"os/user"

	"github.com/Jonaires777/src/constants"
	"github.com/Jonaires777/src/filemanager"
	"github.com/Jonaires777/src/repl"
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "debug" {
		fmt.Println("Rodando modo debug...")
		if err := filemanager.PrintSuperblock(); err != nil {
			fmt.Println("Erro ao imprimir superbloco:", err)
		}

		if err := filemanager.PrintBitmap(); err != nil {
			fmt.Println("Erro ao imprimir bitmap:", err)
		}

		if err := filemanager.PrintInodeTable(); err != nil {
			fmt.Println("Erro ao imprimir tabela de inodes:", err)
		}
		return
	}

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

	fmt.Println("Use 'help' to see the available commands\n")

	repl.Start()
}

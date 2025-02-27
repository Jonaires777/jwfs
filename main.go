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

	fmt.Println("Use the following commands to interact with the file system:")
	fmt.Println("create <filename> <size> - create a new file with the given size")
	fmt.Println("remove <filename> - remove a file")
	fmt.Println("list - list all files")
	fmt.Println("order <filename> - order a file")
	fmt.Println("read <filename> - read a file")
	fmt.Println("concat <filename1> <filename2> <newFile> - concatenate two files into a new file")
	fmt.Println("exit - exit the program\n")

	repl.Start()
}

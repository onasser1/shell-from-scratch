package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// Ensures gofmt doesn't remove the "fmt" import in stage 1 (feel free to remove this!)
var _ = fmt.Print

func mainLoop() {
	fmt.Print("$ ")
	cmdInput, err := bufio.NewReader(os.Stdin).ReadString('\n')
	if err != nil {
		fmt.Println("error reading command input")
	}
	cmd := strings.TrimSuffix(cmdInput, "\n")
	fmt.Printf("%s: command not found\n", cmd)
}

func main() {
	for {
		mainLoop()
	}
}

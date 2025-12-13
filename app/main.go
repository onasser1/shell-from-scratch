package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// Ensures gofmt doesn't remove the "fmt" import in stage 1 (feel free to remove this!)
var _ = fmt.Print

func echoFunc(cmdList []string) {
	if len(cmdList) == 1 {
		fmt.Println()
		return
	}
	for i := 1; i < len(cmdList)-1; i++ {
		fmt.Printf("%s ", cmdList[i])
	}
	fmt.Print(cmdList[len(cmdList)-1])
}

func typeFunc(cmdList []string) {
	if len(cmdList) != 2 {
		fmt.Println("Please provide at least one command after type.")
		return
	}
	trimmedCommand := strings.TrimSuffix(cmdList[1], "\n")
	switch trimmedCommand {
	case "echo", "exit", "type":
		fmt.Printf("%s is a shell builtin\n", strings.TrimSuffix(trimmedCommand, "\n"))
	default:
		fmt.Printf("%s: not found\n", trimmedCommand)
	}
}

func mainLoop() {
	fmt.Print("$ ")
	cmdInput, err := bufio.NewReader(os.Stdin).ReadString('\n')

	if err != nil {
		fmt.Println("error reading command input")
	}
	cmdList := strings.Split(cmdInput, " ")
	if len(cmdList) == 0 {
		return
	}

	trimmedCommand := strings.TrimSuffix(cmdList[0], "\n")
	switch trimmedCommand {
	case "exit":
		os.Exit(127)
	case "echo":
		echoFunc(cmdList)
	case "type":
		typeFunc(cmdList)
	default:
		fmt.Printf("%s: command not found\n", trimmedCommand)
	}

}

func main() {
	for {
		mainLoop()
	}
}

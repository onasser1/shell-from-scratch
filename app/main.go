package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

const (
	// PATH              = "/usr/local/bin:/usr/bin:/usr/sbin/:/bin"
	PathListSeparator = ":"
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
		LookForDirectoriesTypeFunc(trimmedCommand)
	}
}

func LookForDirectoriesTypeFunc(tCmd string) {
	PATH := os.Getenv("PATH")
	directories := strings.Split(PATH, PathListSeparator)
	ReadDirsTypeFunc(directories, tCmd)
}

func ReadDirsTypeFunc(directories []string, commandName string) {
	var found bool
	for _, dir := range directories {
		if strings.Contains(dir, "/var/run") || strings.Contains(dir, "/Users/omar") {
			continue
		}
		err := os.MkdirAll(dir, os.ModePerm)
		if err != nil {
			fmt.Println("error creating non-existent directories", dir)
		}
		entries, err := os.ReadDir(dir)
		if err != nil {
			fmt.Println("error reading directory")
		}
		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}
			if entry.Name() == commandName {
				found = true
				if isExecutable(entry) {
					fmt.Printf("%s is %s/%s\n", entry.Name(), dir, commandName)
					return
				}
			}
		}
	}
	if !found {
		fmt.Printf("%s: not found\n", commandName)
	}
}

func ExecFunc(cmdList []string) {
	tCmd := strings.TrimSuffix(cmdList[0], "\n")
	LookForDirectoriesExecProgram(tCmd, cmdList)
}

func LookForDirectoriesExecProgram(tCmd string, args []string) {
	PATH := os.Getenv("PATH")
	directories := strings.Split(PATH, PathListSeparator)
	ReadDirsExecProgram(directories, tCmd, args)
}

func ReadDirsExecProgram(directories []string, commandName string, args []string) {
	var execPerm os.FileMode = 0755
	var found bool
	for _, dir := range directories {
		if strings.Contains(dir, "/var/run") || strings.Contains(dir, "/Users/omar") {
			continue
		}
		err := os.MkdirAll(dir, os.ModePerm)
		if err != nil {
			fmt.Println("error creating non-existent directories", dir)
		}
		entries, err := os.ReadDir(dir)
		if err != nil {
			fmt.Println("error reading directory")
		}
		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}
			if entry.Name() == commandName {
				found = true
				if isExecutable(entry) {
					trim := strings.TrimSuffix(strings.Join(args, " "), "\n")
					trimmedSlice := strings.Split(trim, " ")
					cmd := exec.Command(commandName, trimmedSlice[1:]...)
					out, err := cmd.CombinedOutput()
					if err != nil {
						return
					}
					fmt.Printf("%s", out)
					return
				} else {
					fPath := dir + entry.Name()
					MakeExecutable(fPath, execPerm)
				}
			}
		}
	}
	if !found {
		fmt.Printf("%s: not found\n", commandName)
	}
}

func isExecutable(entry os.DirEntry) bool {
	entryInfo, err := entry.Info()
	if err != nil {
		fmt.Println("error retrieving entry information")
	}
	return strings.Contains(entryInfo.Mode().String(), "x")
}

func MakeExecutable(filePath string, mode os.FileMode) {
	if err := os.Chmod(filePath, mode); err != nil {
		return
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
		ExecFunc(cmdList)
	}

}

func main() {
	for {
		mainLoop()
	}
}

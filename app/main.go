package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const (
	// PATH              = "/usr/local/bin:/usr/bin:/usr/sbin/:/bin"
	PathListSeparator = ":"
)

// Ensures gofmt doesn't remove the "fmt" import
var _ = fmt.Print

func pwdFunc() error {
	currentPath, err := filepath.Abs("")
	if err != nil {
		return fmt.Errorf("error retrieving path: %s", err)
	}
	fmt.Println(currentPath)
	return nil
}

func echoFunc(cmdList []string) error {
	if len(cmdList) == 1 {
		fmt.Println()
		return nil
	}
	for i := 1; i < len(cmdList)-1; i++ {
		fmt.Printf("%s ", cmdList[i])
	}
	fmt.Print(cmdList[len(cmdList)-1])
	return nil
}

func typeFunc(cmdList []string) {
	if len(cmdList) != 2 {
		fmt.Println("Please provide at least one command after type.")
		return
	}
	trimmedCommand := strings.TrimSuffix(cmdList[1], "\n")
	switch trimmedCommand {
	case "echo", "exit", "type", "pwd":
		fmt.Printf("%s is a shell builtin\n", strings.TrimSuffix(trimmedCommand, "\n"))
	default:
		LookForDirectoriesTypeFunc(trimmedCommand)
	}
}

func LookForDirectoriesTypeFunc(tCmd string) error {
	PATH := os.Getenv("PATH")
	directories := strings.Split(PATH, PathListSeparator)
	if err := ReadDirsTypeFunc(directories, tCmd); err != nil {
		return fmt.Errorf("%s", err)
	}
	return nil
}

func ReadDirsTypeFunc(directories []string, commandName string) error {
	var found bool
	for _, dir := range directories {
		if strings.Contains(dir, "/var/run") || strings.Contains(dir, "/Users/omar") {
			continue
		}
		err := os.MkdirAll(dir, os.ModePerm)
		if err != nil {
			return fmt.Errorf("error creating non-existent directories: %s", err)
		}
		entries, err := os.ReadDir(dir)
		if err != nil {
			return fmt.Errorf("error reading directory: %s", err)
		}
		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}
			if entry.Name() == commandName {
				found = true
				if isExecutable(entry) {
					fmt.Printf("%s is %s/%s\n", entry.Name(), dir, commandName)
					return nil
				}
			}
		}
	}
	// **TODO**: The following if block is redundant and can be removed and return nil immediately with the print statement.
	if !found {
		fmt.Printf("%s: not found\n", commandName)
	}
	return nil
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

func ReadDirsExecProgram(directories []string, commandName string, args []string) error {
	var execPerm os.FileMode = 0755
	var found bool
	for _, dir := range directories {
		if strings.Contains(dir, "/var/run") || strings.Contains(dir, "/Users/omar") {
			continue
		}
		err := os.MkdirAll(dir, os.ModePerm)
		if err != nil {
			return fmt.Errorf("error creating non-existent directories: %s", err)
		}
		entries, err := os.ReadDir(dir)
		if err != nil {
			return fmt.Errorf("error reading directory: %s", err)
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
						return fmt.Errorf("%s", err)
					}
					fmt.Printf("%s", out)
					// **TODO**: Maybe the next return nil statement is redundant too. Removal to be considered.
					return nil
				} else {
					fPath := dir + entry.Name()
					if err := MakeExecutable(fPath, execPerm); err != nil {
						return fmt.Errorf("%s", err)
					}
				}
			}
		}
	}
	if !found {
		fmt.Printf("%s: not found\n", commandName)
	}
	// **TODO**: Maybe the next return nil statement is redundant too. Removal to be considered.
	return nil
}

func isExecutable(entry os.DirEntry) bool {
	entryInfo, err := entry.Info()
	if err != nil {
		fmt.Println("error retrieving entry information")
	}
	return strings.Contains(entryInfo.Mode().String(), "x")
}

func MakeExecutable(filePath string, mode os.FileMode) error {
	if err := os.Chmod(filePath, mode); err != nil {
		return err
	}
	return nil
}

func mainLoop() error {
	fmt.Print("$ ")
	cmdInput, err := bufio.NewReader(os.Stdin).ReadString('\n')

	if err != nil {
		return fmt.Errorf("error reading command input: %s", err)
	}
	cmdList := strings.Split(cmdInput, " ")
	if len(cmdList) == 0 {
		return errors.New("invalid input")
	}

	trimmedCommand := strings.TrimSuffix(cmdList[0], "\n")
	switch trimmedCommand {
	case "exit":
		os.Exit(127)
	case "echo":
		echoFunc(cmdList)
	case "type":
		typeFunc(cmdList)
	case "pwd":
		pwdFunc()
	default:
		ExecFunc(cmdList)
	}
	return nil
}

func main() {
	for {
		err := mainLoop()
		if err != nil {
			break
		}
	}
}

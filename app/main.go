package main

import (
	"bufio"
	"bytes"
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

func redirect(args []string) error {
	if len(args) < 3 {
		return fmt.Errorf("Invalid args\n")
	}
	outFile := strings.TrimSuffix(args[len(args)-1], "\n")
	matcher := args[len(args)-2]

	buf := ExecFunc(args[:len(args)-2], true)

	if matcher != "" && buf != nil {
		err := os.WriteFile(outFile, []byte(buf.Bytes()), 0644)
		if err != nil {
			fmt.Printf("Error %s", err)
		}
	}
	return nil
}

// cdFunc() should at least takes Path as an argument.
func cdFunc(cmdList []string) error {
	var err error
	if len(cmdList) == 1 {
		absPath, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("err: %s", err)
		}
		return changeDirectoryFunc(absPath)
	}

	pathArgs := cmdList[1:]
	normalizedPathArg := strings.TrimSpace(strings.Join(pathArgs, ""))

	if normalizedPathArg == "" || normalizedPathArg == "~" {
		normalizedPathArg, _ = os.UserHomeDir()
	}
	err = changeDirectoryFunc(normalizedPathArg)

	if err != nil {
		fmt.Print(err)
	}
	return nil
}

func changeDirectoryFunc(normalizedPathArg string) error {
	_, err := os.ReadDir(normalizedPathArg)
	if err != nil {
		return fmt.Errorf("cd: %s: No such file or directory\n", normalizedPathArg)
	}
	err = os.Chdir(normalizedPathArg)
	return nil
}

func echoFunc(cmdList []string) error {
	if len(cmdList) == 1 {
		fmt.Println()
		return nil
	}
	args := strings.TrimSuffix(strings.Join(cmdList[1:], " "), "\n")
	fmt.Println(args)
	return nil
}

func typeFunc(cmdList []string) {
	if len(cmdList) != 2 {
		fmt.Println("Please provide at least one command after type.")
		return
	}
	trimmedCommand := strings.TrimSuffix(cmdList[1], "\n")
	switch trimmedCommand {
	case "echo", "exit", "type", "pwd", "cd":
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

func ExecFunc(cmdList []string, redirectFlag bool) *bytes.Buffer {
	tCmd := strings.TrimSuffix(cmdList[0], "\n")
	return LookForDirectoriesExecProgram(tCmd, cmdList[1:], redirectFlag)
}

func LookForDirectoriesExecProgram(tCmd string, args []string, redirectFlag bool) *bytes.Buffer {
	PATH := os.Getenv("PATH")
	directories := strings.Split(PATH, PathListSeparator)
	buf, err := ReadDirsExecProgram(directories, tCmd, args, redirectFlag)
	if err != nil {
		return &bytes.Buffer{}
	}
	return buf
}

func ReadDirsExecProgram(directories []string, commandName string, args []string, redirectFlag bool) (*bytes.Buffer, error) {
	var execPerm os.FileMode = 0755
	var buf bytes.Buffer
	var found bool
	for _, dir := range directories {
		if strings.Contains(dir, "/var/run") || strings.Contains(dir, "/Users/omar") {
			continue
		}
		err := os.MkdirAll(dir, os.ModePerm)
		if err != nil {
			return &bytes.Buffer{}, fmt.Errorf("error creating non-existent directories: %s", err)
		}
		entries, err := os.ReadDir(dir)
		if err != nil {
			return &bytes.Buffer{}, fmt.Errorf("error reading directory: %s", err)
		}
		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}
			if entry.Name() == commandName {
				found = true
				if isExecutable(entry) {
					trim := strings.TrimSuffix(strings.Join(args, " "), "\n")
					trimmedArgs := strings.Split(trim, " ")

					cmd := exec.Command(commandName, trimmedArgs...)
					if redirectFlag {
						cmd.Stdin = os.Stdin
						cmd.Stdout = &buf
						cmd.Stderr = os.Stderr
					} else {
						cmd.Stdin = os.Stdin
						cmd.Stdout = os.Stdout
						cmd.Stderr = os.Stderr
					}

					err := cmd.Run()
					if err != nil {
						return &bytes.Buffer{}, fmt.Errorf("err: %s", err)
					}
					// combinedOut, err := cmd.CombinedOutput()
					// if err != nil {
					// 	return fmt.Errorf("%s", err)
					// }
					// fmt.Printf("%s", combinedOut)
					// **TODO**: Maybe the next return nil statement is redundant too. Removal to be considered.
					return &buf, nil
				} else {
					fPath := dir + entry.Name()
					if err := MakeExecutable(fPath, execPerm); err != nil {
						return &bytes.Buffer{}, fmt.Errorf("%s", err)
					}
				}
			}
		}
	}
	if !found {
		fmt.Printf("%s: not found\n", commandName)
	}
	// **TODO**: Maybe the next return nil statement is redundant too. Removal to be considered.
	return &buf, nil
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
	switch {
	case trimmedCommand == "exit":
		os.Exit(127)
	case strings.Contains(cmdInput, ">"):
		err = redirect(cmdList)
	case trimmedCommand == "echo":
		err = echoFunc(cmdList)
	case trimmedCommand == "type":
		typeFunc(cmdList)
	case trimmedCommand == "pwd":
		err = pwdFunc()
	case trimmedCommand == "cd":
		err = cdFunc(cmdList)
	default:
		ExecFunc(cmdList, false)
	}
	if err != nil {
		fmt.Print(err)
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

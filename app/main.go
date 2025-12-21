package main

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
)

const (
	PathListSeparator = ":"
)

type Command struct {
	stdoutRedirect bool
	stderrRedirect bool
	redirect       bool
	args           []string
	cmdName        string
}

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

func redirect(c *Command) error {
	var (
		matcher string
		idx     int
	)
	if len(c.args) < 3 {
		return fmt.Errorf("Invalid args\n")
	}
	redirectionCharacters := []string{">", "1>", "2>"}
	// This is a very weak assumption. here we always assume that the input is ** ** ** ** > path. but what if we didn't get that formula?
	outFile := strings.TrimSuffix(c.args[len(c.args)-1], "\n")
	for i, _ := range redirectionCharacters {
		idx = slices.Index(c.args, redirectionCharacters[i])
		if idx != -1 {
			matcher = c.args[idx]
			if matcher == "2>" {
				c.stderrRedirect = true
			} else {
				c.stdoutRedirect = true
			}
			c.redirect = true
			break
		}
	}
	// c.args[:len(c.args)-2]
	buf := ExecFunc(c)
	cleanedBuf := strings.ReplaceAll(buf.String(), "'", "")

	if matcher != "" && buf != nil {
		err := os.WriteFile(outFile, []byte(cleanedBuf), 0644)
		if err != nil {
			fmt.Printf("Error %s", err)
		}
	}
	return nil
}

// cdFunc() should at least takes Path as an argument.
func cdFunc(c *Command) error {
	var err error
	if len(c.args) == 1 {
		absPath, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("err: %s", err)
		}
		return changeDirectoryFunc(absPath)
	}

	pathArgs := c.args[1:]
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

func echoFunc(c *Command) error {
	if len(c.args) == 1 {
		fmt.Println()
		return nil
	}
	args := strings.TrimSuffix(strings.Join(c.args[1:], " "), "\n")
	fmt.Println(strings.ReplaceAll(args, "'", ""))
	return nil
}

func typeFunc(c *Command) {
	if len(c.args) != 2 {
		fmt.Println("Please provide at least one command after type.")
		return
	}
	trimmedCommand := strings.TrimSuffix(c.args[1], "\n")
	switch trimmedCommand {
	case "echo", "exit", "type", "pwd", "cd":
		fmt.Printf("%s is a shell builtin\n", strings.TrimSuffix(trimmedCommand, "\n"))
	default:
		if err := LookForDirectoriesTypeFunc(c); err != nil {
			fmt.Printf("%s: command not found\n", trimmedCommand)
		}
	}
}

func LookForDirectoriesTypeFunc(c *Command) error {
	PATH := os.Getenv("PATH")
	directories := strings.Split(PATH, PathListSeparator)
	if err := ReadDirsTypeFunc(directories, c); err != nil {
		return fmt.Errorf("%s", err)
	}
	return nil
}

func ReadDirsTypeFunc(directories []string, c *Command) error {
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
			if entry.Name() == c.cmdName {
				found = true
				if isExecutable(entry) {
					fmt.Printf("%s is %s/%s\n", entry.Name(), dir, c.cmdName)
					return nil
				}
			}
		}
	}
	// **TODO**: The following if block is redundant and can be removed and return nil immediately with the print statement.
	if !found {
		fmt.Printf("%s: not found\n", c.cmdName)
	}
	return nil
}

func ExecFunc(c *Command) *bytes.Buffer {
	return LookForDirectoriesExecProgram(c)
}

func LookForDirectoriesExecProgram(c *Command) *bytes.Buffer {
	PATH := os.Getenv("PATH")
	directories := strings.Split(PATH, PathListSeparator)
	buf, _ := ReadDirsExecProgram(directories, c)
	return buf
}

func ReadDirsExecProgram(directories []string, c *Command) (*bytes.Buffer, error) {
	var (
		execPerm os.FileMode = 0755
		buf      bytes.Buffer
		found    bool
		args     []string
	)
	// Assign commandName to the first argument in the input.
	commandName := strings.TrimSuffix(c.args[0], "\n")
	args = c.args[1:]
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
					if c.redirect {
						// Assign args to the input starting from the second element to the end excluding last two elements.
						// Following the good user input that the last two elements are the redirection character and file path.
						args = c.args[1 : len(c.args)-2]
					}
					trimmedInput := strings.TrimSuffix(strings.Join(args, " "), "\n")
					trimmedArgs := strings.Split(trimmedInput, " ")

					cmd := exec.Command(commandName, trimmedArgs...)
					cmd.Stdin = os.Stdin
					cmd.Stdout = os.Stdout
					cmd.Stderr = os.Stderr

					if c.stdoutRedirect {
						cmd.Stdout = &buf
					}
					if c.stderrRedirect {
						cmd.Stderr = &buf
					}
					err := cmd.Run()
					if err != nil {
						return &buf, fmt.Errorf("err: %s", err)
					}
					// combinedOut, err := cmd.CombinedOutput()
					// if err != nil {
					// 	return fmt.Errorf("%s", err)
					// }
					// fmt.Printf("%s", combinedOut)
					// TODO: Maybe the next return nil statement is redundant too. Removal to be considered.
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
		return false
	}
	return strings.Contains(entryInfo.Mode().String(), "x")
}

func MakeExecutable(filePath string, mode os.FileMode) error {
	if err := os.Chmod(filePath, mode); err != nil {
		return err
	}
	return nil
}

func mainLoop(c *Command) error {
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
	c.cmdName = trimmedCommand
	c.args = cmdList
	c.stdoutRedirect = false
	c.stderrRedirect = false
	c.redirect = false
	switch {
	case trimmedCommand == "exit":
		os.Exit(127)
	case strings.Contains(cmdInput, ">"):
		err = redirect(c)
	case trimmedCommand == "echo":
		err = echoFunc(c)
	case trimmedCommand == "type":
		typeFunc(c)
	case trimmedCommand == "pwd":
		err = pwdFunc()
	case trimmedCommand == "cd":
		err = cdFunc(c)
	default:
		ExecFunc(c)
	}
	if err != nil {
		fmt.Print(err)
	}
	return nil
}

func main() {
	c := &Command{}
	for {
		err := mainLoop(c)
		if err != nil {
			break
		}
	}
}

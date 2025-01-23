package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
)

// executeCommand executes a given command with its arguments.
func executeCommand(command []string, filePath string) error {
	executablePath, args := prepareCommandArgs(command, filePath)

	if executablePath == "" {
		logger.Println("Skipping execution of empty command.")
		return nil
	}

	cmd := exec.Command(executablePath, args...)

	if filePath != "" {
		cmd.Dir = filepath.Dir(filePath)
	}

	return executeCmdAndWait(cmd)
}

// prepareCommandArgs prepares the command arguments, replacing placeholders and resolving executable path.
func prepareCommandArgs(command []string, filePath string) (string, []string) {
	if len(command) == 0 {
		// Return empty strings if the command is empty
		return "", []string{}
	}
	
	executablePath := command[0]

	// Resolve executable path if not absolute
	if !filepath.IsAbs(executablePath) {
		if absPath, err := resolveExecutablePath(executablePath); err == nil {
			executablePath = absPath
		}
	}

	// Replace placeholder with actual file path and event type
	args := make([]string, 0, len(command)-1)
	for i, arg := range command {
		if i == 0 {
			continue // Skip the executable itself
		}
		if arg == "{filepath}" {
			args = append(args, filePath)
		} else {
			args = append(args, arg)
		}
	}

	return executablePath, args
}

// resolveExecutablePath finds the absolute path of an executable.
func resolveExecutablePath(executableName string) (string, error) {
	// Check in the same directory as the filewatcher
	filewatcherDir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		return "", fmt.Errorf("error getting filewatcher directory: %w", err)
	}
	fullExecutablePath := filepath.Join(filewatcherDir, executableName)
	if _, err := os.Stat(fullExecutablePath); err == nil {
		return fullExecutablePath, nil
	}

	// Check in system PATH
	if path, err := exec.LookPath(executableName); err == nil {
		return path, nil
	}

	return "", fmt.Errorf("executable %s not found", executableName)
}

// executeCmdAndWait executes a command and waits for it to complete, capturing stdout and stderr.
func executeCmdAndWait(cmd *exec.Cmd) error {
	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("error creating stdout pipe: %w", err)
	}
	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("error creating stderr pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("error starting command: %w", err)
	}

	var stdoutWg sync.WaitGroup
	stdoutWg.Add(2)
	go logCmdOutput(stdoutPipe, &stdoutWg, false) // isErrorStream = false
	go logCmdOutput(stderrPipe, &stdoutWg, true)  // isErrorStream = true

	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("error waiting for command to complete: %w", err)
	}

	stdoutWg.Wait()
	return nil
}

// logCmdOutput scans and logs the output from a command (stdout or stderr).
func logCmdOutput(pipe io.ReadCloser, wg *sync.WaitGroup, isErrorStream bool) {
	defer wg.Done()
	scanner := bufio.NewScanner(pipe)
	for scanner.Scan() {
		if isErrorStream {
			logger.Printf("Stderr: %s", scanner.Text())
		} else {
			logger.Printf("Stdout: %s", scanner.Text())
		}
	}
}

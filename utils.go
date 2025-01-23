package main

import (
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
)

// isExcludedPath checks if a given path should be excluded based on the config.
func isExcludedPath(path string, config *Config) bool {
	// Convert the path to an absolute path
	absPath, err := filepath.Abs(path)
	if err != nil {
		logger.Printf("Error getting absolute path for %s: %v", path, err)
		return false // Don't exclude if we can't get the absolute path
	}

	for _, excludePattern := range config.ExcludePaths {
		// Remove * and trim spaces from the exclude pattern
		cleanedExcludePattern := strings.TrimSpace(strings.ReplaceAll(excludePattern, "*", ""))

		// Exact Path Match (case-insensitive)
		if strings.EqualFold(absPath, cleanedExcludePattern) {
			return true
		}

		// Substring Match (case-insensitive)
		if strings.Contains(strings.ToLower(absPath), strings.ToLower(cleanedExcludePattern)) {
			return true
		}
	}

	return false
}

// isAllowedFileType checks if the file extension is in the allowed list
func isAllowedFileType(filename string, allowedTypes []string) bool {
	// If allowedTypes is empty, allow all file types
	if len(allowedTypes) == 0 {
		return true
	}

	ext := strings.ToLower(filepath.Ext(filename))
	for _, t := range allowedTypes {
		if ext == strings.ToLower(t) {
			return true
		}
	}
	return false
}

// setupSignalHandling sets up a signal handler for graceful shutdown.
func setupSignalHandling(config *Config) {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

	go func() {
		sig := <-sigCh
		logger.Printf("Received signal: %v. Shutting down...", sig)
		executeShutdownCommand(config)
		os.Exit(0)
	}()
}

// executeStartupCommand executes the initialization command if specified in the config.
func executeStartupCommand(config *Config) {
	if len(config.InitRun) > 0 {
		logger.Println("Executing initialization command...")
		if err := executeCommand(config.InitRun, ""); err != nil {
			logger.Fatalf("Error executing initialization command: %v", err)
		}
	}
}

// executeShutdownCommand executes the termination command if specified in the config.
func executeShutdownCommand(config *Config) {
	if len(config.ExitRun) > 0 {
		logger.Println("Executing termination command...")
		if err := executeCommand(config.ExitRun, ""); err != nil {
			logger.Printf("Error executing termination command: %v", err)
		}
	}
}

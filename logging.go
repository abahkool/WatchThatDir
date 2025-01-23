package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
)

// initLogging initializes the logger based on configuration.
func initLogging(config *Config) {
	if config.EnableLog {
		if err := initFileLogger(config.LogPath); err != nil {
			log.Fatalf("Error initializing logger: %v", err)
		}
	} else {
		logger = log.New(os.Stdout, "", log.LstdFlags) // Default logger writes to standard output
	}
}

// initFileLogger initializes the logger to write to the specified file path.
func initFileLogger(logPath string) error {
	logDir := filepath.Dir(logPath)
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return fmt.Errorf("error creating log directory: %w", err)
	}

	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("error opening log file: %w", err)
	}

	logger = log.New(logFile, "", log.LstdFlags)
	return nil
}

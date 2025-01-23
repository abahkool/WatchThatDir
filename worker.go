package main

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
)

// setupWorkerPool creates and starts the worker pool.
func setupWorkerPool(config *Config) (chan string, *sync.WaitGroup) {
	if config.MaxWorkers <= 0 {
		config.MaxWorkers = runtime.NumCPU()
	}

	taskQueue := make(chan string, 100)
	var workerWg sync.WaitGroup

	for i := 0; i < config.MaxWorkers; i++ {
		workerWg.Add(1)
		go worker(taskQueue, &workerWg, config, i+1)
	}

	return taskQueue, &workerWg
}

// worker function to process files from the task queue.
func worker(taskQueue chan string, wg *sync.WaitGroup, config *Config, workerID int) {
	defer wg.Done()
	logger.Printf("Worker %d starting", workerID)

	for filePathWithEvent := range taskQueue {
		logger.Printf("Worker %d: Processing file: %s", workerID, filePathWithEvent)

		// Parse the event type from the suffixed filePath
		filePath, eventType := parseEventType(filePathWithEvent)

		if err := processFile(filePath, config, eventType); err != nil {
			logger.Printf("Worker %d: Error processing file %s: %v", workerID, filePath, err)
		} else {
			logger.Printf("Worker %d: Successfully processed file: %s", workerID, filePath)
		}
	}

	logger.Printf("Worker %d exiting", workerID)
}

// parseEventType extracts the file path and event type from the suffixed string.
func parseEventType(filePathWithEvent string) (string, EventType) {
	parts := strings.Split(filePathWithEvent, "?event=")
	if len(parts) != 2 {
		return filePathWithEvent, "" // Return the original string and empty event type if no suffix
	}
	return parts[0], EventType(parts[1])
}

// processFile handles execution of commands and post-processing for a single file.
func processFile(filePath string, config *Config, eventType EventType) error {
	var cmd []string

	// Select the command based on the event type
	switch eventType {
	case CreateEvent:
		cmd = config.OnCreateRun
	case RenameEvent:
		cmd = config.OnRenameRun
	case WriteEvent:
		cmd = config.OnModifyRun
	case RemoveEvent:
		cmd = config.OnRemoveRun
	default:
		return fmt.Errorf("unknown event type: %s", eventType)
	}

	// Execute the command
	if err := executeCommand(cmd, filePath); err != nil {
		return fmt.Errorf("error executing command for file %s: %w", filePath, err)
	}

	// Handle post-processing only if event type is not Remove
	if eventType != RemoveEvent {
		return handlePostProcessing(filePath, config)
	}

	// Handle post-processing
	return nil
}

// handlePostProcessing performs actions on the file after the command has been executed.
func handlePostProcessing(filePath string, config *Config) error {
	switch config.PostProcessAction {
	case PostProcessActionDoNothing:
		logger.Println("File processed (no action taken):", filePath)
	case PostProcessActionMove:
		if err := moveFileToCompletionDir(filePath, config); err != nil {
			return err
		}
	case PostProcessActionDelete:
		if err := deleteFile(filePath); err != nil {
			return err
		}
	default:
		return fmt.Errorf("invalid post process action setting in config.yaml: %d", config.PostProcessAction)
	}
	return nil
}

// moveFileToCompletionDir moves the processed file to the completion directory.
func moveFileToCompletionDir(filePath string, config *Config) error {
	destPath := filepath.Join(config.ProcessedPath, filepath.Base(filePath))

	// Get the absolute path of the destination
	absDestPath, err := filepath.Abs(destPath)
	if err != nil {
		return fmt.Errorf("error getting absolute path for destination %s: %w", destPath, err)
	}

	if err := os.MkdirAll(config.ProcessedPath, 0755); err != nil {
		return fmt.Errorf("error creating completion directory %s: %w", config.ProcessedPath, err)
	}

	if err := os.Rename(filePath, absDestPath); err != nil {
		return fmt.Errorf("error moving file: %w", err)
	}

	logger.Println("Moved file to:", absDestPath) // Log the absolute destination path
	return nil
}

// deleteFile deletes the processed file.
func deleteFile(filePath string) error {
	if err := os.Remove(filePath); err != nil {
		return fmt.Errorf("error deleting file: %w", err)
	}
	logger.Println("Deleted file:", filePath)
	return nil
}

// processExistingFiles scans the watch directory and processes files that match the allowed types.
func processExistingFiles(config *Config, taskQueue chan string) {
	err := filepath.Walk(config.TargetPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Check if the path should be excluded
		if isExcludedPath(path, config) {
			if info.IsDir() {
				logger.Printf("Skipping excluded directory: %s", path)
				return filepath.SkipDir // Skip the entire directory
			} else {
				logger.Printf("Skipping excluded file: %s", path)
				return nil
			}
		}

		if !info.IsDir() && isAllowedFileType(path, config.FileTypes) {
			// Get the absolute path
			absPath, err := filepath.Abs(path)
			if err != nil {
				return fmt.Errorf("error getting absolute path for %s: %w", path, err)
			}

			logger.Println("Processing existing file:", absPath)

			// Simulate a Create event
			if shouldProcessEvent(absPath, config) {
				taskQueue <- absPath + "?event=create"
			}
		}
		return nil
	})

	if err != nil {
		logger.Printf("Error walking the path %s: %v", config.TargetPath, err)
	}
}

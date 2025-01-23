package main

import (
	"os"
	"sync"
	"time"

	"github.com/rjeczalik/notify"
)

var (
	lastEventTimes      = make(map[string]time.Time)
	lastEventTimesMutex sync.Mutex
)

// initializeWatcher sets up the directory watcher.
// initializeWatcher sets up the directory watcher.
func initializeWatcher(config *Config) {
	// Moved outside -> watcherChannel := make(chan notify.EventInfo, 100)
	if err := notify.Watch(config.TargetPath+"/...", watcherChannel, notify.Create, notify.Write, notify.Remove, notify.Rename); err != nil {
		logger.Fatalf("Error setting up watch: %v", err)
	}
	// Moved outside -> return watcherChannel
}

// handleEvents is the main loop for processing file system events.
func handleEvents(watcherChannel chan notify.EventInfo, taskQueue chan string, config *Config) {
	for event := range watcherChannel {
		eventPath := event.Path()
		switch event.Event() {
		case notify.Create:
			handleCreateEvent(eventPath, taskQueue, config, watcherChannel)
		case notify.Rename:
			handleRenameEvent(eventPath, taskQueue, config, watcherChannel)
		case notify.Write:
			handleWriteEvent(eventPath, taskQueue, config)
		case notify.Remove:
			handleRemoveEvent(eventPath, taskQueue, config)
		}
	}
}

// handleCreateEvent handles file/directory creation events.
func handleCreateEvent(eventPath string, taskQueue chan string, config *Config, watcherChannel chan notify.EventInfo) {
	if isExcludedPath(eventPath, config) {
		logger.Printf("Skipping excluded path: %s", eventPath)
		return
	}

	fi, err := os.Stat(eventPath)
	if err != nil {
		logger.Printf("Error stating file %s: %v", eventPath, err)
		return
	}

	if fi.IsDir() {
		logger.Println("Detected new directory:", eventPath)
		watchNewDirectory(eventPath, watcherChannel)
	} else if fi.Mode().IsRegular() && isAllowedFileType(eventPath, config.FileTypes) {
		logger.Println("New file created:", eventPath)
		// Execute command specific to Create event
		if len(config.OnCreateRun) > 0 {
			if shouldProcessEvent(eventPath, config) {
				taskQueue <- eventPath + "?event=create"
			}
		}
	}
}

// handleRenameEvent handles file/directory renaming events.
func handleRenameEvent(eventPath string, taskQueue chan string, config *Config, watcherChannel chan notify.EventInfo) {
	if isExcludedPath(eventPath, config) {
		logger.Printf("Skipping excluded path: %s", eventPath)
		return
	}

	fi, err := os.Stat(eventPath)
	if err != nil {
		logger.Printf("Error stating file %s: %v", eventPath, err)
		return
	}

	if fi.IsDir() {
		logger.Println("Detected renamed directory:", eventPath)
		watchNewDirectory(eventPath, watcherChannel)
	} else if fi.Mode().IsRegular() && isAllowedFileType(eventPath, config.FileTypes) {
		logger.Println("File renamed:", eventPath)
		// Execute command specific to Rename event
		if len(config.OnRenameRun) > 0 {
			if shouldProcessEvent(eventPath, config) {
				taskQueue <- eventPath + "?event=rename"
			}
		}
	}
}

// handleWriteEvent handles file write events.
func handleWriteEvent(eventPath string, taskQueue chan string, config *Config) {
	if isAllowedFileType(eventPath, config.FileTypes) {
		logger.Println("File modified:", eventPath)
		// Execute command specific to Write event
		if len(config.OnModifyRun) > 0 {
			if shouldProcessEvent(eventPath, config) {
				taskQueue <- eventPath + "?event=write"
			}
		}
	}
}

// handleRemoveEvent handles file removal events.
func handleRemoveEvent(eventPath string, taskQueue chan string, config *Config) {
	logger.Printf("File or directory removed: %s", eventPath)

	// Execute command specific to Remove event
	if len(config.OnRemoveRun) > 0 {
		// Add the event path to the task queue with the "remove" event marker
		taskQueue <- eventPath + "?event=remove"
	}
}

// watchNewDirectory starts watching a new directory recursively.
func watchNewDirectory(dirPath string, watcherChannel chan notify.EventInfo) {
	if err := notify.Watch(dirPath+"/...", watcherChannel, notify.Create, notify.Write, notify.Remove, notify.Rename); err != nil {
		logger.Printf("Error watching new directory: %v", err)
	} else {
		logger.Println("Now watching new directory:", dirPath)
	}
}

// shouldProcessEvent checks if an event should be processed based on the debounce configuration.
func shouldProcessEvent(eventPath string, config *Config) bool {
	lastEventTimesMutex.Lock()
	defer lastEventTimesMutex.Unlock()

	now := time.Now()
	lastEventTime, exists := lastEventTimes[eventPath]

	if !exists || now.Sub(lastEventTime) >= time.Duration(config.Debounce)*time.Millisecond {
		lastEventTimes[eventPath] = now
		return true
	}

	logger.Printf("Debouncing event for %s", eventPath)
	return false
}

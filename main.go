package main

import (
	"log"
	"os"
	"reflect"
	"sync"
	"time"

	"github.com/rjeczalik/notify"
)

// --- Global Variables ---
var logger *log.Logger
var watcherChannel chan notify.EventInfo
var watcherMutex sync.Mutex
var taskQueue chan string         // Now a global variable
var workerWg *sync.WaitGroup // Also made global

func main() {
	// 1. Load Configuration
	config, err := loadConfig("config.yaml")
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	// 2. Initialize Logger
	initLogging(config)

	// 3. Create TargetPath if it doesn't exist
	if err := os.MkdirAll(config.TargetPath, 0755); err != nil {
		logger.Fatalf("Error creating target directory %s: %v", config.TargetPath, err)
	}

	// 4. Execute Initialization Command
	executeStartupCommand(config)

	// 5. Set Up Signal Handling
	setupSignalHandling(config)

	// 6. Watcher Initialization
	watcherChannel = make(chan notify.EventInfo, 100)
	initializeWatcher(config)
	defer func() {
		watcherMutex.Lock()
		defer watcherMutex.Unlock()
		if watcherChannel != nil {
			notify.Stop(watcherChannel)
		}
	}()

	// 7. Worker Pool Setup
	taskQueue, workerWg = setupWorkerPool(config) // Initialized here

	// 8. Process Existing Files (if enabled)
	if config.ProcessOnStart {
		processExistingFiles(config, taskQueue)
	}

	// 9. Event Handling
	logger.Println("Watching for file changes in:", config.TargetPath)
	go handleEvents(watcherChannel, taskQueue, config)

	// 10. Config Reloading
	if config.ReloadConfig > 0 {
		go periodicConfigReload(config)
	}

	// 11. Start Watcher Recovery Routine
	go periodicWatcherRecovery(config)

	// 12. Keep the Main Goroutine Alive
	<-make(chan struct{})

	// 13. Clean Up
	close(taskQueue)
	workerWg.Wait()
}

func periodicConfigReload(config *Config) {
	logger.Println("Configuration is set to reload every ", config.ReloadConfig, " miliseconds")
	ticker := time.NewTicker(time.Duration(config.ReloadConfig) * time.Millisecond)
	defer ticker.Stop()

	// Store the initial configuration values
	oldConfigValues := make(map[string]interface{})
	configVal := reflect.ValueOf(config).Elem()
	configType := configVal.Type()
	for i := 0; i < configVal.NumField(); i++ {
		field := configType.Field(i)
		value := configVal.Field(i).Interface()
		oldConfigValues[field.Name] = value
	}

	for range ticker.C {
		newConfig, err := loadConfig("config.yaml")
		if err != nil {
			logger.Fatalf("Error reloading config: %v", err)
			continue
		}

		// Compare old and new values and log changes
		newConfigVal := reflect.ValueOf(newConfig).Elem()
		for i := 0; i < newConfigVal.NumField(); i++ {
			fieldName := configType.Field(i).Name
			newValue := newConfigVal.Field(i).Interface()
			oldValue := oldConfigValues[fieldName]

			if !reflect.DeepEqual(oldValue, newValue) {
				logger.Printf("Config change detected - %s: %v -> %v", fieldName, oldValue, newValue)
				oldConfigValues[fieldName] = newValue // Update old value
			}
		}

		// Update the global config variable
		*config = *newConfig

		// Update ticker if ReloadConfig changed
		if config.ReloadConfig != newConfig.ReloadConfig {
			ticker.Stop()
			ticker = time.NewTicker(time.Duration(newConfig.ReloadConfig) * time.Millisecond)
		}
	}
}

// isTargetAccessible checks if the target path is accessible.
func isTargetAccessible(config *Config) bool {
	_, err := os.Stat(config.TargetPath)
	return err == nil
}

// reinitializeWatcher reinitializes the file system watcher.
func reinitializeWatcher(config *Config) {
	watcherMutex.Lock()
	defer watcherMutex.Unlock()

	if watcherChannel != nil {
		notify.Stop(watcherChannel)
	}

	watcherChannel = make(chan notify.EventInfo, 100)
	initializeWatcher(config)
	go handleEvents(watcherChannel, taskQueue, config) // Now taskQueue is accessible
	logger.Println("Watcher reinitialized successfully.")
}

// periodicWatcherRecovery periodically checks the accessibility of the target path and reinitializes the watcher if necessary.
func periodicWatcherRecovery(config *Config) {
	logger.Println("Watcher recovery routine started. Checking accessibility every", config.CheckInterval, "seconds")
	ticker := time.NewTicker(time.Duration(config.CheckInterval) * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		if !isTargetAccessible(config) {
			logger.Printf("Target path %s is inaccessible. Stopping watcher.", config.TargetPath)

			watcherMutex.Lock()
			if watcherChannel != nil {
				notify.Stop(watcherChannel)
				watcherChannel = nil
			}
			watcherMutex.Unlock()

			// Keep checking for accessibility until it's restored
			for !isTargetAccessible(config) {
				time.Sleep(time.Duration(config.CheckInterval) * time.Second)
			}

			logger.Printf("Target path %s is accessible again. Reinitializing watcher.", config.TargetPath)
			reinitializeWatcher(config)
		}
	}
}

/*
func periodicConfigReload_old(config *Config) {
	logger.Println("Configuration is set to reload every ", config.ReloadConfig, " miliseconds")
	ticker := time.NewTicker(time.Duration(config.ReloadConfig) * time.Millisecond)
	defer ticker.Stop()

	for range ticker.C {
		newConfig, err := loadConfig("config.yaml")
		if err != nil {
			logger.Printf("Error reloading config: %v", err)
			continue
		}

		// Update the global config variable
		*config = *newConfig
	}
}
*/
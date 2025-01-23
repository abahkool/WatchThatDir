package main

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Config defines the structure for application configuration.
type Config struct {
	TargetPath        string   `yaml:"target_path"`
	ProcessedPath     string   `yaml:"processed_path"`
	MaxWorkers        int      `yaml:"max_workers"`
	PostProcessAction int      `yaml:"post_process"`
	FileTypes         []string `yaml:"file_type"`
	ProcessOnStart    bool     `yaml:"process_on_start"`
	LogPath           string   `yaml:"logfile_path"`
	EnableLog         bool     `yaml:"enable_logging"`
	InitRun           []string `yaml:"init_run"`
	ExitRun           []string `yaml:"exit_run"`
	OnCreateRun       []string `yaml:"oncreate_run"`
	OnModifyRun       []string `yaml:"onmodify_run"`
	OnRenameRun       []string `yaml:"onrename_run"`
	OnRemoveRun       []string `yaml:"onremove_run"`
	Debounce          int      `yaml:"debounce"`
	ExcludePaths      []string `yaml:"exclude_path"`
	ReloadConfig      int      `yaml:"reload_config"`
	CheckInterval     int      `yaml:"check_interval"`
}

// EventType defines the type for different file system events.
type EventType string

// Constants for different event types.
const (
	CreateEvent EventType = "create"
	RenameEvent EventType = "rename"
	WriteEvent  EventType = "write"
	RemoveEvent EventType = "remove"
)

// Constants for on_completion actions.
const (
	PostProcessActionDoNothing = 0
	PostProcessActionMove      = 1
	PostProcessActionDelete    = -1
)

// loadConfig loads the configuration from the specified YAML file, sets default values,
// and creates a default config file if it doesn't exist.
func loadConfig(filename string) (*Config, error) {
	// Default configuration values
	config := Config{
		TargetPath:        "targetpath",
		ProcessedPath:     "completed",
		MaxWorkers:        1,
		PostProcessAction: PostProcessActionDoNothing,
		FileTypes:         nil, // Empty to allow all file types
		ProcessOnStart:    true,
		LogPath:           "FileEventsHandler.log",
		EnableLog:         false,
		InitRun:           nil,
		ExitRun:           nil,
		OnCreateRun:       nil,
		OnModifyRun:       nil,
		OnRenameRun:       nil,
		OnRemoveRun:       nil,
		Debounce:          100,
		ExcludePaths:      nil,
		ReloadConfig:      0,
		CheckInterval:     5,
	}

	data, err := os.ReadFile(filename)
	if err != nil {
		if os.IsNotExist(err) {
			// Config file doesn't exist, create it with default values
			data, err = yaml.Marshal(&config)
			if err != nil {
				return nil, fmt.Errorf("error marshalling default config: %w", err)
			}
			err = os.WriteFile(filename, data, 0644)
			if err != nil {
				return nil, fmt.Errorf("error creating default config file: %w", err)
			}
			fmt.Println("Config file not found. Created a new one with default values.")
			return &config, nil
		}
		// Error reading config file (other than not existing)
		return nil, fmt.Errorf("error reading config file: %w", err)
	}

	// Unmarshal config file data into the config struct
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling config: %w", err)
	}

	// Validate post_process value
	if config.PostProcessAction != PostProcessActionDoNothing &&
		config.PostProcessAction != PostProcessActionMove &&
		config.PostProcessAction != PostProcessActionDelete {
		return nil, fmt.Errorf("invalid post_process value: %d", config.PostProcessAction)
	}

	return &config, nil
}

// loadConfig loads the configuration from the specified YAML file.
func loadConfig_old(filename string) (*Config, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("error reading config file: %w", err)
	}

	var config Config
	//Set default value
	config.CheckInterval = 5
	config.ReloadConfig = 0
	config.ExcludePaths = nil
	config.Debounce = 10
	config.OnRemoveRun = nil
	config.OnRenameRun = nil
	config.OnModifyRun = nil
	config.OnCreateRun = nil
	config.InitRun = nil
	config.ExitRun = nil
	config.EnableLog = false
	config.LogPath = "WatchThatDir"
	config.ProcessOnStart = true
	config.FileTypes = nil //empty to allow all file types
	config.PostProcessAction = 0 //do nothing after process the file
	config.MaxWorkers = 1
	config.ProcessedPath = "completed"
	config.TargetPath = "targetpath"

	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling config: %w", err)
	}

	// Validate on_completion value
	if config.PostProcessAction != PostProcessActionDoNothing && config.PostProcessAction != PostProcessActionMove && config.PostProcessAction != PostProcessActionDelete {
		return nil, fmt.Errorf("invalid post_process value: %d", config.PostProcessAction)
	}

	return &config, nil
}

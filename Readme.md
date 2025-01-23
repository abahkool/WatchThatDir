# WatchThatDir

WatchThatDir is a simple command-line utility built in Go that monitors a directory you specify and automatically reacts to file events. Think of it as a little helper that keeps an eye on your files and does something when they changed.

## 1. Introduction: Automate the Boring Stuff with Your Files

Ever wished your computer could automatically do things when you add, modify, or delete files in a specific folder?  WatchThatDir makes that wish a reality. It's a lightweight application that watches a directory of your choice and triggers actions based on what happens to the files within.  

For example, you could set it up to:

*   **Automatically convert images** as soon as they're added to a folder.
*   **Process invoices** and update your accounting system instantly when a new PDF arrives.
*   **Trigger backups** whenever a critical document is modified.
*   **Execute scripts** for any file change scenario you can imagine!

It's all about automating those little file-related tasks that can eat up your time.

## 2. Target Audience: Who Can Benefit from This?

WatchThatDir is designed for anyone who wants a bit more automation in their workflow. This includes:

*   **Developers:** Automate build processes, code generation, or deployment tasks triggered by file changes.
*   **System Administrators:** Monitor critical directories for unauthorized modifications, trigger alerts, or automate system maintenance tasks.
*   **Content Creators:** Streamline workflows by automatically processing images, videos, or audio files as they are created.
*   **Data Analysts:** Process incoming data files in real-time, trigger data analysis pipelines, and generate reports automatically.
*   **Power Users:** Anyone who wants to automate repetitive file-related tasks and regain control of their digital environment.
*   **Anyone Who Deals with Lots of Files:** If you find yourself regularly doing the same things with files, this tool can help streamline your process.

## 3. Advantages: Why Use WatchThatDir?

There are many great file watching tools out there, but here's what makes WatchThatDir an alternative choice:

*   **Multiplatform:** Support Windows, Linux and MacOS.
*   **Simple to Use:** Configuration is handled through a straightforward `config.yaml` file.
*   **Customizable:** Full control over what commands are executed for each type of file event (Create, Rename, Modify, Remove).
*   **Live Config Reloading:** Change the `config.yaml` file, and the changes are picked up without restarting the application.
*   **Efficient:** Built with Go, should be fast and handles multiple file changes at once.
*   **Handles Network Hiccups:**  Detect directory inaccessibility if occured and resume watching once it back online.
*   **Precise Targeting:** Exclude specific paths and filter by file type, ensuring you're only monitoring what matters.
*   **Open Source:** You can see exactly how it works and modify the codes if needed.

## 4. Configuration: The `config.yaml` File Explained

All the settings for WatchThatDir live in a file named `config.yaml`. Here's a breakdown of what each setting does:

```yaml
target_path: "/path/to/watch"         # The directory to monitor for file changes.
processed_path: "/path/to/processed"  # Directory to move processed files to (if post_process is set to 1).
max_workers: 4                        # Number of worker threads for concurrent processing (default: number of CPU cores).
post_process: 0                       # Action after processing: 0 (Do Nothing), 1 (Move), -1 (Delete).
file_type: [".txt", ".jpg"]           # File extensions to process (empty = all types).
process_on_start: true                # true: Process existing files in target_path as newly created files during WatchThatDir startup
logfile_path: "watcher.log"           # Path to the log file.
enable_logging: true                  # Enable (true) or disable (false) logging.
reload_config: 5000                   # Interval in milliseconds for reloading the config file (0 = disable).
check_interval: 5                     # How often (in seconds) to check if the target_path is accessible.
init_run:                             # Command to execute on application startup.
 - "your-executable"
 - "arg1"
 - "arg2"
 - "arg..."
exit_run:                             # Command to execute on application shutdown.
 - "your-executable"
 - "arg1"
 - "arg2"
 - "arg..."
oncreate_run:                         # Command to run when a file is created.
 - "your-executable"
 - "some_args"
 - "{filepath}"                       # {filepath} = captured file path to be processed by your-executable
onmodify_run:                         # Command to run when a file is modified.
 - "your-executable"
 - "{filepath}" 
 - "some_args"
onrename_run:                         # Command to run when a file is renamed.
 - "your-executable"
 - "{filepath}"
onremove_run:                         # Command to run when a file is removed.
 - "your-executable"
 - "{filepath}"
debounce: 250                         # Debounce time in milliseconds.
exclude_path:                         # Paths to exclude (supports direct and substring match).
 - "/path/to/exclude"
 - "/another/path/to/exclude"
```

**A Closer Look:**

  * **`target_path`:** The most important setting - the directory you want to monitor.
  * **`processed_path`:**  If you want processed files to be moved, put the destination directory here.
  * **`max_workers`:** Controls how many files are processed simultaneously.
  * **`post_process`:**  Choose what happens to a file *after* your commands have been executed.
  * **`file_type`:**  A list of file extensions (like `.txt` or `.jpg`). If it's empty, all file types are processed.
  * **`process_on_start`:**  Set this to `true` if you want to process files that are already in `target_path` when WatchThatDir starts.
  * **`logfile_path`:** Where the application's log messages will be saved.
  * **`enable_logging`:**  Turn logging on or off.
  * **`init_run`:** A command (and its arguments) that runs once when the application starts.
  * **`exit_run`:** A command that runs when the application is shutting down.
  * **`oncreate_run`**, **`onmodify_run`**, **`onrename_run`**, **`onremove_run`:** These are the core of the application. Define what commands you want to run for each file event. Use `{filepath}` as a placeholder for the file that triggered the event.
  * **`debounce`:** Helps avoid processing the same file multiple times if it's rapidly changed.
  * **`exclude_path`:** A list of paths you want WatchThatDir to ignore. Useful for temporary folders or system files. Supports both **exact** and **substring** matching of paths.
  * **`reload_config`:** How often the application should check if `config.yaml` has changed. Set to `0` to disable automatic reloading.
  * **`check_interval`:**  How often (in seconds) the application should check if the `target_path` is accessible (especially useful for network drives).

The `init_run`, `exit_run`, `onmodify_run`, `oncreate_run`, `onrename_run` and `onremove_run` section in these YAML configuration allows you to specify a command that will be automatically executed when triggered. This command, along with its arguments, should be provided as a list within the `*_run:` field.  The first element of the list represents the command itself, followed by subsequent elements that represent the arguments to be passed to that command. For instance, if you wanted to execute a Python script named `my_script.py` with arguments `arg1` and `arg2`, your `*_run:` would look like: `["python", "<path_to_the_script>/my_script.py", "arg1", "arg2"]`. It's important to remember that each argument, including flags and their values, should be separate list elements.

## 5\. Building and Running the Application

To get WatchThatDir up and running, you'll need:

1.  **Go (Golang):** Download and install it from [https://golang.org/dl/](https://www.google.com/url?sa=E&source=gmail&q=https://golang.org/dl/)
2.  **Git:** To clone the repository.

**Steps:**

1.  **Clone the repository:**
    ```bash
    git clone <repository_url>
    ```
2.  **Go to the project directory:**
    ```bash
    cd <repository_directory>
    ```
3.  **Build the application:**
    ```bash
    go build
    ```

This will create an executable file (e.g., `WatchThatDir` or `WatchThatDir.exe`) in the project directory.

**Running the Application:**

1.  Make sure you have a `config.yaml` file in the same directory as the executable.
2.  Run the executable from your terminal:
    ```bash
    ./WatchThatDir
    ```

## 6\. Conclusion

WatchThatDir is a simple tool for automating file-related tasks with ease. Give it a try and see how it can simplify your workflow\!


### Credits
Author: \@abahcool | Ai Helper: Gemini 2.0 | Editor: VS Code @ Windows 11
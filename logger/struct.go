package logger

// Logger contains all potential settings available for setting up the rs/zerolog instance
type Logger struct {
	LogFile     string `json:"log_file"`  // Full path and filename of log file; required by LogToFile
	LogLevel    string `json:"log_level"` // One of rs/zerolog support log levels: trace, debug, etc
	LogToStdout bool   `json:"to_stdout"` // Log to STDOUT
	LogToStderr bool   `json:"to_stderr"` // Log to STDERR
	LogToFile   bool   `json:"to_file"`   // Enable logging to file; requires that LogFile is set.
}

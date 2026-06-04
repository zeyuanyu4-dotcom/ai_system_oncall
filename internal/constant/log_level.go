package constant

// Log levels
const (
	LogLevelError = "ERROR"
	LogLevelWarn  = "WARN"
	LogLevelInfo  = "INFO"
	LogLevelDebug = "DEBUG"
)

// ValidLogLevels returns all valid log levels
func ValidLogLevels() []string {
	return []string{LogLevelError, LogLevelWarn, LogLevelInfo, LogLevelDebug}
}

// IsValidLogLevel checks if the log level is valid
func IsValidLogLevel(level string) bool {
	for _, l := range ValidLogLevels() {
		if l == level {
			return true
		}
	}
	return false
}

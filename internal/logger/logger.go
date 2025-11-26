package logger

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"
)

// Logger handles application logging with color support and file output
type Logger struct {
	file  *os.File
	debug bool
}

// New creates a new Logger instance
func New(debug bool) (*Logger, error) {
	l := &Logger{debug: debug}
	if debug {
		exePath, err := os.Executable()
		if err != nil {
			return nil, fmt.Errorf("could not get executable path: %v", err)
		}
		exeDir := filepath.Dir(exePath)
		logDir := filepath.Join(exeDir, "logs")

		if err := os.MkdirAll(logDir, 0755); err != nil {
			return nil, fmt.Errorf("could not create log directory: %v", err)
		}
		timestamp := time.Now().Format("2006-01-02_15-04-05")
		path := filepath.Join(logDir, fmt.Sprintf("%s.log", timestamp))
		f, err := os.Create(path)
		if err != nil {
			return nil, fmt.Errorf("could not create log file: %v", err)
		}
		l.file = f

		// Redirect standard library log to this file as well
		log.SetOutput(f)

		fmt.Printf("Debug logging to: %s\n", path)
	}
	return l, nil
}

// Close closes the log file if it exists
func (l *Logger) Close() {
	if l.file != nil {
		l.file.Close()
	}
}

// Debug logs a message in gray color to console and to file
func (l *Logger) Debug(msg string) {
	if !l.debug {
		return
	}
	ts := time.Now().Format("15:04:05.000")

	// Print to console with gray color (ANSI escape code \033[90m)
	fmt.Printf("\033[90m[%s] %s\033[0m\n", ts, msg)

	if l.file != nil {
		l.file.WriteString(fmt.Sprintf("[%s] %s\n", ts, msg))
	}
}

// Warn logs a message in yellow color to console and to file
func (l *Logger) Warn(msg string) {
	ts := time.Now().Format("15:04:05.000")

	// Print to console with yellow color (ANSI escape code \033[33m)
	fmt.Printf("\033[33m[%s] %s\033[0m\n", ts, msg)

	if l.file != nil {
		l.file.WriteString(fmt.Sprintf("[%s] WARN: %s\n", ts, msg))
	}
}

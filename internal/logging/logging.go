package logging

import (
	"fmt"
	"log"
)

func Info(category, format string, args ...interface{}) {
	// TODO: [Refactor] Replace log.Printf with a structured logging library like zap or slog for better log management.
	log.Printf("[INFO][%s] %s", category, fmt.Sprintf(format, args...))
}

func Error(category, format string, args ...interface{}) {
	// TODO: [Refactor] Replace log.Printf with a structured logging library like zap or slog for better log management.
	log.Printf("[ERROR][%s] %s", category, fmt.Sprintf(format, args...))
}

func Fatal(category, format string, args ...interface{}) {
	// TODO: [Refactor] Replace log.Fatalf with a structured logging library like zap or slog for better log management.
	log.Fatalf("[FATAL][%s] %s", category, fmt.Sprintf(format, args...))
}

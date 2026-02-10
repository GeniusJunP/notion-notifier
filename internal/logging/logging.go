package logging

import (
	"fmt"
	"log"
)

func Info(category, format string, args ...interface{}) {
	log.Printf("[INFO][%s] %s", category, fmt.Sprintf(format, args...))
}

func Error(category, format string, args ...interface{}) {
	log.Printf("[ERROR][%s] %s", category, fmt.Sprintf(format, args...))
}

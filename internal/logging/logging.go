package logging

import (
	"fmt"
	"log/slog"
	"os"
)

var logger = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
	Level: slog.LevelInfo,
}))

func Info(category, format string, args ...interface{}) {
	logger.Info(fmt.Sprintf(format, args...), slog.String("category", category))
}

func Error(category, format string, args ...interface{}) {
	logger.Error(fmt.Sprintf(format, args...), slog.String("category", category))
}

func Fatal(category, format string, args ...interface{}) {
	logger.Error(fmt.Sprintf(format, args...), slog.String("category", category))
	os.Exit(1)
}

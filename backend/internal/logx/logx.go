package logx

import (
	"log/slog"
	"os"
	"strings"
	"sync"
)

var (
	mu     sync.RWMutex
	logger *slog.Logger
)

func Init(level string, env string) {
	lvl := slog.LevelInfo
	switch strings.ToLower(level) {
	case "debug":
		if env != "production" {
			lvl = slog.LevelDebug
		}
	case "warn", "warning":
		lvl = slog.LevelWarn
	case "error":
		lvl = slog.LevelError
	}
	h := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: lvl})
	mu.Lock()
	logger = slog.New(h)
	slog.SetDefault(logger)
	mu.Unlock()
}

func L() *slog.Logger {
	mu.RLock()
	defer mu.RUnlock()
	if logger == nil {
		return slog.Default()
	}
	return logger
}

func Info(msg string, args ...any) {
	if l := L(); l != nil {
		l.Info(msg, args...)
	}
}

func Warn(msg string, args ...any) {
	if l := L(); l != nil {
		l.Warn(msg, args...)
	}
}

func Error(msg string, args ...any) {
	if l := L(); l != nil {
		l.Error(msg, args...)
	}
}

func Debug(msg string, args ...any) {
	if l := L(); l != nil {
		l.Debug(msg, args...)
	}
}

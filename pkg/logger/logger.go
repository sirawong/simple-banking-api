package logger

import (
	"context"
	"log/slog"
	"os"
	"strings"
	"sync"
)

type loggerKey struct{}

type Logger struct {
	baseLogger *slog.Logger
}

var (
	globalLogger *Logger
	once         sync.Once
)

type Options struct {
	Level string // debug | info | warn | error
}

func New(opts Options) (*Logger, error) {
	level := parseLevel(opts.Level)

	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: level})

	return &Logger{baseLogger: slog.New(handler)}, nil
}

// @wire:ignore
func ProvideGlobalLogger() *Logger {
	once.Do(func() {
		globalLogger, _ = New(Options{Level: os.Getenv("LOG_LEVEL")})
	})
	return globalLogger
}

func GetGlobalLogger() *Logger {
	if globalLogger == nil {
		ProvideGlobalLogger()
	}
	return globalLogger
}

func (l *Logger) WithContext(ctx context.Context, fields ...any) context.Context {
	enriched := l.baseLogger.With(fields...)
	return context.WithValue(ctx, loggerKey{}, enriched)
}

func (l *Logger) Context(ctx context.Context) *slog.Logger {
	if lg, ok := ctx.Value(loggerKey{}).(*slog.Logger); ok {
		return lg
	}
	return l.baseLogger
}

func (l *Logger) Info(msg string, args ...any)  { l.baseLogger.Info(msg, args...) }
func (l *Logger) Debug(msg string, args ...any) { l.baseLogger.Debug(msg, args...) }
func (l *Logger) Warn(msg string, args ...any)  { l.baseLogger.Warn(msg, args...) }
func (l *Logger) Error(msg string, err error, args ...any) {
	l.baseLogger.Error(msg, append(args, "error", err)...)
}

func Info(msg string, args ...any)  { GetGlobalLogger().Info(msg, args...) }
func Debug(msg string, args ...any) { GetGlobalLogger().Debug(msg, args...) }
func Warn(msg string, args ...any)  { GetGlobalLogger().Warn(msg, args...) }
func Error(msg string, err error, args ...any) {
	GetGlobalLogger().Error(msg, err, args...)
}

func parseLevel(s string) slog.Level {
	switch strings.ToLower(s) {
	case "debug":
		return slog.LevelDebug
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

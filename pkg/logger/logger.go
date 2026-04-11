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

// Options configures the logger output and level.
type Options struct {
	Level            string // debug | info | warn | error
	Output           string // stdout | cloudwatch (default: stdout)
	CloudWatchGroup  string // required when Output=cloudwatch
	CloudWatchStream string // optional; defaults to hostname-YYYY-MM-DD
	AWSRegion        string // AWS region for CloudWatch
}

// New creates a Logger from explicit options.
func New(opts Options) (*Logger, error) {
	level := parseLevel(opts.Level)

	var handler slog.Handler
	if strings.ToLower(opts.Output) == "cloudwatch" {
		cwHandler, err := newCloudWatchHandler(context.Background(), opts, level)
		if err != nil {
			return nil, err
		}
		handler = cwHandler
	} else {
		handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: level})
	}

	return &Logger{baseLogger: slog.New(handler)}, nil
}

// ProvideGlobalLogger reads LOG_LEVEL / LOG_OUTPUT / CLOUDWATCH_* from environment
// and initialises the singleton logger exactly once.
func ProvideGlobalLogger() *Logger {
	once.Do(func() {
		opts := Options{
			Level:            os.Getenv("LOG_LEVEL"),
			Output:           os.Getenv("LOG_OUTPUT"),
			CloudWatchGroup:  os.Getenv("CLOUDWATCH_LOG_GROUP"),
			CloudWatchStream: os.Getenv("CLOUDWATCH_LOG_STREAM"),
			AWSRegion:        os.Getenv("AWS_REGION"),
		}
		var err error
		globalLogger, err = New(opts)
		if err != nil {
			// Fallback to stdout so the app still boots
			globalLogger, _ = New(Options{Level: opts.Level, Output: "stdout"})
			globalLogger.Warn("cloudwatch logger init failed, falling back to stdout", "error", err.Error())
		}
	})
	return globalLogger
}

// GetGlobalLogger returns the singleton, initialising it on first call.
func GetGlobalLogger() *Logger {
	if globalLogger == nil {
		ProvideGlobalLogger()
	}
	return globalLogger
}

// --- Context helpers ---

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

// --- Structured log methods ---

func (l *Logger) Info(msg string, args ...any)  { l.baseLogger.Info(msg, args...) }
func (l *Logger) Debug(msg string, args ...any) { l.baseLogger.Debug(msg, args...) }
func (l *Logger) Warn(msg string, args ...any)  { l.baseLogger.Warn(msg, args...) }
func (l *Logger) Error(msg string, err error, args ...any) {
	l.baseLogger.Error(msg, append(args, "error", err)...)
}

func (l *Logger) InfoContext(ctx context.Context, msg string, args ...any) {
	l.Context(ctx).Info(msg, args...)
}
func (l *Logger) DebugContext(ctx context.Context, msg string, args ...any) {
	l.Context(ctx).Debug(msg, args...)
}
func (l *Logger) WarnContext(ctx context.Context, msg string, args ...any) {
	l.Context(ctx).Warn(msg, args...)
}
func (l *Logger) ErrorContext(ctx context.Context, msg string, err error, args ...any) {
	l.Context(ctx).Error(msg, append(args, "error", err)...)
}

// --- Package-level helpers (use global logger) ---

func Info(msg string, args ...any)  { GetGlobalLogger().Info(msg, args...) }
func Debug(msg string, args ...any) { GetGlobalLogger().Debug(msg, args...) }
func Warn(msg string, args ...any)  { GetGlobalLogger().Warn(msg, args...) }
func Error(msg string, err error, args ...any) {
	GetGlobalLogger().Error(msg, err, args...)
}
func WithFields(ctx context.Context, fields ...any) context.Context {
	return GetGlobalLogger().WithContext(ctx, fields...)
}

// --- Internal helpers ---

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

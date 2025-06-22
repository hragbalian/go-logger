package logger

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/fatih/color"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	// Color functions
	red    = color.New(color.FgRed).SprintFunc()
	yellow = color.New(color.FgYellow).SprintFunc()
	green  = color.New(color.FgGreen).SprintFunc()
	blue   = color.New(color.FgBlue).SprintFunc()
	cyan   = color.New(color.FgCyan).SprintFunc()
	white  = color.New(color.FgWhite).SprintFunc()
)

// Logger wraps zap.Logger with additional functionality
type Logger struct {
	*zap.Logger
	sugar *zap.SugaredLogger
}

// Config holds logger configuration
type Config struct {
	Level      string
	Format     string
	EnableFile bool
	FilePath   string
}

// NewLogger creates a new logger instance with color support
func NewLogger(config Config) (*Logger, error) {
	level, err := zapcore.ParseLevel(config.Level)
	if err != nil {
		return nil, fmt.Errorf("invalid log level: %w", err)
	}

	// Create console encoder with colors
	consoleEncoderConfig := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		FunctionKey:    zapcore.OmitKey,
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    colorLevelEncoder,
		EncodeTime:     zapcore.TimeEncoderOfLayout("2006-01-02 15:04:05"),
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	// Create file encoder without colors
	fileEncoderConfig := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		FunctionKey:    zapcore.OmitKey,
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.TimeEncoderOfLayout("2006-01-02 15:04:05"),
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	var cores []zapcore.Core

	// Console core with colors
	consoleEncoder := zapcore.NewConsoleEncoder(consoleEncoderConfig)
	consoleCore := zapcore.NewCore(
		consoleEncoder,
		zapcore.AddSync(os.Stdout),
		level,
	)
	cores = append(cores, consoleCore)

	// File core if enabled
	if config.EnableFile {
		if err := os.MkdirAll(filepath.Dir(config.FilePath), 0755); err != nil {
			return nil, fmt.Errorf("failed to create log directory: %w", err)
		}

		fileWriter, err := os.OpenFile(config.FilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			return nil, fmt.Errorf("failed to open log file: %w", err)
		}

		fileEncoder := zapcore.NewJSONEncoder(fileEncoderConfig)
		fileCore := zapcore.NewCore(
			fileEncoder,
			zapcore.AddSync(fileWriter),
			level,
		)
		cores = append(cores, fileCore)
	}

	// Combine cores
	core := zapcore.NewTee(cores...)

	// Create logger with caller information
	zapLogger := zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))

	return &Logger{
		Logger: zapLogger,
		sugar:  zapLogger.Sugar(),
	}, nil
}

// colorLevelEncoder adds colors to log levels
func colorLevelEncoder(level zapcore.Level, enc zapcore.PrimitiveArrayEncoder) {
	var levelStr string
	switch level {
	case zapcore.DebugLevel:
		levelStr = cyan("[DEBUG]")
	case zapcore.InfoLevel:
		levelStr = green("[INFO]")
	case zapcore.WarnLevel:
		levelStr = yellow("[WARN]")
	case zapcore.ErrorLevel:
		levelStr = red("[ERROR]")
	case zapcore.DPanicLevel:
		levelStr = red("[DPANIC]")
	case zapcore.PanicLevel:
		levelStr = red("[PANIC]")
	case zapcore.FatalLevel:
		levelStr = red("[FATAL]")
	default:
		levelStr = fmt.Sprintf("[%s]", level.String())
	}
	enc.AppendString(levelStr)
}

// Sugar returns the sugared logger
func (l *Logger) Sugar() *zap.SugaredLogger {
	return l.sugar
}

// WithField adds a field to the logger
func (l *Logger) WithField(key string, value any) *Logger {
	return &Logger{
		Logger: l.Logger.With(zap.Any(key, value)),
		sugar:  l.sugar.With(key, value),
	}
}

// WithFields adds multiple fields to the logger
func (l *Logger) WithFields(fields map[string]any) *Logger {
	zapFields := make([]zap.Field, 0, len(fields))
	for k, v := range fields {
		zapFields = append(zapFields, zap.Any(k, v))
	}
	return &Logger{
		Logger: l.Logger.With(zapFields...),
		sugar:  l.sugar.With(fields),
	}
}

// Convenience methods with colors
func (l *Logger) Success(msg string, fields ...zap.Field) {
	l.Info(green("✓ ")+msg, fields...)
}

func (l *Logger) Progress(msg string, fields ...zap.Field) {
	l.Info(blue("→ ")+msg, fields...)
}

func (l *Logger) Warning(msg string, fields ...zap.Field) {
	l.Warn(yellow("⚠ ")+msg, fields...)
}

func (l *Logger) Failure(msg string, fields ...zap.Field) {
	l.Error(red("✗ ")+msg, fields...)
}

// Structured logging methods
func (l *Logger) LogEventProcessed(eventID int, eventName string) {
	l.Success("Event processed",
		zap.Int("event_id", eventID),
		zap.String("event_name", eventName),
	)
}

func (l *Logger) LogFileDownloaded(fileName, filePath string, fileSize int64) {
	l.Success("File downloaded",
		zap.String("file_name", fileName),
		zap.String("file_path", filePath),
		zap.Int64("file_size", fileSize),
	)
}

func (l *Logger) LogAPIRequest(url string, statusCode int, duration string) {
	if statusCode >= 200 && statusCode < 300 {
		l.Progress("API request completed",
			zap.String("url", url),
			zap.Int("status_code", statusCode),
			zap.String("duration", duration),
		)
	} else {
		l.Warning("API request failed",
			zap.String("url", url),
			zap.Int("status_code", statusCode),
			zap.String("duration", duration),
		)
	}
}

func (l *Logger) LogDatabaseOperation(operation, table string, count int) {
	l.Progress("Database operation completed",
		zap.String("operation", operation),
		zap.String("table", table),
		zap.Int("count", count),
	)
}

func (l *Logger) LogScrapeSession(sessionID, mode string, stats map[string]int) {
	fields := []zap.Field{
		zap.String("session_id", sessionID),
		zap.String("mode", mode),
	}

	for key, value := range stats {
		fields = append(fields, zap.Int(key, value))
	}

	l.Success("Scrape session completed", fields...)
}

// Sync flushes any buffered log entries
func (l *Logger) Sync() error {
	return l.Logger.Sync()
}

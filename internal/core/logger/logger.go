package core_logger

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type loggerContextKey struct{}

// Logger key is a struct in purpose to not rewrite the key
var key = loggerContextKey{}

// Logger is a custom logger with ZapLogger and log file
type Logger struct {
	*zap.Logger
	file *os.File
}

func (l *Logger) With(fields ...zap.Field) *Logger {
	return &Logger{
		Logger: l.Logger.With(fields...),
		file:   l.file,
	}
}

func (l *Logger) Close() error {
	if err := l.file.Close(); err != nil {
		return fmt.Errorf("Faild to close log file %s: %w", l.file.Name(), err)
	}
	return nil
}

func NewLogger(config LoggerConfig) (*Logger, error) {
	// Logger level from config unmarshall to zapLevel
	zapLevel := zap.NewAtomicLevel()
	if err := zapLevel.UnmarshalText([]byte(config.Level)); err != nil {
		return nil, fmt.Errorf("unmarshall log level: %w", err)
	}

	// Make dir to folder from config
	if err := os.MkdirAll(config.Folder, 0755); err != nil {
		return nil, fmt.Errorf("create logger folder: %w", err)
	}

	timestamp := time.Now().UTC().Format("2006-01-02T15-04-05.000000")
	filePath := filepath.Join(config.Folder, timestamp)
	logFile, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, fmt.Errorf("open logger file: %w", err)
	}

	// Zap logger initalisation
	zapConfig := zap.NewDevelopmentEncoderConfig()
	zapConfig.EncodeTime = zapcore.TimeEncoderOfLayout("2006-01-02T15-04-05.000000")
	zapEncoder := zapcore.NewConsoleEncoder(zapConfig)

	core := zapcore.NewTee(
		zapcore.NewCore(zapEncoder, zapcore.AddSync(os.Stdout), zapLevel),
		zapcore.NewCore(zapEncoder, zapcore.AddSync(logFile), zapLevel),
	)

	logger := zap.New(core, zap.AddCaller())

	return &Logger{
		Logger: logger,
		file:   logFile,
	}, nil

}

func loggerToContext(ctx context.Context, logger *Logger) context.Context {
	return context.WithValue(ctx, key, logger)
}

func loggerFromContext(ctx context.Context) *Logger {
	logger, ok := ctx.Value(key).(*Logger)
	if !ok {
		panic("no logger in context")
	}

	return logger
}

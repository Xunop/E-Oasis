package log

import (
	"os"
	"strings"

	"github.com/Xunop/e-oasis/config"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

var Logger *zap.Logger

func Info(msg string, fields ...zap.Field) {
	Logger.Info(msg, fields...)
}

func Error(msg string, fields ...zap.Field) {
	Logger.Error(msg, fields...)
}

func Debug(msg string, fields ...zap.Field) {
	Logger.Debug(msg, fields...)
}

func Warn(msg string, fields ...zap.Field) {
	Logger.Warn(msg, fields...)
}

func Fatal(msg string, fields ...zap.Field) {
	Logger.Fatal(msg, fields...)
}

func NewLogger() *zap.Logger {
	rotationLog := &lumberjack.Logger{
		Filename:   config.Opts.LogFile,
		MaxSize:    config.Opts.LogFileMaxSize, // megabytes
		MaxBackups: config.Opts.LogFileMaxBackups,
		MaxAge:     config.Opts.LogFileMaxAge, // days
		Compress:   config.Opts.LogCompress,
	}

	return newZap(rotationLog)
}

func newZap(rotationLog *lumberjack.Logger) *zap.Logger {
	encodeConfig := zap.NewProductionEncoderConfig()
	encodeConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	fileEncoder := zapcore.NewJSONEncoder(encodeConfig)
	consoleEncoder := zapcore.NewConsoleEncoder(encodeConfig)

	consoleWriter := zapcore.AddSync(os.Stdout)
	rotationWrite := zapcore.AddSync(rotationLog)

	var defaultLogLevel zapcore.Level
	switch strings.ToLower(config.Opts.LogLevel) {
	case "debug":
		defaultLogLevel = zapcore.DebugLevel
	case "info":
		defaultLogLevel = zapcore.InfoLevel
	case "warn":
		defaultLogLevel = zapcore.WarnLevel
	case "error":
		defaultLogLevel = zapcore.ErrorLevel
	}

	consoleCore := zapcore.NewCore(consoleEncoder, consoleWriter, defaultLogLevel)
	rotationCore := zapcore.NewCore(fileEncoder, rotationWrite, defaultLogLevel)

	core := zapcore.NewTee(consoleCore, rotationCore)

	return zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1), zap.AddStacktrace(zapcore.ErrorLevel))
}

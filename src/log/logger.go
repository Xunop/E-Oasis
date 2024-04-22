package log

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

var Logger *zap.Logger

func init() {
	Logger = newLogger()
}

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

func newLogger() *zap.Logger {
	// TODO: Read config
	filename := "logs.log"

	rotationLog := &lumberjack.Logger{
		Filename:   filename,
		MaxSize:    10, // megabytes
		MaxBackups: 3,
		MaxAge:     28, // days
	}

	return newZap(rotationLog)
}

func newZap(rotationLog *lumberjack.Logger) *zap.Logger {
	config := zap.NewProductionEncoderConfig()
	config.EncodeTime = zapcore.ISO8601TimeEncoder

	fileEncoder := zapcore.NewJSONEncoder(config)
	consoleEncoder := zapcore.NewConsoleEncoder(config)

	consoleWriter := zapcore.AddSync(os.Stdout)
	rotationWrite := zapcore.AddSync(rotationLog)

	defaultLogLevel := zapcore.InfoLevel

	consoleCore := zapcore.NewCore(consoleEncoder, consoleWriter, defaultLogLevel)
	rotationCore := zapcore.NewCore(fileEncoder, rotationWrite, defaultLogLevel)

	core := zapcore.NewTee(consoleCore, rotationCore)

	return zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1), zap.AddStacktrace(zapcore.ErrorLevel))
}

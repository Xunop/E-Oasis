package log

import (
	"os"
	"testing"

	"gopkg.in/natefinch/lumberjack.v2"
)

// Test the log rotation, the log file should be rotated when it reaches the maximum size
func TestLogRotation(t *testing.T) {
	dir := os.TempDir()
	defer os.RemoveAll(dir)

	filename := dir + "/foobar.log"

	rotationLog := &lumberjack.Logger{
		Filename:   filename,
		MaxSize:    1, // megabytes
		MaxBackups: 3,
		MaxAge:     1, // days
	}
	defer rotationLog.Close()
	logger := newZap(rotationLog)
	defer logger.Sync()
	oneMegabyte := 1024 * 1024
	// Write 1MiB of data
	// should create a new file
	rotationLog.Write(make([]byte, oneMegabyte))
	logger.Info("This log should be in a new file")
	// Get file size
	fileInfo, err := os.Stat(filename)
	if err != nil {
		t.Fatal(err)
	}
	if fileInfo.Size() > int64(oneMegabyte) {
		t.Fatalf("File size %d is greater than expected %d", fileInfo.Size(), oneMegabyte)
	}
}

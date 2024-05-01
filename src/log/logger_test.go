package log

import (
	"os"
	"testing"

	"gopkg.in/natefinch/lumberjack.v2"
)

// Test the log rotation, the log file should be rotated when it reaches the maximum size
func TestLogRotation(t *testing.T) {
	// /tmp/e-oasis/
	dir := os.TempDir()
	dir += "/e-oasis-test"
	// Create a directory if not exists
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		err := os.Mkdir(dir, 0755)
		if err != nil {
			t.Fatal(err)
		}
	}
	defer os.RemoveAll(dir)

	filename := dir + "/foobar.log"
	// filename := "./foobar.log"

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
	log1 := "This log should be in a old file"
	log2 := "This log should be in a new file(new foobar.log)"
	logger.Info(log1)
	rotationLog.Write(make([]byte, oneMegabyte))
	logger.Info(log2)
	// Get file size
	fileInfo, err := os.Stat(filename)
	if err != nil {
		t.Fatal(err)
	}
	if fileInfo.Size() > int64(oneMegabyte) {
		t.Errorf("File size %d is greater than expected %d", fileInfo.Size(), oneMegabyte)
	}
	// Get all file in dir
	files, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("Files in dir: %s", files[0].Name())
	t.Logf("Files in dir: %s", files[1].Name())
	if len(files) != 2 {
		t.Errorf("Expected 2 files in dir, got %d", len(files))
	}
}

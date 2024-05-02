package util

import (
	"fmt"
	"os"
	"testing"
)

func TestGenerateNewFileName(t *testing.T) {
	dir := os.TempDir()
	fileDir := dir + "/e-oasis-test-util"
	fileLoc := fileDir + "/test.epub"
	if _, err := os.Stat(fileDir); os.IsNotExist(err) {
		err := os.Mkdir(fileDir, 0755)
		if err != nil {
			t.Fatalf("Error create tempDir: %s, err: %v", fileDir, err)
		}
	}
	defer os.RemoveAll(fileDir)

	if _, err := os.Create(fileLoc); err != nil {
		t.Fatalf("Error create file: %s", fileLoc)
	}

	for i := 1; i < 15; i++ {
		newFile := GenerateNewFileName(fileLoc)
		t.Logf("New filename: %s", newFile)
		expected := fmt.Sprintf("%s/test_%d.epub", fileDir, i)
		if newFile != expected {
			t.Errorf("Error generate new filename, expected: %s, but got: %s", expected, newFile)
		}
		if _, err := os.Create(newFile); err != nil {
			t.Errorf("Error create new file: %s, err: %v", newFile, err)
		}
	}
}

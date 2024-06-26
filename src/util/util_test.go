package util

import (
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"os"
	"sync"
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

func TestGenerateNewDirName(t *testing.T) {
	dir := os.TempDir()
	fileDir := dir + "/e-oasis-test-util"
	curDir := fileDir + "/test"
	if _, err := os.Stat(fileDir); os.IsNotExist(err) {
		err := os.Mkdir(fileDir, 0755)
		if err != nil {
			t.Fatalf("Error create tempDir: %s, err: %v", fileDir, err)
		}
	}
	defer os.RemoveAll(fileDir)

	if err := os.MkdirAll(curDir, os.ModePerm); err != nil {
		t.Fatalf("Error create dir: %s", curDir)
	}

	for i := 1; i < 15; i++ {
		newDir := GenerateNewDirName(curDir)
		t.Logf("New dirname: %s", newDir)
		expected := fmt.Sprintf("%s/test_%d", fileDir, i)
		if newDir != expected {
			t.Errorf("Error generate new dirname, expected: %s, but got: %s", expected, newDir)
		}
		if err := os.Mkdir(newDir, 0755); err != nil {
			t.Errorf("Error create new dir: %s, err: %v", newDir, err)
		}
	}
}

// createTempImage creates a temporary image file for testing purposes.
func createTempImage(extension string) (string, error) {
	img := image.NewRGBA(image.Rect(0, 0, 100, 100))
	for y := 0; y < 100; y++ {
		for x := 0; x < 100; x++ {
			img.Set(x, y, color.RGBA{uint8(x), uint8(y), 0, 255})
		}
	}

	tempFile, err := os.CreateTemp("", "test-*."+extension)
	if err != nil {
		return "", err
	}
	defer tempFile.Close()

	switch extension {
	case "jpg", "jpeg":
		err = jpeg.Encode(tempFile, img, nil)
	case "png":
		err = png.Encode(tempFile, img)
	}

	if err != nil {
		return "", err
	}

	return tempFile.Name(), nil
}

func TestImageToWebp(t *testing.T) {
	formats := []string{"jpg", "jpeg", "png"}

	for _, format := range formats {
		t.Run(fmt.Sprintf("Test %s to WebP", format), func(t *testing.T) {
			tempFile, err := createTempImage(format)
			if err != nil {
				t.Fatalf("Failed to create temporary %s file: %v", format, err)
			}
			defer os.Remove(tempFile)

			done := make(chan bool)

			go func(tempFile string) {
				defer close(done)
				// 75 is the quality of the WebP image
				outputFileName := ImageToWebp(tempFile, 75)
				if outputFileName == "" {
					t.Error("Expected output file name, got empty string")
					return
				}

				if _, err := os.Stat(outputFileName); os.IsNotExist(err) {
					t.Errorf("Expected WebP file %s to exist, but it does not", outputFileName)
				} else {
					os.Remove(outputFileName)
				}
			}(tempFile)

			<-done
		})
	}
}

func TestImageToWebpConcurrent(t *testing.T) {
	waitGroup := sync.WaitGroup{}

	waitGroup.Add(10)
	for i := 1; i <= 10; i++ {

		tempFile, err := createTempImage("jpg")
		if err != nil {
			t.Fatalf("Failed to create temporary %s file: %v", "jpg", err)
		}

		go func(i int) {
			defer waitGroup.Done()
			outputFileName := ImageToWebp(tempFile, 75)
			if outputFileName == "" {
				t.Error("Expected output file name, got empty string")
				return
			}

			if _, err := os.Stat(outputFileName); os.IsNotExist(err) {
				t.Errorf("Expected WebP file %s to exist, but it does not", outputFileName)
			} else {
				// Remove file
				os.Remove(outputFileName)
			}
		}(i)
	}

	waitGroup.Wait()
}

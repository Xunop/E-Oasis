package util

import (
	"crypto/rand"
	"fmt"
	"image"
	_ "image/jpeg" // Register the JPEG format
	_ "image/png" // Register the PNG format
	"math/big"
	"net/mail"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"unicode"

	"github.com/google/uuid"
    "github.com/chai2010/webp"
)

// ConvertStringToInt32 converts a string to int32.
func ConvertStringToInt32(src string) (int32, error) {
	parsed, err := strconv.ParseInt(src, 10, 32)
	if err != nil {
		return 0, err
	}
	return int32(parsed), nil
}

// HasPrefixes returns true if the string s has any of the given prefixes.
func HasPrefixes(src string, prefixes ...string) bool {
	for _, prefix := range prefixes {
		if strings.HasPrefix(src, prefix) {
			return true
		}
	}
	return false
}

// ValidateEmail validates the email.
func ValidateEmail(email string) bool {
	if _, err := mail.ParseAddress(email); err != nil {
		return false
	}
	return true
}

func GenUUID() string {
	return uuid.New().String()
}

var letters = []rune("0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

// RandomString returns a random string with length n.
func RandomString(n int) (string, error) {
	var sb strings.Builder
	sb.Grow(n)
	for i := 0; i < n; i++ {
		// The reason for using crypto/rand instead of math/rand is that
		// the former relies on hardware to generate random numbers and
		// thus has a stronger source of random numbers.
		randNum, err := rand.Int(rand.Reader, big.NewInt(int64(len(letters))))
		if err != nil {
			return "", err
		}
		if _, err := sb.WriteRune(letters[randNum.Uint64()]); err != nil {
			return "", err
		}
	}
	return sb.String(), nil
}

// GenerateNewFileName is a helper function to generate a new file name
func GenerateNewFileName(filePath string) string {
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return filePath // file does not exist, return the same name
	}

	dir := filepath.Dir(filePath)
	base := filepath.Base(filePath)
	ext := filepath.Ext(base)
	fileName := strings.TrimSuffix(base, ext)

	existingFiles, err := filepath.Glob(filepath.Join(dir, fileName+"_*[0-9]"+ext))
	if err != nil {
		return filePath
	}

	index := 1
	for _, existingFile := range existingFiles {
		existingBase := filepath.Base(existingFile)
		existingName := strings.TrimSuffix(existingBase, ext)
		var existingIndex int
		fileName = strings.Split(existingName, "_")[0]
		existingIndex, err = strconv.Atoi(strings.Split(existingName, "_")[1])
		if err == nil && existingIndex >= index {
			index = existingIndex + 1
		}
	}
	newFileName := fmt.Sprintf("%s_%d%s", fileName, index, ext)
	return filepath.Join(dir, newFileName)
}

// GenerateNewDir is a helper function to generate a new dir name
// If dir exist, return dir_1/dir_2
func GenerateNewDirName(filePath string) string {
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return filePath
	}

	dir := filepath.Dir(filePath)
	base := filepath.Base(filePath)

	// If the file path matches the pattern
	existingDirs, err := filepath.Glob(filepath.Join(dir, base+"_*[0-9]"))
	if err != nil {
		return filePath
	}

	index := 1
	for _, existingDir := range existingDirs {
		existingBase := filepath.Base(existingDir)
		// existingName := strings.TrimSuffix(existingBase, ext)
		var existingIndex int
		lastIndex := len(strings.Split(existingBase, "_")) - 1
		existingIndex, err = strconv.Atoi(strings.Split(existingBase, "_")[lastIndex])
		if err == nil && existingIndex >= index {
			index = existingIndex + 1
		}
	}
	newFileName := fmt.Sprintf("%s_%d", base, index)
	return filepath.Join(dir, newFileName)
}

// IsChinese checks if the string contains Chinese characters
func IsChinese(s string) bool {
	for _, r := range s {
		if unicode.Is(unicode.Scripts["Han"], r) {
			return true
		}
	}
	return false
}

// ImageToWebp converts an image to webp format.
// quality is a float32 value between 0 and 100.
// I recommend using a quality of 75. It is a good balance between quality and file size.
func ImageToWebp(src string, quality float32) string {
    file, err := os.Open(src)
    if err != nil {
        fmt.Println("Error opening source file:", err)
        return ""
    }
    defer file.Close()

    img, _, err := image.Decode(file)
    if err != nil {
        fmt.Println("Error decoding image:", err)
        return ""
    }

    // Remove extension from file name
    fileName := strings.TrimSuffix(src, filepath.Ext(src))
    outputFileName := fileName + ".webp"
    outputFile, err := os.Create(outputFileName)
    if err != nil {
        fmt.Println("Error creating output file:", err)
        return ""
    }
    defer outputFile.Close()

	// Lossless compression
    err = webp.Encode(outputFile, img, &webp.Options{Lossless: false, Quality: quality})
    if err != nil {
        fmt.Println("Error encoding image to WebP:", err)
        return ""
    }

    return outputFileName
}

// CheckFileType checks if the file type is in the list of types
func CheckFileType(file string, types ...string) bool {
	// Check type by fileSignatures
	return false
}

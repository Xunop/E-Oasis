package storage // import "github.com/Xunop/e-oasis/internal/storage"

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/Xunop/e-oasis/internal/config"
	"github.com/Xunop/e-oasis/internal/log"
	"github.com/Xunop/e-oasis/internal/util"
	"go.uber.org/zap"
)

type LocalStorage struct {
	// Path to the storage directory
	Path string
	// User ID
	UID int
}

func (s *LocalStorage) Save(data []byte) error {
	return nil
}

func (s *LocalStorage) Load() ([]byte, error) {
	return nil, nil
}

func storeFile(reader io.Reader, fileName string, uid int) (string, error) {
	// Check if the file type is supported
	ext := filepath.Ext(fileName)
	if !config.CheckSupportedTypes(ext[1:]) {
		return "", fmt.Errorf("Unsupported file type: %s", ext)
	}

	// Generate the storage path
	fileBase := strings.TrimSuffix(filepath.Base(fileName), ext)
	bookPath := fmt.Sprintf("%s/%d/books/%s", config.Opts.Data, uid, fileBase)
	bookPath = util.GenerateNewDirName(bookPath)

	// Create directories if not exist
	if err := os.MkdirAll(bookPath, os.ModePerm); err != nil {
		return "", fmt.Errorf("Failed to create directories: %v", err)
	}

	// Calculate hash and save the file
	hash := sha256.New()
	filePath := filepath.Join(bookPath, fileBase+ext)
	outFile, err := os.Create(filePath)
	if err != nil {
		return "", fmt.Errorf("Failed to create file: %v", err)
	}
	defer outFile.Close()

	// Copy data to the file and calculate the hash
	if _, err := io.Copy(io.MultiWriter(outFile, hash), reader); err != nil {
		return "", fmt.Errorf("Failed to write file: %v", err)
	}

	fileHash := hex.EncodeToString(hash.Sum(nil))
	log.Debug("Stored file", zap.String("path", filePath), zap.String("hash", fileHash))

	return filePath, nil
}

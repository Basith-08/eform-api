package storage

import (
	"fmt"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"eform/backend/internal/domain"

	"github.com/google/uuid"
)

type DocumentRule struct {
	Extensions map[string]struct{}
	MaxBytes   int64
}

type LocalStorage struct {
	root  string
	rules map[string]DocumentRule
}

func NewLocalStorage(root string, maxBytes int64) (*LocalStorage, error) {
	if err := os.MkdirAll(root, 0o755); err != nil {
		return nil, err
	}

	return &LocalStorage{
		root: root,
		rules: map[string]DocumentRule{
			domain.DocumentTypeCV: {
				Extensions: map[string]struct{}{".pdf": {}, ".doc": {}, ".docx": {}},
				MaxBytes:   maxBytes,
			},
			domain.DocumentTypeKTP: {
				Extensions: map[string]struct{}{".png": {}, ".jpg": {}, ".jpeg": {}},
				MaxBytes:   maxBytes,
			},
			domain.DocumentTypeNPWP: {
				Extensions: map[string]struct{}{".png": {}, ".jpg": {}, ".jpeg": {}},
				MaxBytes:   maxBytes,
			},
		},
	}, nil
}

func (s *LocalStorage) SaveDocument(userID uuid.UUID, documentType string, file *multipart.FileHeader) (domain.Document, error) {
	rule, ok := s.rules[documentType]
	if !ok {
		return domain.Document{}, domain.NewAppError(http.StatusBadRequest, "unsupported document type", nil)
	}

	extension := strings.ToLower(filepath.Ext(file.Filename))
	if _, ok := rule.Extensions[extension]; !ok {
		return domain.Document{}, domain.NewAppError(http.StatusBadRequest, fmt.Sprintf("%s file format is not allowed", documentType), nil)
	}

	if file.Size > rule.MaxBytes {
		return domain.Document{}, domain.NewAppError(http.StatusBadRequest, fmt.Sprintf("%s file exceeds the 1 MB limit", documentType), nil)
	}

	src, err := file.Open()
	if err != nil {
		return domain.Document{}, err
	}
	defer src.Close()

	subDir := filepath.Join(s.root, userID.String())
	if err := os.MkdirAll(subDir, 0o755); err != nil {
		return domain.Document{}, err
	}

	generatedName := fmt.Sprintf("%s-%d%s", documentType, time.Now().UnixNano(), extension)
	fullPath := filepath.Join(subDir, generatedName)

	dst, err := os.Create(fullPath)
	if err != nil {
		return domain.Document{}, err
	}
	defer dst.Close()

	if _, err = dst.ReadFrom(src); err != nil {
		return domain.Document{}, err
	}

	return domain.Document{
		UserID:    userID,
		Type:      documentType,
		FileName:  file.Filename,
		FilePath:  strings.TrimPrefix(fullPath, "./"),
		MimeType:  file.Header.Get("Content-Type"),
		SizeBytes: file.Size,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}, nil
}

func (s *LocalStorage) DeleteFile(path string) error {
	if path == "" {
		return nil
	}

	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return err
	}

	return nil
}

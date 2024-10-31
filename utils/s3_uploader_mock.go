package utils

import (
	"context"
	"os"
	"sync"
)

// MockS3Uploader est une implémentation mock de S3Uploader pour les tests
type MockS3Uploader struct {
	Uploads []string
	mu      sync.Mutex
}

func NewMockS3Uploader() *MockS3Uploader {
	return &MockS3Uploader{
		Uploads: []string{},
	}
}

func (m *MockS3Uploader) UploadFile(ctx context.Context, file *os.File, key string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Uploads = append(m.Uploads, key)
	return nil
}

func (m *MockS3Uploader) GetUploads() []string {
	m.mu.Lock()
	defer m.mu.Unlock()
	return append([]string(nil), m.Uploads...)
}

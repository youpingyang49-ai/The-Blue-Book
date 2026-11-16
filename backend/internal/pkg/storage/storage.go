package storage

import (
	"context"
	"crypto/rand"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

// Storage is a simple local-disk object store (dev stand-in for MinIO/OSS/S3).
type Storage struct {
	dir     string
	baseURL string
}

func New(dir, baseURL string) (*Storage, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, err
	}
	return &Storage{dir: dir, baseURL: baseURL}, nil
}

func (s *Storage) Upload(ctx context.Context, key string, reader io.Reader, size int64, contentType string) error {
	dst, err := os.Create(filepath.Join(s.dir, key))
	if err != nil {
		return err
	}
	defer dst.Close()
	_, err = io.Copy(dst, reader)
	return err
}

func (s *Storage) PresignedGetURL(ctx context.Context, key string) (string, error) {
	return s.baseURL + "/uploads/" + key, nil
}

func NewKey(ext string) string {
	b := make([]byte, 12)
	_, _ = rand.Read(b)
	return fmt.Sprintf("%d-%x%s", time.Now().UnixNano(), b, ext)
}

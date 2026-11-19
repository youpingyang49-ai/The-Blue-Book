package storage

import (
	"context"
	"crypto/rand"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// Storage is the object-storage abstraction (local disk or MinIO).
type Storage interface {
	Upload(ctx context.Context, key string, reader io.Reader, size int64, contentType string) error
	PresignedGetURL(ctx context.Context, key string) (string, error)
}

func NewKey(ext string) string {
	b := make([]byte, 12)
	_, _ = rand.Read(b)
	return fmt.Sprintf("%d-%x%s", time.Now().UnixNano(), b, ext)
}

// ---- Local disk ----

type LocalStorage struct {
	dir     string
	baseURL string
}

func NewLocal(dir, baseURL string) (*LocalStorage, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, err
	}
	return &LocalStorage{dir: dir, baseURL: baseURL}, nil
}

func (s *LocalStorage) Upload(ctx context.Context, key string, reader io.Reader, size int64, contentType string) error {
	dst, err := os.Create(filepath.Join(s.dir, key))
	if err != nil {
		return err
	}
	defer dst.Close()
	_, err = io.Copy(dst, reader)
	return err
}

func (s *LocalStorage) PresignedGetURL(ctx context.Context, key string) (string, error) {
	return s.baseURL + "/uploads/" + key, nil
}

// ---- MinIO ----

type MinioStorage struct {
	client *minio.Client
	bucket string
}

func NewMinio(endpoint, accessKey, secretKey, bucket string, useSSL bool) (*MinioStorage, error) {
	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		return nil, err
	}
	s := &MinioStorage{client: client, bucket: bucket}

	ctx := context.Background()
	exists, err := client.BucketExists(ctx, bucket)
	if err != nil {
		return nil, err
	}
	if !exists {
		if err := client.MakeBucket(ctx, bucket, minio.MakeBucketOptions{}); err != nil {
			return nil, err
		}
	}
	return s, nil
}

func (s *MinioStorage) Upload(ctx context.Context, key string, reader io.Reader, size int64, contentType string) error {
	_, err := s.client.PutObject(ctx, s.bucket, key, reader, size, minio.PutObjectOptions{ContentType: contentType})
	return err
}

func (s *MinioStorage) PresignedGetURL(ctx context.Context, key string) (string, error) {
	u, err := s.client.PresignedGetObject(ctx, s.bucket, key, 7*24*time.Hour, nil)
	if err != nil {
		return "", err
	}
	return u.String(), nil
}

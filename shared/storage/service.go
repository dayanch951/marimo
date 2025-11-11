package storage

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// StorageService handles file storage operations
type StorageService struct {
	client     *minio.Client
	bucketName string
	useLocal   bool
	localPath  string
}

// FileInfo represents uploaded file information
type FileInfo struct {
	ID          string    `json:"id"`
	Filename    string    `json:"filename"`
	OriginalName string   `json:"original_name"`
	Size        int64     `json:"size"`
	ContentType string    `json:"content_type"`
	URL         string    `json:"url"`
	UploadedAt  time.Time `json:"uploaded_at"`
}

// NewStorageService creates a new storage service
func NewStorageService() (*StorageService, error) {
	useLocal := os.Getenv("USE_LOCAL_STORAGE") == "true"

	if useLocal {
		localPath := os.Getenv("LOCAL_STORAGE_PATH")
		if localPath == "" {
			localPath = "./uploads"
		}

		// Create local directory if not exists
		if err := os.MkdirAll(localPath, 0755); err != nil {
			return nil, fmt.Errorf("failed to create local storage directory: %w", err)
		}

		return &StorageService{
			useLocal:  true,
			localPath: localPath,
		}, nil
	}

	// MinIO/S3 configuration
	endpoint := getEnv("MINIO_ENDPOINT", "localhost:9000")
	accessKey := getEnv("MINIO_ACCESS_KEY", "minioadmin")
	secretKey := getEnv("MINIO_SECRET_KEY", "minioadmin")
	useSSL := os.Getenv("MINIO_USE_SSL") == "true"
	bucketName := getEnv("MINIO_BUCKET", "marimo-files")

	// Initialize MinIO client
	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to initialize MinIO client: %w", err)
	}

	// Create bucket if not exists
	ctx := context.Background()
	exists, err := client.BucketExists(ctx, bucketName)
	if err != nil {
		return nil, fmt.Errorf("failed to check bucket existence: %w", err)
	}

	if !exists {
		err = client.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{})
		if err != nil {
			return nil, fmt.Errorf("failed to create bucket: %w", err)
		}
		log.Printf("Bucket %s created successfully", bucketName)
	}

	return &StorageService{
		client:     client,
		bucketName: bucketName,
		useLocal:   false,
	}, nil
}

// UploadFile uploads a file to storage
func (s *StorageService) UploadFile(ctx context.Context, reader io.Reader, originalFilename string, contentType string, size int64) (*FileInfo, error) {
	// Generate unique filename
	ext := filepath.Ext(originalFilename)
	filename := fmt.Sprintf("%s%s", uuid.New().String(), ext)

	if s.useLocal {
		return s.uploadLocal(reader, filename, originalFilename, contentType, size)
	}

	return s.uploadMinio(ctx, reader, filename, originalFilename, contentType, size)
}

// uploadLocal saves file to local filesystem
func (s *StorageService) uploadLocal(reader io.Reader, filename, originalFilename, contentType string, size int64) (*FileInfo, error) {
	filePath := filepath.Join(s.localPath, filename)

	file, err := os.Create(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	written, err := io.Copy(file, reader)
	if err != nil {
		return nil, fmt.Errorf("failed to write file: %w", err)
	}

	return &FileInfo{
		ID:           filename,
		Filename:     filename,
		OriginalName: originalFilename,
		Size:         written,
		ContentType:  contentType,
		URL:          fmt.Sprintf("/files/%s", filename),
		UploadedAt:   time.Now(),
	}, nil
}

// uploadMinio uploads file to MinIO/S3
func (s *StorageService) uploadMinio(ctx context.Context, reader io.Reader, filename, originalFilename, contentType string, size int64) (*FileInfo, error) {
	info, err := s.client.PutObject(ctx, s.bucketName, filename, reader, size, minio.PutObjectOptions{
		ContentType: contentType,
		UserMetadata: map[string]string{
			"original-filename": originalFilename,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to upload file: %w", err)
	}

	// Generate presigned URL (valid for 7 days)
	url, err := s.client.PresignedGetObject(ctx, s.bucketName, filename, 7*24*time.Hour, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to generate URL: %w", err)
	}

	return &FileInfo{
		ID:           filename,
		Filename:     filename,
		OriginalName: originalFilename,
		Size:         info.Size,
		ContentType:  contentType,
		URL:          url.String(),
		UploadedAt:   time.Now(),
	}, nil
}

// DownloadFile retrieves a file from storage
func (s *StorageService) DownloadFile(ctx context.Context, filename string) (io.ReadCloser, *FileInfo, error) {
	if s.useLocal {
		return s.downloadLocal(filename)
	}

	return s.downloadMinio(ctx, filename)
}

// downloadLocal reads file from local filesystem
func (s *StorageService) downloadLocal(filename string) (io.ReadCloser, *FileInfo, error) {
	filePath := filepath.Join(s.localPath, filename)

	file, err := os.Open(filePath)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to open file: %w", err)
	}

	stat, err := file.Stat()
	if err != nil {
		file.Close()
		return nil, nil, fmt.Errorf("failed to stat file: %w", err)
	}

	return file, &FileInfo{
		ID:          filename,
		Filename:    filename,
		Size:        stat.Size(),
		UploadedAt:  stat.ModTime(),
	}, nil
}

// downloadMinio downloads file from MinIO/S3
func (s *StorageService) downloadMinio(ctx context.Context, filename string) (io.ReadCloser, *FileInfo, error) {
	object, err := s.client.GetObject(ctx, s.bucketName, filename, minio.GetObjectOptions{})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get object: %w", err)
	}

	stat, err := object.Stat()
	if err != nil {
		object.Close()
		return nil, nil, fmt.Errorf("failed to stat object: %w", err)
	}

	return object, &FileInfo{
		ID:           filename,
		Filename:     filename,
		OriginalName: stat.UserMetadata["original-filename"],
		Size:         stat.Size,
		ContentType:  stat.ContentType,
		UploadedAt:   stat.LastModified,
	}, nil
}

// DeleteFile removes a file from storage
func (s *StorageService) DeleteFile(ctx context.Context, filename string) error {
	if s.useLocal {
		filePath := filepath.Join(s.localPath, filename)
		return os.Remove(filePath)
	}

	return s.client.RemoveObject(ctx, s.bucketName, filename, minio.RemoveObjectOptions{})
}

// ListFiles lists all files in storage
func (s *StorageService) ListFiles(ctx context.Context, prefix string) ([]FileInfo, error) {
	var files []FileInfo

	if s.useLocal {
		pattern := filepath.Join(s.localPath, prefix+"*")
		matches, err := filepath.Glob(pattern)
		if err != nil {
			return nil, err
		}

		for _, match := range matches {
			stat, err := os.Stat(match)
			if err != nil {
				continue
			}

			files = append(files, FileInfo{
				ID:         filepath.Base(match),
				Filename:   filepath.Base(match),
				Size:       stat.Size(),
				UploadedAt: stat.ModTime(),
			})
		}

		return files, nil
	}

	// List objects from MinIO
	objectCh := s.client.ListObjects(ctx, s.bucketName, minio.ListObjectsOptions{
		Prefix:    prefix,
		Recursive: true,
	})

	for object := range objectCh {
		if object.Err != nil {
			return nil, object.Err
		}

		files = append(files, FileInfo{
			ID:          object.Key,
			Filename:    object.Key,
			Size:        object.Size,
			ContentType: object.ContentType,
			UploadedAt:  object.LastModified,
		})
	}

	return files, nil
}

// GetFileURL generates a temporary URL for file access
func (s *StorageService) GetFileURL(ctx context.Context, filename string, expires time.Duration) (string, error) {
	if s.useLocal {
		return fmt.Sprintf("/files/%s", filename), nil
	}

	url, err := s.client.PresignedGetObject(ctx, s.bucketName, filename, expires, nil)
	if err != nil {
		return "", fmt.Errorf("failed to generate presigned URL: %w", err)
	}

	return url.String(), nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

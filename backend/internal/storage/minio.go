package storage

import (
	"context"
	"fmt"
	"io"
	"log"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// Docs: https://github.com/minio/minio-go/blob/master/examples/s3/makebucket.go

type MinIOStorage struct {
	client *minio.Client
	bucket string
}

func NewMinIOStorage(endpoint, accessKey, secretKey, bucket string, useSSL bool, region string) (*MinIOStorage, error) {
	ctx := context.Background()

	minioClient, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create MinIO client: %w", err)
	}

	exists, err := minioClient.BucketExists(ctx, bucket)
	if err != nil {
		return nil, fmt.Errorf("failed to check bucket existence: %w", err)
	}

	if !exists {
		opts := minio.MakeBucketOptions{Region: region}
		if err := minioClient.MakeBucket(ctx, bucket, opts); err != nil {
			return nil, fmt.Errorf("failed to create bucket: %w", err)
		}
		log.Printf("Successfully created bucket %s\n", bucket)
	} else {
		log.Printf("Bucket %s already exists\n", bucket)
	}

	return &MinIOStorage{client: minioClient, bucket: bucket}, nil
}

func (m *MinIOStorage) SaveFile(ctx context.Context, objectName string, reader io.Reader, size int64, contentType string) error {
	info, err := m.client.PutObject(ctx, m.bucket, objectName, reader, size, minio.PutObjectOptions{ContentType: contentType})
	if err != nil {
		return fmt.Errorf("failed to upload file: %w", err)
	}
	log.Printf("Successfully uploaded %s of size %d\n", objectName, info.Size)
	return nil
}

func (m *MinIOStorage) GetFile(ctx context.Context, objectName string) (io.ReadCloser, error) {
	obj, err := m.client.GetObject(ctx, m.bucket, objectName, minio.GetObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get file: %w", err)
	}
	return obj, nil
}

func (m *MinIOStorage) GetFileRange(ctx context.Context, objectName string, start, end int64) (io.ReadCloser, error) {
	opts := minio.GetObjectOptions{}
	if err := opts.SetRange(start, end); err != nil {
		return nil, fmt.Errorf("failed to set range: %w", err)
	}

	obj, err := m.client.GetObject(ctx, m.bucket, objectName, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to get file range: %w", err)
	}
	return obj, nil
}

func (m *MinIOStorage) DeleteFile(ctx context.Context, objectName string) error {
	if err := m.client.RemoveObject(ctx, m.bucket, objectName, minio.RemoveObjectOptions{}); err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}
	return nil
}

func (m *MinIOStorage) GetFileInfo(ctx context.Context, objectName string) (minio.ObjectInfo, error) {
	info, err := m.client.StatObject(ctx, m.bucket, objectName, minio.StatObjectOptions{})
	if err != nil {
		return minio.ObjectInfo{}, fmt.Errorf("failed to get file info: %w", err)
	}
	return info, nil
}

// MinIOObject represents a MinIO object for storage analysis
type MinIOObject struct {
	Key  string
	Size int64
}

// ListAllObjects lists all objects in the bucket for storage analysis
func (m *MinIOStorage) ListAllObjects(ctx context.Context) ([]MinIOObject, error) {
	var objects []MinIOObject

	// Create a channel to receive objects
	objectCh := m.client.ListObjects(ctx, m.bucket, minio.ListObjectsOptions{
		Recursive: true,
	})

	for object := range objectCh {
		if object.Err != nil {
			return nil, fmt.Errorf("failed to list objects: %w", object.Err)
		}

		objects = append(objects, MinIOObject{
			Key:  object.Key,
			Size: object.Size,
		})
	}

	return objects, nil
}

package storage

import (
	"context"
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

// NewMinIOStorage creates a new MinIO storage client
func NewMinIOStorage(endpoint, accessKey, secretKey, bucket string, useSSL bool, region string) (*MinIOStorage, error) {
	ctx := context.Background()
	opts := minio.MakeBucketOptions{
		Region: region,
	}

	// 1. Create MinIO client using minio.New()
	minioClient, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		log.Fatalln(err)
	}

	// 2. Check if bucket exists with client.BucketExists()
	// 3. If not exists, create with client.MakeBucket()
	err = minioClient.MakeBucket(ctx, bucket, opts)
	if err != nil {
		// Check to see if we already own this bucket (which happens if you run this twice)
		exists, errBucketExists := minioClient.BucketExists(ctx, bucket)
		if errBucketExists == nil && exists {
			log.Printf("We already own %s\n", bucket)
		} else {
			log.Fatalln(err)
			return nil, err
		}
	} else {
		log.Printf("Successfully created %s\n", bucket)
	}

	// 4. Return MinIOStorage instance
	return &MinIOStorage{
		client: minioClient,
		bucket: bucket,
	}, nil
}

// SaveFile saves a file to MinIO
func (m *MinIOStorage) SaveFile(ctx context.Context, objectName string, reader io.Reader, size int64, contentType string) error {
	// 1. Use client.PutObject() to upload the file
	info, err := m.client.PutObject(ctx, m.bucket, objectName, reader, size, minio.PutObjectOptions{ContentType: contentType})
	if err != nil {
		log.Fatalln(err)
		return err
	}
	log.Printf("Successfully uploaded %s of size %d\n", objectName, info.Size)
	return nil
}

// GetFile retrieves a file from MinIO
func (m *MinIOStorage) GetFile(ctx context.Context, objectName string) (io.ReadCloser, error) {
	// 1. Use client.GetObject() to download the file
	minioObj, err := m.client.GetObject(ctx, m.bucket, objectName, minio.GetObjectOptions{})
	if err != nil {
		log.Fatalln(err)
		return nil, err
	}
	// 2. Return the io.ReadCloser (don't read entire file into memory)
	// 3. Caller is responsible for closing the reader
	return minioObj, nil
}

// GetFileRange retrieves a specific byte range of a file (for streaming)
func (m *MinIOStorage) GetFileRange(ctx context.Context, objectName string, start, end int64) (io.ReadCloser, error) {
	// 1. Create GetObjectOptions and set the byte range
	opts := minio.GetObjectOptions{}

	// 2. Set the range using SetRange method
	// This tells MinIO to only return bytes from 'start' to 'end' (both inclusive)
	if err := opts.SetRange(start, end); err != nil {
		return nil, err
	}

	// 3. Get the object with range options
	// MinIO will only return the requested byte range, not the entire file
	minioObj, err := m.client.GetObject(ctx, m.bucket, objectName, opts)
	if err != nil {
		log.Fatalln(err)
		return nil, err
	}

	// 4. Return the reader - it will only contain the requested bytes
	// The caller is responsible for closing this reader when done
	return minioObj, nil
}

// DeleteFile deletes a file from MinIO
func (m *MinIOStorage) DeleteFile(ctx context.Context, objectName string) error {
	// 1. Use client.RemoveObject() to delete the file
	err := m.client.RemoveObject(ctx, m.bucket, objectName, minio.RemoveObjectOptions{})
	if err != nil {
		log.Fatalln(err)
		return err
	}
	return nil
}

// GetFileInfo gets metadata about a file without downloading it
func (m *MinIOStorage) GetFileInfo(ctx context.Context, objectName string) (minio.ObjectInfo, error) {
	// 1. Use client.StatObject() to get file info
	info, err := m.client.StatObject(ctx, m.bucket, objectName, minio.StatObjectOptions{})
	if err != nil {
		log.Fatalln(err)
		return minio.ObjectInfo{}, err
	}
	// 2. Return ObjectInfo which includes size, content-type, etc.
	return info, nil
}

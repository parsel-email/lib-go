package helpers

import (
	"bytes"
	"context"
	"log"
	"testing"

	"github.com/minio/minio-go/v7"
	client "github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

func TestCreateMinioContainer(t *testing.T) {
	ctx := context.Background()

	// Create container
	container, err := CreateMinioContainer(ctx)
	if err != nil {
		t.Fatalf("could not create minio container: %v", err)
	}

	// Ensure cleanup
	t.Cleanup(func() {
		if err := container.Terminate(ctx); err != nil {
			t.Errorf("failed to close minio client: %v", err)
		}
	})

	// Verify connection string is not empty
	if container.Endpoint == "" {
		t.Error("expected connection string to not be empty")
	}

	minioClient, err := client.New(container.Endpoint, &client.Options{
		Creds:  credentials.NewStaticV4(container.AccessKeyID, container.SecretAccessKey, ""),
		Secure: false,
	})
	if err != nil {
		log.Fatalf("Failed to create MinIO client: %v", err)
	}

	// Verify bucket exists and is accessible
	exists, err := minioClient.BucketExists(ctx, "testbucket")
	if err != nil {
		t.Fatalf("failed to check if bucket exists: %v", err)
	}
	if !exists {
		t.Error("expected test bucket to exist")
	}

	// Test uploading a small object
	bucketName := "testbucket"
	objectName := "test.txt"
	contentType := "text/plain"
	testData := []byte("hello world")
	_, err = minioClient.PutObject(ctx, bucketName, objectName, bytes.NewReader(testData), int64(len(testData)), minio.PutObjectOptions{ContentType: contentType})
	if err != nil {
		t.Fatalf("failed to upload test object: %v", err)
	}
}

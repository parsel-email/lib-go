package helpers

import (
	"context"
	"log"

	client "github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

type MinioContainer struct {
	testcontainers.Container
	Endpoint        string
	AccessKeyID     string
	SecretAccessKey string
	Port            string
}

func CreateMinioContainer(ctx context.Context) (*MinioContainer, error) {

	// Start MinIO container
	minioContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        "minio/minio:latest",
			ExposedPorts: []string{"9000/tcp"},
			Cmd:          []string{"server", "/data"},
			Env: map[string]string{
				"MINIO_ROOT_USER":     "minioadmin",
				"MINIO_ROOT_PASSWORD": "minioadmin",
			},
			WaitingFor: wait.ForListeningPort("9000/tcp"),
		},
		Started: true,
	})
	if err != nil {
		log.Fatalf("Failed to start MinIO container: %v", err)
	}

	// Get the container's host and port
	host, err := minioContainer.Host(ctx)
	if err != nil {
		log.Fatalf("Failed to get container host: %v", err)
	}
	port, err := minioContainer.MappedPort(ctx, "9000")
	if err != nil {
		log.Fatalf("Failed to get container port: %v", err)
	}

	// Connect to MinIO
	endpoint := host + ":" + port.Port()
	accessKeyID := "minioadmin"
	secretAccessKey := "minioadmin"
	useSSL := false

	minioClient, err := client.New(endpoint, &client.Options{
		Creds:  credentials.NewStaticV4(accessKeyID, secretAccessKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		log.Fatalf("Failed to create MinIO client: %v", err)
	}

	// Use MinIO client for testing
	// For example, create a bucket
	bucketName := "testbucket"
	err = minioClient.MakeBucket(ctx, bucketName, client.MakeBucketOptions{})
	if err != nil {
		log.Fatalf("Failed to create bucket: %v", err)
	}

	log.Println("MinIO testcontainer setup complete")

	// Return the MinIO container
	return &MinioContainer{
		Container:       minioContainer,
		Endpoint:        endpoint,
		AccessKeyID:     accessKeyID,
		SecretAccessKey: secretAccessKey,
		Port:            port.Port(),
	}, nil
}

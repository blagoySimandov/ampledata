package api

import (
	"context"
	"fmt"
	"time"

	"cloud.google.com/go/storage"
	"github.com/google/uuid"
)

type SignedURLRequest struct {
	ContentType string `json:"contentType"`
	Length      int    `json:"length"`
}

type SignedURLResponse struct {
	URL   string `json:"url"`
	JobID string `json:"jobId"`
}

var WHITELISTED_CONTENT_TYPES = []string{
	"text/csv",
	"application/json",
}

var (
	bucketName string = "ampledata-enrichment-uploads"
	method     string = "PUT"
)

func generateSignedURL(objectName string, contentType string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	client, err := storage.NewClient(ctx)
	if err != nil {
		return "", fmt.Errorf("client connection error: %w", err)
	}
	defer client.Close() // no need for err check

	expirationTime := time.Now().Add(90 * time.Second)
	bucket := client.Bucket(bucketName)
	url, err := bucket.SignedURL(objectName, &storage.SignedURLOptions{
		Expires:     expirationTime,
		Method:      method,
		ContentType: contentType,
	})
	if err != nil {
		return "", fmt.Errorf("failed to generate signed URL: %w", err)
	}
	return url, nil
}

// extension is with the dot. Example ".csv"
func generateJobId(extension string) string {
	filename := uuid.New().String() + extension
	return filename
}

package utils

import (
	"context"
	"fmt"
	"io"
	"mime"
	"path/filepath"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"

	"chronote/config"
)

func UploadPostcardObject(objectKey, filename string, body io.Reader, contentType string) (string, error) {
	cfg := config.AppConfig.S3
	if contentType == "" {
		contentType = detectContentType(filename)
	}
	_, err := config.S3Client.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket:      aws.String(cfg.BucketName),
		Key:         aws.String(objectKey),
		Body:        body,
		ContentType: aws.String(contentType),
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload to S3: %w", err)
	}
	return buildObjectURL(cfg.Endpoint, cfg.BucketName, objectKey, cfg.UseSSL), nil
}

func DeleteObject(objectKey string) error {
	if objectKey == "" {
		return nil
	}
	cfg := config.AppConfig.S3
	_, err := config.S3Client.DeleteObject(context.TODO(), &s3.DeleteObjectInput{
		Bucket: aws.String(cfg.BucketName),
		Key:    aws.String(objectKey),
	})
	if err != nil {
		return fmt.Errorf("failed to delete from S3: %w", err)
	}
	return nil
}

func detectContentType(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))
	if ext == "" {
		return "application/octet-stream"
	}
	if contentType := mime.TypeByExtension(ext); contentType != "" {
		return contentType
	}
	return "application/octet-stream"
}

func buildObjectURL(endpoint, bucketName, objectKey string, useSSL bool) string {
	protocol := "http"
	if useSSL {
		protocol = "https"
	}
	return fmt.Sprintf("%s://%s/%s/%s", protocol, endpoint, bucketName, objectKey)
}

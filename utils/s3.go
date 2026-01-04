package utils

import (
	"chronote/config"
	"context"
	"fmt"
	"mime/multipart"
	"path/filepath"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// UploadAvatar uploads an avatar file to S3 and returns the file URL
func UploadAvatar(file *multipart.FileHeader, userID uint) (string, error) {
	cfg := config.AppConfig.S3

	// Open the uploaded file
	src, err := file.Open()
	if err != nil {
		return "", fmt.Errorf("failed to open file: %w", err)
	}
	defer src.Close()

	// Validate file size (2MB = 2 * 1024 * 1024 bytes)
	const maxFileSize = 2 * 1024 * 1024
	if file.Size > maxFileSize {
		return "", fmt.Errorf("file size exceeds 2MB limit")
	}

	// Validate file extension
	ext := strings.ToLower(filepath.Ext(file.Filename))
	allowedExtensions := map[string]bool{
		".jpg":  true,
		".jpeg": true,
		".png":  true,
		".gif":  true,
		".webp": true,
	}
	if !allowedExtensions[ext] {
		return "", fmt.Errorf("invalid file type: only jpg, jpeg, png, gif, webp are allowed")
	}

	// Generate object key: avatars/{userID}/{timestamp}.{ext}
	timestamp := time.Now().Unix()
	objectKey := fmt.Sprintf("avatars/%d/%d%s", userID, timestamp, ext)

	// Determine content type
	contentType := getContentType(ext)

	// Upload to S3
	_, err = config.S3Client.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket:      aws.String(cfg.BucketName),
		Key:         aws.String(objectKey),
		Body:        src,
		ContentType: aws.String(contentType),
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload to S3: %w", err)
	}

	// Generate file URL
	protocol := "http"
	if cfg.UseSSL {
		protocol = "https"
	}
	fileURL := fmt.Sprintf("%s://%s/%s/%s", protocol, cfg.Endpoint, cfg.BucketName, objectKey)

	return fileURL, nil
}

// DeleteAvatar deletes an avatar file from S3
func DeleteAvatar(avatarURL string) error {
	if avatarURL == "" {
		return nil
	}

	cfg := config.AppConfig.S3

	// Extract object key from URL
	// URL format: http://localhost:9000/chronote-avatars/avatars/{userID}/{timestamp}.{ext}
	parts := strings.Split(avatarURL, fmt.Sprintf("/%s/", cfg.BucketName))
	if len(parts) < 2 {
		return fmt.Errorf("invalid avatar URL format")
	}
	objectKey := parts[1]

	// Delete from S3
	_, err := config.S3Client.DeleteObject(context.TODO(), &s3.DeleteObjectInput{
		Bucket: aws.String(cfg.BucketName),
		Key:    aws.String(objectKey),
	})
	if err != nil {
		return fmt.Errorf("failed to delete from S3: %w", err)
	}

	return nil
}

// getContentType returns the MIME type for a given file extension
func getContentType(ext string) string {
	contentTypes := map[string]string{
		".jpg":  "image/jpeg",
		".jpeg": "image/jpeg",
		".png":  "image/png",
		".gif":  "image/gif",
		".webp": "image/webp",
	}
	if ct, ok := contentTypes[ext]; ok {
		return ct
	}
	return "application/octet-stream"
}

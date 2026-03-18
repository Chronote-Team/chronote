package utils

import (
	"bytes"
	"chronote/config"
	"context"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// UploadAvatar uploads an avatar file to S3 and returns the file URL
func UploadAvatar(file *multipart.FileHeader, userID uint) (string, error) {
	cfg := config.AppConfig.S3
	if file == nil {
		return "", errors.New("头像文件不能为空")
	}

	// Validate file size (2MB = 2 * 1024 * 1024 bytes)
	const maxFileSize = 2 * 1024 * 1024
	if file.Size <= 0 {
		return "", errors.New("头像文件不能为空")
	}
	if file.Size > maxFileSize {
		return "", errors.New("头像文件大小超出限制")
	}

	// Open the uploaded file
	src, err := file.Open()
	if err != nil {
		return "", errors.New("读取头像文件失败")
	}
	defer src.Close()

	data, err := io.ReadAll(io.LimitReader(src, maxFileSize+1))
	if err != nil {
		return "", errors.New("读取头像文件失败")
	}
	if len(data) == 0 {
		return "", errors.New("头像文件不能为空")
	}
	if int64(len(data)) > maxFileSize {
		return "", errors.New("头像文件大小超出限制")
	}

	contentType := strings.ToLower(http.DetectContentType(data))
	allowedContentTypes := map[string]string{
		"image/jpeg": ".jpg",
		"image/png":  ".png",
		"image/gif":  ".gif",
		"image/webp": ".webp",
	}
	ext, ok := allowedContentTypes[contentType]
	if !ok {
		return "", errors.New("头像文件类型无效")
	}

	// Generate object key: avatars/{userID}/{timestamp}.{ext}
	timestamp := time.Now().Unix()
	objectKey := fmt.Sprintf("avatars/%d/%d%s", userID, timestamp, ext)

	// Upload to S3
	_, err = config.S3Client.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket:      aws.String(cfg.BucketName),
		Key:         aws.String(objectKey),
		Body:        bytes.NewReader(data),
		ContentType: aws.String(contentType),
	})
	if err != nil {
		return "", errors.New("上传头像失败")
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

package infra

import (
	"bytes"
	"context"
	"fmt"
	"strings"

	s3lib "github.com/aws/aws-sdk-go-v2/service/s3"
)

type S3Storage struct {
	client  *s3lib.Client
	bucket  string
	baseURL string
}

func NewS3Storage(client *s3lib.Client, bucket, baseURL string) *S3Storage {
	return &S3Storage{client: client, bucket: bucket, baseURL: strings.TrimRight(baseURL, "/")}
}

func (s *S3Storage) Upload(key, filename string, data []byte, contentType string) (string, error) {
	if s.client != nil && s.bucket != "" {
		_, err := s.client.PutObject(context.Background(), &s3lib.PutObjectInput{
			Bucket:      &s.bucket,
			Key:         &key,
			Body:        bytes.NewReader(data),
			ContentType: &contentType,
		})
		if err != nil {
			return "", err
		}
	}
	if s.baseURL != "" {
		return s.baseURL + "/" + key, nil
	}
	return fmt.Sprintf("https://cdn.example.com/%s", key), nil
}

func (s *S3Storage) Delete(key string) error {
	if s.client == nil || s.bucket == "" {
		return nil
	}
	_, err := s.client.DeleteObject(context.Background(), &s3lib.DeleteObjectInput{
		Bucket: &s.bucket,
		Key:    &key,
	})
	return err
}

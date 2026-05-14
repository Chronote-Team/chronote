package s3

import (
	"context"
	"strings"
	"time"

	platformconfig "chronote-refactor/internal/platform/config"

	awsv2 "github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	s3lib "github.com/aws/aws-sdk-go-v2/service/s3"
)

func NewClient(cfg *platformconfig.Config) (*s3lib.Client, error) {
	endpoint := strings.TrimSpace(cfg.S3.Endpoint)
	if endpoint != "" && !strings.Contains(endpoint, "://") {
		scheme := "http://"
		if cfg.S3.UseSSL {
			scheme = "https://"
		}
		endpoint = scheme + endpoint
	}

	awsCfg, err := awsconfig.LoadDefaultConfig(
		context.Background(),
		awsconfig.WithRegion(cfg.S3.Region),
		awsconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(cfg.S3.AccessKeyID, cfg.S3.SecretAccessKey, "")),
	)
	if err != nil {
		return nil, err
	}

	return s3lib.NewFromConfig(awsCfg, func(options *s3lib.Options) {
		if endpoint != "" {
			options.BaseEndpoint = awsv2.String(endpoint)
		}
		options.UsePathStyle = true
	}), nil
}

type Presigner struct {
	client *s3lib.PresignClient
	bucket string
}

func NewPresigner(client *s3lib.Client, bucket string) *Presigner {
	return &Presigner{client: s3lib.NewPresignClient(client), bucket: bucket}
}

func (p *Presigner) PresignGetObject(ctx context.Context, objectKey string, ttl time.Duration) (string, error) {
	output, err := p.client.PresignGetObject(ctx, &s3lib.GetObjectInput{
		Bucket: &p.bucket,
		Key:    &objectKey,
	}, func(options *s3lib.PresignOptions) {
		options.Expires = ttl
	})
	if err != nil {
		return "", err
	}
	return output.URL, nil
}

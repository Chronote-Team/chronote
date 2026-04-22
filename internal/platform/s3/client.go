package s3

import (
	"context"
	"strings"

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

package s3

import (
	"context"
	"errors"

	s3lib "github.com/aws/aws-sdk-go-v2/service/s3"
)

type Checker struct {
	client *s3lib.Client
}

func NewChecker(client *s3lib.Client) *Checker {
	return &Checker{client: client}
}

func (c *Checker) Check(ctx context.Context) error {
	if c == nil || c.client == nil {
		return errors.New("client not initialized")
	}
	_, err := c.client.ListBuckets(ctx, &s3lib.ListBucketsInput{})
	return err
}

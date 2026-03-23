package server

import (
	"context"
	"socket-flow/internal/config"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/pkg/errors"
)

func initMinioS3Client(ctx context.Context, cfg config.MinioConfig) (*minio.Client, error) {
	client, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKeyID, cfg.SecretAccessKey, ""),
		Secure: cfg.UseSSL,
		Region: cfg.Region,
	})
	if err != nil {
		return nil, errors.Wrap(err, "minio.New")
	}

	checkCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if _, err := client.ListBuckets(checkCtx); err != nil {
		return nil, errors.Wrap(err, "minio ping (ListBuckets)")
	}

	return client, nil
}

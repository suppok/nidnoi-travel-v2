// Package objectstorage stores media in Cloudflare R2 (S3-compatible). Only the backend holds the
// credentials; objects are read publicly through the R2 custom domain (R2_PUBLIC_URL).
package objectstorage

import (
	"bytes"
	"context"
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"

	"backend/pkg/core/config"
)

// ErrNotConfigured is returned by Put and Delete when the R2_* variables are not all set.
var ErrNotConfigured = errors.New("object storage is not configured")

// PutInput describes one object. Keys are immutable: replacing a file always uses a new key.
type PutInput struct {
	Key                string
	Body               []byte
	ContentType        string
	ContentDisposition string // optional, e.g. `attachment; filename*=UTF-8''...`
}

type ObjectStorage interface {
	Put(ctx context.Context, in PutInput) error
	Delete(ctx context.Context, key string) error
	// PublicURL is the URL the object is served from (R2_PUBLIC_URL + "/" + key).
	PublicURL(key string) string
	// PublicBaseURL is R2_PUBLIC_URL without a trailing slash ("" when not set).
	PublicBaseURL() string
}

// immutable objects: every key is content-addressed, so browsers and the CDN may cache forever.
const cacheControl = "public, max-age=31536000, immutable"

type r2Storage struct {
	client    *s3.Client
	bucket    string
	publicURL string
}

// New returns the R2 client, or a storage that fails with ErrNotConfigured when cfg is incomplete.
func New(cfg config.R2Config) ObjectStorage {
	if !cfg.Configured() {
		return disabledStorage{publicURL: cfg.PublicURL}
	}
	client := s3.New(s3.Options{
		Region:       "auto",
		BaseEndpoint: aws.String(fmt.Sprintf("https://%s.r2.cloudflarestorage.com", cfg.AccountID)),
		Credentials:  credentials.NewStaticCredentialsProvider(cfg.AccessKeyID, cfg.SecretAccessKey, ""),
	})
	return &r2Storage{client: client, bucket: cfg.Bucket, publicURL: cfg.PublicURL}
}

func (s *r2Storage) Put(ctx context.Context, in PutInput) error {
	input := &s3.PutObjectInput{
		Bucket:        aws.String(s.bucket),
		Key:           aws.String(in.Key),
		Body:          bytes.NewReader(in.Body),
		ContentLength: aws.Int64(int64(len(in.Body))),
		ContentType:   aws.String(in.ContentType),
		CacheControl:  aws.String(cacheControl),
	}
	if in.ContentDisposition != "" {
		input.ContentDisposition = aws.String(in.ContentDisposition)
	}
	if _, err := s.client.PutObject(ctx, input); err != nil {
		return fmt.Errorf("r2 put %s: %w", in.Key, err)
	}
	return nil
}

func (s *r2Storage) Delete(ctx context.Context, key string) error {
	if _, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{Bucket: aws.String(s.bucket), Key: aws.String(key)}); err != nil {
		return fmt.Errorf("r2 delete %s: %w", key, err)
	}
	return nil
}

func (s *r2Storage) PublicURL(key string) string { return s.publicURL + "/" + key }
func (s *r2Storage) PublicBaseURL() string       { return s.publicURL }

type disabledStorage struct{ publicURL string }

func (disabledStorage) Put(context.Context, PutInput) error  { return ErrNotConfigured }
func (disabledStorage) Delete(context.Context, string) error { return ErrNotConfigured }
func (d disabledStorage) PublicURL(key string) string        { return d.publicURL + "/" + key }
func (d disabledStorage) PublicBaseURL() string              { return d.publicURL }
